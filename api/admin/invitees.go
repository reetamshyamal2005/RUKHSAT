package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
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

		// Map columns flexibly (substring match)
		nameIdx, emailIdx := -1, -1
		for i, h := range header {
			hClean := strings.ToLower(strings.TrimSpace(h))
			if strings.Contains(hClean, "name") {
				nameIdx = i
			} else if strings.Contains(hClean, "email") || strings.Contains(hClean, "mail") {
				emailIdx = i
			}
		}

		if nameIdx == -1 || emailIdx == -1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "CSV must contain columns representing 'Name' and 'Email'"})
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
					"likesReading":   "",
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
			common.InvalidateStudentsCache()
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

		if format := r.URL.Query().Get("format"); format == "csv" || format == "xlsx" {
			w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			w.Header().Set("Content-Disposition", "attachment; filename=rukhsat-rsvp-export.xlsx")

			f := excelize.NewFile()
			sheetName := "Sheet1"
			// Write Headers
			headers := []interface{}{"Timestamp", "Email address", "Name", "Dept", "Contact Number", "Year", "Food Preference", "Likes Reading", "Unique_id"}
			f.SetSheetRow(sheetName, "A1", &headers)

			rowIdx := 2
			for _, s := range students {
				if !s.Verified {
					continue // Only export verified RSVPs for the scanner app
				}

				foodPrefStr := "Vegetarian"
				if s.FoodPreference == "non-veg" {
					foodPrefStr = "Non-Vegetarian"
				}

				likesReadingStr := "No"
				if s.LikesReading == "yes" {
					likesReadingStr = "Yes"
				}

				timestamp := ""
				if !s.LastVerificationSent.IsZero() {
					timestamp = s.LastVerificationSent.Format("02/01/2006 15:04:05")
				}

				uniqueId := s.UniqueCode
				row := []interface{}{
					timestamp,
					s.Email,
					s.Name,
					"IT",
					s.Phone,
					"4th",
					foodPrefStr,
					likesReadingStr,
					uniqueId,
				}
				cell := fmt.Sprintf("A%d", rowIdx)
				f.SetSheetRow(sheetName, cell, &row)
				rowIdx++
			}

			if err := f.Write(w); err != nil {
				fmt.Printf("Error writing excel file: %v\n", err)
			}
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

		common.InvalidateStudentsCache()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Invitee deleted successfully"})

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}
