package file_extract

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/UnicomAI/wanwu/pkg/log"
)

type TarGzFileExtractServiceService struct {
}

var tarGzFileExtractServiceService = &TarGzFileExtractServiceService{}

func init() {
	AddFileExtractService(tarGzFileExtractServiceService)
}

func (t TarGzFileExtractServiceService) ExtractFileType() string {
	return ".tar.gz"
}

func (t TarGzFileExtractServiceService) ExtractFile(ctx context.Context, localFilePath string, destDir string) (extractDir string, err error) {
	// Open .tar.gz file
	file, err := os.Open(localFilePath)
	if err != nil {
		return "", err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Errorf("error closing file: %v", err)
		}
	}()
	// Create gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		log.Errorf("error gzip new reader file: %v", err)
		return "", err
	}
	defer func() {
		err = gzipReader.Close()
		if err != nil {
			log.Errorf("error gzip close reader: %v", err)
		}
	}()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)
	// Make sure the target directory exists
	if err = os.MkdirAll(destDir, 0755); err != nil {
		log.Errorf("error make dir: %v", err)
		return "", err
	}

	// Traverse files in tarball
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // All files have been traversed
		}
		if err != nil {
			log.Errorf("error tar reader: %v", err)
			return "", err
		}
		// Get the path of the file
		path := filepath.Join(destDir, header.Name)
		// Create folders or write files based on file type
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			log.Errorf("error make dir: %v", err)
			return "", err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return "", err
			}
		case tar.TypeReg:
			err = writeFile(path, header, tarReader)
			if err != nil {
				return "", err
			}
		default:
			// Ignore other types of files
			continue
		}
	}
	return destDir, nil
}

// writeFile writes to file
func writeFile(filePath string, header *tar.Header, tarReader *tar.Reader) error {
	// Open file for writing
	writer, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
	if err != nil {
		return err
	}
	defer func() {
		err = writer.Close()
		log.Errorf("error make dir: %v", err)
	}()

	// Read file contents from tarball and write to file
	if _, err := io.Copy(writer, tarReader); err != nil {
		return err
	}
	return nil
}
