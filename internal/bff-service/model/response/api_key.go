package response

// UserInfoByApiKey 通过apikey获取的用户信息
type UserInfoByApiKey struct {
	UserID   string `json:"userId"`
	OrgID    string `json:"orgId"`
	Username string `json:"username"`
}

type APIKeyDetailResponse struct {
	KeyID     string `json:"keyId" `    // ID
	Key       string `json:"key"`       // 生成的ApiKey
	Creator   string `json:"creator"`   // 创建者
	Name      string `json:"name"`      // 名称
	Desc      string `json:"desc"`      // 描述
	ExpiredAt string `json:"expiredAt"` // 过期时间
	CreatedAt string `json:"createdAt"` // 创建ApiKey的时间
	Status    bool   `json:"status"`    // 状态
}
