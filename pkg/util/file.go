package util

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
)

const (
	kb               = 1024
	mb               = kb * 1024
	MaxScanTokenSize = 1024 * 1024 // Set the maximum token size to 1 MB
)

var specialFileExtList = []string{".tar.gz"}

type FileMergeResult struct {
	TotalSuccessCount int64
	TotalLineCount    int64
	TotalByteCount    int64
	FilePath          string
}

func FileExt(filePath string) string {
	if len(filePath) == 0 {
		return ""
	}
	for _, ext := range specialFileExtList {
		if strings.HasSuffix(filePath, ext) {
			return ext
		}
	}
	return filepath.Ext(filePath)
}

// ToFileSizeStr fileSize unit is B, conversion rules: less than 1M is KB, greater than or equal to 1M, the unit is M, keep two decimal places
func ToFileSizeStr(fileSize int64) string {
	if fileSize < mb {
		return fmt.Sprintf("%.2f KB", float64(fileSize)/float64(kb))
	} else {
		return fmt.Sprintf("%.2f MB", float64(fileSize)/float64(mb))
	}
}

func FileExist(filePath string) (bool, error) {
	if len(filePath) == 0 {
		return false, nil
	}
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func DirFileList(dir string, subDir bool, fullPath bool) ([]string, error) {
	var fileNameList []string
	// Read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir (%v) err: %v", dir, err)
	}

	// Traverse all files and subdirectories in a directory
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			// handling errors
			log.Errorf("read dir (%v) entry err: %v", dir, err)
			continue
		}

		// Determine whether it is a file
		if !info.IsDir() {
			if fullPath {
				fileNameList = append(fileNameList, dir+"/"+entry.Name())
			} else {
				fileNameList = append(fileNameList, entry.Name())
			}
		} else if !subDir { //No need to verify the underlying directory
			continue
		} else {
			list, err := DirFileList(dir+"/"+entry.Name(), subDir, fullPath)
			if err != nil {
				return nil, err
			} else {
				fileNameList = append(fileNameList, list...)
			}
		}
	}

	return fileNameList, nil
}

// MergeFile merge files
func MergeFile(filePathList []string, mergeFilePath string) (*FileMergeResult, error) {
	// Create or open a file
	//0644, indicating that the file owner can read and write, and users in the same group and other users can only read.
	dir := filepath.Dir(mergeFilePath)
	exist, err := FileExist(dir)
	if err != nil {
		return nil, err
	}
	if !exist {
		err = os.MkdirAll(filepath.Dir(mergeFilePath), 0755)
		if err != nil {
			return nil, err
		}
	}

	destinationFile, err := os.OpenFile(mergeFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open merge file (%v) err: %v", mergeFilePath, err)
	}
	defer func() {
		if err := destinationFile.Close(); err != nil {
			log.Errorf("close merge file (%v) err: %v", mergeFilePath, err)
		}
	}()

	var totalByteCount int64
	for _, fileInfo := range filePathList {
		byteCount, err := AppendFileStream(fileInfo, destinationFile)
		if err != nil {
			return nil, fmt.Errorf("merge file (%v) err: %v", mergeFilePath, err)
		}
		totalByteCount += byteCount
	}
	return &FileMergeResult{
		TotalByteCount: totalByteCount,
		FilePath:       mergeFilePath,
	}, nil
}

func DeleteDirFile(fileDir string) error {
	err := os.RemoveAll(fileDir)
	if err != nil {
		return fmt.Errorf("delete dir (%v) err: %v", fileDir, err)
	}
	return nil
}

func DeleteFile(file string) error {
	err := os.Remove(file)
	if err != nil {
		return fmt.Errorf("delete file (%v) err: %v", file, err)
	}
	return nil
}

func AppendFileStream(filePath string, destinationFile *os.File) (int64, error) {
	// Open the source file for reading
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("open append file (%v) err: %v", filePath, err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			log.Errorf("close append file (%v) err: %v", filePath, err)
		}
	}()
	fileReader := bufio.NewReader(sourceFile)
	byteCount, err := appendFile(fileReader, destinationFile)
	if err != nil {
		return 0, fmt.Errorf("append file (%v) to (%v) err: %v", filePath, destinationFile.Name(), err)
	}
	log.Infof("append file (%v) to (%v) succeed, bytes: %v", filePath, destinationFile.Name(), byteCount)
	return byteCount, nil
}

func appendFile(reader *bufio.Reader, destinationFile *os.File) (byteCount int64, error error) {
	buf := make([]byte, MaxScanTokenSize)
	for {
		n, err := reader.Read(buf)
		if FileEOF(err) { // Check if end of file is reached
			break
		}
		if err != nil {
			log.Errorf("Error reading file: %s", err)
			return -1, err
		}
		line := buf[:n]
		bytesWritten, err := destinationFile.Write(line)
		if err != nil {
			log.Errorf("appendFile error %s", err)
			return -1, err
		}
		byteCount += int64(bytesWritten)
	}
	return byteCount, nil
}

func FileEOF(err error) bool {
	return errors.Is(err, io.EOF) || (err != nil && err.Error() == "EOF")
}
