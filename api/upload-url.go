package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"rukhsat/common"
)

type uploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
}

// sanitizeFilename strips spaces and special characters from the file key to prevent URL signature mismatches
func sanitizeFilename(filename string) string {
	clean := strings.ReplaceAll(filename, " ", "-")
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_\.]`)
	return reg.ReplaceAllString(clean, "")
}

// UploadUrlHandler handles POST /api/upload-url
func UploadUrlHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Secret")
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

	// Validate optional admin security token to prevent spam uploads
	adminSecret := os.Getenv("ADMIN_SECRET")
	if adminSecret != "" && r.Header.Get("X-Admin-Secret") != adminSecret {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized action"})
		return
	}

	var req uploadRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Filename == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Filename is required"})
		return
	}

	// Retrieve S3 client singleton
	svc, bucketName, err := common.GetS3Client()
	if err != nil || svc == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Backblaze B2 storage is not configured properly or failed to initialize."})
		return
	}

	// Create a unique clean file key using Unix timestamp and sanitized filename
	cleanFilename := sanitizeFilename(req.Filename)
	uniqueKey := fmt.Sprintf("%d-%s", time.Now().Unix(), cleanFilename)

	putReq, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(uniqueKey),
	})

	// Generate the S3 presigned URL (valid for 30 minutes)
	urlStr, err := putReq.Presign(30 * time.Minute)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Presigning URL failed: " + err.Error()})
		return
	}

	endpoint := strings.TrimSpace(os.Getenv("B2_ENDPOINT"))
	// Construct public direct file link (to be saved in DB; we sign this URL at GET time)
	publicURL := fmt.Sprintf("https://%s/%s/%s", endpoint, bucketName, uniqueKey)

	json.NewEncoder(w).Encode(map[string]string{
		"uploadUrl": urlStr,
		"publicUrl": publicURL,
		"key":       uniqueKey,
	})
}
