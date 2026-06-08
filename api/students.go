package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rukhsat/common"
)

// StudentsHandler handles GET /api/students?query=...
func StudentsHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	query := r.URL.Query().Get("query")
	if len(query) < 2 {
		json.NewEncoder(w).Encode([]common.Student{})
		return
	}

	collection, err := common.GetCollection("students")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Case-insensitive regex match for name
	filter := bson.M{
		"name": bson.M{
			"$regex": primitive.Regex{Pattern: query, Options: "i"},
		},
	}

	// Limit suggestions to 8 and return only required details
	opts := options.Find().
		SetProjection(bson.M{"name": 1, "email": 1, "rsvpStatus": 1}).
		SetLimit(8)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to search database: " + err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var students []common.Student = []common.Student{}
	if err := cursor.All(ctx, &students); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to decode database results: " + err.Error()})
		return
	}

	json.NewEncoder(w).Encode(students)
}

