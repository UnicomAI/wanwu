package request

import "fmt"

// 消息中心请求参数的上限与取值边界（超限在 Check() 拒绝，不落库）。
const (
	// NoticeKeywordMaxLen 关键字模糊匹配最大长度（按字符数计）
	NoticeKeywordMaxLen = 50
	// NoticePageSizeMax 单页最大条数
	NoticePageSizeMax = 100
	// NoticeDeleteIDsMax 单次批量删除的最大消息数
	NoticeDeleteIDsMax = 100
	// NoticeCategoryMinVal 消息类别下界（0 全部）
	NoticeCategoryMinVal = 0
	// NoticeCategoryMaxVal 消息类别上界（3 工单）
	NoticeCategoryMaxVal = 3
)

// NoticeListReq 消息中心整页列表请求。
// 悬浮面板的未读列表也复用本接口（onlyUnread=true + 较小 pageSize）。
type NoticeListReq struct {
	// Category 0 全部 / 1 公告 / 2 产品服务 / 3 工单
	Category int `json:"category" form:"category"`
	// OnlyUnread true 仅未读；false 已读+未读混排（默认）
	OnlyUnread bool `json:"onlyUnread" form:"onlyUnread"`
	// Keyword 标题/正文模糊匹配；% 与 _ 按普通字符处理（服务端转义）
	Keyword string `json:"keyword" form:"keyword"`
	PageSearch
}

func (r *NoticeListReq) Check() error {
	if r.Category < NoticeCategoryMinVal || r.Category > NoticeCategoryMaxVal {
		return fmt.Errorf("invalid category %v, must be in [%v, %v]",
			r.Category, NoticeCategoryMinVal, NoticeCategoryMaxVal)
	}
	if len([]rune(r.Keyword)) > NoticeKeywordMaxLen {
		return fmt.Errorf("keyword too long, max %v characters", NoticeKeywordMaxLen)
	}
	if r.PageSize > NoticePageSizeMax {
		return fmt.Errorf("pageSize too large, max %v", NoticePageSizeMax)
	}
	return nil
}

// NoticeReadReq 单条消息标记已读
type NoticeReadReq struct {
	MessageId string `json:"messageId" form:"messageId" validate:"required"`
	CommonCheck
}

// NoticeDeleteReq 批量删除消息（只从自己当前组织上下文的列表移除）
type NoticeDeleteReq struct {
	Ids []string `json:"ids" form:"ids" validate:"required"`
}

func (r *NoticeDeleteReq) Check() error {
	if len(r.Ids) == 0 {
		return fmt.Errorf("ids empty")
	}
	if len(r.Ids) > NoticeDeleteIDsMax {
		return fmt.Errorf("too many ids, max %v", NoticeDeleteIDsMax)
	}
	return nil
}
