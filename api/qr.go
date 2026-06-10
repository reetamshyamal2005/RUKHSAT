package handler

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"os"
	"path/filepath"

	"github.com/skip2/go-qrcode"
)

// QrHandler handles GET /api/qr?data=...
// It generates a QR code with a logo in the center.
func QrHandler(w http.ResponseWriter, r *http.Request) {
	data := r.URL.Query().Get("data")
	if data == "" {
		http.Error(w, "Missing data parameter", http.StatusBadRequest)
		return
	}

	// 1. Generate QR Code image
	// We use High recovery level to ensure the QR remains scannable even with a logo in the center
	q, err := qrcode.New(data, qrcode.High)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	qrImg := q.Image(300)

	// 2. Try to load the logo
	logoPath := filepath.Join("images", "qr-logo.png")
	logoFile, err := os.Open(logoPath)
	
	var finalImg image.Image = qrImg

	if err == nil {
		defer logoFile.Close()
		logoImg, err := png.Decode(logoFile)
		if err == nil {
			// 3. Overlay logo onto QR code
			finalImg = overlayLogo(qrImg, logoImg)
		}
	}

	// 4. Serve the final image
	w.Header().Set("Content-Type", "image/png")
	// Cache for a long time since QR for a unique ID won't change
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	png.Encode(w, finalImg)
}

func overlayLogo(qrImg image.Image, logoImg image.Image) image.Image {
	qrBounds := qrImg.Bounds()
	
	// Create a new RGBA image to draw on
	canvas := image.NewRGBA(qrBounds)
	draw.Draw(canvas, qrBounds, qrImg, image.Point{}, draw.Src)

	// Calculate logo size (roughly 20% of QR size)
	qrSize := qrBounds.Dx()
	logoSize := qrSize * 22 / 100
	
	logoBounds := logoImg.Bounds()
	
	// Center point
	center := qrSize / 2
	halfLogo := logoSize / 2
	
	destRect := image.Rect(
		center-halfLogo,
		center-halfLogo,
		center+halfLogo,
		center+halfLogo,
	)

	// 1. Draw a white background plate behind the logo to ensure scannability
	// We make the plate slightly larger than the logo
	plateRect := image.Rect(
		destRect.Min.X-2,
		destRect.Min.Y-2,
		destRect.Max.X+2,
		destRect.Max.Y+2,
	)
	draw.Draw(canvas, plateRect, &image.Uniform{color.White}, image.Point{}, draw.Src)

	// 2. Draw the logo in the center
	// Since we aren't using a scaling library, we'll just draw the logoImg centered.
	// We expect the logoImg to be roughly the right size (e.g. 60-70px for a 300px QR)
	// Calculate the offset for the logoImg source to center it
	offsetX := (logoBounds.Dx() - logoSize) / 2
	offsetY := (logoBounds.Dy() - logoSize) / 2

	draw.Draw(canvas, destRect, logoImg, image.Point{logoBounds.Min.X + offsetX, logoBounds.Min.Y + offsetY}, draw.Over)

	return canvas
}
