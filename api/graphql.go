package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"rukhsat/common"
)

// generateSecureToken creates a cryptographically secure random 32-character hex token
func generateSecureToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

// sanitizeFilename strips spaces and special characters from the file key
func sanitizeFilename(filename string) string {
	clean := strings.ReplaceAll(filename, " ", "-")
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_\.]`)
	return reg.ReplaceAllString(clean, "")
}

type graphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

// Schema instance variable
var schema graphql.Schema

func init() {
	// 1. Define Student Type
	studentType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Student",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.NewNonNull(graphql.ID),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						if s, ok := p.Source.(common.Student); ok {
							return s.ID.Hex(), nil
						}
						return nil, nil
					},
				},
				"name": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"email": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"rsvpStatus": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"foodPreference": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"verified": &graphql.Field{
					Type: graphql.NewNonNull(graphql.Boolean),
				},
			},
		},
	)

	// 2. Define Media Type
	mediaType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Media",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.NewNonNull(graphql.ID),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						if m, ok := p.Source.(common.Media); ok {
							return m.ID.Hex(), nil
						}
						return nil, nil
					},
				},
				"url": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"title": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"type": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"category": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"description": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"duration": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"createdAt": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						if m, ok := p.Source.(common.Media); ok {
							return m.CreatedAt.Format(time.RFC3339), nil
						}
						return nil, nil
					},
				},
			},
		},
	)

	// 3. Define UploadURLResponse Type
	uploadURLResponseType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "UploadURLResponse",
			Fields: graphql.Fields{
				"uploadUrl": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"publicUrl": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
				"key": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
		},
	)

	// 4. Define Root Query
	rootQuery := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "RootQuery",
			Fields: graphql.Fields{
				// Search students by query name (used in RSVP form)
				"searchStudents": &graphql.Field{
					Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(studentType))),
					Args: graphql.FieldConfigArgument{
						"query": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						query := p.Args["query"].(string)
						if len(query) < 2 {
							return []common.Student{}, nil
						}

						students, err := common.GetCachedStudents()
						if err != nil {
							return nil, err
						}

						var matches []common.Student
						queryLower := strings.ToLower(query)
						for _, s := range students {
							if strings.Contains(strings.ToLower(s.Name), queryLower) {
								matches = append(matches, s)
								if len(matches) >= 8 {
									break
								}
							}
						}
						return matches, nil
					},
				},

				// Get all media (for Photos & Videos vault)
				"listMedia": &graphql.Field{
					Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(mediaType))),
					Args: graphql.FieldConfigArgument{
						"type": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						mediaList, err := common.GetCachedMedia()
						if err != nil {
							return nil, err
						}

						mediaTypeVal, _ := p.Args["type"].(string)
						var filtered []common.Media
						for _, m := range mediaList {
							if mediaTypeVal == "" || m.Type == mediaTypeVal {
								mappedMedia := m
								parts := strings.Split(mappedMedia.URL, "/")
								if len(parts) > 0 {
									key := parts[len(parts)-1]
									mappedMedia.URL = "/api/view?key=" + key
								}
								filtered = append(filtered, mappedMedia)
							}
						}

						return filtered, nil
					},
				},

				// List all invitees (Admin Console)
				"listInvitees": &graphql.Field{
					Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(studentType))),
					Args: graphql.FieldConfigArgument{
						"secret": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						secret := p.Args["secret"].(string)
						adminSecret := os.Getenv("ADMIN_SECRET")
						if adminSecret != "" && secret != adminSecret {
							return nil, errors.New("unauthorized action: invalid admin secret")
						}

						students, err := common.GetCachedStudents()
						if err != nil {
							return nil, err
						}
						return students, nil
					},
				},
			},
		},
	)

	// 5. Define Root Mutation
	rootMutation := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "RootMutation",
			Fields: graphql.Fields{
				// Submit RSVP (sends verification email)
				"submitRSVP": &graphql.Field{
					Type: graphql.NewNonNull(studentType),
					Args: graphql.FieldConfigArgument{
						"studentId": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.ID),
						},
						"rsvpStatus": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"foodPreference": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						studentIdStr := p.Args["studentId"].(string)
						rsvpStatus := p.Args["rsvpStatus"].(string)
						foodPreference := ""
						if fPref, ok := p.Args["foodPreference"].(string); ok {
							foodPreference = fPref
						}

						if rsvpStatus != "confirmed" && rsvpStatus != "declined" {
							return nil, errors.New("rsvpStatus must be 'confirmed' or 'declined'")
						}

						studentObjID, err := primitive.ObjectIDFromHex(studentIdStr)
						if err != nil {
							return nil, errors.New("invalid student ID format")
						}

						collection, err := common.GetCollection("students")
						if err != nil {
							return nil, err
						}

						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()

						var student common.Student
						err = collection.FindOne(ctx, bson.M{"_id": studentObjID}).Decode(&student)
						if err != nil {
							return nil, errors.New("student not found in database")
						}

						if student.Verified {
							return nil, errors.New("RSVP has already been completed and verified for this student")
						}

						// 5-minute cooling off period check
						if !student.LastVerificationSent.IsZero() && time.Since(student.LastVerificationSent) < 5*time.Minute {
							remaining := 5*time.Minute - time.Since(student.LastVerificationSent)
							mins := int(remaining.Minutes())
							secs := int(remaining.Seconds()) % 60
							return nil, fmt.Errorf("a verification email was already sent recently. Please wait %d minutes and %d seconds before requesting a new link", mins, secs)
						}

						token := generateSecureToken()
						update := bson.M{
							"$set": bson.M{
								"verificationToken":    token,
								"pendingRsvpStatus":    rsvpStatus,
								"pendingFoodPref":      foodPreference,
								"lastVerificationSent": time.Now(),
							},
						}
						_, err = collection.UpdateOne(ctx, bson.M{"_id": studentObjID}, update)
						if err != nil {
							return nil, err
						}

						common.InvalidateStudentsCache()

						// Compose and send SMTP verification email
						// We use a fallback address if request headers are missing
						host := "rukhsat.vercel.app"
						if req, ok := p.Context.Value("http_request").(*http.Request); ok && req != nil {
							if fwdHost := req.Header.Get("X-Forwarded-Host"); fwdHost != "" {
								host = fwdHost
							} else {
								host = req.Host
							}
						}
						verifyURL := fmt.Sprintf("https://%s/api/verify?token=%s", host, token)
						if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
							verifyURL = fmt.Sprintf("http://%s/api/verify?token=%s", host, token)
						}

						emailBody := fmt.Sprintf(`
						<!DOCTYPE html>
						<html>
						<head>
							<meta charset="utf-8">
							<title>Verify Your RSVP - Rukhsat '26</title>
							<style>
								body { background-color: #fdfbf7; color: #2d2926; font-family: 'Inter', sans-serif; margin: 0; padding: 20px; }
								.card { max-width: 500px; margin: 40px auto; background: #ffffff; border: 1px solid #c2945d; box-shadow: 0 10px 25px rgba(45,41,38,0.08); padding: 30px; border-radius: 4px; text-align: center; }
								.title { font-family: 'Georgia', serif; font-size: 24px; color: #7c3d49; margin-bottom: 20px; }
								.text { font-size: 15px; line-height: 1.6; color: #5e5854; margin-bottom: 30px; }
								.btn { display: inline-block; background-color: #7c3d49; color: #ffffff !important; text-decoration: none; padding: 12px 30px; border-radius: 30px; font-weight: 600; letter-spacing: 1px; text-transform: uppercase; box-shadow: 0 4px 10px rgba(124, 61, 73, 0.2); margin-bottom: 20px; }
								.btn:hover { background-color: #2d2926; }
								.footer { font-size: 11px; color: #a17743; margin-top: 30px; border-top: 1px dashed rgba(194, 148, 93, 0.25); padding-top: 15px; }
							</style>
						</head>
						<body>
							<div class="card">
								<div class="title">Rukhsat '26 Farewell RSVP</div>
								<p class="text">Hello <strong>%s</strong>,<br><br>We received a reservation request for your farewell entry. Please click the button below to verify your email and complete your RSVP details.</p>
								<a href="%s" class="btn">Verify RSVP Now</a>
								<p class="text" style="font-size:12px; margin-top:15px; color:#a17743;">If the button above does not work, copy and paste this link in your browser:<br><a href="%s" style="color:#7c3d49;">%s</a></p>
								<div class="footer">
									Rukhsat © Class of 2026. Made with love for our seniors.
								</div>
							</div>
						</body>
						</html>
						`, student.Name, verifyURL, verifyURL, verifyURL)

						err = common.SendEmail(student.Email, "Verify Your RSVP - Rukhsat '26", emailBody)
						if err != nil {
							return nil, fmt.Errorf("failed to send SMTP verification email: %v", err)
						}

						// Return student with updated values
						student.RSVPStatus = rsvpStatus
						student.FoodPreference = foodPreference
						return student, nil
					},
				},

				// Generate B2 upload URL
				"generateUploadURL": &graphql.Field{
					Type: graphql.NewNonNull(uploadURLResponseType),
					Args: graphql.FieldConfigArgument{
						"secret": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"filename": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						secret := p.Args["secret"].(string)
						adminSecret := os.Getenv("ADMIN_SECRET")
						if adminSecret != "" && secret != adminSecret {
							return nil, errors.New("unauthorized action: invalid admin secret")
						}

						filename := p.Args["filename"].(string)

						svc, bucketName, err := common.GetS3Client()
						if err != nil || svc == nil {
							return nil, errors.New("B2 storage not configured properly")
						}

						cleanFilename := sanitizeFilename(filename)
						uniqueKey := fmt.Sprintf("%d-%s", time.Now().Unix(), cleanFilename)

						putReq, _ := svc.PutObjectRequest(&s3.PutObjectInput{
							Bucket: aws.String(bucketName),
							Key:    aws.String(uniqueKey),
						})

						urlStr, err := putReq.Presign(30 * time.Minute)
						if err != nil {
							return nil, fmt.Errorf("presigning URL failed: %v", err)
						}

						endpoint := strings.TrimSpace(os.Getenv("B2_ENDPOINT"))
						publicURL := fmt.Sprintf("https://%s/%s/%s", endpoint, bucketName, uniqueKey)

						return map[string]string{
							"uploadUrl": urlStr,
							"publicUrl": publicURL,
							"key":       uniqueKey,
						}, nil
					},
				},

				// Save media metadata to MongoDB
				"saveMediaMetadata": &graphql.Field{
					Type: graphql.NewNonNull(mediaType),
					Args: graphql.FieldConfigArgument{
						"secret": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"url": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"title": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"type": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"category": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"description": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"duration": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						secret := p.Args["secret"].(string)
						adminSecret := os.Getenv("ADMIN_SECRET")
						if adminSecret != "" && secret != adminSecret {
							return nil, errors.New("unauthorized action: invalid admin secret")
						}

						urlStr := p.Args["url"].(string)
						title := p.Args["title"].(string)
						mediaTypeVal := p.Args["type"].(string)
						category := p.Args["category"].(string)
						description := p.Args["description"].(string)
						duration := ""
						if dur, ok := p.Args["duration"].(string); ok {
							duration = dur
						}

						if mediaTypeVal != "video" && mediaTypeVal != "photo" {
							return nil, errors.New("type must be 'video' or 'photo'")
						}

						collection, err := common.GetCollection("media")
						if err != nil {
							return nil, err
						}

						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()

						media := common.Media{
							ID:          primitive.NewObjectID(),
							URL:         urlStr,
							Title:       title,
							Type:        mediaTypeVal,
							Category:    category,
							Description: description,
							Duration:    duration,
							CreatedAt:   time.Now(),
						}

						_, err = collection.InsertOne(ctx, media)
						if err != nil {
							return nil, err
						}

						common.InvalidateMediaCache()
						return media, nil
					},
				},

				// Delete media from MongoDB and Backblaze S3
				"deleteMedia": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
					Args: graphql.FieldConfigArgument{
						"secret": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"id": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.ID),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						secret := p.Args["secret"].(string)
						adminSecret := os.Getenv("ADMIN_SECRET")
						if adminSecret != "" && secret != adminSecret {
							return nil, errors.New("unauthorized action: invalid admin secret")
						}

						mediaIdStr := p.Args["id"].(string)
						objID, err := primitive.ObjectIDFromHex(mediaIdStr)
						if err != nil {
							return nil, errors.New("invalid media id format")
						}

						collection, err := common.GetCollection("media")
						if err != nil {
							return nil, err
						}

						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()

						var media common.Media
						err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&media)
						if err != nil {
							return nil, errors.New("media asset not found")
						}

						// Delete metadata
						_, err = collection.DeleteOne(ctx, bson.M{"_id": objID})
						if err != nil {
							return nil, err
						}

						common.InvalidateMediaCache()

						// Delete file asynchronously from S3
						go func(fileURL string) {
							svc, bucketName, err := common.GetS3Client()
							if err != nil || svc == nil {
								return
							}
							parts := strings.Split(fileURL, "/")
							if len(parts) > 0 {
								key := parts[len(parts)-1]
								svc.DeleteObject(&s3.DeleteObjectInput{
									Bucket: aws.String(bucketName),
									Key:    aws.String(key),
								})
							}
						}(media.URL)

						return "success", nil
					},
				},

				// Delete invitee student from MongoDB
				"deleteInvitee": &graphql.Field{
					Type: graphql.NewNonNull(graphql.String),
					Args: graphql.FieldConfigArgument{
						"secret": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"id": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.ID),
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						secret := p.Args["secret"].(string)
						adminSecret := os.Getenv("ADMIN_SECRET")
						if adminSecret != "" && secret != adminSecret {
							return nil, errors.New("unauthorized action: invalid admin secret")
						}

						studentIdStr := p.Args["id"].(string)
						objID, err := primitive.ObjectIDFromHex(studentIdStr)
						if err != nil {
							return nil, errors.New("invalid student id format")
						}

						collection, err := common.GetCollection("students")
						if err != nil {
							return nil, err
						}

						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()

						_, err = collection.DeleteOne(ctx, bson.M{"_id": objID})
						if err != nil {
							return nil, err
						}

						common.InvalidateStudentsCache()
						return "success", nil
					},
				},
			},
		},
	)

	// 6. Build final schema
	var err error
	schema, err = graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    rootQuery,
			Mutation: rootMutation,
		},
	)
	if err != nil {
		panic("Failed to compile GraphQL schema: " + err.Error())
	}
}

// GraphqlHandler handles POST /api/graphql
func GraphqlHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Secret")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only POST queries are allowed"})
		return
	}

	var reqBody graphQLRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON request payload"})
		return
	}

	// Setup context with HTTP request info (to resolve host header dynamically in resolvers)
	ctx := context.WithValue(r.Context(), "http_request", r)

	// Execute GraphQL Query
	result := graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  reqBody.Query,
		VariableValues: reqBody.Variables,
		OperationName:  reqBody.OperationName,
		Context:        ctx,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
