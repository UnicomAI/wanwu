package imaging

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"strings"

	"github.com/disintegration/imaging"
)

const defaultQuality = 85

// Resize implements the Resize method of the Service interface
func Resize(input io.Reader, width, height int) ([]byte, error) {
	// 1. Decode images from Reader
	img, format, err := image.Decode(input)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	// 2. Use imaging library to resize
	// imaging.Resize will handle the case where width or height is 0 (maintaining aspect ratio)
	resizedImg := imaging.Resize(img, width, height, imaging.Lanczos)
	// 3. Get the output format
	outputFormat, err := getOutputFormat(format)
	if err != nil {
		return nil, err
	}
	// 5. Encoding images
	resultBytes, err := encodeImage(resizedImg, outputFormat, defaultQuality)
	if err != nil {
		return nil, err
	}
	return resultBytes, nil
}

// --- Here are some internal helper functions ---

// getOutputFormat determines the output format.
// If no target format is provided, the original format is used.
func getOutputFormat(originalFormat string) (string, error) {
	// Here originalFormat comes from image.Decode and is a MIME type subpart
	// And imaging requires format strings like "jpeg", "png"
	formatKey := strings.ToLower(originalFormat)

	switch formatKey {
	case "jpeg", "jpg":
		return "jpeg", nil
	case "png", "gif", "bmp", "tiff":
		return formatKey, nil
	default:
		// Fallback to JPEG by default
		return "jpeg", nil
	}
}

// encodeImage encodes image.Image into a byte slice in the specified format
func encodeImage(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer

	var err error
	switch format {
	case "jpeg":
		err = imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(quality))
	case "png":
		// PNG is usually lossless, ignoring the quality parameter
		err = imaging.Encode(&buf, img, imaging.PNG)
	case "gif":
		err = imaging.Encode(&buf, img, imaging.GIF)
	case "bmp":
		err = imaging.Encode(&buf, img, imaging.BMP)
	case "tiff":
		err = imaging.Encode(&buf, img, imaging.TIFF)
	default:
		// For unfamiliar formats, use JPEG as a fallback
		err = imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(quality))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image to %s: %w", format, err)
	}
	return buf.Bytes(), nil
}
