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

// Resize 实现了 Service 接口的 Resize 方法 [EN] Resize implements the Resize method of the Service interface
func Resize(input io.Reader, width, height int) ([]byte, error) {
	// 1. 从 Reader 解码图像 [EN] 1. Decode images from Reader
	img, format, err := image.Decode(input)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	// 2. 使用 imaging 库调整大小 [EN] 2. Use imaging library to resize
	// imaging.Resize 会处理 width 或 height 为 0 的情况（保持宽高比） [EN] imaging.Resize will handle the case where width or height is 0 (maintaining aspect ratio)
	resizedImg := imaging.Resize(img, width, height, imaging.Lanczos)
	// 3. 获取输出格式 [EN] 3. Get the output format
	outputFormat, err := getOutputFormat(format)
	if err != nil {
		return nil, err
	}
	// 5. 编码图像 [EN] 5. Encoding images
	resultBytes, err := encodeImage(resizedImg, outputFormat, defaultQuality)
	if err != nil {
		return nil, err
	}
	return resultBytes, nil
}

// --- 以下是一些内部辅助函数 --- [EN] --- Here are some internal helper functions ---

// getOutputFormat 确定输出格式。 [EN] getOutputFormat determines the output format.
// 如果没有提供目标格式，则使用原始格式。 [EN] If no target format is provided, the original format is used.
func getOutputFormat(originalFormat string) (string, error) {
	// 这里 originalFormat 来自 image.Decode，是一个 MIME 类型子部分 [EN] Here originalFormat comes from image.Decode and is a MIME type subpart
	// 而 imaging 需要的是像 "jpeg", "png" 这样的格式字符串 [EN] And imaging requires format strings like "jpeg", "png"
	formatKey := strings.ToLower(originalFormat)

	switch formatKey {
	case "jpeg", "jpg":
		return "jpeg", nil
	case "png", "gif", "bmp", "tiff":
		return formatKey, nil
	default:
		// 默认回退到 JPEG [EN] Fallback to JPEG by default
		return "jpeg", nil
	}
}

// encodeImage 将 image.Image 编码为指定格式的字节切片 [EN] encodeImage encodes image.Image into a byte slice in the specified format
func encodeImage(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer

	var err error
	switch format {
	case "jpeg":
		err = imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(quality))
	case "png":
		// PNG 通常是无损的，忽略 quality 参数 [EN] PNG is usually lossless, ignoring the quality parameter
		err = imaging.Encode(&buf, img, imaging.PNG)
	case "gif":
		err = imaging.Encode(&buf, img, imaging.GIF)
	case "bmp":
		err = imaging.Encode(&buf, img, imaging.BMP)
	case "tiff":
		err = imaging.Encode(&buf, img, imaging.TIFF)
	default:
		// 对于不认识的格式，使用 JPEG 作为兜底 [EN] For unfamiliar formats, use JPEG as a fallback
		err = imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(quality))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image to %s: %w", format, err)
	}
	return buf.Bytes(), nil
}
