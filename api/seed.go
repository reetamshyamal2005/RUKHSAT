package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"rukhsat/common"
)

// Handler is the entrypoint for Vercel Serverless Function /api/seed
func Handler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	collection, err := common.GetCollection("students")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to connect to database: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Clear existing students
	_, err = collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to clear collection: " + err.Error()})
		return
	}

	// Insert sample students
	sampleStudents := []interface{}{
		common.Student{
			ID:             primitive.NewObjectID(),
			Name:           "Reetam Shyamal",
			Email:          "reetamshyamal2005@gmail.com",
			RSVPStatus:     "pending",
			FoodPreference: "",
			Verified:       false,
		},
		common.Student{
			ID:             primitive.NewObjectID(),
			Name:           "Satyam Kumar",
			Email:          "satyam.kumar.dev@gmail.com",
			RSVPStatus:     "pending",
			FoodPreference: "",
			Verified:       false,
		},
		common.Student{
			ID:             primitive.NewObjectID(),
			Name:           "Aditya Vardhan",
			Email:          "aditya.vardhan@example.com",
			RSVPStatus:     "pending",
			FoodPreference: "",
			Verified:       false,
		},
		common.Student{
			ID:             primitive.NewObjectID(),
			Name:           "Anjali Sharma",
			Email:          "anjali.sharma@example.com",
			RSVPStatus:     "pending",
			FoodPreference: "",
			Verified:       false,
		},
		common.Student{
			ID:             primitive.NewObjectID(),
			Name:           "Vikram Aditya Singh",
			Email:          "vikram.singh@example.com",
			RSVPStatus:     "pending",
			FoodPreference: "",
			Verified:       false,
		},
	}

	_, err = collection.InsertMany(ctx, sampleStudents)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to seed students: " + err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Database seeded successfully with 5 sample students.",
	})
}

