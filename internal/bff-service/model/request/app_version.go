package request

type GetAppVersionListRequest struct {
	AppId   string `form:"appId" json:"appId" validate:"required"`     // 对应应用 ID
	AppType string `form:"appType" json:"appType" validate:"required"` // 支持的类型
}

func (g *GetAppVersionListRequest) Check() error {
	return nil
}

type UpdateAppVersionRequest struct {
	AppType     string `json:"appType" validate:"required"`     // 应用类型
	AppId       string `json:"appId" validate:"required"`       // 应用 ID
	Desc        string `json:"desc"`                            // 描述
	Extra       string `json:"extra"`                           // 描述
	PublishType string `json:"publishType" validate:"required"` // 发布类型(public:系统公开发布,organization:组织公开发布,private:私密发布)
}

func (u *UpdateAppVersionRequest) Check() error {
	return nil
}

type RollbackAppVersionRequest struct {
	AppType string `json:"appType" validate:"required"`
	AppID   string `json:"appId" validate:"required"`
	Version string `json:"version" validate:"required"` // 目标回滚版本
}

func (r *RollbackAppVersionRequest) Check() error {
	return nil
}
