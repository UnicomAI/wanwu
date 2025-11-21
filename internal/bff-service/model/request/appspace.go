package request

type DeleteAppSpaceAppRequest struct {
	AppId   string `json:"appId" validate:"required"`   // Application ID
	AppType string `json:"appType" validate:"required"` // Application type
}

func (req DeleteAppSpaceAppRequest) Check() error {
	return nil
}

type GetAppSpaceAppListRequest struct {
	Name    string `form:"name" json:"name"`
	AppType string `form:"appType" json:"appType"`
}

type PublishAppRequest struct {
	AppId       string `json:"appId"`       // Application ID
	AppType     string `json:"appType"`     // Application type
	PublishType string `json:"publishType"` // Release type (public: system public release, organization: organization public release, private: private release)
}

func (req PublishAppRequest) Check() error {
	return nil
}

type UnPublishAppRequest struct {
	AppId   string `json:"appId"`   // Application ID
	AppType string `json:"appType"` // Application type
}

func (req UnPublishAppRequest) Check() error {
	return nil
}

type GetApiBaseUrlRequest struct {
	AppId   string `form:"appId" json:"appId" validate:"required"`     // Application ID
	AppType string `form:"appType" json:"appType" validate:"required"` // Application type
}

func (req GetApiBaseUrlRequest) Check() error {
	return nil
}

func (o *GetAppSpaceAppListRequest) Check() error {
	return nil
}
