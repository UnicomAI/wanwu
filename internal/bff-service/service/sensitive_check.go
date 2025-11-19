package service

import (
	"fmt"

	queue_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/queue-util"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	safety_service "github.com/UnicomAI/wanwu/api/proto/safety-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/pkg/ahocorasick"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
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

// 构建敏感词字典 [EN] Build a dictionary of sensitive words
func BuildSensitiveDict(ctx *gin.Context, tableIds []string) ([]ahocorasick.DictConfig, error) {
	var dicts []ahocorasick.DictConfig
	resp, err := safety.GetSensitiveWordTableListByIDs(ctx.Request.Context(), &safety_service.GetSensitiveWordTableListByIDsReq{
		TableIds: tableIds,
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
	// 检测内存中的敏感词表 [EN] Detect sensitive word lists in memory
	dictStatus, err := ahocorasick.CheckDictStatus(dicts)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFSensitiveWordCheck, err.Error())
	}
	// 拼接id,version与内存不匹配的tableID [EN] Splicing id, version and tableID that do not match memory
	var needLoadTableIDs []string
	var ret []ahocorasick.DictConfig // 本次build最终在内存中的dicts [EN] The final dicts in memory of this build
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
	// 访问safey 更新词表信息 [EN] Visit safey to update vocabulary information
	tableWithWords, err := safety.GetSensitiveWordTableListWithWordsByIDs(ctx.Request.Context(), &safety_service.GetSensitiveWordTableListByIDsReq{
		TableIds: needLoadTableIDs,
	})
	if err != nil {
		return nil, err
	}
	// 重新构建version不匹配的词表 [EN] Rebuild the word list with mismatched versions
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

// ProcessSensitiveWords 中间处理函数，负责敏感词检测并返回处理后的通道 [EN] ProcessSensitiveWords intermediate processing function, responsible for detecting sensitive words and returning the processed channel
func ProcessSensitiveWords(ctx *gin.Context, rawCh <-chan string, matchDicts []ahocorasick.DictConfig, chatSrv chatService) <-chan string {
	outputCh := make(chan string, 128)
	go func() {
		defer util.PrintPanicStack()
		defer close(outputCh)
		// 初始化队列 [EN] Initialize queue
		var id string
		var matchResults []ahocorasick.MatchResult
		var err error
		contentQueue := queue_util.NewOverridableQueue(defaultCheckWindowSize)
		rawQueue := queue_util.NewBoundedQueue(defaultRawCacheSize)
		for raw := range rawCh {
			currId, currContent := chatSrv.parseContent(raw)
			log.Debugf("[%v] raw (%v) parse id (%v) content (%v)", chatSrv.serviceType(), raw, currId, currContent)
			id = currId
			contentQueue.EnQueue(currContent)
			if rawQueue.IsFull() {
				// 校验敏感词 [EN] Check sensitive words
				content := contentQueue.AllValue()
				matchResults, err = ahocorasick.ContentMatch(content, matchDicts, true)
				log.Debugf("[%v] content (%v) check %+v sensitive results: %+v", chatSrv.serviceType(), content, matchDicts, matchResults)
				if err != nil {
					log.Errorf("[%v] content (%v) check sensitive err: %v", chatSrv.serviceType(), content, err)
				} else if len(matchResults) > 0 {
					break
				}
				// 输出队列内容 [EN] Output queue contents
				for !rawQueue.IsEmpty() {
					if dequeue, ok := rawQueue.Dequeue(); ok {
						outputCh <- dequeue
					}
				}
			}
			rawQueue.Enqueue(raw)
		}

		// 处理剩余内容 [EN] Process the rest
		if len(matchResults) == 0 {
			content := contentQueue.AllValue()
			matchResults, err = ahocorasick.ContentMatch(content, matchDicts, true)
			log.Debugf("[%v] rest content (%v) check %+v sensitive results: %+v", chatSrv.serviceType(), content, matchDicts, matchResults)
			if err != nil {
				log.Errorf("[%v] content (%v) check sensitive err: %v", chatSrv.serviceType(), content, err)
			}
		}

		// 检测到敏感词 [EN] Sensitive word detected
		if len(matchResults) > 0 {
			if matchResults[0].Reply != "" {
				for _, sensitiveMsg := range chatSrv.buildSensitiveResp(id, matchResults[0].Reply) {
					outputCh <- sensitiveMsg
					return
				}
			}
			for _, sensitiveMsg := range chatSrv.buildSensitiveResp(id, gin_util.I18nKey(ctx, "bff_sensitive_check_resp_default_reply")) {
				outputCh <- sensitiveMsg
				return
			}
		}

		// 返回剩余内容 [EN] Return remaining content
		valueList := rawQueue.AllValue()
		if len(valueList) > 0 {
			for _, value := range valueList {
				outputCh <- value
			}
		}
	}()
	return outputCh
}
