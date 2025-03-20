package request

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xiaoxiongmao5/we-api/utils"
	"github.com/xiaoxiongmao5/we-api/xlog"
	"github.com/xiaoxiongmao5/we-api/xnet/xresty"
)

type FetchOpts struct {
	Host     string                 `json:"host"`
	Url      string                 `json:"url"`
	Method   string                 `json:"method"`
	Data     map[string]interface{} `json:"data"`
	GetData  map[string]interface{} `json:"get_data"`
	PostData any                    `json:"post_data"`
	Headers  map[string]string      `json:"headers"`
	// NeedTmOut bool                   `json:"-"`
	// TmOut     time.Duration          `json:"-"`
}

func Fetch[T interface{}](reqOpts FetchOpts) (*T, error) {
	ctx := context.Background()
	logger := utils.Logf(ctx, "Fetch")

	uri := reqOpts.Host + reqOpts.Url

	logger.Info("request params",
		xlog.String("uri", uri),
		xlog.String("method", reqOpts.Method),
		xlog.Any("get_data", reqOpts.GetData),
		xlog.Any("post_data", reqOpts.PostData))

	reqIns := xresty.New().SetTimeout(10 * time.Second).R()

	reqIns.Header.Add("Content-Type", "application/json")

	if reqOpts.Headers != nil {
		logger.Info("request headers", xlog.Any("headers", reqOpts.Headers))
		for k, v := range reqOpts.Headers {
			reqIns.Header.Add(k, v)
		}
	}

	if len(reqOpts.GetData) > 0 {
		converted := mapInterfaceToMapString(reqOpts.GetData)
		reqIns.SetQueryParams(converted)
	}

	if reqOpts.Method == "POST" {
		jsonData, err := json.Marshal(reqOpts.PostData)
		if err != nil {
			logger.Error("json.Marshal(reqOpts.PostData) error", xlog.Err(err))
			return nil, err
		}

		reqIns.SetBody(jsonData)
	}

	// 记录请求开始时间
	startTime := time.Now().Unix()

	res, err := reqIns.SetContext(ctx).Execute(reqOpts.Method, uri)

	// 计算请求执行时间
	execTime := time.Now().Unix() - startTime

	resBody := res.Body()

	if err != nil {
		logger.Error("request error", xlog.Err(err),
			xlog.String("uri", uri),
			xlog.Int64("execTime", execTime),
			xlog.Any("body", resBody))
		return nil, err
	}

	if statusCode := res.StatusCode(); statusCode != http.StatusOK && statusCode != http.StatusCreated {
		logger.Error("request error: request status is not 200",
			xlog.String("status", res.Status()),
			xlog.String("uri", uri),
			xlog.Int64("execTime", execTime),
			xlog.String("body", string(resBody)))

		return nil, errors.New("request code:" + res.Status())
	}

	var ret *T
	if err = json.Unmarshal(resBody, &ret); err != nil {
		logger.Error("request body json.Unmarshal error",
			xlog.Err(err),
			xlog.String("uri", uri),
			xlog.Int64("execTime", execTime),
			xlog.String("body", string(resBody)))
		return nil, err
	}

	return ret, nil
}

func FetchStream[T interface{}](reqOpts FetchOpts, resChan chan *T, errChan chan error) {
	var err error
	defer func() {
		errChan <- err
		close(errChan)
		if _, ok := <-resChan; ok {
			close(resChan)
		}
	}()
	ctx := context.Background()
	logger := utils.Logf(ctx, "FetchStream")

	uri := reqOpts.Host + reqOpts.Url

	logger.Info("request params",
		xlog.String("uri", uri),
		xlog.String("method", reqOpts.Method),
		xlog.Any("get_data", reqOpts.GetData),
		xlog.Any("post_data", reqOpts.PostData))

	reqIns := xresty.New().R().
		SetDoNotParseResponse(true) // 禁止自动解析

	reqIns.Header.Add("Content-Type", "application/json")

	if reqOpts.Headers != nil {
		logger.Info("request headers", xlog.Any("headers", reqOpts.Headers))
		for k, v := range reqOpts.Headers {
			reqIns.Header.Add(k, v)
		}
	}

	if len(reqOpts.GetData) > 0 {
		converted := mapInterfaceToMapString(reqOpts.GetData)
		reqIns.SetQueryParams(converted)
	}

	if reqOpts.Method == "POST" {
		jsonData, err := json.Marshal(reqOpts.PostData)
		if err != nil {
			logger.Error("json.Marshal(reqOpts.PostData) error", xlog.Err(err))
			return
		}

		reqIns.SetBody(jsonData)
	}

	// 记录请求开始时间
	startTime := time.Now().Unix()

	res, err := reqIns.SetContext(ctx).Execute(reqOpts.Method, uri)
	defer res.RawBody().Close() //关闭响应体

	if err != nil {
		logger.Error("request error", xlog.Err(err),
			xlog.String("uri", uri))
		return
	}

	if statusCode := res.StatusCode(); statusCode != http.StatusOK && statusCode != http.StatusCreated {
		logger.Error("request error: request status is not 200",
			xlog.String("status", res.Status()),
			xlog.String("uri", uri))

		err = errors.New("request code:" + res.Status())
		return
	}

	reader := bufio.NewReader(res.RawBody()) // 使用 bufio.Reader 方便按行读取 (如果流是按行分隔的，例如 SSE)

	for {
		line, err := reader.ReadBytes('\n') // 按行读取，可根据实际流格式调整分隔符
		if err != nil {
			// 计算请求执行时间
			execTime := time.Now().Unix() - startTime
			if err == io.EOF { // 流结束
				close(resChan)
				logger.Info("Stream finished",
					xlog.String("uri", uri),
					xlog.Int64("execTime", execTime))
				break
			}
			// 读取错误
			logger.Error("reading stream error", xlog.Err(err),
				xlog.String("uri", uri),
				xlog.Int64("execTime", execTime))
			break
		}

		// 处理每一行数据 (流式处理的核心逻辑)
		logger.Info("stream data", xlog.String("line", string(line)))
		line = []byte(strings.TrimPrefix(string(line), "data: ")) // 去掉 SSE 格式的前缀
		if str := string(line); str == "\n" || str == "\r\n" || str == "" {
			continue
		}
		if strings.HasPrefix(string(line), "[DONE]") {
			close(resChan)
			logger.Info("Stream finished", xlog.String("uri", uri))
			return
		}

		var ret *T
		err = json.Unmarshal(line, &ret)

		if err != nil {
			logger.Error("stream data json.Unmarshal error", xlog.Err(err),
				xlog.String("uri", uri))
			return
		} else {
			resChan <- ret
		}

	}
}

func FetchStreamBase(reqOpts FetchOpts, resChan chan []byte, errChan chan error) {
	var err error
	defer func() {
		errChan <- err
		close(errChan)
		if _, ok := <-resChan; ok {
			close(resChan)
		}
	}()
	ctx := context.Background()
	logger := utils.Logf(ctx, "FetchStream")

	uri := reqOpts.Host + reqOpts.Url

	logger.Info("request params",
		xlog.String("uri", uri),
		xlog.String("method", reqOpts.Method),
		xlog.Any("get_data", reqOpts.GetData),
		xlog.Any("post_data", reqOpts.PostData))

	reqIns := xresty.New().R().
		SetDoNotParseResponse(true) // 禁止自动解析

	reqIns.Header.Add("Content-Type", "application/json")

	if reqOpts.Headers != nil {
		logger.Info("request headers", xlog.Any("headers", reqOpts.Headers))
		for k, v := range reqOpts.Headers {
			reqIns.Header.Add(k, v)
		}
	}

	if len(reqOpts.GetData) > 0 {
		converted := mapInterfaceToMapString(reqOpts.GetData)
		reqIns.SetQueryParams(converted)
	}

	if reqOpts.Method == "POST" {
		jsonData, err := json.Marshal(reqOpts.PostData)
		if err != nil {
			logger.Error("json.Marshal(reqOpts.PostData) error", xlog.Err(err))
			return
		}

		reqIns.SetBody(jsonData)
	}

	// 记录请求开始时间
	startTime := time.Now().Unix()

	res, err := reqIns.SetContext(ctx).Execute(reqOpts.Method, uri)
	defer res.RawBody().Close() //关闭响应体

	if err != nil {
		logger.Error("request error", xlog.Err(err),
			xlog.String("uri", uri))
		return
	}

	if statusCode := res.StatusCode(); statusCode != http.StatusOK && statusCode != http.StatusCreated {
		logger.Error("request error: request status is not 200",
			xlog.String("status", res.Status()),
			xlog.String("uri", uri))

		err = errors.New("request code:" + res.Status())
		return
	}

	reader := bufio.NewReader(res.RawBody()) // 使用 bufio.Reader 方便按行读取 (如果流是按行分隔的，例如 SSE)

	for {
		line, err := reader.ReadBytes('\n') // 按行读取，可根据实际流格式调整分隔符
		if err != nil {
			// 计算请求执行时间
			execTime := time.Now().Unix() - startTime
			if err == io.EOF { // 流结束
				close(resChan)
				logger.Info("Stream finished",
					xlog.String("uri", uri),
					xlog.Int64("execTime", execTime))
				break
			}
			// 读取错误
			logger.Error("reading stream error", xlog.Err(err),
				xlog.String("uri", uri),
				xlog.Int64("execTime", execTime))
			break
		}

		// 处理每一行数据 (流式处理的核心逻辑)
		logger.Info("stream data", xlog.String("line", string(line)))
		resChan <- line
	}
}

func mapInterfaceToMapString(m map[string]interface{}) map[string]string {
	converted := make(map[string]string)
	for k, v := range m {
		switch value := v.(type) {
		case string:
			converted[k] = value
		case int:
			converted[k] = strconv.Itoa(value)
		case float64:
			converted[k] = fmt.Sprintf("%f", value)
		default:
			converted[k] = fmt.Sprintf("%v", value)
		}
	}

	return converted
}
