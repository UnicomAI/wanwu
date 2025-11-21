package request

type ImportOrUpdateModelRequest struct {
	ModelConfig
}

func (o *ImportOrUpdateModelRequest) Check() error {
	return o.ModelConfig.Check()
}

type DeleteModelRequest struct {
	BaseModelRequest
}

func (o *DeleteModelRequest) Check() error {
	return nil
}

type GetModelRequest struct {
	BaseModelRequest
}

func (o *GetModelRequest) Check() error {
	return nil
}

type ListModelsRequest struct {
	ModelType   string `json:"modelType" form:"modelType" `    // Model type
	Provider    string `json:"provider" form:"provider"`       // model supplier
	DisplayName string `json:"displayName" form:"displayName"` // Model display name
	IsActive    bool   `json:"isActive" form:"isActive"`       // Enabled status (true: enabled)
}

func (o *ListModelsRequest) Check() error {
	return nil
}

type ListTypeModelsRequest struct {
	ModelType string `json:"modelType" form:"modelType" ` // Model type
}

func (o *ListTypeModelsRequest) Check() error {
	return nil
}

type ModelStatusRequest struct {
	BaseModelRequest
	IsActive bool `json:"isActive"` // Enabled status (true: enabled, false: disabled)
}

func (o *ModelStatusRequest) Check() error {
	return nil
}

type GetModelByIdRequest struct {
	BaseModelRequest
}

func (o *GetModelByIdRequest) Check() error {
	return nil
}
