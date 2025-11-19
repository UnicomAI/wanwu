package util

// BuildDocRespStatus rag 1.0 状态码和rag 2.0状态码转换关系如下： [EN] The conversion relationship between BuildDocRespStatus rag 1.0 status code and rag 2.0 status code is as follows:
// rag 2.0  --->     rag 1.0
// 1     --->     1
// 10     --->     1
// 2     --->     2
// 31-35   --->      3
// 4      --->      4
// 51-56    --->     5
// 61-69    --->     5
func BuildDocRespStatus(number int) int {
	if number < 10 {
		return number
	} else if (number/10)%10 == 6 { //用户责任导致的错误码为61,62...使用5返回前端 [EN] The error code caused by user responsibility is 61,62...Use 5 to return to the front end
		return 5
	} else {
		return (number / 10) % 10
	}
}

func BuildDocReqStatusList(status int) []int {
	var statusList []int
	switch status {
	case -1: //-1 查全部 [EN] -1 Check all
		break
	case 1:
		statusList = append(statusList, []int{1, 10}...)
	case 3:
		statusList = append(statusList, []int{31, 32, 33, 34, 35}...)
	case 5:
		statusList = append(statusList, []int{5, 51, 52, 53, 54, 55, 56, 61, 62}...)
	default:
		statusList = append(statusList, status)
	}
	return statusList
}

// BuildDocErrMessage 构造文档错误信息 [EN] BuildDocErrMessage constructs document error message
func BuildDocErrMessage(status int) string {
	//判断：如果是status属于(51,52,53,54,55,56)，说明是RAG本身导致的解析异常，此时给errMsg写入一个默认值“文件解析服务异常” [EN] Judgment: If the status belongs to (51,52,53,54,55,56), it means that the parsing exception is caused by RAG itself. At this time, write a default value "File parsing service exception" to errMsg.
	//判断：如果是status属于(61,62)，说明是用户责任导致的异常，此时分别写入errMsg，提示用户修改文档 [EN] Judgment: If the status belongs to (61,62), it means that the exception is caused by the user's responsibility. At this time, errMsg is written respectively to prompt the user to modify the document.
	switch status {
	case 51:
		return KnowledgeDocVectorDuplicateErr
	case 52:
		return KnowledgeDocDuplicateErr
	case 53:
		return KnowledgeDocDownloadErr
	case 54:
		return KnowledgeDocSplitErr
	case 55:
		return KnowledgeDocEmbeddingErr
	case 56:
		return KnowledgeDocTextErr
	case 61:
		return KnowledgeDocEmptyFileContentErr
	case 62:
		return KnowledgeDocFileUnUsableErr
	default:
		break
	}
	return ""
}

func BuildAnalyzingStatus() []int {
	var stList []int
	stList = append(stList, 3, 31, 32, 33, 34, 35)
	return stList
}
