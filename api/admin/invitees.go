package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rukhsat/common"
)

// Handler handles GET, POST, and DELETE requests for /api/admin/invitees
func Handler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Secret")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Validate admin security token
	adminSecret := os.Getenv("ADMIN_SECRET")
	token := r.Header.Get("X-Admin-Secret")
	if token == "" {
		token = r.URL.Query().Get("secret")
	}

	if adminSecret != "" && token != adminSecret {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized action"})
		return
	}

	collection, err := common.GetCollection("students")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch r.Method {
	case "POST":
		// Parse CSV Import
		// We expect multipart/form-data with key "file"
		err := r.ParseMultipartForm(10 << 20) // 10MB limit
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse form: " + err.Error()})
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "No file uploaded under key 'file'"})
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		// Read header
		header, err := reader.Read()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read CSV header: " + err.Error()})
			return
		}

		// Map columns
		nameIdx, emailIdx := -1, -1
		for i, h := range header {
			hClean := strings.ToLower(strings.TrimSpace(h))
			if hClean == "name" {
				nameIdx = i
			} else if hClean == "email" {
				emailIdx = i
			}
		}

		if nameIdx == -1 || emailIdx == -1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "CSV must contain 'name' and 'email' columns"})
			return
		}

		// Read records and prepare bulk operations
		var bulkOps []mongo.WriteModel
		importedCount := 0

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Error reading CSV row: " + err.Error()})
				return
			}

			if len(record) <= nameIdx || len(record) <= emailIdx {
				continue
			}

			name := strings.TrimSpace(record[nameIdx])
			email := strings.ToLower(strings.TrimSpace(record[emailIdx]))

			if name == "" || email == "" {
				continue
			}

			// Upsert based on email to prevent duplicate invitees
			filter := bson.M{"email": email}
			update := bson.M{
				"$setOnInsert": bson.M{
					"name":           name,
					"rsvpStatus":     "pending",
					"foodPreference": "",
					"verified":       false,
				},
			}
			model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true)
			bulkOps = append(bulkOps, model)
			importedCount++
		}

		if len(bulkOps) > 0 {
			_, err = collection.BulkWrite(ctx, bulkOps)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Failed to write database records: " + err.Error()})
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "success",
			"imported": importedCount,
		})

	case "GET":
		// Query list of invitees
		opts := options.Find().SetSort(bson.M{"name": 1})
		cursor, err := collection.Find(ctx, bson.M{}, opts)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to query students: " + err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var students []common.Student = []common.Student{}
		if err := cursor.All(ctx, &students); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to decode students: " + err.Error()})
			return
		}

		if r.URL.Query().Get("format") == "csv" {
			// Set headers for file download
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", "attachment; filename=rukhsat-rsvp-export.csv")

			writer := csv.NewWriter(w)
			writer.Write([]string{"Name", "Email", "RSVP Status", "Food Preference", "Verified"})

			for _, s := range students {
				verifiedStr := "No"
				if s.Verified {
					verifiedStr = "Yes"
				}
				writer.Write([]string{
					s.Name,
					s.Email,
					s.RSVPStatus,
					s.FoodPreference,
					verifiedStr,
				})
			}
			writer.Flush()
			return
		}

		// Default: Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(students)

	case "DELETE":
		// Delete an invitee by ID
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing student id parameter"})
			return
		}

		objID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid student id format"})
			return
		}

		_, err = collection.DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete student: " + err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Invitee deleted successfully"})

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}
