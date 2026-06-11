package handler

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"rukhsat/common"
)

// VerifyHandler handles GET /api/verify?token=...
func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Verification token is required"})
		return
	}

	collection, err := common.GetCollection("students")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Find the student matching this token
	var student common.Student
	err = collection.FindOne(ctx, bson.M{"verificationToken": token}).Decode(&student)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `
		<html>
		<body style="font-family:sans-serif; text-align:center; padding-top:100px; background:#fdfbf7;">
			<h2 style="color:#7c3d49;">Invalid or Expired Token</h2>
			<p style="color:#5e5854;">We could not find a pending RSVP matching this verification link.</p>
			<a href="/rsvp.html" style="color:#c2945d; text-decoration:none; font-weight:bold;">Return to RSVP Desk</a>
		</body>
		</html>
		`)
		return
	}

	// Update student to confirmed/declined status and remove temporary verification fields
	// Generate Unique Code
	finalRSVPStatus := student.PendingRSVPStatus
	finalFoodPreference := student.PendingFoodPref
	finalLikesReading := student.PendingLikesReading
	finalPhone := student.PendingPhone

	uniqueCode := student.UniqueCode
	if uniqueCode == "" && finalRSVPStatus == "confirmed" {
		// Use SHA-256 to generate a strong, unique, and non-predictable ID
		data := fmt.Sprintf("%s-%s-%d", student.ID.Hex(), student.Email, time.Now().UnixNano())
		hash := sha256.Sum256([]byte(data))
		uniqueCode = fmt.Sprintf("%x", hash)
	}

	update := bson.M{
		"$set": bson.M{
			"rsvpStatus":     finalRSVPStatus,
			"foodPreference": finalFoodPreference,
			"likesReading":   finalLikesReading,
			"phone":          finalPhone,
			"uniqueCode":     uniqueCode,
			"verified":       true,
		},
		"$unset": bson.M{
			"verificationToken":   "",
			"pendingRsvpStatus":   "",
			"pendingFoodPref":     "",
			"pendingLikesReading": "",
			"pendingPhone":        "",
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": student.ID}, update)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update RSVP status: " + err.Error()})
		return
	}

	// Invalidate the cache since database state changed
	common.InvalidateStudentsCache()

	// Compose elegant HTML email body
	rsvpText := "Yes, Count Me In! (Attending)"
	foodText := "Vegetarian"
	if finalFoodPreference == "non-veg" {
		foodText = "Non-Vegetarian"
	}
	if finalRSVPStatus == "declined" {
		rsvpText = "No, I Can't Make It"
		foodText = "N/A"
	}

	foodPrefHtml := ""
	if finalRSVPStatus == "confirmed" {
		foodPrefHtml = fmt.Sprintf(`<p class="text" style="font-size: 15px; margin: 5px 0;"><strong>Meal Preference:</strong> %s</p>`, foodText)
	}

	qrHtml := ""
	if finalRSVPStatus == "confirmed" && uniqueCode != "" {
		// Use local QR generator with logo overlay
		host := r.Host
		scheme := "https"
		if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
			scheme = "http"
		}
		qrUrl := fmt.Sprintf("%s://%s/api/qr?data=%s", scheme, host, uniqueCode)

		qrHtml = fmt.Sprintf(`
			<div style="margin: 30px auto; padding: 20px; background: #fff; border: 2px solid #7c3d49; border-radius: 8px; max-width: 250px;">
				<h4 style="margin: 0 0 15px 0; color: #7c3d49; font-family: 'Georgia', serif;">Your Entry Pass</h4>
				<img src="%s" alt="QR Code" style="width: 200px; height: 200px; display: block; margin: 0 auto;">
			</div>
			<p class="text">Please show this QR code at the entrance desk.</p>
		`, qrUrl)
	}

	emailBody := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>RSVP Confirmed - Rukhsat '26</title>
		<style>
			body { background-color: #fdfbf7; color: #2d2926; font-family: 'Inter', sans-serif; margin: 0; padding: 20px; }
			.card { max-width: 500px; margin: 40px auto; background: #ffffff; border: 1px solid #c2945d; box-shadow: 0 10px 25px rgba(45,41,38,0.08); padding: 30px; border-radius: 4px; text-align: center; }
			.title { font-family: 'Georgia', serif; font-size: 24px; color: #7c3d49; margin-bottom: 20px; }
			.text { font-size: 15px; line-height: 1.6; color: #5e5854; margin-bottom: 15px; }
			.details-box { background: #eff2ef; border: 1px dashed #6b7a67; border-radius: 6px; padding: 20px; margin: 25px auto; max-width: 380px; text-align: left; }
			.footer { font-size: 11px; color: #a17743; margin-top: 30px; border-top: 1px dashed rgba(194, 148, 93, 0.25); padding-top: 15px; }
		</style>
	</head>
	<body>
		<div class="card">
			<div class="title">Rukhsat '26 RSVP Registered</div>
			<p class="text">Hello <strong>%s</strong>,<br><br>Your RSVP response has been successfully registered and verified. Here are your selection details:</p>

			<div class="details-box">
				<p class="text" style="font-size: 15px; margin: 5px 0;"><strong>Name:</strong> %s</p>
				<p class="text" style="font-size: 15px; margin: 5px 0;"><strong>Attendance:</strong> %s</p>
				%s
			</div>

			%s

			<p class="text">We look forward to sharing these memorable farewell moments with you! Stay tuned for entry and counter food details.</p>
			<div class="footer">
				Rukhsat © Class of 2026. Keep these memories alive forever.
			</div>
		</div>
	</body>
	</html>
	`, student.Name, student.Name, rsvpText, foodPrefHtml, qrHtml)

	// Send confirmation email
	err = common.SendEmail(student.Email, "RSVP Registered - Rukhsat '26", emailBody)
	if err != nil {
		fmt.Printf("Warning: Failed to send RSVP registration mail: %v\n", err)
	}

	// Redirect to verification success screen
	http.Redirect(w, r, "/verification-success.html", http.StatusFound)
}
