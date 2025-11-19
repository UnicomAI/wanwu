package request

type DeleteAppSpaceAppRequest struct {
	AppId   string `json:"appId" validate:"required"`   // 应用ID [EN] Application ID
	AppType string `json:"appType" validate:"required"` // 应用类型 [EN] Application type
}

func (req DeleteAppSpaceAppRequest) Check() error {
	return nil
}

type GetAppSpaceAppListRequest struct {
	Name    string `form:"name" json:"name"`
	AppType string `form:"appType" json:"appType"`
}

type PublishAppRequest struct {
	AppId       string `json:"appId"`       // 应用ID [EN] Application ID
	AppType     string `json:"appType"`     // 应用类型 [EN] Application type
	PublishType string `json:"publishType"` // 发布类型(public:系统公开发布,organization:组织公开发布,private:私密发布) [EN] Release type (public: system public release, organization: organization public release, private: private release)
}

func (req PublishAppRequest) Check() error {
	return nil
}

type UnPublishAppRequest struct {
	AppId   string `json:"appId"`   // 应用ID [EN] Application ID
	AppType string `json:"appType"` // 应用类型 [EN] Application type
}

func (req UnPublishAppRequest) Check() error {
	return nil
}

type GetApiBaseUrlRequest struct {
	AppId   string `form:"appId" json:"appId" validate:"required"`     // 应用ID [EN] Application ID
	AppType string `form:"appType" json:"appType" validate:"required"` // 应用类型 [EN] Application type
}

func (req GetApiBaseUrlRequest) Check() error {
	return nil
}

func (o *GetAppSpaceAppListRequest) Check() error {
	return nil
}
