package request

type GetExplorationAppListRequest struct {
	Name       string `form:"name" json:"name"`             // 应用名称 [EN] Application name
	AppType    string `form:"appType" json:"appType"`       // 应用类型 [EN] Application type
	SearchType string `form:"searchType" json:"searchType"` // 搜索类型(all(全部),favorite(显示收藏的),private(显示私密发布的),history(历史应用)) [EN] Search type (all (all), favorite (show favorites), private (show privately published), history (history applications))
}

func (g GetExplorationAppListRequest) Check() error {
	return nil
}

type ChangeExplorationAppFavoriteRequest struct {
	AppId      string `json:"appId" validate:"required"`   // 应用id [EN] application id
	AppType    string `json:"appType" validate:"required"` // 应用类型 [EN] Application type
	IsFavorite bool   `json:"isFavorite"`                  // 是否收藏 [EN] Whether to collect
}

func (c ChangeExplorationAppFavoriteRequest) Check() error {
	return nil
}
