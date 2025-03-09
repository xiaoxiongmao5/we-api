package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoxiongmao5/we-api/service/ai"
	"github.com/xiaoxiongmao5/we-api/xlog"
)

func GetAiSvr(ctx context.Context, model string) ai.DoCompletions {
	switch model {
	case "gpt-3.5-turbo", "gpt-4o", "":
		return ai.NewOpenAiSvr(ctx)
	case "gemini-2.0-flash-exp":
		return ai.NewGeminiSvr(ctx)
	default:
		return ai.NewOpenAiSvr(ctx)
	}
}

func completions(c *gin.Context) {
	var req ai.OpenAiReq
	err := c.ShouldBind(&req)
	if err != nil {
		jsonRender(c, 400, gin.H{"message": err.Error()})
		return
	}

	svr := GetAiSvr(c.Request.Context(), req.Model)
	authorization := c.Request.Header.Get("Authorization")
	req.AuthHeader = authorization
	if req.Stream {
		resChan := make(chan *ai.OpenAiRes[ai.ChoiceStream])
		errChan := make(chan error)
		go func() {
			svr.DoStream(req, resChan, errChan)
		}()
		streamRender(c, resChan, errChan)
		return
	}

	// 非流式
	if res, err := svr.Do(req); err != nil {
		jsonRender(c, 500, gin.H{"message": err.Error()})
	} else {
		jsonRender(c, http.StatusOK, res)
	}
}

func jsonRender(c *gin.Context, stateCode int, data interface{}) {
	c.JSON(stateCode, data)
}
func streamRender(c *gin.Context, resChan chan *ai.OpenAiRes[ai.ChoiceStream], errChan chan error) {
	// 设置响应头部，关键是 Content-Type: text/event-stream
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive") // 可选，通常默认 keep-alive
	// 获取 ResponseWriter
	writer := c.Writer

	for {
		select {
		case res, ok := <-resChan:
			if !ok {
				fmt.Fprintf(writer, "[DONE]\n\n")
				writer.(http.Flusher).Flush()
				return
			}

			line, _ := json.Marshal(res)
			// SSE 格式: data: your_data\n\n
			fmt.Fprintf(writer, "data: %s\n\n", line)

			// 强制刷新缓冲区，确保数据立即发送到客户端
			writer.(http.Flusher).Flush()

			// 检查客户端是否断开连接，如果断开则退出循环
			if c.Request.Context().Err() != nil {
				fmt.Println("Client disconnected")
				return
			}
		case err := <-errChan:
			if err != nil {
				fmt.Fprintf(writer, "data: error: %s\n\n", err.Error())
				writer.(http.Flusher).Flush()
				return
			}
		}
	}
}

func streamHandler(c *gin.Context) {
	// 设置响应头部，关键是 Content-Type: text/event-stream
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive") // 可选，通常默认 keep-alive

	clientChan := make(chan string)
	defer close(clientChan)

	go func() {
		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("Data chunk %d at %s", i+1, time.Now().Format(time.RFC3339))
			clientChan <- message
			time.Sleep(1 * time.Second) // 模拟数据生成间隔
		}
	}()

	// 获取 ResponseWriter
	writer := c.Writer

	// 循环从 channel 读取数据并写入 ResponseWriter
	for message := range clientChan {
		// SSE 格式: data: your_data\n\n
		fmt.Fprintf(writer, "data: %s\n\n", message)

		// 强制刷新缓冲区，确保数据立即发送到客户端
		writer.(http.Flusher).Flush()

		// 检查客户端是否断开连接，如果断开则退出循环
		if c.Request.Context().Err() != nil {
			fmt.Println("Client disconnected")
			return
		}
	}

	fmt.Println("Stream finished")
}

func main() {
	err := xlog.StdConfig().Build()
	if err != nil {
		fmt.Printf("xlog.StdConfig().Build with error(%s)\n", err)
		os.Exit(-1)
	}

	r := gin.Default()

	r.POST("/v1/chat/completions", completions)

	r.Static("/static", "./static")

	r.GET("/stream", streamHandler)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Run("127.0.0.1:8080")
}
