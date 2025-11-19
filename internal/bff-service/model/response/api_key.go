package response

type ApiResponse struct {
	ApiID     string `json:"apiId" `    // ApiID
	ApiKey    string `json:"apiKey"`    // 生成的ApiKey [EN] Generated ApiKey
	CreatedAt string `json:"createdAt"` // 创建ApiKey的时间 [EN] Time when ApiKey was created
}
