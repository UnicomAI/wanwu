package response

type ApiResponse struct {
	ApiID     string `json:"apiId" `    // ApiID
	ApiKey    string `json:"apiKey"`    // Generated ApiKey
	CreatedAt string `json:"createdAt"` // Time when ApiKey was created
}
