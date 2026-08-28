package request

type ToolBoxDetailReq struct {
	BoxID    string `form:"box_id"`
	BoxType  string `form:"box_type"`
	ToolID   string `form:"tool_id"` // 可选：按 action 的 operationId 精确过滤
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
}

func (r *ToolBoxDetailReq) Check() error {
	return nil
}
