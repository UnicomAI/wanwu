package service

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	knowledgeBase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	safety_service "github.com/UnicomAI/wanwu/api/proto/safety-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func CreateRag(ctx *gin.Context, userId, orgId string, req request.AppBriefConfig) (*request.RagReq, error) {
	resp, err := rag.CreateRag(ctx.Request.Context(), &rag_service.CreateRagReq{
		AppBrief: appBriefConfigModel2Proto(req),
		Identity: &rag_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &request.RagReq{
		RagID: resp.RagId,
	}, err
}

func UpdateRag(ctx *gin.Context, req request.RagUpdateReq, userId, orgId string) error {
	existingRag, err := rag.GetRagDetail(ctx.Request.Context(), &rag_service.RagDetailReq{RagId: req.RagID})
	if err != nil {
		return err
	}
	if err := util.ValidateBriefUpdate(&req.Name, existingRag.BriefConfig.Name, &req.Desc, existingRag.BriefConfig.Desc, util.SubjectRag); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
	}
	_, err = rag.UpdateRag(ctx.Request.Context(), &rag_service.UpdateRagReq{
		RagId:    req.RagID,
		AppBrief: appBriefConfigModel2Proto(req.AppBriefConfig),
		Identity: &rag_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func UpdateRagConfig(ctx *gin.Context, req request.RagConfig) error {
	var modelConfig *common.AppModelConfig
	var err error
	if req.ModelConfig != nil {
		if req.ModelConfig.ModelId == "" {
			modelConfig = &common.AppModelConfig{}
		} else {
			modelConfig, err = appModelConfigModel2Proto(*req.ModelConfig)
			if err != nil {
				return err
			}
		}
	}

	var rerankConfig *common.AppModelConfig
	if req.RerankConfig != nil {
		if req.RerankConfig.ModelId == "" {
			rerankConfig = &common.AppModelConfig{}
		} else {
			rerankConfig, err = appModelConfigModel2Proto(*req.RerankConfig)
			if err != nil {
				return err
			}
		}
	}

	var qaRerankConfig *common.AppModelConfig
	if req.QARerankConfig != nil {
		if req.QARerankConfig.ModelId == "" {
			qaRerankConfig = &common.AppModelConfig{}
		} else {
			qaRerankConfig, err = appModelConfigModel2Proto(*req.QARerankConfig)
			if err != nil {
				return err
			}
		}
	}

	var knowledgeBaseConfig *rag_service.RagKnowledgeBaseConfig
	if req.KnowledgeBaseConfig != nil {
		knowledgeBaseConfig = ragKBConfigToProto(*req.KnowledgeBaseConfig)
	}

	var qaKnowledgeBaseConfig *rag_service.RagQAKnowledgeBaseConfig
	if req.QAKnowledgeBaseConfig != nil {
		qaKnowledgeBaseConfig = ragQAKBConfigToProto(*req.QAKnowledgeBaseConfig)
	}

	var sensitiveConfig *rag_service.RagSensitiveConfig
	if req.SafetyConfig != nil {
		sensitiveConfig = ragSensitiveConfigToProto(*req.SafetyConfig)
	}

	visionConfig := &rag_service.RagVisionConfig{}
	if req.VisionConfig != nil {
		visionConfig = &rag_service.RagVisionConfig{
			PicNum: req.VisionConfig.PicNum,
		}
	}
	_, err = rag.UpdateRagConfig(ctx.Request.Context(), &rag_service.UpdateRagConfigReq{
		RagId:                 req.RagID,
		ModelConfig:           modelConfig,
		RerankConfig:          rerankConfig,
		QArerankConfig:        qaRerankConfig,
		KnowledgeBaseConfig:   knowledgeBaseConfig,
		QAknowledgeBaseConfig: qaKnowledgeBaseConfig,
		SensitiveConfig:       sensitiveConfig,
		VisionConfig:          visionConfig,
		RecommendQuestion:     req.RecommendQuestion,
	})
	return err
}

func RagUpload(ctx *gin.Context, userId, orgId string, req request.RagUploadParams) (*response.RagUploadResponse, error) {
	form, err := ctx.MultipartForm()
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_file_upload_save", err.Error())
	}
	files := form.File["files"]
	if len(files) <= 0 {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_file_upload_check", fmt.Errorf("file is empty").Error())
	}
	value := form.Value
	if len(value) > 0 {
		markdown := value["markdown"]
		if len(markdown) > 0 {
			if markdown[0] == "true" {
				req.Markdown = true
			}
		}
	}

	var fileHandleList []multipart.File
	defer func() {
		if len(fileHandleList) > 0 {
			for _, file := range fileHandleList {
				err2 := file.Close()
				if err2 != nil {
					log.Errorf("RagUpload close error %s", err2)
				}
			}
		}
	}()
	params, fileNameList, fileHandleList, err := buildRagUploadHttpParams(files, fileHandleList)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_file_upload_save", err.Error())
	}
	resultArray := safe_go_util.NewSafeArray[*response.RagUploadResponseWithErr]()
	safe_go_util.SageGoWaitGroup(buildPostFileFun(ctx, req.Markdown, fileNameList, resultArray, params)...)

	return buildRagUploadResponse(resultArray)
}

func buildPostFileFun(ctx *gin.Context, markDown bool, fileNameList []string, resultArray *safe_go_util.SafeArray[*response.RagUploadResponseWithErr], params []*http_client.HttpRequestParams) []func() {
	var funcList []func()
	for index, param := range params {
		funcList = append(funcList, func() {
			//调用rag上传文档
			result, err := ragKnowHttp.PostFile(ctx, param)
			uploadFile, err := buildRagUploadFile(markDown, fileNameList[index], result, err)
			if err != nil {
				resultArray.Append(response.RagUploadError(index, err))
			} else {
				resultArray.Append(response.RagUploadSuccess(index, uploadFile))
			}
		})
	}
	return funcList
}
func buildRagUploadHttpParams(files []*multipart.FileHeader, fileHandleList []multipart.File) ([]*http_client.HttpRequestParams, []string, []multipart.File, error) {
	var params []*http_client.HttpRequestParams
	var fileNameList []string
	knowledgeConfig := config.Cfg().RagKnowledgeConfig
	var timeout int64 = 120
	if knowledgeConfig.UploadTime > 0 {
		timeout = knowledgeConfig.UploadTime
	}

	for _, file := range files {
		// 打开上传的文件
		fileHandle, err1 := file.Open()
		if err1 != nil {
			return nil, nil, fileHandleList, err1
		}
		fileHandleList = append(fileHandleList, fileHandle)
		fileNameList = append(fileNameList, file.Filename)
		randomFileName := util.NewRandomFile(file.Filename)
		paramsMap := make(map[string]string)
		paramsMap["bucket_name"] = knowledgeConfig.UploadBucket
		params = append(params, &http_client.HttpRequestParams{
			Url:        knowledgeConfig.UploadEndpoint + knowledgeConfig.UploadUri,
			MonitorKey: "rag-file-upload",
			LogLevel:   http_client.LogAll,
			Params:     paramsMap,
			FileParams: []*http_client.HttpRequestFileParams{{FileName: randomFileName, FileData: fileHandle}},
			Timeout:    time.Duration(timeout) * time.Second,
		})
	}
	return params, fileNameList, fileHandleList, nil
}

func buildRagUploadResponse(resultArray *safe_go_util.SafeArray[*response.RagUploadResponseWithErr]) (*response.RagUploadResponse, error) {
	if resultArray.Length() == 0 {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_file_upload_save", "empty url")
	}
	if resultArray.All(func(resp *response.RagUploadResponseWithErr) bool {
		return resp.Error != nil
	}) {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_file_upload_save", "upload error")
	}
	var result = &response.RagUploadResponse{}
	resultArray.Range(func(resp *response.RagUploadResponseWithErr) {
		if resp.Error != nil {
			return
		}
		result.FileList = append(result.FileList, resp.RagUploadFile)
	})
	return result, nil
}

func buildRagUploadFile(markdown bool, fileName string, result []byte, err error) (*response.RagUploadFile, error) {
	if err != nil {
		return nil, err
	}
	uploadResult := &response.RagUploadResult{}
	err = json.Unmarshal(result, uploadResult)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_file_upload_save", err.Error())
	}
	if len(uploadResult.Error) > 0 {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_file_upload_save", uploadResult.Error)
	}
	if len(uploadResult.DownloadLink) == 0 {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_file_upload_save", "empty url")
	}
	var uploadUrl = uploadResult.DownloadLink
	if markdown {
		uploadUrl = util.MdImageUrl(fileName, uploadResult.DownloadLink)
	}
	return &response.RagUploadFile{FileUrl: uploadUrl}, nil
}

func ragSensitiveConfigToProto(req request.AppSafetyConfig) *rag_service.RagSensitiveConfig {
	var sensitiveTableIds []string
	for _, v := range req.Tables {
		sensitiveTableIds = append(sensitiveTableIds, v.TableId)
	}
	sensitiveConfig := &rag_service.RagSensitiveConfig{
		Enable:   req.Enable,
		TableIds: sensitiveTableIds,
	}
	return sensitiveConfig
}

func ragKBConfigToProto(knowledgeConfig request.AppKnowledgebaseConfig) *rag_service.RagKnowledgeBaseConfig {
	result := &rag_service.RagKnowledgeBaseConfig{
		PerKnowledgeConfigs: make([]*rag_service.RagPerKnowledgeConfig, 0, len(knowledgeConfig.Knowledgebases)),
	}
	for _, knowledge := range knowledgeConfig.Knowledgebases {
		// 初始化单个知识库配置
		perConfig := &rag_service.RagPerKnowledgeConfig{
			KnowledgeId: knowledge.ID,
			GraphSwitch: knowledge.GraphSwitch,
		}
		// 构建元数据过滤条件（如果启用）
		if metaFilter := buildRagMetaFilter(knowledge.MetaDataFilterParams); metaFilter != nil {
			perConfig.RagMetaFilter = metaFilter
		}
		// 单个知识库配置添加到result
		result.PerKnowledgeConfigs = append(result.PerKnowledgeConfigs, perConfig)
	}
	result.GlobalConfig = buildRagGlobalConfig(knowledgeConfig.Config)
	return result
}

func ragQAKBConfigToProto(qaKnowledgeConfig request.AppQAKnowledgebaseConfig) *rag_service.RagQAKnowledgeBaseConfig {
	result := &rag_service.RagQAKnowledgeBaseConfig{
		PerKnowledgeConfigs: make([]*rag_service.RagPerQAKnowledgeConfig, 0, len(qaKnowledgeConfig.Knowledgebases)),
	}
	for _, knowledge := range qaKnowledgeConfig.Knowledgebases {
		// 初始化单个问答库配置
		perConfig := &rag_service.RagPerQAKnowledgeConfig{
			KnowledgeId: knowledge.ID,
		}
		// 构建元数据过滤条件（如果启用）
		if metaFilter := buildRagMetaFilter(knowledge.MetaDataFilterParams); metaFilter != nil {
			perConfig.RagMetaFilter = metaFilter
		}
		// 单个知识库配置添加到result
		result.PerKnowledgeConfigs = append(result.PerKnowledgeConfigs, perConfig)
	}
	result.GlobalConfig = buildRagQAGlobalConfig(qaKnowledgeConfig.Config)
	return result
}

// 构建单个知识库的元数据过滤条件
func buildRagMetaFilter(params *request.MetaDataFilterParams) *rag_service.RagMetaFilter {
	// 检查过滤参数是否有效（未启用则返回nil）
	if params == nil {
		return nil
	}
	if params.MetaFilterParams == nil {
		return &rag_service.RagMetaFilter{
			FilterEnable:    params.FilterEnable,
			FilterLogicType: params.FilterLogicType,
			FilterItems:     make([]*rag_service.RagMetaFilterItem, 0),
		}
	}
	// 转换过滤条件项
	filterItems := make([]*rag_service.RagMetaFilterItem, 0, len(params.MetaFilterParams))
	for _, metaParam := range params.MetaFilterParams {
		filterItems = append(filterItems, &rag_service.RagMetaFilterItem{
			Key:       metaParam.Key,
			Type:      metaParam.Type,
			Value:     metaParam.Value,
			Condition: metaParam.Condition,
		})
	}
	return &rag_service.RagMetaFilter{
		FilterEnable:    params.FilterEnable,
		FilterLogicType: params.FilterLogicType,
		FilterItems:     filterItems,
	}
}

func buildRagGlobalConfig(kbConfig request.AppKnowledgebaseParams) *rag_service.RagGlobalConfig {
	return &rag_service.RagGlobalConfig{
		MaxHistory:        kbConfig.MaxHistory,
		Threshold:         kbConfig.Threshold,
		TopK:              kbConfig.TopK,
		MatchType:         kbConfig.MatchType,
		KeywordPriority:   kbConfig.KeywordPriority,
		PriorityMatch:     kbConfig.PriorityMatch,
		SemanticsPriority: kbConfig.SemanticsPriority,
		TermWeight:        kbConfig.TermWeight,
		TermWeightEnable:  kbConfig.TermWeightEnable,
		UseGraph:          kbConfig.UseGraph,
	}
}

func buildRagQAGlobalConfig(kbConfig request.AppQAKnowledgebaseParams) *rag_service.RagQAGlobalConfig {
	return &rag_service.RagQAGlobalConfig{
		MaxHistory:        kbConfig.MaxHistory,
		Threshold:         kbConfig.Threshold,
		TopK:              kbConfig.TopK,
		MatchType:         kbConfig.MatchType,
		KeywordPriority:   kbConfig.KeywordPriority,
		PriorityMatch:     kbConfig.PriorityMatch,
		SemanticsPriority: kbConfig.SemanticsPriority,
	}
}

func DeleteRag(ctx *gin.Context, req request.RagReq) error {
	_, err := rag.DeleteRag(ctx.Request.Context(), &rag_service.RagDeleteReq{
		RagId: req.RagID,
	})
	return err
}

func GetRag(ctx *gin.Context, req request.RagReq, needPublished bool) (*response.RagInfo, error) {
	resp, err := rag.GetRagDetail(ctx.Request.Context(), &rag_service.RagDetailReq{
		RagId:   req.RagID,
		Publish: util.IfElse(needPublished, int32(1), int32(0)),
		Version: req.Version,
	})
	if err != nil {
		return nil, err
	}
	modelConfig, rerankConfig, qaRerankConfig, err := appModelRerankProto2Model(ctx, resp)
	if err != nil {
		log.Errorf("ragId: %v gets config fail: %v", req.RagID, err.Error())
	}
	appInfo, _ := app.GetAppInfo(ctx.Request.Context(), &app_service.GetAppInfoReq{AppId: req.RagID, AppType: constant.AppTypeRag})
	ragInfo := &response.RagInfo{
		RagID:                 resp.RagId,
		AppBriefConfig:        appBriefConfigProto2Model(ctx, resp.BriefConfig, constant.AppTypeRag),
		ModelConfig:           modelConfig,
		RerankConfig:          rerankConfig,
		QARerankConfig:        qaRerankConfig,
		KnowledgeBaseConfig:   ragKBConfigProto2Model(ctx, resp.KnowledgeBaseConfig),
		QAKnowledgeBaseConfig: ragKBQAConfigProto2Model(ctx, resp.QAknowledgeBaseConfig),
		SafetyConfig:          ragSafetyConfigProto2Model(ctx, resp.SensitiveConfig),
		AppPublishConfig:      request.AppPublishConfig{PublishType: appInfo.GetPublishType()},
		VisionConfig:          ragVisionConfigProto2Model(resp.VisionConfig),
		RecommendQuestion:     resp.RecommendQuestion,
	}

	return ragInfo, nil
}

func appModelRerankProto2Model(ctx *gin.Context, resp *rag_service.RagInfo) (request.AppModelConfig, request.AppModelConfig, request.AppModelConfig, error) {
	var modelConfig, rerankConfig, qaRerankConfig request.AppModelConfig
	if resp.ModelConfig.ModelId != "" {
		modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: resp.ModelConfig.ModelId})
		if err != nil {
			return request.AppModelConfig{}, request.AppModelConfig{}, request.AppModelConfig{}, err
		}
		modelConfig, err = appModelConfigProto2Model(resp.ModelConfig, modelInfo.DisplayName)
		if err != nil {
			return request.AppModelConfig{}, request.AppModelConfig{}, request.AppModelConfig{}, err
		}
	}
	if resp.RerankConfig.ModelId != "" {
		rerankInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: resp.RerankConfig.ModelId})
		if err != nil {
			return request.AppModelConfig{}, request.AppModelConfig{}, request.AppModelConfig{}, err
		}
		rerankConfig, err = appModelConfigProto2Model(resp.RerankConfig, rerankInfo.DisplayName)
		if err != nil {
			return request.AppModelConfig{}, request.AppModelConfig{}, request.AppModelConfig{}, err
		}
	}
	if resp.QArerankConfig.ModelId != "" {
		qaRerankInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: resp.QArerankConfig.ModelId})
		if err != nil {
			return request.AppModelConfig{}, request.AppModelConfig{}, request.AppModelConfig{}, err
		}
		qaRerankConfig, err = appModelConfigProto2Model(resp.QArerankConfig, qaRerankInfo.DisplayName)
		if err != nil {
			return request.AppModelConfig{}, request.AppModelConfig{}, request.AppModelConfig{}, err
		}
	}
	return modelConfig, rerankConfig, qaRerankConfig, nil
}

func ragSafetyConfigProto2Model(ctx *gin.Context, sensitiveCfg *rag_service.RagSensitiveConfig) request.AppSafetyConfig {
	var sensitiveTableList []request.SensitiveTable
	tableIds := sensitiveCfg.GetTableIds()

	if len(tableIds) != 0 {
		sensitiveWordTable, _ := safety.GetSensitiveWordTableListByIDs(ctx.Request.Context(), &safety_service.GetSensitiveWordTableListByIDsReq{TableIds: tableIds})

		if sensitiveWordTable != nil {
			for _, table := range sensitiveWordTable.List {
				sensitiveTableList = append(sensitiveTableList, request.SensitiveTable{
					TableId:   table.TableId,
					TableName: table.TableName,
				})
			}
		}
	}

	enable := sensitiveCfg.Enable
	if len(sensitiveTableList) == 0 {
		enable = false
	}

	safetyConfig := request.AppSafetyConfig{
		Enable: enable,
		Tables: sensitiveTableList,
	}
	return safetyConfig
}

func ragVisionConfigProto2Model(visionConfig *rag_service.RagVisionConfig) request.VisionConfig {
	if visionConfig == nil {
		return request.VisionConfig{}
	}
	return request.VisionConfig{
		PicNum: visionConfig.PicNum,
	}
}
func ragKBConfigProto2Model(ctx *gin.Context, kbConfig *rag_service.RagKnowledgeBaseConfig) request.AppKnowledgebaseConfig {
	if kbConfig == nil {
		return request.AppKnowledgebaseConfig{
			Knowledgebases: make([]request.AppKnowledgeBase, 0),
			Config:         request.AppKnowledgebaseParams{},
		}
	}
	knowledgeList := make([]request.AppKnowledgeBase, 0, len(kbConfig.PerKnowledgeConfigs))

	// 转换每个知识库的单独配置
	for _, perConfig := range kbConfig.PerKnowledgeConfigs {
		kbInfo, err := knowledgeBase.SelectKnowledgeDetailById(ctx.Request.Context(), &knowledgeBase_service.KnowledgeDetailSelectReq{
			KnowledgeId: perConfig.KnowledgeId,
		})
		if err != nil {
			log.Errorf("select knowledge detail error: %v", err)
			continue
		}
		// 基础信息映射
		share := kbInfo.ShareCount > 1
		var orgName string
		if share {
			orgInfo, err := iam.GetOrgInfo(ctx.Request.Context(), &iam_service.GetOrgInfoReq{OrgId: kbInfo.CreateOrgId})
			if err != nil {
				log.Errorf("get org info error: %v", err)
			} else {
				orgName = buildShareOrgName(share, orgInfo.Name)
			}
		}
		knowledge := request.AppKnowledgeBase{
			ID:          perConfig.KnowledgeId,
			Name:        kbInfo.Name,
			GraphSwitch: kbInfo.GraphSwitch,
			External:    kbInfo.External,
			Category:    kbInfo.Category,
			OrgName:     orgName,
			Share:       share,
			Avatar:      cacheKnowledgeAvatar(ctx, kbInfo.AvatarPath, kbInfo.Category),
			Description: kbInfo.Description,
		}
		// 转换元数据过滤配置
		metaFilter := perConfig.RagMetaFilter
		knowledge.MetaDataFilterParams = convertRagMetaFilterToParams(metaFilter)

		knowledgeList = append(knowledgeList, knowledge)
	}
	globalConfig := kbConfig.GlobalConfig
	if globalConfig == nil {
		globalConfig = &rag_service.RagGlobalConfig{}
	}
	appConfig := request.AppKnowledgebaseParams{
		MaxHistory:        globalConfig.MaxHistory,
		Threshold:         globalConfig.Threshold,
		TopK:              globalConfig.TopK,
		MatchType:         globalConfig.MatchType,
		KeywordPriority:   globalConfig.KeywordPriority,
		PriorityMatch:     globalConfig.PriorityMatch,
		SemanticsPriority: globalConfig.SemanticsPriority,
		TermWeight:        globalConfig.TermWeight,
		TermWeightEnable:  globalConfig.TermWeightEnable,
		UseGraph:          globalConfig.UseGraph,
	}
	return request.AppKnowledgebaseConfig{
		Knowledgebases: knowledgeList,
		Config:         appConfig,
	}
}

func ragKBQAConfigProto2Model(ctx *gin.Context, kbConfig *rag_service.RagQAKnowledgeBaseConfig) request.AppQAKnowledgebaseConfig {
	if kbConfig == nil {
		return request.AppQAKnowledgebaseConfig{
			Knowledgebases: make([]request.AppQAKnowledgeBase, 0),
			Config:         request.AppQAKnowledgebaseParams{},
		}
	}
	knowledgeList := make([]request.AppQAKnowledgeBase, 0, len(kbConfig.PerKnowledgeConfigs))

	// 转换每个问答库的单独配置
	for _, perConfig := range kbConfig.PerKnowledgeConfigs {
		kbInfo, err := knowledgeBase.SelectKnowledgeDetailById(ctx.Request.Context(), &knowledgeBase_service.KnowledgeDetailSelectReq{
			KnowledgeId: perConfig.KnowledgeId,
		})
		if err != nil {
			log.Errorf("select qa detail error: %v", err)
			continue
		}
		// 基础信息映射
		share := kbInfo.ShareCount > 1
		var orgName string
		if share {
			orgInfo, err := iam.GetOrgInfo(ctx.Request.Context(), &iam_service.GetOrgInfoReq{OrgId: kbInfo.CreateOrgId})
			if err != nil {
				log.Errorf("get org info error: %v", err)
			} else {
				orgName = buildShareOrgName(share, orgInfo.Name)
			}
		}
		knowledge := request.AppQAKnowledgeBase{
			ID:          perConfig.KnowledgeId,
			Name:        kbInfo.Name,
			Category:    kbInfo.Category,
			OrgName:     orgName,
			Share:       share,
			Avatar:      cacheKnowledgeAvatar(ctx, kbInfo.AvatarPath, kbInfo.Category),
			Description: kbInfo.Description,
		}
		// 转换元数据过滤配置
		metaFilter := perConfig.RagMetaFilter
		knowledge.MetaDataFilterParams = convertRagMetaFilterToParams(metaFilter)

		knowledgeList = append(knowledgeList, knowledge)
	}
	globalConfig := kbConfig.GlobalConfig
	if globalConfig == nil {
		globalConfig = &rag_service.RagQAGlobalConfig{}
	}
	appConfig := request.AppQAKnowledgebaseParams{
		MaxHistory:        globalConfig.MaxHistory,
		Threshold:         globalConfig.Threshold,
		TopK:              globalConfig.TopK,
		MatchType:         globalConfig.MatchType,
		KeywordPriority:   globalConfig.KeywordPriority,
		PriorityMatch:     globalConfig.PriorityMatch,
		SemanticsPriority: globalConfig.SemanticsPriority,
	}
	return request.AppQAKnowledgebaseConfig{
		Knowledgebases: knowledgeList,
		Config:         appConfig,
	}
}

func convertRagMetaFilterToParams(metaFilter *rag_service.RagMetaFilter) *request.MetaDataFilterParams {
	if metaFilter == nil {
		return nil
	}
	// 转换过滤条件项
	filterParams := make([]*request.MetaFilterParams, 0, len(metaFilter.FilterItems))
	for _, item := range metaFilter.FilterItems {
		filterParams = append(filterParams, &request.MetaFilterParams{
			Key:       item.Key,
			Type:      item.Type,
			Value:     item.Value,
			Condition: item.Condition,
		})
	}
	return &request.MetaDataFilterParams{
		FilterEnable:     metaFilter.FilterEnable,
		FilterLogicType:  metaFilter.FilterLogicType,
		MetaFilterParams: filterParams, // 映射过滤条件列表
	}
}

func CopyRag(ctx *gin.Context, userId, orgId string, req request.RagReq) (*request.RagReq, error) {
	resp, err := rag.CopyRag(ctx.Request.Context(), &rag_service.CopyRagReq{
		Identity: &rag_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
		RagId: req.RagID,
	})
	if err != nil {
		return nil, err
	}
	return &request.RagReq{
		RagID: resp.RagId,
	}, err
}
