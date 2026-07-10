package middleware

import (
	"encoding/json"
	"io"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func Record(ctx *gin.Context) {
	var req string
	var err error
	traceID := trace_util.GetTraceID(ctx)
	if ctx.ContentType() == gin.MIMEJSON {
		if req, err = requestBody(ctx); err != nil {
			log.Errorf("[%v] | %s | %v | %v", ctx.Request.Method, traceID, requestFullPath(ctx), err)
			gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "", err.Error())
			ctx.Abort()
			return
		}
	}
	ctx.Next()

	resp := ctx.GetString(gin_util.RESULT)
	log.Debugf("[%v] | %s | %v | %v | %v", ctx.Request.Method, traceID, requestFullPath(ctx), req, resp)
}

func requestFullPath(ctx *gin.Context) string {
	if ctx.Request.URL.RawQuery != "" {
		return ctx.Request.URL.Path + "?" + ctx.Request.URL.RawQuery
	}
	return ctx.Request.URL.Path
}

func getFieldValue(ctx *gin.Context, fieldName string) string {
	//尝试从query中获取field
	value := ctx.Query(fieldName)
	if len(value) > 0 {
		return value
	}
	// 再从 form（multipart/url-encoded）提取
	value = ctx.PostForm(fieldName)
	if len(value) > 0 {
		return value
	}
	if binding.MIMEJSON != ctx.ContentType() {
		return ""
	}
	//获取原始数据
	body, err := requestBody(ctx)
	if err != nil || len(body) == 0 {
		return ""
	}
	//构造参数对应map
	paramsMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(body), &paramsMap)
	if err != nil {
		return ""
	}
	//获取对应filed的值
	fieldValue := paramsMap[fieldName]
	if fieldValue == nil {
		return ""
	}
	retValue, ok := fieldValue.(string)
	if !ok {
		return ""
	}
	return retValue
}

func requestBody(ctx *gin.Context) (string, error) {
	var body []byte
	var err error
	if cb, ok := ctx.Get(gin.BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body == nil {
		body, err = io.ReadAll(ctx.Request.Body)
		if err != nil {
			return "", err
		}
		ctx.Set(gin.BodyBytesKey, body)
	}

	// avoid err: unexpected end of JSON input
	if strings.TrimSpace(string(body)) == "" {
		return "", nil
	}

	kv := make(map[string]interface{})
	if err = json.Unmarshal(body, &kv); err != nil {
		return "", err
	}
	if b, err := json.Marshal(kv); err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}
