package handler

import (
	"image"
	"image/color"
	stdDraw "image/draw"
	"image/png"
	"net/http"
	"os"
	"path/filepath"

	"github.com/skip2/go-qrcode"
	"golang.org/x/image/draw"
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

	// 2. Try to load and resize the logo
	logoPath := "qr-logo.png" 
	logoFile, err := os.Open(logoPath)
	
	if err != nil {
		// Fallback
		logoPath = filepath.Join("api", "qr-logo.png")
		logoFile, err = os.Open(logoPath)
	}

	var finalImg image.Image = qrImg

	if err == nil {
		defer logoFile.Close()
		logoImg, err := png.Decode(logoFile)
		if err == nil {
			// Resize logo to 70x70 for consistent look
			resizedLogo := image.NewRGBA(image.Rect(0, 0, 70, 70))
			draw.BiLinear.Scale(resizedLogo, resizedLogo.Bounds(), logoImg, logoImg.Bounds(), draw.Over, nil)
			
			// 3. Overlay logo onto QR code
			finalImg = overlayCircularLogo(qrImg, resizedLogo)
		}
	}

	// 4. Serve the final image
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	png.Encode(w, finalImg)
}

// circle structure for mask
type circle struct {
	p image.Point
	r int
}

func (c circle) ColorModel() color.Model { return color.AlphaModel }
func (c circle) Bounds() image.Rectangle { return image.Rect(c.p.X-c.r, c.p.Y-c.r, c.p.X+c.r, c.p.Y+c.r) }
func (c circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.p.X)+0.5, float64(y-c.p.Y)+0.5, float64(c.r)
	if xx*xx+yy*yy < rr*rr {
		return color.Alpha{255}
	}
	return color.Alpha{0}
}

func overlayCircularLogo(qrImg image.Image, logoImg image.Image) image.Image {
	qrBounds := qrImg.Bounds()
	canvas := image.NewRGBA(qrBounds)
	stdDraw.Draw(canvas, qrBounds, qrImg, image.Point{}, stdDraw.Src)

	logoBounds := logoImg.Bounds()
	qrSize := qrBounds.Dx()
	center := qrSize / 2
	
	// White circular background for scannability
	radius := (logoBounds.Dx() / 2) + 3
	stdDraw.DrawMask(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, circle{image.Point{center, center}, radius}, image.Point{}, stdDraw.Over)

	// Draw the resized logo
	halfLogo := logoBounds.Dx() / 2
	destRect := image.Rect(
		center-halfLogo,
		center-halfLogo,
		center+halfLogo,
		center+halfLogo,
	)
	stdDraw.Draw(canvas, destRect, logoImg, image.Point{}, stdDraw.Over)

	return canvas
}
