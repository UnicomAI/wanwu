package util

import (
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
)

const (
	KnowledgeImportFileFormatErr    = "know_doc_unsupported_file_format"
	KnowledgeImportFileSizeErr      = "know_doc_file_size_exceed"
	KnowledgeImportSameNameErr      = "know_same_name_validation_fail"
	KnowledgeImportInvalidNameErr   = "know_doc_invalid_file_name"
	KnowledgeImportNameTooLongErr   = "know_doc_file_name_too_long"
	KnowledgeDocLastFailureErr      = "know_doc_last_failure_info"
	KnowledgeDocParsingServiceErr   = "know_doc_parsing_service_error"
	KnowledgeDocVectorDuplicateErr  = "know_doc_vector_duplicate_error"
	KnowledgeDocDuplicateErr        = "know_doc_duplicate_error"
	KnowledgeDocDownloadErr         = "know_doc_download_error"
	KnowledgeDocSplitErr            = "know_doc_split_error"
	KnowledgeDocEmbeddingErr        = "know_doc_embedding_error"
	KnowledgeDocTextErr             = "know_doc_text_error"
	KnowledgeDocEmptyFileContentErr = "know_doc_empty_file_content"
	KnowledgeDocFileUnUsableErr     = "know_doc_file_unusable"
)

// IsPreImportCheckFailErr 导入前校验失败的文档从未推送到 RAG，删除时无需调用 RAG 删除接口
func IsPreImportCheckFailErr(errMsg string) bool {
	switch errMsg {
	case KnowledgeImportFileFormatErr, KnowledgeImportFileSizeErr,
		KnowledgeImportSameNameErr, KnowledgeImportInvalidNameErr,
		KnowledgeImportNameTooLongErr:
		return true
	default:
		return false
	}
}

func ErrCode(code err_code.Code) error {
	return grpc_util.ErrorStatusWithKey(code, "")
}

func ErrStatus(code err_code.Code, status *err_code.Status) error {
	return grpc_util.ErrorStatusWithKey(code, status.TextKey, status.Args...)
}

func ToErrStatus(key string, args ...string) *err_code.Status {
	return &err_code.Status{
		TextKey: key,
		Args:    args,
	}
}
