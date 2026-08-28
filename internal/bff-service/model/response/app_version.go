package response

type AppVersionInfo struct {
	Version     string `json:"version"`
	Desc        string `json:"desc"`
	Extra       string `json:"extra"`
	CreatedAt   string `json:"createdAt"`
	PublishType string `json:"publishType"`
}
