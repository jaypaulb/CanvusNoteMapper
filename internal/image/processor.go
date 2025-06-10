package image

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"strings"

	"github.com/nfnt/resize"
)

const (
	MaxImageSize = 20 * 1024 * 1024 // 20MB in bytes
	MaxDimension = 2048             // Maximum width/height
	Quality      = 85               // JPEG quality (0-100)
)

// ProcessImage takes raw image bytes and returns optimized image bytes
// that are within size limits while maintaining quality
func ProcessImage(input []byte) ([]byte, string, error) {
	// First try to decode the image
	img, format, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		log.Printf("[ProcessImage] Failed to decode image: %v", err)
		return nil, "", err
	}

	// Get original dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	log.Printf("[ProcessImage] Original image: %dx%d, format: %s", width, height, format)

	// Calculate new dimensions if needed
	newWidth, newHeight := width, height
	if width > MaxDimension || height > MaxDimension {
		if width > height {
			newWidth = MaxDimension
			newHeight = int(float64(height) * float64(MaxDimension) / float64(width))
		} else {
			newHeight = MaxDimension
			newWidth = int(float64(width) * float64(MaxDimension) / float64(height))
		}
		log.Printf("[ProcessImage] Resizing to: %dx%d", newWidth, newHeight)
	}

	// Resize if needed
	if newWidth != width || newHeight != height {
		img = resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)
	}

	// Encode with appropriate format and quality
	var output bytes.Buffer
	var mimeType string
	if strings.EqualFold(format, "jpeg") || strings.EqualFold(format, "jpg") {
		err = jpeg.Encode(&output, img, &jpeg.Options{Quality: Quality})
		mimeType = "image/jpeg"
		log.Printf("[ProcessImage] Set MIME type to: %s", mimeType)
	} else {
		// Default to PNG for other formats
		err = png.Encode(&output, img)
		mimeType = "image/png"
		log.Printf("[ProcessImage] Set MIME type to: %s", mimeType)
	}
	if err != nil {
		log.Printf("[ProcessImage] Failed to encode image: %v", err)
		return nil, "", err
	}

	// Check final size
	outputBytes := output.Bytes()
	log.Printf("[ProcessImage] Final image size: %d bytes", len(outputBytes))

	// If still too large, reduce quality further
	if len(outputBytes) > MaxImageSize && strings.EqualFold(format, "jpeg") {
		log.Printf("[ProcessImage] Image still too large, reducing quality further")
		quality := Quality
		for len(outputBytes) > MaxImageSize && quality > 10 {
			quality -= 10
			output.Reset()
			err = jpeg.Encode(&output, img, &jpeg.Options{Quality: quality})
			if err != nil {
				log.Printf("[ProcessImage] Failed to encode with reduced quality: %v", err)
				return nil, "", err
			}
			outputBytes = output.Bytes()
			log.Printf("[ProcessImage] Reduced quality to %d, new size: %d bytes", quality, len(outputBytes))
		}
	}

	log.Printf("[ProcessImage] Returning MIME type: %s", mimeType)
	return outputBytes, mimeType, nil
}
