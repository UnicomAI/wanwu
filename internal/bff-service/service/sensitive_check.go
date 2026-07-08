package service

import (
	"fmt"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	safety_service "github.com/UnicomAI/wanwu/api/proto/safety-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/pkg/ahocorasick"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	queue_util "github.com/UnicomAI/wanwu/pkg/queue-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	defaultCheckWindowSize = 20
	defaultRawCacheSize    = 3
)

type chatService interface {
	serviceType() string
	buildSensitiveResp(id, content string) []string
	parseContent(raw string) (id, content string)
}

type SensitiveChecker struct {
	PersonalTableIds []string
	ChatSrv          chatService
	Enable           bool
}

func CreateSensitiveChecker(personalTableIds []string, chatSrv chatService, enable bool) *SensitiveChecker {
	return &SensitiveChecker{
		PersonalTableIds: personalTableIds,
		ChatSrv:          chatSrv,
		Enable:           enable,
	}
}

func (c *SensitiveChecker) Check(ctx *gin.Context, query string, executor func() (ch <-chan string, callback func(string, string), err error)) (<-chan string, error) {
	if !c.Enable {
		ch, _, err := executor()
		if err != nil {
			return nil, err
		}
		return ch, nil
	}
	//1.查询敏感词表
	matchDicts, err := BuildSensitiveDict(ctx, c.PersonalTableIds, c.Enable)
	if err != nil {
		return nil, err
	}
	//2.同步敏感词检测
	err = SyncSensitiveCheck(query, matchDicts)
	if err != nil {
		return nil, err
	}
	//3.任务执行
	rawCh, callback, err := executor()
	if err != nil {
		return nil, err
	}
	//4.敏感词过滤(必须过滤，全局敏感词)
	outputCh := ProcessSensitiveWordsWithCallback(ctx, rawCh, matchDicts, c.ChatSrv, callback)
	return outputCh, nil
}

func SyncSensitiveCheck(query string, matchDicts []ahocorasick.DictConfig) error {
	matchResults, err := ahocorasick.ContentMatch(query, matchDicts, true)
	if err != nil {
		return err
	}
	if len(matchResults) > 0 {
		if matchResults[0].Reply != "" {
			return grpc_util.ErrorStatusWithKey(err_code.Code_BFFSensitiveWordCheck, "bff_sensitive_check_req", matchResults[0].Reply)
		}
		return grpc_util.ErrorStatusWithKey(err_code.Code_BFFSensitiveWordCheck, "bff_sensitive_check_req_default_reply")
	}
	return nil
}

// 构建敏感词字典
func BuildSensitiveDict(ctx *gin.Context, personalTableIds []string, enable bool) ([]ahocorasick.DictConfig, error) {
	var tableIDs []string
	if enable {
		tableIDs = personalTableIds
	}
	// safety服务获取全局敏感词
	globalTables, err := safety.GetGlobalSensitiveWordTableList(ctx.Request.Context(), &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	for _, table := range globalTables.List {
		tableIDs = append(tableIDs, table.TableId)
	}
	var dicts []ahocorasick.DictConfig
	resp, err := safety.GetSensitiveWordTableListByIDs(ctx.Request.Context(), &safety_service.GetSensitiveWordTableListByIDsReq{
		TableIds: tableIDs,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.List) == 0 {
		return nil, nil
	}
	for _, dict := range resp.List {
		dicts = append(dicts, ahocorasick.DictConfig{
			DictID:  dict.TableId,
			Version: dict.Version,
		})
	}
	// 检测内存中的敏感词表
	dictStatus, err := ahocorasick.CheckDictStatus(dicts)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFSensitiveWordCheck, err.Error())
	}
	// 拼接id,version与内存不匹配的tableID
	var needLoadTableIDs []string
	var ret []ahocorasick.DictConfig // 本次build最终在内存中的dicts
	for _, dict := range dictStatus {
		if !dict.Status {
			needLoadTableIDs = append(needLoadTableIDs, dict.DictCfg.DictID)
		} else {
			ret = append(ret, ahocorasick.DictConfig{
				DictID:  dict.DictCfg.DictID,
				Version: dict.DictCfg.Version,
			})
		}
	}
	// 访问safey 更新词表信息
	tableWithWords, err := safety.GetSensitiveWordTableListWithWordsByIDs(ctx.Request.Context(), &safety_service.GetSensitiveWordTableListByIDsReq{
		TableIds: needLoadTableIDs,
	})
	if err != nil {
		return nil, err
	}
	// 重新构建version不匹配的词表
	for _, table := range tableWithWords.Details {
		dict := ahocorasick.DictConfig{
			DictID:  table.Table.TableId,
			Version: table.Table.Version,
		}
		if err := ahocorasick.BuildDict(dict, table.Table.Reply, table.SensitiveWords); err != nil {
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("build dict id %v & dict version %v err: %v", dict.DictID, dict.Version, err))
		}
		ret = append(ret, ahocorasick.DictConfig{
			DictID:  table.Table.TableId,
			Version: table.Table.Version,
		})
	}
	return ret, nil
}

// ProcessSensitiveWords 中间处理函数，负责敏感词检测并返回处理后的通道。
// 当下游（前端）断开后 outputCh 无人消费，为避免背压阻塞上游 gRPC 消费和 SSE 会话发布，
// outputCh 的写入均采用非阻塞方式：缓冲区满时丢弃消息而非阻塞。
func ProcessSensitiveWords(ctx *gin.Context, rawCh <-chan string, matchDicts []ahocorasick.DictConfig, chatSrv chatService) <-chan string {
	return ProcessSensitiveWordsWithCallback(ctx, rawCh, matchDicts, chatSrv, nil)
}

// ProcessSensitiveWordsWithCallback 中间处理函数，负责敏感词检测并返回处理后的通道
func ProcessSensitiveWordsWithCallback(ctx *gin.Context, rawCh <-chan string, matchDicts []ahocorasick.DictConfig, chatSrv chatService, callback func(string, string)) <-chan string {
	// 无敏感词字典时直接返回原始通道，跳过检测
	if len(matchDicts) == 0 {
		return rawCh
	}

	outputCh := make(chan string, 128)
	go func() {
		defer util.PrintPanicStack()
		defer close(outputCh)
		var id, content string
		// contentQueue: 滑动窗口队列，累积最近M条内容用于检测跨消息拆分的敏感词
		contentQueue := queue_util.NewOverridableQueue(defaultCheckWindowSize)

		for raw := range rawCh {
			currId, currContent := chatSrv.parseContent(raw)
			id = currId
			contentQueue.EnQueue(currContent)

			content = contentQueue.AllValue()
			matchResults, err := ahocorasick.ContentMatch(content, matchDicts, true)
			if err != nil {
				log.Errorf("[%v] content (%v) check sensitive err: %v", chatSrv.serviceType(), content, err)
				select {
				case outputCh <- raw:
				default:
					//	log.Warnf("[%v] outputCh full, dropping message", chatSrv.serviceType())
				}
				continue
			}
			if len(matchResults) > 0 {
				log.Warnf("[%v] content (%v) check sensitive match results: %+v", chatSrv.serviceType(), content, matchResults)
				if matchResults[0].Reply != "" {
					for _, sensitiveMsg := range chatSrv.buildSensitiveResp(id, matchResults[0].Reply) {
						select {
						case outputCh <- sensitiveMsg:
							if callback != nil {
								callback(currId, sensitiveMsg)
							}
							return
						default:
							log.Warnf("[%v] outputCh full, dropping sensitive reply", chatSrv.serviceType())
						}
					}
				}
				for _, sensitiveMsg := range chatSrv.buildSensitiveResp(id, gin_util.I18nKey(ctx, "bff_sensitive_check_resp_default_reply")) {
					select {
					case outputCh <- sensitiveMsg:
						if callback != nil {
							callback(currId, sensitiveMsg)
						}
						return
					default:
						log.Warnf("[%v] outputCh full, dropping sensitive default reply", chatSrv.serviceType())
					}
				}
			}

			select {
			case outputCh <- raw:
			default:
				//log.Warnf("[%v] outputCh full, dropping message", chatSrv.serviceType())
			}
		}

	}()
	return outputCh
}
