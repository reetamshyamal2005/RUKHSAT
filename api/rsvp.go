package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"rukhsat/common"
)

type rsvpRequest struct {
	StudentID      string `json:"studentId"`
	RSVPStatus     string `json:"rsvpStatus"`     // confirmed | declined
	FoodPreference string `json:"foodPreference"` // veg | non-veg
}

// generateSecureToken creates a cryptographically secure random 32-character hex token
func generateSecureToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback in case of CSPRNG failure
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

// Handler handles POST /api/rsvp
func Handler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req rsvpRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON request payload"})
		return
	}

	if req.StudentID == "" || req.RSVPStatus == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "studentId and rsvpStatus are required"})
		return
	}

	if req.RSVPStatus != "confirmed" && req.RSVPStatus != "declined" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "rsvpStatus must be 'confirmed' or 'declined'"})
		return
	}

	collection, err := common.GetCollection("students")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error: " + err.Error()})
		return
	}

	studentObjID, err := primitive.ObjectIDFromHex(req.StudentID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid student ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the student
	var student common.Student
	err = collection.FindOne(ctx, bson.M{"_id": studentObjID}).Decode(&student)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Student not found in database"})
		return
	}

	// Generate verification token
	token := generateSecureToken()

	// Update the student with pending RSVP details and the verification token
	update := bson.M{
		"$set": bson.M{
			"verificationToken": token,
			"pendingRsvpStatus": req.RSVPStatus,
			"pendingFoodPref":   req.FoodPreference,
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": studentObjID}, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update database: " + err.Error()})
		return
	}

	// Prepare Verification Email details
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	verifyURL := fmt.Sprintf("%s://%s/api/verify?token=%s", scheme, r.Host, token)

	emailBody := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Verify Your RSVP - Rukhsat '26</title>
		<style>
			body {
				background-color: #fdfbf7;
				color: #2d2926;
				font-family: 'Inter', sans-serif;
				margin: 0;
				padding: 20px;
			}
			.card {
				max-width: 500px;
				margin: 40px auto;
				background: #ffffff;
				border: 1px solid #c2945d;
				box-shadow: 0 10px 25px rgba(45,41,38,0.08);
				padding: 30px;
				border-radius: 4px;
				text-align: center;
			}
			.title {
				font-family: 'Georgia', serif;
				font-size: 24px;
				color: #7c3d49;
				margin-bottom: 20px;
			}
			.text {
				font-size: 15px;
				line-height: 1.6;
				color: #5e5854;
				margin-bottom: 30px;
			}
			.btn {
				display: inline-block;
				background-color: #7c3d49;
				color: #ffffff !important;
				text-decoration: none;
				padding: 12px 30px;
				border-radius: 30px;
				font-weight: 600;
				letter-spacing: 1px;
				text-transform: uppercase;
				box-shadow: 0 4px 10px rgba(124, 61, 73, 0.2);
				margin-bottom: 20px;
			}
			.btn:hover {
				background-color: #2d2926;
			}
			.footer {
				font-size: 11px;
				color: #a17743;
				margin-top: 30px;
				border-top: 1px dashed rgba(194, 148, 93, 0.25);
				padding-top: 15px;
			}
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

	// Send verification email via Gmail SMTP
	err = common.SendEmail(student.Email, "Verify Your RSVP - Rukhsat '26", emailBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to send email via SMTP: " + err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Verification email sent successfully.",
	})
}

