package controllers

import (
	"context"
	"fmt"
	"gamepulse/logic"
	"gamepulse/logic/chat_agent/subagents"
	chattools "gamepulse/logic/chat_agent/tools"
	"gamepulse/models"
	"gamepulse/pkg/snowflake"
	"gamepulse/setting"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const conversationIDKey string = "conversation_id"

type chatGraphState struct {
	History      []*schema.Message
	ToolMessages []*schema.Message
}

// ChatStreamHandler 手写流式对话
func ChatStreamHandler(c *gin.Context) {
	requestStartAt := time.Now()

	var req models.ChatStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("ChatStreamHandler with invalid param", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// Step 1. 如果请求体里没有 conversationID
	//         就尝试从路由参数里取回本轮对话的会话ID
	if req.ConversationID == 0 {
		routeConversationID := strings.TrimSpace(c.Param("id"))
		if routeConversationID != "" && routeConversationID != "new" {
			parsedConversationID, err := strconv.ParseInt(routeConversationID, 10, 64)
			if err != nil {
				zap.L().Error("strconv.ParseInt(routeConversationID) failed",
					zap.String("routeConversationID", routeConversationID),
					zap.Error(err))
			} else {
				req.ConversationID = parsedConversationID
			}
		}
	}

	// Step 2. 如果前端没有返回 conversationID 说明是第一次对话
	//         需要使用雪花算法创建会话ID
	if req.ConversationID == 0 {
		req.ConversationID = snowflake.GenID()
	}

	// Step 3. 根据 Gin 的请求上下文创建 context
	//         并把会话 ID 传下去
	ctx := context.WithValue(c.Request.Context(), conversationIDKey, req.ConversationID)

	/* Eino知识补充
	1. Component: 可替换、可组合的能力单元 需要自己管理对话历史、编排调用流程、处理流式输出
		1) ChatModel: 调用大语言模型
		2) Tool: 执行特定任务
		3) Retriever: 检索信息
		4) Loader: 加载数据
	2. Agent:
		1) 完整的运行时框架 通过 Runner 统一管理执行过程
		2) 标准的事件流输出 Run() -> AsyncIterator[*AgentEvent] 支持流式、中断、恢复
		3) 可扩展能力 可以添加 tools、middleware、interrupt 等
		4) 开箱即用 创建 Agent 后直接运行 无需关心内部细节
	3. 我们选择使用 Agent
	*/

	// 下面开始完整的多轮对话逻辑

	// Step 4. 根据对话ID加载历史对话
	history := logic.LoadHistoryByConversationID(req.ConversationID)
	zap.L().Info("chat stream request started",
		zap.Int64("conversationID", req.ConversationID),
		zap.Int("historyLength", len(history)),
		zap.Int("queryLength", len(req.Query)))

	// Step 5. 将当前用户的 query 放入历史对话
	history = append(history, schema.UserMessage(req.Query))

	// Step 6. 初始化 SSE 响应头
	logic.InitSSE(c)

	// Step 7. 创建 Runner 管理 Agent
	chatModel, err := subagents.AnswerModel()
	if err != nil {
		zap.L().Error("subagents.AnswerModel failed",
			zap.Int64("conversationID", req.ConversationID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "初始化智能体失败"})
		return
	}

	messages := make([]*schema.Message, 0, len(history)+1)
	messages = append(messages, schema.SystemMessage(subagents.AnswerInstruction))
	messages = append(messages, history...)

	// Step 8. 运行智能体
	//         events 是 *AsyncIterator[*AgentEvent]，由 runner.Run() 返回
	streamOpenStartAt := time.Now()
	stream, err := chatModel.Stream(ctx, messages)
	if err != nil {
		zap.L().Error("chatModel.Stream failed",
			zap.Int64("conversationID", req.ConversationID),
			zap.Error(err))
		_ = logic.WriteSSEError(c, err)
		return
	}
	zap.L().Info("chatModel.Stream returned",
		zap.Int64("conversationID", req.ConversationID),
		zap.Duration("streamOpenCost", time.Since(streamOpenStartAt)),
		zap.Duration("requestCostBeforeStreamRead", time.Since(requestStartAt)))

	// Step 9. 根据返回的事件类型进行流式响应
	content, err := logic.StreamAndCollectAssistantFromModelStream(c, stream)
	if err != nil {
		zap.L().Error("logic.StreamAndCollectAssistantFromModelStream failed",
			zap.Int64("conversationID", req.ConversationID),
			zap.Error(err))
		return
	}

	// Step 10. 将 assistant 的完整回复追加到历史对话
	if strings.TrimSpace(content) != "" {
		history = append(history, schema.AssistantMessage(content, nil))
	}

	// Step 11. 将最新的 history 保存到内存
	logic.SaveHistoryByConversationID(req.ConversationID, history)

	// Step 12. 通知前端本轮流式响应结束
	if err = logic.WriteSSEDone(c, req.ConversationID); err != nil {
		zap.L().Error("logic.WriteSSEDone failed",
			zap.Int64("conversationID", req.ConversationID),
			zap.Error(err))
	}

	zap.L().Info("chat stream request completed",
		zap.Int64("conversationID", req.ConversationID),
		zap.Int("contentLength", len(content)),
		zap.Duration("totalCost", time.Since(requestStartAt)))
}

func GraphDemoHandler(c *gin.Context) {
	var req models.ChatStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("GraphHandler with invalid param", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 1. 处理 conversationID
	if req.ConversationID == 0 {
		routeConversationID := strings.TrimSpace(c.Param("id"))
		if routeConversationID != "" && routeConversationID != "new" {
			parsedConversationID, err := strconv.ParseInt(routeConversationID, 10, 64)
			if err != nil {
				zap.L().Error("strconv.ParseInt(routeConversationID) failed",
					zap.String("routeConversationID", routeConversationID),
					zap.Error(err))
			} else {
				req.ConversationID = parsedConversationID
			}
		}
	}

	if req.ConversationID == 0 {
		req.ConversationID = snowflake.GenID()
	}

	// 2. 加载历史
	history := logic.LoadHistoryByConversationID(req.ConversationID)
	// 注意：这里不要 append system message
	// history 只保存真实对话
	history = append(history, schema.UserMessage(req.Query))
	zap.L().Info("graph chat request started",
		zap.Int64("conversationID", req.ConversationID),
		zap.Int("historyLength", len(history)),
		zap.Int("queryLength", len(req.Query)))

	ctx := c.Request.Context()

	/*
		History      原始历史 + 本轮用户 query
		DecisionMsg  chat_model_1 输出的 assistant tool_call 消息
		ToolMessages tool_node 执行后的 tool result 消息

			START
			   ↓
			 prepare_decision_messages
			   ↓
			 chat_model_1
			   ↓
			 tool_node
			   ↓
			 prepare_answer_messages
			   ↓
			 chat_model_2
			   ↓
			 END
	*/
	type chatGraphState struct {
		History      []*schema.Message
		DecisionMsg  *schema.Message
		ToolMessages []*schema.Message
	}
	// 3. 创建 Graph
	graph := compose.NewGraph[[]*schema.Message, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *chatGraphState {
			return &chatGraphState{}
		}),
	)

	const (
		nodePrepareDecision = "prepare_decision_messages"
		nodeChatModel1      = "chat_model_1"
		nodeTool            = "tool_node"
		nodePrepareAnswer   = "prepare_answer_messages"
		nodePrepareNoTool   = "prepare_no_tool_answer_messages"
		nodeChatModel2      = "chat_model_2"
	)

	// 4. 准备 chat_model_1 的输入
	// 输入：history []*schema.Message
	// 输出：带 decision system prompt 的 []*schema.Message
	prepareDecision := compose.InvokableLambda(
		func(ctx context.Context, history []*schema.Message) ([]*schema.Message, error) {
			// 保存原始 history 到图状态，后面 answer 阶段要用
			err := compose.ProcessState[*chatGraphState](ctx,
				func(ctx context.Context, st *chatGraphState) error {
					st.History = append([]*schema.Message{}, history...)
					return nil
				})
			if err != nil {
				return nil, err
			}

			msgs := make([]*schema.Message, 0, len(history)+1)

			msgs = append(msgs, schema.SystemMessage(`
你是游戏社区舆情智能体的“工具决策节点”。

你的任务：
1. 判断用户问题是否需要查询真实社区数据。
2. 如果需要，必须调用合适的 tool。
3. 你只负责 tool call 决策和参数生成。
4. 不要生成最终自然语言回答。

当用户问：
- 某社区最近有什么舆论
- 某游戏社区风评怎么样
- 最近有哪些负面帖子
- 玩家在讨论什么

你应该优先调用社区/帖子舆情查询工具。
`))

			msgs = append(msgs, history...)
			return msgs, nil
		},
	)

	// 5. 保存 chat_model_1 的输出
	saveDecisionMsg := func(ctx context.Context, out *schema.Message, st *chatGraphState) (*schema.Message, error) {
		st.DecisionMsg = out
		return out, nil
	}

	// 6. 工具执行后，准备 chat_model_2 的输入
	// 输入：tool_node 输出的 []*schema.Message
	// 输出：给 answer model 的完整 messages
	prepareAnswer := compose.InvokableLambda(
		func(ctx context.Context, toolMessages []*schema.Message) ([]*schema.Message, error) {
			var history []*schema.Message
			var decisionMsg *schema.Message

			err := compose.ProcessState[*chatGraphState](ctx, func(ctx context.Context, st *chatGraphState) error {
				history = append([]*schema.Message{}, st.History...)
				decisionMsg = st.DecisionMsg
				st.ToolMessages = append([]*schema.Message{}, toolMessages...)
				return nil
			})
			if err != nil {
				return nil, err
			}

			msgs := make([]*schema.Message, 0, len(history)+len(toolMessages)+3)

			msgs = append(msgs, schema.SystemMessage(`
你是游戏社区舆情智能体“游小脉”。

你现在已经拿到了工具返回的数据。
请基于历史消息和工具结果，给用户生成最终自然语言回答。

要求：
1. 不要编造工具结果里没有的数据。
2. 优先总结舆情趋势、情绪倾向、风险点、代表帖子。
3. 如果工具结果为空，要明确说明数据不足。
4. 回答要像产品里的 AI 助手，不要暴露内部 tool call 细节。
`))

			// 原始历史：用户问了什么、多轮上下文是什么
			msgs = append(msgs, history...)

			// OpenAI tool calling 的上下文里，通常应该保留 assistant tool_call message
			if decisionMsg != nil {
				msgs = append(msgs, decisionMsg)
			}

			// 工具返回结果
			msgs = append(msgs, toolMessages...)

			return msgs, nil
		},
	)

	// 7. 准备未调用工具时的回答输入
	prepareNoToolAnswer := compose.InvokableLambda(
		func(ctx context.Context, decisionMsg *schema.Message) ([]*schema.Message, error) {
			var history []*schema.Message

			err := compose.ProcessState[*chatGraphState](ctx, func(ctx context.Context, st *chatGraphState) error {
				history = append([]*schema.Message{}, st.History...)
				if st.DecisionMsg == nil {
					st.DecisionMsg = decisionMsg
				}
				return nil
			})
			if err != nil {
				return nil, err
			}

			msgs := make([]*schema.Message, 0, len(history)+1)
			msgs = append(msgs, schema.SystemMessage(`
You are GamePulse's game community sentiment assistant.
Answer the user directly from the conversation history.
If the user asks for real community data but no tool result is available, clearly say the available data is insufficient and do not invent facts.`))
			msgs = append(msgs, history...)
			return msgs, nil
		},
	)

	// 8. 创建工具
	postTool := chattools.ChatPostInfoTool()

	toolInfo, err := postTool.Info(ctx)
	if err != nil {
		zap.L().Error("postTool.Info failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "工具信息初始化失败"})
		return
	}

	// 9. 只把工具绑定给 chat_model_1
	// chat_model_1 负责 tool call，所以它需要知道工具。
	// chat_model_2 只负责总结，不需要绑定工具。
	cfg := setting.Conf.LLMConfig
	temperature := float32(0.8)
	config := &openai.ChatModelConfig{
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		BaseURL:     cfg.BaseURL,
		Temperature: &temperature,
	}
	chatModel1, err := openai.NewChatModel(ctx, config)
	if err != nil {
		zap.L().Error("GraphChat1.NewChatModel failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model init failed"})
		return
	}
	chatModel2, err := openai.NewChatModel(ctx, config)
	if err != nil {
		zap.L().Error("GraphChat2.NewChatModel failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model init failed"})
		return
	}
	err = chatModel1.BindTools([]*schema.ToolInfo{toolInfo})
	if err != nil {
		zap.L().Error("GraphChat1.BindTools failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "工具绑定模型失败"})
		return
	}
	routeAfterDecision := compose.NewGraphBranch(
		func(ctx context.Context, msg *schema.Message) (string, error) {
			if msg != nil && len(msg.ToolCalls) > 0 {
				return nodeTool, nil
			}
			return nodePrepareNoTool, nil
		},
		map[string]bool{
			nodeTool:          true,
			nodePrepareNoTool: true,
		},
	)
	toolNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{postTool},
	})
	if err != nil {
		zap.L().Error("compose.NewToolNode failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "工具节点创建失败"})
		return
	}
	// 10. 添加节点
	_ = graph.AddLambdaNode(nodePrepareDecision, prepareDecision)

	_ = graph.AddChatModelNode(
		nodeChatModel1,
		chatModel1,
		compose.WithStatePostHandler(saveDecisionMsg),
	)

	_ = graph.AddToolsNode(nodeTool, toolNode)

	_ = graph.AddLambdaNode(nodePrepareAnswer, prepareAnswer)

	_ = graph.AddLambdaNode(nodePrepareNoTool, prepareNoToolAnswer)

	_ = graph.AddChatModelNode(nodeChatModel2, chatModel2)

	// 11. 添加边
	_ = graph.AddEdge(compose.START, nodePrepareDecision)
	_ = graph.AddEdge(nodePrepareDecision, nodeChatModel1)
	_ = graph.AddBranch(nodeChatModel1, routeAfterDecision)
	_ = graph.AddEdge(nodeTool, nodePrepareAnswer)
	_ = graph.AddEdge(nodePrepareAnswer, nodeChatModel2)
	_ = graph.AddEdge(nodePrepareNoTool, nodeChatModel2)
	_ = graph.AddEdge(nodeChatModel2, compose.END)

	// 12. 编译
	runnable, err := graph.Compile(ctx)
	if err != nil {
		zap.L().Error("graph compile failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "图编译失败"})
		return
	}

	// 13. 先用 Invoke 跑通，不要一上来 Stream
	out, err := runnable.Invoke(ctx, history)
	if err != nil {
		zap.L().Error("graph invoke failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "图运行失败"})
		return
	}

	// 14. 保存本轮用户消息和模型回答
	// 这里你后面可以改成你的 logic.SaveMessage(...)
	// logic.SaveChatMessage(req.ConversationID, schema.UserMessage(req.Query))
	// logic.SaveChatMessage(req.ConversationID, out)

	c.JSON(http.StatusOK, gin.H{
		"conversation_id": req.ConversationID,
		"answer":          out.Content,
	})
}

// GraphHandler 使用 Eino 图进行流式对话
//func GraphHandler(c *gin.Context) {
//	var req models.ChatStreamRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		zap.L().Error("ChatStreamHandler with invalid param", zap.Error(err))
//		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
//		return
//	}
//
//	// Step 1. 如果请求体里没有 conversationID
//	//         就尝试从路由参数里取回本轮对话的会话ID
//	if req.ConversationID == 0 {
//		routeConversationID := strings.TrimSpace(c.Param("id"))
//		if routeConversationID != "" && routeConversationID != "new" {
//			parsedConversationID, err := strconv.ParseInt(routeConversationID, 10, 64)
//			if err != nil {
//				zap.L().Error("strconv.ParseInt(routeConversationID) failed",
//					zap.String("routeConversationID", routeConversationID),
//					zap.Error(err))
//			} else {
//				req.ConversationID = parsedConversationID
//			}
//		}
//	}
//
//	// Step 2. 如果前端没有返回 conversationID 说明是第一次对话
//	//         需要使用雪花算法创建会话ID
//	if req.ConversationID == 0 {
//		req.ConversationID = snowflake.GenID()
//	}
//
//	// 创建一个图实例
//	// 输入是
//	// 输出是
//	// TODO 调试/阅读文档 这里是怎么实现了一个图状态机的
//	graph := compose.NewGraph[[]*schema.Message, *schema.Message](
//		compose.WithGenLocalState(func(ctx context.Context) *chatGraphState {
//			return &chatGraphState{}
//		}))
//
//	/* 整体流程
//	START
//	  ↓
//	load_history_lambda
//	  ↓
//	decision_template
//	  ↓
//	chat_model_with_tools (chat 1)
//	  ↓
//	branch: 是否有 tool_call
//	      ├── 没有 -> chat_model_answer -> save_answer_to_state -> END
//	      └── 有 -> tools_node
//	                ↓
//	              save_tool_result_to_state
//	                ↓
//	              answer_template
//	                ↓
//	              chat_model_answer
//	                ↓
//	              save_answer_to_state
//	                ↓
//	              END
//	*/
//
//	decisionPrompt := prompt.FromMessages(schema.FString,
//		schema.SystemMessage("你是游戏社区舆情智能体之一，只能在需要时发起 tool call，不要回答最终问题。"),
//		schema.MessagesPlaceholder("history", true),
//	)
//
//	answerPrompt := prompt.FromMessages(schema.FString,
//		schema.SystemMessage("你是游戏社区舆情智能体之一，负责根据历史消息和工具返回结果，给用户生成最终自然语言回答。"),
//		schema.MessagesPlaceholder("history", true),
//	)
//
//	// 将 load_history_lambda + decision_template 实现为一个节点
//	chat1Pre := func(ctx context.Context, history []*schema.Message, st *chatGraphState) ([]*schema.Message, error) {
//		st.History = history
//
//		msgs := make([]*schema.Message, 0, len(history)+1)
//		msgs = append(msgs, schema.SystemMessage("你是游戏社区舆情智能体之一，只能在需要时发起 tool call，不要回答最终问题。"))
//		msgs = append(msgs, history...)
//		return msgs, nil
//	}
//
//	// 将 save_tool_result_to_state 实现
//	toolPre := func(ctx context.Context, msg []*schema.Message, st *chatGraphState) ([]*schema.Message, error) {
//		st.ToolMessages = append(st.ToolMessages, msg...)
//		return msg, nil
//	}
//
//	// 为节点编写名称
//	const (
//		nodeKeyOfLoadHistory = "load_history_lambda"
//		nodeKeyOfDeciTemp    = "decision_template"
//		nodeKeyOfChatModel1  = "chat_node_1"
//		nodeKeyOfBranch      = "branch"
//		nodeKyeOfTool        = "tool_node"
//		nodeKeyOfSaveState   = "save_tool_result_to_state"
//		nodeKeyOfChatModel2  = "chat_node_2"
//		nodeKeyOfSaveAnswer  = "save_answer_to_state"
//	)
//
//	// 工具节点
//	ctx := context.Background()
//	postToolInfo := chattools.ChatPostInfoTool()
//	toolNode1, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
//		Tools: []tool.BaseTool{postToolInfo},
//	})
//	// TODO 将工具提供给模型
//
//	// 增加聊天模型节点 1 负责工具决定调用
//	// 传入组件：model.BaseChatModel
//	// 节点运行时输入输出类型
//	//   输入：[]*schema.Message
//	//   输出：*schema.Message
//	graph.AddChatModelNode(nodeKeyOfChatModel1, logic.GraphChat1)
//	// 增加工具调用节点 1 负责查询某个游戏社区 正/负向 帖子
//	// 传入组件：*compose.ToolsNode
//	// 节点运行时输入输出类型
//	//   输入：*schema.Message
//	//   输出：*schema.Message
//	graph.AddToolsNode(nodeKyeOfTool1, toolNode1)
//	// 增加聊天模型节点 2 负责根据工具调用结果(若调用) 与用户聊天
//	// 传入组件：model.BaseChatModel
//	// 节点运行时输入输出类型
//	//   输入：[]*schema.Message
//	//   输出：*schema.Message
//	graph.AddChatModelNode(nodeKeyOfChatModel2, logic.GraphChat2)
//
//	// 在节点减增加边
//	graph.AddEdge(compose.START, nodeKeyOfChatModel1)
//	graph.AddEdge(nodeKeyOfChatModel1, nodeKyeOfTool1)
//	graph.AddEdge(nodeKyeOfTool1, nodeKeyOfChatModel2)
//	graph.AddEdge(nodeKeyOfChatModel2, compose.END)
//
//	// 编译 Graph[I, O] to Runnable[I, O]
//	runnable, err := graph.Compile(ctx)
//	if err != nil {
//		zap.L().Error("图实例编译为运行时错误", zap.Error(err))
//		return
//	}
//	out, err := runnable.Invoke(ctx, history)
//	if err != nil {
//		zap.L().Error("运行时调用错误", zap.Error(err))
//		return
//	}
//
//	fmt.Printf("/n")
//	fmt.Printf(out.Content)
//}

// ChainHandler 使用 Eino 链进行流式对话
func ChainHandler(c *gin.Context) {
	var req models.ChatStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("ChatStreamHandler with invalid param", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// Step 1. 如果请求体里没有 conversationID
	//         就尝试从路由参数里取回本轮对话的会话ID
	if req.ConversationID == 0 {
		routeConversationID := strings.TrimSpace(c.Param("id"))
		if routeConversationID != "" && routeConversationID != "new" {
			parsedConversationID, err := strconv.ParseInt(routeConversationID, 10, 64)
			if err != nil {
				zap.L().Error("strconv.ParseInt(routeConversationID) failed",
					zap.String("routeConversationID", routeConversationID),
					zap.Error(err))
			} else {
				req.ConversationID = parsedConversationID
			}
		}
	}

	// Step 2. 如果前端没有返回 conversationID 说明是第一次对话
	//         需要使用雪花算法创建会话ID
	if req.ConversationID == 0 {
		req.ConversationID = snowflake.GenID()
	}

	// Step 4. 根据对话ID加载历史对话
	history := logic.LoadHistoryByConversationID(req.ConversationID)
	history = append(history, schema.SystemMessage("你是一个舆情助手，负责回答社区的游戏舆情问题，你的名字是游小脉"))
	zap.L().Info("chat stream request started",
		zap.Int64("conversationID", req.ConversationID),
		zap.Int("historyLength", len(history)),
		zap.Int("queryLength", len(req.Query)))

	// Step 5. 将当前用户的 query 放入历史对话
	history = append(history, schema.UserMessage(req.Query))

	//// 创建历史消息
	//history := make([]*schema.Message, 0, 20)
	//history = append(history,
	//	schema.SystemMessage("你是一个舆情助手，负责回答社区的游戏舆情问题，你的名字是游小脉"),
	//	schema.UserMessage("你好，你是谁？"),
	//)

	// 创建模型
	ctx := context.Background()
	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: "http://localhost:11434/v1",
		Model:   "gemma4:e2b",
	})
	if err != nil {
		fmt.Printf("模型初始化报错")
		return
	}

	/* 创建工具
	需求分析：当用户发送一条信息过来时，需要大模型判断是否需要调用工具
	例如
	step1. 用户发送”鸣潮最近有什么舆论“ ，大模型分析语义，大模型决定调用 [数据库查询工具]
	step2. 如何调用 -> 大模型

	*/

	// 1. 创建链
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendChatModel(model)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		fmt.Printf("链编译错误")
		return
	}

	stream, err := runnable.Stream(ctx, history)
	if err != nil {
		fmt.Printf("运行时错误")
		return
	}

	// Step 9. 根据返回的事件类型进行流式响应
	content, err := logic.StreamAndCollectAssistantFromModelStream(c, stream)
	if err != nil {
		zap.L().Error("logic.StreamAndCollectAssistantFromModelStream failed",
			zap.Int64("conversationID", req.ConversationID),
			zap.Error(err))
		return
	}

	// Step 10. 将 assistant 的完整回复追加到历史对话
	if strings.TrimSpace(content) != "" {
		history = append(history, schema.AssistantMessage(content, nil))
	}

	// Step 11. 将最新的 history 保存到内存
	logic.SaveHistoryByConversationID(req.ConversationID, history)

	// Step 12. 通知前端本轮流式响应结束
	if err = logic.WriteSSEDone(c, req.ConversationID); err != nil {
		zap.L().Error("logic.WriteSSEDone failed",
			zap.Int64("conversationID", req.ConversationID),
			zap.Error(err))
	}

	zap.L().Info("chat stream request completed",
		zap.Int64("conversationID", req.ConversationID),
		zap.Int("contentLength", len(content)))
}
