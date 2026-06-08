package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rukhsat/common"
)

// Handler handles GET, POST, and DELETE requests for /api/media
func Handler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Secret")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	collection, err := common.GetCollection("media")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch r.Method {
	case "GET":
		// Query list of media sorted by newest first
		opts := options.Find().SetSort(bson.M{"createdAt": -1})
		cursor, err := collection.Find(ctx, bson.M{}, opts)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch media list: " + err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var mediaList []common.Media = []common.Media{}
		if err := cursor.All(ctx, &mediaList); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to decode media results: " + err.Error()})
			return
		}

		// Since Backblaze B2 bucket is Private, we route requests through our CDN-cached proxy endpoint `/api/view?key=...`
		for i := range mediaList {
			parts := strings.Split(mediaList[i].URL, "/")
			if len(parts) > 0 {
				key := parts[len(parts)-1]
				mediaList[i].URL = "/api/view?key=" + key
			}
		}

		json.NewEncoder(w).Encode(mediaList)

	case "POST":
		// Save metadata (Admin action)
		adminSecret := os.Getenv("ADMIN_SECRET")
		if adminSecret != "" && r.Header.Get("X-Admin-Secret") != adminSecret {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized action"})
			return
		}

		var req common.Media
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.URL == "" || req.Title == "" || req.Type == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "url, title, and type are required"})
			return
		}

		if req.Type != "video" && req.Type != "photo" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "type must be 'video' or 'photo'"})
			return
		}

		// Set timestamps and generate ID
		req.ID = primitive.NewObjectID()
		req.CreatedAt = time.Now()

		_, err = collection.InsertOne(ctx, req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save media metadata: " + err.Error()})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"media":  req,
		})

	case "DELETE":
		// Delete media (Admin action)
		adminSecret := os.Getenv("ADMIN_SECRET")
		if adminSecret != "" && r.Header.Get("X-Admin-Secret") != adminSecret {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized action"})
			return
		}

		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing media id parameter"})
			return
		}

		objID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid media id format"})
			return
		}

		// Find media to get file URL
		var media common.Media
		err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&media)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Media asset not found"})
			return
		}

		// Delete from MongoDB
		_, err = collection.DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete media metadata: " + err.Error()})
			return
		}

		// Delete file from Backblaze B2 S3 storage bucket asynchronously
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

		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Media deleted successfully from database and storage",
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}
