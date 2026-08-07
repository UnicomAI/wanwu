package request

// StatisticSortReq 列表排序（嵌入分页列表请求）。
// sortField 仅支持数值类指标（Tokens/调用次数/失败次数/失败率/耗时等）；未命中白名单时回退默认排序。
type StatisticSortReq struct {
	SortField string `json:"sortField" example:"callCount"` // 见各列表接口 Description 中的可选值
	SortOrder string `json:"sortOrder" example:"desc"`      // asc|desc，默认 desc
}

// ModelStatisticV2Req 模型统计 V2 公共筛选（概览/趋势/排行/导出基底）
type ModelStatisticV2Req struct {
	StatisticFilter
	StartDate string   `json:"startDate" validate:"required"` // yyyy-MM-dd
	EndDate   string   `json:"endDate"   validate:"required"`
	ModelType string   `json:"modelType" validate:"required"` // llm|embedding|...
	Models    []string `json:"models"`                        // modelId 列表，空=全部
	ViewScope string   `json:"viewScope" validate:"required"` // published|used
}

func (r *ModelStatisticV2Req) Check() error {
	return checkStatisticViewScope(r.ViewScope)
}

// ModelStatisticV2ChartReq 趋势 + 排行（合并）
type ModelStatisticV2ChartReq struct {
	ModelStatisticV2Req
	Limit int `json:"limit"` // 排行 TopN，默认 5
}

func (r *ModelStatisticV2ChartReq) Check() error {
	return r.ModelStatisticV2Req.Check()
}

// ModelStatisticV2ListReq 调用统计主表（分页）
type ModelStatisticV2ListReq struct {
	ModelStatisticV2Req
	StatisticSortReq
	PageNo   int `json:"pageNo"   validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

func (r *ModelStatisticV2ListReq) Check() error {
	return r.ModelStatisticV2Req.Check()
}

// ModelStatisticV2UserListReq 用户使用统计（分页，需 modelId）
type ModelStatisticV2UserListReq struct {
	ModelStatisticV2Req
	ModelId string `json:"modelId" validate:"required"`
	StatisticSortReq
	PageNo   int `json:"pageNo"   validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

func (r *ModelStatisticV2UserListReq) Check() error {
	return r.ModelStatisticV2Req.Check()
}

// ModelStatisticV2AppListReq 应用使用统计（分页，需 modelId）
type ModelStatisticV2AppListReq struct {
	ModelStatisticV2Req
	ModelId string `json:"modelId" validate:"required"`
	StatisticSortReq
	PageNo   int `json:"pageNo"   validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

func (r *ModelStatisticV2AppListReq) Check() error {
	return r.ModelStatisticV2Req.Check()
}

// ModelStatisticV2RecordReq 调用明细列表（分页，不支持用户排序；后端固定 created_at 降序）
type ModelStatisticV2RecordReq struct {
	ModelStatisticV2Req
	PageNo   int `json:"pageNo"   validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

func (r *ModelStatisticV2RecordReq) Check() error {
	return r.ModelStatisticV2Req.Check()
}

// --- 导出（无分页，可带排序）---

type ModelStatisticV2ExportListReq struct {
	ModelStatisticV2Req
	StatisticSortReq
}

func (r *ModelStatisticV2ExportListReq) Check() error {
	return r.ModelStatisticV2Req.Check()
}

type ModelStatisticV2UserExportReq struct {
	ModelStatisticV2Req
	ModelId string `json:"modelId" validate:"required"`
	StatisticSortReq
}

func (r *ModelStatisticV2UserExportReq) Check() error {
	return r.ModelStatisticV2Req.Check()
}

type ModelStatisticV2AppExportReq struct {
	ModelStatisticV2Req
	ModelId string `json:"modelId" validate:"required"`
	StatisticSortReq
}

func (r *ModelStatisticV2AppExportReq) Check() error {
	return r.ModelStatisticV2Req.Check()
}

type ModelStatisticV2RecordExportReq struct {
	ModelStatisticV2Req
}

func (r *ModelStatisticV2RecordExportReq) Check() error {
	return r.ModelStatisticV2Req.Check()
}

// ModelStatisticV2SelectReq 模型下拉
type ModelStatisticV2SelectReq struct {
	StatisticFilter
	ModelType string `json:"modelType" validate:"required"`
	ViewScope string `json:"viewScope" validate:"required"` // published|used
}

func (r *ModelStatisticV2SelectReq) Check() error {
	return checkStatisticViewScope(r.ViewScope)
}
