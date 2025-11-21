package middleware

import (
	"errors"
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

const (
	KnowledgeView   int32 = 0
	KnowledgeEdit   int32 = 10
	KnowledgeGrant  int32 = 20
	KnowledgeSystem int32 = 30
)

// AuthKnowledgeDoc verifies knowledge base permissions
func AuthKnowledgeDoc(fieldName string, permissionType int32) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()
		//1. Get the value
		value := getFieldValue(ctx, fieldName)
		if len(value) == 0 {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("docId is required"))
			ctx.Abort()
			return
		}
		//2. Get the knowledge base id based on docId
		knowledgeId, err := searchKnowledgeId(ctx, value)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
		//3. Verify user authorization permissions
		err = knowledgeGrantUser(ctx, knowledgeId, permissionType)
		//4.Exception handling
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
	}
}

// AuthKnowledgeIfHas verifies knowledge base permissions, allowing fields to be empty
func AuthKnowledgeIfHas(fieldName string, permissionType int32) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()
		//1. Get the value
		value := getFieldValue(ctx, fieldName)
		if len(value) == 0 {
			ctx.Next()
			return
		}
		//2. Verify user authorization permissions
		err := knowledgeGrantUser(ctx, value, permissionType)
		//3.Return results
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
	}
}

// AuthKnowledge verifies knowledge base permissions
func AuthKnowledge(fieldName string, permissionType int32) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()
		//1. Get the value
		value := getFieldValue(ctx, fieldName)
		if len(value) == 0 {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("knowledgeId is required"))
			ctx.Abort()
			return
		}
		//2. Verify user authorization permissions
		err := knowledgeGrantUser(ctx, value, permissionType)
		//3.Return results
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
	}
}

func searchKnowledgeId(ctx *gin.Context, docId string) (string, error) {
	docInfo, err := service.GetDocDetail(ctx, "", "", docId)
	if err != nil {
		return "", err
	}
	return docInfo.KnowledgeId, nil
}

func knowledgeGrantUser(ctx *gin.Context, knowledgeId string, permissionType int32) error {
	// userID
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	// orgID
	orgID := getOrgID(ctx)
	if len(orgID) == 0 {
		return errors.New("X-Org-Id is empty")
	}

	// check user knowledge permission
	if err = service.CheckKnowledgeUserPermission(ctx, userID, orgID, knowledgeId, permissionType); err != nil {
		return err
	}
	return nil
}
