package logic

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	chatHistoryMap = make(map[int64][]*schema.Message)
	chatHistoryMu  sync.RWMutex
)

// LoadHistoryByConversationID 根据对话ID加载内存中的历史消息
func LoadHistoryByConversationID(conversationID int64) []*schema.Message {
	// Step 1. 先根据对话ID读取内存中的历史消息
	chatHistoryMu.RLock()
	history := chatHistoryMap[conversationID]
	chatHistoryMu.RUnlock()

	// Step 2. 如果这是第一次对话 直接返回空切片
	if len(history) == 0 {
		return make([]*schema.Message, 0, 16)
	}

	// Step 3. 返回一份切片副本 避免外部直接修改 map 中的底层数组
	copied := make([]*schema.Message, len(history))
	copy(copied, history)
	return copied
}

// SaveHistoryByConversationID 根据对话ID保存最新的历史消息
func SaveHistoryByConversationID(conversationID int64, history []*schema.Message) {
	// Step 1. 对非法的 conversationID 直接忽略
	if conversationID == 0 {
		return
	}

	// Step 2. 保存一份切片副本 避免外部继续 append 时影响内存中的历史
	copied := make([]*schema.Message, len(history))
	copy(copied, history)

	chatHistoryMu.Lock()
	chatHistoryMap[conversationID] = copied
	chatHistoryMu.Unlock()
}

// InitSSE 初始化 SSE 响应头
func InitSSE(c *gin.Context) {
	// Step 1. 告诉前端当前响应是 SSE 流
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// Step 2. 提前把响应头刷给前端
	c.Status(http.StatusOK)
	c.Writer.Flush()
}

// WriteSSEDelta 将大模型返回的增量内容写给前端
func WriteSSEDelta(c *gin.Context, delta string) error {
	if strings.TrimSpace(delta) == "" {
		return nil
	}
	return writeSSEEvent(c, "delta", gin.H{"delta": delta})
}

// WriteSSEError 将错误事件写给前端
func WriteSSEError(c *gin.Context, err error) error {
	if err == nil {
		return nil
	}
	return writeSSEEvent(c, "error", gin.H{"error": err.Error()})
}

// WriteSSEDone 通知前端本轮流式响应已经结束
func WriteSSEDone(c *gin.Context, conversationID int64) error {
	return writeSSEEvent(c, "done", gin.H{"conversation_id": conversationID})
}

// StreamAndCollectAssistantFromEvents 一边将 SSE 写给前端 一边收集完整回答
func StreamAndCollectAssistantFromEvents(c *gin.Context, events *adk.AsyncIterator[*adk.AgentEvent]) (string, error) {
	var sb strings.Builder

	for {
		event, ok := events.Next()
		if !ok {
			break
		}

		// Step 1. Runner 层如果抛错 直接透传给前端
		if event.Err != nil {
			zap.L().Error("runner event error", zap.Error(event.Err))
			_ = WriteSSEError(c, event.Err)
			return "", event.Err
		}

		// Step 2. 过滤掉没有消息输出的事件
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}

		msgOutput := event.Output.MessageOutput

		// Step 3. 当前接口只把 assistant 的消息返回给前端
		if msgOutput.Role != schema.Assistant {
			continue
		}

		// Step 4. 如果是流式输出 就持续 Recv 直到 EOF
		if msgOutput.IsStreaming && msgOutput.MessageStream != nil {
			msgOutput.MessageStream.SetAutomaticClose()

			for {
				chunk, err := msgOutput.MessageStream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					zap.L().Error("msgOutput.MessageStream.Recv failed", zap.Error(err))
					_ = WriteSSEError(c, err)
					return sb.String(), err
				}

				if chunk != nil && chunk.Content != "" {
					sb.WriteString(chunk.Content)
					if err = WriteSSEDelta(c, chunk.Content); err != nil {
						zap.L().Error("WriteSSEDelta failed", zap.Error(err))
						return sb.String(), err
					}
				}
			}
			continue
		}

		// Step 5. 如果不是流式输出 就直接取完整消息
		if msgOutput.Message != nil && msgOutput.Message.Content != "" {
			sb.WriteString(msgOutput.Message.Content)
			if err := WriteSSEDelta(c, msgOutput.Message.Content); err != nil {
				zap.L().Error("WriteSSEDelta failed", zap.Error(err))
				return sb.String(), err
			}
		}
	}

	return sb.String(), nil
}

func StreamAndCollectAssistantFromModelStream(c *gin.Context, stream *schema.StreamReader[*schema.Message]) (string, error) {
	var sb strings.Builder
	streamReadStartAt := time.Now()
	firstChunkAt := time.Time{}
	chunkCount := 0

	if stream == nil {
		return "", nil
	}

	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			zap.L().Error("stream.Recv failed", zap.Error(err))
			_ = WriteSSEError(c, err)
			return sb.String(), err
		}

		if chunk != nil && chunk.Content != "" {
			chunkCount++
			if firstChunkAt.IsZero() {
				firstChunkAt = time.Now()
				zap.L().Info("first model chunk received",
					zap.Duration("firstChunkCost", firstChunkAt.Sub(streamReadStartAt)),
					zap.Int("chunkLength", len(chunk.Content)))
			}

			zap.L().Info("model chunk received",
				zap.Int("chunkIndex", chunkCount),
				zap.Int("chunkLength", len(chunk.Content)),
				zap.Duration("elapsedSinceStreamReadStart", time.Since(streamReadStartAt)))

			sb.WriteString(chunk.Content)
			if err = WriteSSEDelta(c, chunk.Content); err != nil {
				zap.L().Error("WriteSSEDelta failed", zap.Error(err))
				return sb.String(), err
			}
		}
	}

	if !firstChunkAt.IsZero() {
		zap.L().Info("model stream completed",
			zap.Int("chunkCount", chunkCount),
			zap.Int("contentLength", sb.Len()),
			zap.Duration("firstChunkCost", firstChunkAt.Sub(streamReadStartAt)),
			zap.Duration("totalStreamReadCost", time.Since(streamReadStartAt)))
	} else {
		zap.L().Warn("model stream completed without chunk",
			zap.Duration("totalStreamReadCost", time.Since(streamReadStartAt)))
	}

	return sb.String(), nil
}

func writeSSEEvent(c *gin.Context, eventName string, data any) error {
	// Step 1. 如果前端已经断开连接 就不继续写了
	select {
	case <-c.Request.Context().Done():
		err := c.Request.Context().Err()
		zap.L().Warn("client connection closed", zap.String("eventName", eventName), zap.Error(err))
		return err
	default:
	}

	// Step 2. 按 SSE 的格式把事件写给前端
	c.SSEvent(eventName, data)
	c.Writer.Flush()
	return nil
}
