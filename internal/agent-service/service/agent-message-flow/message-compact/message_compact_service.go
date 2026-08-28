package message_compact

import (
	"slices"

	tokenizer_service "github.com/UnicomAI/wanwu/internal/agent-service/service/tokenizer-service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/cloudwego/eino/schema"
)

func Compact(multiAgent bool, messages []*schema.Message, userOtherMessages []*schema.Message, tokenLimit int) (retMessages []*schema.Message) {
	if multiAgent {
		//添加 用户消息
		retMessages = append(retMessages, messages...)
		//添加 用户其他消息
		retMessages = append(retMessages, userOtherMessages...)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			var dataMessages []*schema.Message
			dataMessages = append(dataMessages, messages...)
			dataMessages = append(dataMessages, userOtherMessages...)
			retMessages = dataMessages
			return
		}
	}()
	// 消息数量小于等于2，标识系统提示词+用户消息，则不进行压缩
	if len(messages) <= 2 {
		return messages
	}
	//提取用户问题消息
	var userMessages = messages[len(messages)-1:]
	//提取其他消息
	dataMessages := messages[:len(messages)-1]
	tokenizer := tokenizer_service.NewTokenizer(tokenizer_service.DefaultTokenizer)
	var costTokens = userMessagesTokenCost(tokenizer, userMessages)

	var compactMessages []*schema.Message
	//系统提示词，判断消耗token大小决定添加消息
	for _, message := range dataMessages {
		tokens, err := tokenizer.CountTokens(buildTokenMessage(message))
		if err != nil {
			log.Errorf("CountTokens err: %v", err)
		}
		if message.Role == schema.System {
			retMessages = append(retMessages, message)
			costTokens = costTokens + tokens
		} else {
			compactMessages = append(compactMessages, message)
		}
	}
	//压缩历史，获取最新消息
	historyMessages := latestHistoryMessages(tokenizer, compactMessages, costTokens, tokenLimit)
	//添加 历史消息
	if len(historyMessages) > 0 {
		retMessages = append(retMessages, historyMessages...)
	}
	//添加 用户消息
	retMessages = append(retMessages, userMessages...)
	//添加 用户其他消息
	retMessages = append(retMessages, userOtherMessages...)
	return retMessages
}

func userMessagesTokenCost(tokenizer tokenizer_service.Tokenizer, userMessages []*schema.Message) int {
	var costTokens = 0
	for _, message := range userMessages {
		tokens, err := tokenizer.CountTokens(buildTokenMessage(message))
		if err != nil {
			log.Errorf("CountTokens err: %v", err)
			continue
		}
		costTokens = costTokens + tokens
	}
	return costTokens
}

// 获取最新消息
func latestHistoryMessages(tokenizer tokenizer_service.Tokenizer, compactMessages []*schema.Message, costTokens, tokenLimit int) []*schema.Message {
	if tokenLimit <= 0 {
		return compactMessages
	}
	if costTokens >= tokenLimit {
		return make([]*schema.Message, 0)
	}
	hisLen := len(compactMessages)
	var reverseMessages []*schema.Message
	for i := hisLen - 1; i >= 0; i-- {
		message := compactMessages[i]
		tokens, err := tokenizer.CountTokens(buildTokenMessage(message))
		if err != nil {
			log.Errorf("CountTokens err: %v", err)
			continue
		}
		costTokens = costTokens + tokens
		if costTokens > tokenLimit {
			//如果是用户问题，则把上一步的模型答案也成对移除了
			reverseMessages = removePairMessage(message, reverseMessages)
			break
		}
		reverseMessages = append(reverseMessages, message)
	}
	//反转
	slices.Reverse(reverseMessages)
	return reverseMessages
}

func buildTokenMessage(message *schema.Message) string {
	return string(message.Role) + ": " + message.Content
}

// removePairMessage 移除模型回答的答案, 保证问答成对移除
func removePairMessage(message *schema.Message, reverseMessages []*schema.Message) []*schema.Message {
	if message.Role == schema.User && len(reverseMessages) > 0 {
		lastMessage := reverseMessages[len(reverseMessages)-1]
		if lastMessage.Role == schema.Assistant {
			reverseMessages = reverseMessages[:len(reverseMessages)-1]
		}
	}
	return reverseMessages
}
