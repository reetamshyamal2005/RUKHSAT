package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/bson"

	"rukhsat/api/common"
)

// readInvitationCard attempts to locate and read the invitation card file
func readInvitationCard() ([]byte, string, error) {
	paths := []string{
		"images/invitation_card.png",
		"../images/invitation_card.png",
		"../../images/invitation_card.png",
		"images/invitation_card.pdf",
		"../images/invitation_card.pdf",
		"../../images/invitation_card.pdf",
		"images/farewell'26.png",
		"../images/farewell'26.png",
		"../../images/farewell'26.png",
	}

	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, filepath.Base(p), nil
		}
	}
	return nil, "", fmt.Errorf("could not locate invitation_card or farewell'26.png in common paths")
}

// Handler handles GET /api/verify?token=...
func Handler(w http.ResponseWriter, r *http.Request) {
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
	finalRSVPStatus := student.PendingRSVPStatus
	finalFoodPreference := student.PendingFoodPref

	update := bson.M{
		"$set": bson.M{
			"rsvpStatus":     finalRSVPStatus,
			"foodPreference": finalFoodPreference,
			"verified":       true,
		},
		"$unset": bson.M{
			"verificationToken": "",
			"pendingRsvpStatus": "",
			"pendingFoodPref":   "",
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": student.ID}, update)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update RSVP status: " + err.Error()})
		return
	}

	// If the user declined, send a simple email confirmation and redirect
	if finalRSVPStatus == "declined" {
		declineEmailBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><style>body{background-color:#fdfbf7;color:#2d2926;font-family:sans-serif;padding:20px;}.card{max-width:500px;margin:40px auto;background:#fff;border:1px solid #c2945d;padding:30px;text-align:center;}</style></head>
		<body>
			<div class="card">
				<h2 style="color:#7c3d49;">RSVP Confirmation</h2>
				<p>Hello <strong>%s</strong>,</p>
				<p>We have registered your RSVP response: <strong>Declined</strong>.</p>
				<p>We're sorry you won't be able to make it, but we wish you all the best and hope you stay connected!</p>
			</div>
		</body>
		</html>
		`, student.Name)

		_ = common.SendEmail(student.Email, "RSVP Registered - Rukhsat '26", declineEmailBody)

		// Redirect to success page
		http.Redirect(w, r, "/verification-success.html", http.StatusFound)
		return
	}

	// If confirmed, generate dynamic entry QR code
	qrData := fmt.Sprintf("RUKHSAT '26 ENTRY TICKET\nName: %s\nFood: %s\nStatus: Verified Attendee\nID: %s", 
		student.Name, finalFoodPreference, student.ID.Hex())
	
	qrBytes, err := qrcode.Encode(qrData, qrcode.Medium, 256)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to generate QR code: " + err.Error()})
		return
	}

	// Read invitation card from filesystem
	cardBytes, cardName, err := readInvitationCard()
	if err != nil {
		// Log warning but continue, we can send QR code even if attachment fails
		fmt.Printf("Warning: Failed to load invitation card attachment: %v\n", err)
	}

	// Encode QR code bytes to base64 so we can embed it inline in the HTML body
	qrBase64 := base64.StdEncoding.EncodeToString(qrBytes)

	// Compose Email Body
	htmlBody := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Your Farewell Entry Ticket - Rukhsat '26</title>
		<style>
			body { background-color: #fdfbf7; color: #2d2926; font-family: 'Inter', sans-serif; margin: 0; padding: 20px; }
			.card { max-width: 600px; margin: 30px auto; background: #ffffff; border: 1px solid #c2945d; padding: 30px; border-radius: 4px; text-align: center; }
			.title { font-family: 'Georgia', serif; font-size: 24px; color: #7c3d49; margin-bottom: 20px; }
			.text { font-size: 15px; line-height: 1.6; color: #5e5854; margin-bottom: 20px; text-align: left; }
			.details-box { background: #eff2ef; border: 1px dashed #6b7a67; border-radius: 6px; padding: 15px; margin: 20px 0; text-align: left; }
			.details-row { display: flex; justify-content: space-between; margin-bottom: 8px; font-size: 14px; }
			.details-label { font-weight: bold; color: #6b7a67; }
			.qr-img { margin: 20px auto; display: block; border: 4px solid #2d2926; border-radius: 4px; }
			.footer { font-size: 11px; color: #a17743; margin-top: 30px; border-top: 1px dashed rgba(194, 148, 93, 0.25); padding-top: 15px; }
		</style>
	</head>
	<body>
		<div class="card">
			<div class="title">Rukhsat '26 Invitation Card</div>
			<p class="text" style="text-align:center;">Congratulations! Your RSVP is verified.</p>
			
			<div class="details-box">
				<div class="details-row"><span class="details-label">Guest Name:</span><span>%s</span></div>
				<div class="details-row"><span class="details-label">Attendance:</span><span style="color:#6b7a67; font-weight:bold;">Yes, Attending</span></div>
				<div class="details-row"><span class="details-label">Meal Option:</span><span style="text-transform: uppercase;">%s</span></div>
			</div>

			<p class="text" style="text-align:center;">Here is your entry ticket QR code. Please keep this email safe and present this QR code at the entry gate and food counters:</p>
			
			<img class="qr-img" src="data:image/png;base64,%s" alt="Entry QR Code" width="180" height="180">
			
			<p class="text" style="font-size:13px; text-align:center; font-style:italic;">Your personalized invitation card has been attached to this email.</p>
			
			<div class="footer">
				Rukhsat © Class of 2026. Keep these memories alive forever.
			</div>
		</div>
	</body>
	</html>
	`, student.Name, finalFoodPreference, qrBase64)

	// Send confirmation email with invitation attachment
	if len(cardBytes) > 0 {
		err = common.SendEmailWithAttachment(student.Email, "Your Farewell Invitation Ticket - Rukhsat '26", htmlBody, cardBytes, cardName)
	} else {
		// Fallback: send without attachment if invitation card is missing from repository
		err = common.SendEmail(student.Email, "Your Farewell Invitation Ticket - Rukhsat '26", htmlBody)
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "RSVP was verified but failed to send invitation mail: " + err.Error()})
		return
	}

	// Redirect to verification success screen
	http.Redirect(w, r, "/verification-success.html", http.StatusFound)
}

