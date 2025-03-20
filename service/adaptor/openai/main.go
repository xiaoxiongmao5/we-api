package openai

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaoxiongmao5/we-api/common"
	"github.com/xiaoxiongmao5/we-api/common/render"
	"github.com/xiaoxiongmao5/we-api/relay/model"
)

const (
	dataPrefix       = "data: "
	dataPrefixLength = len(dataPrefix)
	done             = "[DONE]"
)

func StreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, string, *model.Usage) {
	common.SetEventStreamHeaders(c)

	doneRendered := false
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		data := scanner.Text()
		// ignore blank line or wrong format
		if len(data) < dataPrefixLength {
			continue
		}

		if data[:dataPrefixLength] != dataPrefix && !strings.HasPrefix(data, done) {
			continue
		}

		data = strings.TrimPrefix(data, dataPrefix)

		if strings.HasPrefix(data, done) {
			doneRendered = true
			render.StringData(c, data)
			continue
		}

		var streamResponse ChatCompletionsStreamResponse
		err := json.Unmarshal([]byte(data), &streamResponse)
		if err != nil {
			// [TODO]添加日志
			render.StringData(c, data)
			continue
		}

		render.StringData(c, data)
	}

	// 返回扫描过程中发生的任何错误，如果是io.EOF时, err 返回 nil
	if err := scanner.Err(); err != nil {
		// [TODO]记日志
	}

	if !doneRendered {
		render.Done(c)
	}

	err := resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), "", nil
	}

	return nil, "", nil
}

func Handler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}

	c.Writer.WriteHeader(resp.StatusCode)

	_, err := io.Copy(c.Writer, resp.Body)
	if err != nil {
		return ErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}

	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	return nil, nil
}
