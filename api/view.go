package handler

import (
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"rukhsat/common"
)

// Handler handles GET /api/view?key=...
func Handler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	svc, bucketName, err := common.GetS3Client()
	if err != nil || svc == nil {
		http.Error(w, "B2 Storage not configured", http.StatusInternalServerError)
		return
	}

	// 1. Get object metadata (head) to check size and content type
	headReq, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	contentLength := *headReq.ContentLength
	contentType := ""
	if headReq.ContentType != nil {
		contentType = *headReq.ContentType
	}

	// Vercel serverless function payload limit is 4.5MB (4,718,592 bytes).
	// If the file is larger than 4MB, redirect to a presigned GET URL to avoid Vercel 502 limit error.
	if contentLength > 4*1024*1024 {
		getReq, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})
		presignedURL, err := getReq.Presign(1 * time.Hour)
		if err != nil {
			http.Error(w, "Failed to generate redirect URL", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, presignedURL, http.StatusFound)
		return
	}

	// 2. Fetch the file contents
	getObj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	defer getObj.Body.Close()

	// 3. Set content type and cache headers
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Cache in browser and Vercel edge CDN for 30 days
	w.Header().Set("Cache-Control", "public, max-age=2592000, s-maxage=2592000, immutable")

	// 4. Stream body to client
	_, err = io.Copy(w, getObj.Body)
	if err != nil {
		return
	}
}
