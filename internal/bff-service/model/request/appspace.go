package request

import (
	"fmt"
	"regexp"
)

var versionRegexp = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

type DeleteAppSpaceAppRequest struct {
	AppId   string `json:"appId" validate:"required"`   // 应用ID
	AppType string `json:"appType" validate:"required"` // 应用类型
}

func (req DeleteAppSpaceAppRequest) Check() error {
	return nil
}

type GetAppSpaceAppListRequest struct {
	Name string `form:"name" json:"name"`
}

type GetAppSpaceAppListByTypeRequest struct {
	Name    string `form:"name" json:"name"`
	AppType string `form:"appType" json:"appType"`
}

type PublishAppRequest struct {
	AppId       string `json:"appId" validate:"required"`   // 应用ID
	AppType     string `json:"appType" validate:"required"` // 应用类型
	Version     string `json:"version" validate:"required"`
	Desc        string `json:"desc"`                            // 描述
	Extra       string `json:"extra"`                           // 扩展信息，具体字段格式各个业务自己定义
	PublishType string `json:"publishType" validate:"required"` // 发布类型(public:系统公开发布,organization:组织公开发布,private:私密发布)
}

func (req PublishAppRequest) Check() error {
	if !versionRegexp.MatchString(req.Version) {
		return fmt.Errorf("version must be in format 'vX.Y.Z'")
	}
	return nil
}

type UnPublishAppRequest struct {
	AppId   string `json:"appId"`   // 应用ID
	AppType string `json:"appType"` // 应用类型
}

func (req UnPublishAppRequest) Check() error {
	return nil
}

type GetAppBaseUrlRequest struct {
	AppId   string `form:"appId" json:"appId" validate:"required"`     // 应用ID
	AppType string `form:"appType" json:"appType" validate:"required"` // 应用类型
}

func (req GetAppBaseUrlRequest) Check() error {
	return nil
}

func (o *GetAppSpaceAppListRequest) Check() error {
	return nil
}

func (o *GetAppSpaceAppListByTypeRequest) Check() error {
	return nil
}
