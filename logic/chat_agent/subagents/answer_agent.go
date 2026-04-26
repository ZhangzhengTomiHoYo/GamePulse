package subagents

import (
	"context"
	"errors"
	"gamepulse/setting"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"go.uber.org/zap"
)

const AnswerInstruction = `
你是游戏舆情监控社区的对话智能体。你需要根据上下文和检索后的内容，直接回答用户的问题。回答时使用自然语言，不要强制输出 JSON。
`

func AnswerAgent() adk.Agent {
	cm, err := openAIForQwen()
	if err != nil {
		zap.L().Fatal("openAIForQwen failed", zap.Error(err))
	}

	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "AnswerAgent",
		Description: "回应用户的单条消息",
		Instruction: `
你是游戏舆情监控社区的对话智能体。
你需要根据上下文和检索后的内容，直接回答用户的问题。
回答时使用自然语言，不要强制输出 JSON。
`,
		Model: cm,
	})
	if err != nil {
		zap.L().Fatal("adk.NewChatModelAgent failed", zap.Error(err))
	}

	return a
}

func AnswerModel() (model.ToolCallingChatModel, error) {
	return openAIForQwen()
}

func openAIForQwen() (cm model.ToolCallingChatModel, err error) {
	if setting.Conf == nil || setting.Conf.LLMConfig == nil {
		err = errors.New("llm config not initialized")
		zap.L().Error("openAIForQwen failed", zap.Error(err))
		return nil, err
	}

	cfg := setting.Conf.LLMConfig
	if strings.TrimSpace(cfg.APIKey) == "" {
		err = errors.New("llm api key is empty")
		zap.L().Error("openAIForQwen failed", zap.Error(err))
		return nil, err
	}

	temperature := float32(0.2)
	config := &openai.ChatModelConfig{
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		BaseURL:     cfg.BaseURL,
		Temperature: &temperature,
	}

	cm, err = openai.NewChatModel(context.Background(), config)
	if err != nil {
		zap.L().Error("openai.NewChatModel failed", zap.Error(err))
		return nil, err
	}

	return cm, nil
}
