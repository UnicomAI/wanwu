package util

// BuildDocRespStatus 将 rag 2.0 状态码折算回 rag 1.0 粗粒度码
//
// 前端 rag 状态码语义：	0 待处理 1 解析成功 2 正在审核中	3 解析中 4 审核未通过 5 解析失败
//
// rag 2.0 -> rag 1.0 折算关系：
//
//	1       -->  1
//	10      -->  1
//	2       -->  2
//	20      -->  3
//	31-35   -->  3
//	4       -->  4
//	51-56   -->  5
//	61-69   -->  5
func BuildDocRespStatus(number int) int {
	if number < 10 {
		return number
	} else if number == 20 { //rag 2.0 收到 kafka 开始解析，归到"解析中"
		return 3
	} else if (number/10)%10 == 6 { //用户责任导致的错误码为61,62...使用5返回前端
		return 5
	} else {
		return (number / 10) % 10
	}
}

// BuildDocReqStatusList
//
//	-1 全部       -> 不加筛选条件
//	 1 解析成功    -> [1, 10]            1=rag 1.0 历史码（已废弃但保留兼容），10=文档上传成功且完成（终态·成功）
//	 3 解析中      -> [20, 31~35]        20=收到kafka开始解析，31=查重完成，32=下载完成，33=切分完成，34=向量导入完成，35=文本导入完成
//	 5 解析失败    -> [5, 51~56, 61, 62]  5=服务重启时由 stopDocProcess 内部写入的中断码；
//	                                    51=向量库重复查询校验失败，52=文档已存在该知识库，
//	                                    53=文档下载失败，54=切分失败（通用错误），
//	                                    55=向量导入失败，56=文本导入失败，
//	                                    61=切分失败（切分结果为空），62=切分失败（文件不可用）
//	其它          -> 原样透传（如 0 待处理、2 审核中等）
func BuildDocReqStatusList(reqStatusList []int32) []int {
	var statusList []int
	for _, v := range reqStatusList {
		switch v {
		case -1: // 查全部
		case 1: // 解析成功
			statusList = append(statusList, []int{1, 10}...)
		case 3: // 解析中
			statusList = append(statusList, []int{20, 31, 32, 33, 34, 35}...)
		case 5: // 解析失败
			statusList = append(statusList, []int{5, 51, 52, 53, 54, 55, 56, 61, 62}...)
		default:
			statusList = append(statusList, int(v))
		}
	}
	return statusList
}

// BuildDocReqGraphStatusList
//
//	-1 全部       -> 不加筛选条件
//	 0 待处理      -> [0]                 图谱未处理
//	 1 解析中      -> [110~114]           110=图谱开始解析，111=schema 解析成功，112=chunk 文本生成成功，
//	                                    113=提取图谱成功，114=持久化存储成功
//	 2 解析成功    -> [100]               graph 图谱提取成功（终态·成功）
//	3（失败）   -> [101~104, 119]      101=生成图谱获取 chunk 文本失败，102=提取图谱失败，
//	                                    103=图谱持久化存储失败，104=graph schema 解析失败，
//	                                    119=重启打断执行
func BuildDocReqGraphStatusList(reqGraphStatusList []int32) []int {
	var graphStatusList []int
	// 四个状态都有的情况相当于全选
	if checkAllGraphStatus(reqGraphStatusList) {
		return graphStatusList
	}
	for _, v := range reqGraphStatusList {
		switch v {
		case -1: // 查全部
		case 0: // 图谱待处理
			graphStatusList = append(graphStatusList, []int{0}...)
		case 1: // 图谱解析中
			graphStatusList = append(graphStatusList, []int{110, 111, 112, 113, 114}...)
		case 2: // 图谱解析成功
			graphStatusList = append(graphStatusList, []int{100}...)
		default: // 图谱解析失败（rag 失败码 101~104 + 服务重启中断码 119）
			graphStatusList = append(graphStatusList, []int{101, 102, 103, 104, 119}...)
		}
	}
	return graphStatusList
}

func checkAllGraphStatus(reqGraphStatusList []int32) bool {
	if len(reqGraphStatusList) >= 4 {
		set := map[int32]bool{}
		for _, v := range reqGraphStatusList {
			set[v] = true
		}
		if set[0] && set[1] && set[2] && set[3] {
			return true
		}
	}
	return false
}

// BuildDocErrMessage 构造文档错误信息
func BuildDocErrMessage(status int) string {
	//判断：如果是status属于(51,52,53,54,55,56)，说明是RAG本身导致的解析异常，此时给errMsg写入一个默认值“文件解析服务异常”
	//判断：如果是status属于(61,62)，说明是用户责任导致的异常，此时分别写入errMsg，提示用户修改文档
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
	stList = append(stList, 3, 20, 31, 32, 33, 34, 35)
	return stList
}

// BuildDocProgress 将文档原始状态码映射为进度百分比 (0-100)，用于进度条展示。
// 同一进度值同时代表"该阶段完成"和"在下一阶段失败"，进度条停留在最后成功的位置。
func BuildDocProgress(status int) int {
	switch status {
	case 20, 51, 52: // 收到 kafka / 查重阶段失败
		return 5
	case 31, 53: // 查重完成 / 下载阶段失败
		return 20
	case 32, 54, 61, 62: // 下载完成 / 切分阶段失败
		return 40
	case 33, 55: // 切分完成 / 向量化阶段失败
		return 60
	case 34, 56: // 向量导入完成 / 文本导入阶段失败
		return 80
	case 35: // 文本导入完成
		return 95
	case 1, 10: // 全部完成
		return 100
	default: // 0 待处理 / 5 中断 / 未知码
		return 0
	}
}
