package response

type ExplorationAppInfo struct {
	AppBriefInfo
	IsFavorite bool `json:"isFavorite"` // favorite tags
	User       User `json:"user"`       // Author information
}

type User struct {
	UserId   string `json:"userId"`   // User ID
	UserName string `json:"userName"` // Username
}
