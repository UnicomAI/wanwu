package request

// APIKeyStatisticV2Req API Key 统计 V2 公共筛选（无 viewScope）。
type APIKeyStatisticV2Req struct {
	StatisticFilter
	StartDate   string   `json:"startDate" validate:"required"` // yyyy-MM-dd
	EndDate     string   `json:"endDate"   validate:"required"`
	ApiKeyIds   []string `json:"apiKeyIds"`   // 空=全部
	MethodPaths []string `json:"methodPaths"` // 空=全部，如 POST-/v1/chat/completions
}

func (r *APIKeyStatisticV2Req) Check() error { return nil }

// APIKeyStatisticV2ChartReq 趋势 + 排行（合并）
type APIKeyStatisticV2ChartReq struct {
	APIKeyStatisticV2Req
	Limit int `json:"limit"` // 排行 TopN，默认 5
}

func (r *APIKeyStatisticV2ChartReq) Check() error { return nil }

// APIKeyStatisticV2ListReq 调用统计主表（分页，可排序）
type APIKeyStatisticV2ListReq struct {
	APIKeyStatisticV2Req
	StatisticSortReq
	PageNo   int `json:"pageNo"   validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

func (r *APIKeyStatisticV2ListReq) Check() error { return nil }

// APIKeyStatisticV2AppListReq 应用钻取（分页，需 apiKeyId+methodPath 定位主表行；apiKeyId 唯一归属用户/组织）
type APIKeyStatisticV2AppListReq struct {
	APIKeyStatisticV2Req
	ApiKeyId   string `json:"apiKeyId"   validate:"required"`
	MethodPath string `json:"methodPath" validate:"required"`
	StatisticSortReq
	PageNo   int `json:"pageNo"   validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

func (r *APIKeyStatisticV2AppListReq) Check() error { return nil }

// APIKeyStatisticV2ModelListReq 模型钻取（分页，需 apiKeyId+methodPath 定位主表行）
type APIKeyStatisticV2ModelListReq struct {
	APIKeyStatisticV2Req
	ApiKeyId   string `json:"apiKeyId"   validate:"required"`
	MethodPath string `json:"methodPath" validate:"required"`
	StatisticSortReq
	PageNo   int `json:"pageNo"   validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

func (r *APIKeyStatisticV2ModelListReq) Check() error { return nil }

// APIKeyStatisticV2RecordReq 调用明细列表（分页，不支持用户排序；后端固定 created_at 降序）
type APIKeyStatisticV2RecordReq struct {
	APIKeyStatisticV2Req
	PageNo   int `json:"pageNo"   validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

func (r *APIKeyStatisticV2RecordReq) Check() error { return nil }

// --- 导出（无分页，可带排序）---

type APIKeyStatisticV2ExportListReq struct {
	APIKeyStatisticV2Req
	StatisticSortReq
}

func (r *APIKeyStatisticV2ExportListReq) Check() error { return nil }

type APIKeyStatisticV2AppExportReq struct {
	APIKeyStatisticV2Req
	ApiKeyId   string `json:"apiKeyId"   validate:"required"`
	MethodPath string `json:"methodPath" validate:"required"`
	StatisticSortReq
}

func (r *APIKeyStatisticV2AppExportReq) Check() error { return nil }

type APIKeyStatisticV2ModelExportReq struct {
	APIKeyStatisticV2Req
	ApiKeyId   string `json:"apiKeyId"   validate:"required"`
	MethodPath string `json:"methodPath" validate:"required"`
	StatisticSortReq
}

func (r *APIKeyStatisticV2ModelExportReq) Check() error { return nil }

type APIKeyStatisticV2RecordExportReq struct {
	APIKeyStatisticV2Req
}

func (r *APIKeyStatisticV2RecordExportReq) Check() error { return nil }

// APIKeyStatisticV2SelectReq API Key 下拉（组织→用户→API Key 级联）
type APIKeyStatisticV2SelectReq struct {
	StatisticFilter
}

func (r *APIKeyStatisticV2SelectReq) Check() error { return nil }
