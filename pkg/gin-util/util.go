package gin_util

import (
	"encoding/json"
	"fmt"
	"net/http"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/pkg/i18n"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// gin.Contex key
const (
	CLAIMS = "claims"
	STATUS = "STATUS"
	RESULT = "RESULT"

	// http header
	X_LANGUAGE  = "X-Language" // Current language
	X_ORG_ID    = "X-Org-Id"   // current organization
	X_CLIENT_ID = "X-Client-Id"

	// gin.Context
	USER_ID   = "USER_ID"   // current user
	IS_ADMIN  = "IS_ADMIN"  // USER_ID is a built-in administrator role for the current organization X_ORG_ID
	IS_SYSTEM = "IS_SYSTEM" // Whether the current organization X_ORG_ID is [system]

	// openapi related
	APP_ID   = "APP_ID"
	APP_TYPE = "APP_TYPE"

	ANSWER = "ANSWER"

	// OAuth related
	OAUTH_SCOPE     = "SCOPE"
	OAUTH_CLIENT_ID = "OAuth_Client_ID"
)

// http common query key
const (
	PageNo   = "pageNo"
	PageSize = "pageSize"
)

var (
	_v *validator.Validate
)

// --- gin validator ---

type ginValidator struct{}

func (gv *ginValidator) Engine() interface{} {
	return _v
}

func (gv *ginValidator) ValidateStruct(obj interface{}) error {
	return util.Validate(obj)
}

func InitValidator() error {
	if err := util.InitValidator(); err != nil {
		return err
	}
	binding.Validator = &ginValidator{}
	return nil
}

// --- bind ---

type iChecker interface {
	Check() error
}

func Bind(ctx *gin.Context, param iChecker) bool {
	if err := ctx.ShouldBindBodyWith(param, binding.JSON); err != nil {
		ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "", err.Error())
		return false
	}
	if err := param.Check(); err != nil {
		ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "", err.Error())
		return false
	}
	return true
}

func BindUri(ctx *gin.Context, param iChecker) bool {
	if err := ctx.ShouldBindUri(param); err != nil {
		ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "", err.Error())
		return false
	}
	if err := param.Check(); err != nil {
		ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "", err.Error())
		return false
	}
	return true
}

func BindForm(ctx *gin.Context, param iChecker) bool {
	if err := ctx.ShouldBind(param); err != nil {
		ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "", err.Error())
		return false
	}
	if err := param.Check(); err != nil {
		ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "", err.Error())
		return false
	}
	return true
}

func BindQuery(ctx *gin.Context, param iChecker) bool {
	if err := ctx.ShouldBindQuery(param); err != nil {
		ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "", err.Error())
		return false
	}
	if err := param.Check(); err != nil {
		ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "", err.Error())
		return false
	}
	return true
}

// --- response ---

// Response returns 200 and data information, or 400 and err information, err has i18n
func Response(ctx *gin.Context, data interface{}, err error) {
	if err != nil {
		ResponseErr(ctx, err)
		return
	}
	ResponseOKWithData(ctx, data)
}

// ResponseOK returns 200
func ResponseOK(ctx *gin.Context) {
	ResponseDetail(ctx, http.StatusOK, codes.OK, nil, "")
}

// ResponseOKWithData returns 200 and data information
func ResponseOKWithData(ctx *gin.Context, data interface{}) {
	ResponseDetail(ctx, http.StatusOK, codes.OK, data, "")
}

// ResponseDetail returns 400 and err information, err has i18n
func ResponseErr(ctx *gin.Context, err error) {
	ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
}

// ResponseDetail returns httpStatus and err information, err has i18n
func ResponseErrWithStatus(ctx *gin.Context, httpStatus int, err error) {
	st, ok := status.FromError(err)
	if !ok {
		ResponseDetail(ctx, httpStatus, codes.Code(err_code.Code_BFFGeneral), nil, fmt.Sprintf("[i18n] %v", err))
		return
	}
	for _, detail := range st.Details() {
		switch detail := detail.(type) {
		case *err_code.Status:
			ResponseDetail(ctx, httpStatus, st.Code(), nil, I18nCodeOrKey(ctx, err_code.Code(st.Code()), detail.TextKey, detail.Args...))
			return
		}
	}
	ResponseDetail(ctx, httpStatus, st.Code(), nil, fmt.Sprintf("[i18n] %v", st.Message()))
}

// ResponseErrCodeKey returns 400/code and error message, code/key has i18n
func ResponseErrCodeKey(ctx *gin.Context, code err_code.Code, textKey string, args ...string) {
	ResponseDetail(ctx, http.StatusBadRequest, codes.Code(code), nil, I18nCodeOrKey(ctx, code, textKey, args...))
}

// ResponseErrCodeKey returns httpStatus/code and error information, code/key has i18n
func ResponseErrCodeKeyWithStatus(ctx *gin.Context, httpStatus int, code err_code.Code, textKey string, args ...string) {
	ResponseDetail(ctx, httpStatus, codes.Code(code), nil, I18nCodeOrKey(ctx, code, textKey, args...))
}

// ResponseDetail directly returns httpStatus/code/data/msg, msg has no i18n
func ResponseDetail(ctx *gin.Context, httpStatus int, code codes.Code, data interface{}, msg string) {
	resp := &response{
		Code: int64(code),
		Data: data,
		Msg:  msg,
	}
	b, _ := json.Marshal(resp)
	ctx.Set(STATUS, httpStatus)
	ctx.Set(RESULT, string(b))
	ctx.JSON(httpStatus, resp)
}

// ResponseRawByte directly returns []byte data
func ResponseRawByte(ctx *gin.Context, httpStatus int, data []byte) {
	ctx.Set(STATUS, httpStatus)
	ctx.Set(RESULT, string(data))
	ctx.Data(httpStatus, "application/json; charset=utf-8", data)
}

// response is consistent with Response in model/response, which is only used for swagger generation
type response struct {
	Code int64       `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// --- i18n ---

func I18nCode(ctx *gin.Context, code err_code.Code, args ...string) string {
	return I18nCodeOrKey(ctx, code, "", args...)
}

func I18nKey(ctx *gin.Context, key string, args ...string) string {
	return I18nCodeOrKey(ctx, 0, key, args...)
}

func I18nCodeOrKey(ctx *gin.Context, code err_code.Code, key string, args ...string) string {
	return i18n.ByCodeOrKey(getLanguage(ctx), code, key, args)
}

// --- internal ---

func getLanguage(ctx *gin.Context) i18n.Lang {
	// 1. Prioritize the language of the header
	language := ctx.GetHeader(X_LANGUAGE)
	// 2. Secondly, the language set by the user
	if language == "" {
		language = ctx.GetString(X_LANGUAGE)
	}
	// 3. Change the system default language again
	if language == "" {
		language = string(i18n.DefaultLang())
	}
	return i18n.Lang(language)
}
