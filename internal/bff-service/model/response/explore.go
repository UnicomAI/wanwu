package response

type ExplorationAppInfo struct {
	AppBriefInfo
	IsFavorite bool `json:"isFavorite"` // 收藏标签 [EN] favorite tags
	User       User `json:"user"`       // 作者信息 [EN] Author information
}

type User struct {
	UserId   string `json:"userId"`   // 用户ID [EN] User ID
	UserName string `json:"userName"` // 用户名称 [EN] Username
}
