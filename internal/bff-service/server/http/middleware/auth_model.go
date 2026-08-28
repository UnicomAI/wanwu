package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// AuthModelByModelId 校验多个可能的 modelId 字段路径
// fieldPaths: 如 []string{"modelConfig.modelId", "recommendConfig.modelId"}
func AuthModelByModelId(fields []string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()

		modelIds := extractFieldsFromRequest(ctx, fields)
		if len(modelIds) == 0 {
			ctx.Next()
			return
		}

		if err := checkModelPermission(ctx, modelIds); err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// AuthModelByUuid 校验多个可能的 ModelUuid 字段路径
// fieldPaths: 如 []string{"modelConfig.modelId", "recommendConfig.modelId"}
func AuthModelByUuid(fields []string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()

		uuids := extractFieldsFromRequest(ctx, fields)
		if len(uuids) == 0 {
			ctx.Next()
			return
		}

		modelIds, err := service.ListModelIdsByUuids(ctx, uuids)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}

		if err := checkModelPermission(ctx, modelIds); err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// checkModelPermission 检查模型权限
func checkModelPermission(ctx *gin.Context, modelIds []string) error {
	// 获取用户和组织 ID
	userID, _ := getUserID(ctx)
	if userID == "" {
		return grpc_util.ErrorStatus(err_code.Code_BFFAuth, "auth model userID not found")
	}

	orgID, _ := getOrgID(ctx)
	if orgID == "" {
		return grpc_util.ErrorStatus(err_code.Code_BFFAuth, "auth model orgID not found")
	}

	// 校验模型权限
	_, err := service.CheckModelUserPermission(ctx, userID, orgID, modelIds)
	return err
}

// extractFieldsFromRequest 从请求中提取字段值
func extractFieldsFromRequest(ctx *gin.Context, fields []string) []string {
	var values []string

	// 1. 尝试从 Query 参数获取（仅支持顶层字段，不支持嵌套）
	for _, field := range fields {
		// 如果路径不含 "."，说明可能是 query 参数
		if !strings.Contains(field, ".") {
			if val := ctx.Query(field); val != "" {
				values = append(values, val)
			}
		}
	}

	// 2. 从 JSON Body 提取（支持嵌套）
	if ctx.ContentType() == binding.MIMEJSON {
		bodyStr, _ := requestBody(ctx)
		if bodyStr != "" {
			var paramsMap map[string]interface{}
			if json.Unmarshal([]byte(bodyStr), &paramsMap) == nil {
				for _, field := range fields {
					if val, ok := getNestedValue(paramsMap, field); ok {
						strVal, ok := val.(string)
						if !ok {
							continue
						}
						values = append(values, strVal)
					}
				}
			}
		}
	}

	// 3. 去重
	uniqueValues := make(map[string]bool)
	var result []string
	for _, val := range values {
		if !uniqueValues[val] {
			uniqueValues[val] = true
			result = append(result, val)
		}
	}

	return result
}

// extractFieldListFromRequest 从请求中提取字段值列表，支持 string 单值与 []string 数组。
// field: 字段路径，支持顶层（"logIds"）和嵌套（"data.logIds"）。
func extractFieldListFromRequest(ctx *gin.Context, field string) []string {
	var values []string

	// 1. 尝试从 Query 参数获取（仅支持顶层字段，不支持嵌套；query 不支持数组单字段，作单值处理）
	if !strings.Contains(field, ".") {
		if val := ctx.Query(field); val != "" {
			values = append(values, val)
		}
	}

	// 2. 从 JSON Body 提取（支持嵌套，且支持 string 与 []string）
	if ctx.ContentType() == binding.MIMEJSON {
		bodyStr, _ := requestBody(ctx)
		if bodyStr != "" {
			var paramsMap map[string]interface{}
			if json.Unmarshal([]byte(bodyStr), &paramsMap) == nil {
				if val, ok := getNestedValue(paramsMap, field); ok {
					switch v := val.(type) {
					case string:
						if v != "" {
							values = append(values, v)
						}
					case []interface{}:
						for _, item := range v {
							if s, ok := item.(string); ok && s != "" {
								values = append(values, s)
							}
						}
					}
				}
			}
		}
	}

	// 3. 去重
	uniqueValues := make(map[string]bool)
	var result []string
	for _, val := range values {
		if !uniqueValues[val] {
			uniqueValues[val] = true
			result = append(result, val)
		}
	}

	return result
}

func getNestedValue(data map[string]interface{}, path string) (interface{}, bool) {
	keys := strings.Split(path, ".")
	var current interface{} = data

	for _, key := range keys {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[key]
			if current == nil {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return current, true
}
