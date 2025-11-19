package request

type GetExplorationAppListRequest struct {
	Name       string `form:"name" json:"name"`             // Application name
	AppType    string `form:"appType" json:"appType"`       // Application type
	SearchType string `form:"searchType" json:"searchType"` // Search type (all (all), favorite (show favorites), private (show privately published), history (history applications))
}

func (g GetExplorationAppListRequest) Check() error {
	return nil
}

type ChangeExplorationAppFavoriteRequest struct {
	AppId      string `json:"appId" validate:"required"`   // application id
	AppType    string `json:"appType" validate:"required"` // Application type
	IsFavorite bool   `json:"isFavorite"`                  // Whether to collect
}

func (c ChangeExplorationAppFavoriteRequest) Check() error {
	return nil
}
