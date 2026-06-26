package response

type AppVersionInfo struct {
	Version        string `json:"version"`
	Desc           string `json:"desc"`
	CreatedAt      string `json:"createdAt"`
	PublishType    string `json:"publishType"`
	IsClosedSource bool   `json:"isClosedSource"` // 仅 skill 类型：当前最新发布是否闭源
}
