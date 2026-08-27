package csv_util

import (
	"context"
	"encoding/csv"
	"github.com/UnicomAI/wanwu/internal/app-service/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"os"
	"path/filepath"
	"time"
)

const (
	csvSuffix = ".csv"
	// utf8BOM 是 UTF-8 的字节顺序标记，写入文件开头可避免 Excel 打开时中文乱码。
	utf8BOM = "\xEF\xBB\xBF"
)

type ExportCsvParams struct {
	ExportLocalDir  string   //导出的本地目录
	CsvHeader       []string //导出csv的表头
	MinioBucketDir  string   //对象存储前缀
	MinioBucketName string   //对象存储前缀
}

type ExportCsvResult struct {
	TotalCount   int    //文件总行数
	SuccessCount int    //解析成功的数量
	MinioPath    string //文件地址
	FileSize     int64  //文件大小
}

type CsvData[T any] struct {
	DataList []T
	Ext      interface{}
}

func ExportCsvFile[T any](ctx context.Context, exportCsvParams *ExportCsvParams, dataGetter func(ctx context.Context) (CsvData[T], error), lineProcessor func(context.Context, CsvData[T], T) []string) (result *ExportCsvResult, err error) {
	defer util.PrintPanicStackWithCall(func(panicOccur bool, recoverError error) {
		if panicOccur {
			err = recoverError
		}
	})
	//1.查询数据详情
	csvData, err := dataGetter(ctx)
	if err != nil {
		return nil, err
	}
	//2.csv本地文件处理
	exportCsvFilePath, successCount, err := writeCsv(ctx, exportCsvParams, csvData, lineProcessor)
	if err != nil {
		return nil, err
	}
	// 本地文件上传完成后清理（writeCsv 只负责写盘并关闭句柄，不删除文件）
	defer clearLocalCsvFile(exportCsvFilePath)
	//3.上传minio
	filePath, fileSize, err := uploadMinio(ctx, exportCsvParams, exportCsvFilePath)
	if err != nil {
		return nil, err
	}
	return &ExportCsvResult{
		TotalCount:   len(csvData.DataList),
		SuccessCount: successCount,
		MinioPath:    filePath,
		FileSize:     fileSize,
	}, nil
}

func writeCsv[T any](ctx context.Context, exportCsvParams *ExportCsvParams, csvData CsvData[T], lineProcessor func(ctx context.Context, data CsvData[T], lineData T) []string) (string, int, error) {
	exportCsvFilePath, exportCsvFile, err := createLocalCsv(exportCsvParams)
	if err != nil {
		return "", 0, err
	}
	// 写入 UTF-8 BOM，避免 Excel 打开时中文乱码
	if _, err = exportCsvFile.WriteString(utf8BOM); err != nil {
		return "", 0, err
	}
	csvWriter := csv.NewWriter(exportCsvFile)
	csvWriter.Comma = ','

	// 只 flush 并关闭文件句柄；本地文件由调用方在上传 minio 后清理（clearLocalCsvFile）。
	defer func() {
		closeCsvWriter(csvWriter)
		closeLocalCsvFile(exportCsvFile)
	}()

	if len(exportCsvParams.CsvHeader) > 0 {
		// 写入表头
		if err = csvWriter.Write(exportCsvParams.CsvHeader); err != nil {
			return "", 0, err
		}
	}

	// 写入内容
	var successCount = 0
	for _, lineData := range csvData.DataList {
		itemList := lineProcessor(ctx, csvData, lineData)
		if err = csvWriter.Write(itemList); err != nil {
			log.Errorf("write conversation log csv row err: %v", err)
			continue
		}
		successCount++
	}

	return exportCsvFilePath, successCount, nil
}

func uploadMinio(ctx context.Context, exportCsvParams *ExportCsvParams, csvFilePath string) (string, int64, error) {
	//  桶：export-public name: app_log + uuid前8位 +文件名
	_, minioFilePath, fileSize, err := minio.UploadLocalFile(ctx, exportCsvParams.MinioBucketDir, exportCsvParams.MinioBucketName, filepath.Base(csvFilePath), csvFilePath)
	if err != nil {
		log.Errorf("upload file err: %v", err)
		return "", 0, err
	}
	bucket, objectName, _ := minio.SplitFilePath(minioFilePath)
	filePath := bucket + "/" + objectName
	return filePath, fileSize, nil
}

func createLocalCsv(exportCsvParams *ExportCsvParams) (string, *os.File, error) {
	curUuid := util.GenUUID()
	// 写本地 CSV 文件
	exportCsvFilePath := exportCsvParams.ExportLocalDir + curUuid + "_" + time.Now().Format("20060102150405") + csvSuffix
	if err := os.MkdirAll(filepath.Dir(exportCsvFilePath), 0755); err != nil {
		log.Infof("Error create directory: %v", err)
		return "", nil, err
	}
	exportCsvFile, err := os.Create(exportCsvFilePath)
	if err != nil {
		log.Infof("Error opening file: %v", err)
		return "", nil, err
	}
	return exportCsvFilePath, exportCsvFile, nil
}

// closeLocalCsvFile 关闭本地 CSV 文件句柄（写盘后、上传前调用）。
func closeLocalCsvFile(exportCsvFile *os.File) {
	if exportCsvFile == nil {
		return
	}
	if err := exportCsvFile.Close(); err != nil {
		log.Infof("Error closing file: %v", err)
	}
}

// clearLocalCsvFile 删除本地 CSV 文件（上传 minio 完成后调用）。
func clearLocalCsvFile(exportCsvFilePath string) {
	if exportCsvFilePath == "" {
		return
	}
	if err := os.Remove(exportCsvFilePath); err != nil {
		log.Infof("Error remove file: %v", err)
	}
}

func closeCsvWriter(csvWriter *csv.Writer) {
	if csvWriter != nil {
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			log.Errorf("csvWriter error %s", err)
		}
	}
}
