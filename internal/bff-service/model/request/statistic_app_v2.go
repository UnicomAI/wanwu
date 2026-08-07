package request

// AppStatisticV2Req 应用统计 V2 公共筛选（概览/趋势/排行/导出基底）。
// viewScope: published=我发布的(按 module_creator_*); used=我使用的(按调用人 user_id/org_id)。
type AppStatisticV2Req struct {
	StatisticFilter
	StartDate string   `json:"startDate" validate:"required"` // yyyy-MM-dd
	EndDate   string   `json:"endDate"   validate:"required"`
	Source    string   `json:"source"`                        // 调用来源 web|openapi|webURL；空=全部（钻取见子结构必填）
	Module    string   `json:"module"`                        // 板块 wga|skill|knowledge|...，空=全部
	Apps      []string `json:"apps"`                          // appId 列表，空=全部
	ViewScope string   `json:"viewScope" validate:"required"` // published|used
}

func (r *AppStatisticV2Req) Check() error {
	return checkStatisticViewScope(r.ViewScope)
}

// AppStatisticV2ChartReq 趋势 + 排行（合并）
type AppStatisticV2ChartReq struct {
	AppStatisticV2Req
	Limit int `json:"limit"` // 排行 TopN，默认 5
}

func (r *AppStatisticV2ChartReq) Check() error {
	return r.AppStatisticV2Req.Check()
}

// AppStatisticV2ListReq 调用统计主表（分页，可排序）
type AppStatisticV2ListReq struct {
	AppStatisticV2Req
	StatisticSortReq
	PageNo   int `json:"pageNo"   validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

func (r *AppStatisticV2ListReq) Check() error {
	return r.AppStatisticV2Req.Check()
}

// AppStatisticV2UserListReq 按用户钻取（分页；需 source + module + appId + 应用作者 定位主表行）
type AppStatisticV2UserListReq struct {
	AppStatisticV2Req
	Source              string `json:"source" validate:"required"` // 主表行来源，必填
	AppId               string `json:"appId"`                      // 空=板块级行（wga/model/skill/knowledge/prompt）
	ModuleCreatorUserId string `json:"moduleCreatorUserId"`        // 主表行应用作者，定位钻取行
	ModuleCreatorOrgId  string `json:"moduleCreatorOrgId"`         // 主表行应用作者组织，定位钻取行
	StatisticSortReq
	PageNo   int `json:"pageNo"   validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

func (r *AppStatisticV2UserListReq) Check() error {
	return r.AppStatisticV2Req.Check()
}

// AppStatisticV2ModelListReq 按模型钻取（分页；需 source + module + appId + 应用作者 定位主表行）
type AppStatisticV2ModelListReq struct {
	AppStatisticV2Req
	Source              string `json:"source" validate:"required"` // 主表行来源，必填
	AppId               string `json:"appId"`                      // 空=板块级行（wga/model/skill/knowledge/prompt）
	ModuleCreatorUserId string `json:"moduleCreatorUserId"`        // 主表行应用作者，定位钻取行
	ModuleCreatorOrgId  string `json:"moduleCreatorOrgId"`         // 主表行应用作者组织，定位钻取行
	StatisticSortReq
	PageNo   int `json:"pageNo"   validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

func (r *AppStatisticV2ModelListReq) Check() error {
	return r.AppStatisticV2Req.Check()
}

// AppStatisticV2RecordReq 调用明细列表（分页，不支持用户排序；后端固定 created_at 降序）
type AppStatisticV2RecordReq struct {
	AppStatisticV2Req
	AppId    string `json:"appId"` // 单应用筛选，空=不限（仍可用 apps[]）
	PageNo   int    `json:"pageNo"   validate:"required"`
	PageSize int    `json:"pageSize" validate:"required"`
}

func (r *AppStatisticV2RecordReq) Check() error {
	return r.AppStatisticV2Req.Check()
}

// --- 导出（无分页，可带排序）---

type AppStatisticV2ExportListReq struct {
	AppStatisticV2Req
	StatisticSortReq
}

func (r *AppStatisticV2ExportListReq) Check() error {
	return r.AppStatisticV2Req.Check()
}

type AppStatisticV2UserExportReq struct {
	AppStatisticV2Req
	Source              string `json:"source" validate:"required"` // 主表行来源，必填
	AppId               string `json:"appId"`                      // 空=板块级行（wga/model/skill/knowledge/prompt）
	ModuleCreatorUserId string `json:"moduleCreatorUserId"`        // 主表行应用作者，定位钻取行
	ModuleCreatorOrgId  string `json:"moduleCreatorOrgId"`         // 主表行应用作者组织，定位钻取行
	StatisticSortReq
}

func (r *AppStatisticV2UserExportReq) Check() error {
	return r.AppStatisticV2Req.Check()
}

type AppStatisticV2ModelExportReq struct {
	AppStatisticV2Req
	Source              string `json:"source" validate:"required"` // 主表行来源，必填
	AppId               string `json:"appId"`                      // 空=板块级行（wga/model/skill/knowledge/prompt）
	ModuleCreatorUserId string `json:"moduleCreatorUserId"`        // 主表行应用作者，定位钻取行
	ModuleCreatorOrgId  string `json:"moduleCreatorOrgId"`         // 主表行应用作者组织，定位钻取行
	StatisticSortReq
}

func (r *AppStatisticV2ModelExportReq) Check() error {
	return r.AppStatisticV2Req.Check()
}

type AppStatisticV2RecordExportReq struct {
	AppStatisticV2Req
	AppId string `json:"appId"`
}

func (r *AppStatisticV2RecordExportReq) Check() error {
	return r.AppStatisticV2Req.Check()
}

// AppStatisticV2SelectReq 应用下拉（前端传 agent|rag|workflow|knowledge；chatflow 等价 workflow，一并返回 workflow+chatflow）
type AppStatisticV2SelectReq struct {
	StatisticFilter
	Module    string `json:"module" validate:"required"`    // agent|rag|workflow|knowledge（chatflow→workflow）
	ViewScope string `json:"viewScope" validate:"required"` // published|used
}

func (r *AppStatisticV2SelectReq) Check() error {
	return checkStatisticViewScope(r.ViewScope)
}
