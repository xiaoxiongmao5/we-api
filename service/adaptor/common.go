package adaptor

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaoxiongmao5/we-api/meta"
	"github.com/xiaoxiongmao5/we-api/utils"
	"github.com/xiaoxiongmao5/we-api/xlog"
	"github.com/xiaoxiongmao5/we-api/xnet/xresty/xhttp"
)

func DoRequest(c *gin.Context, a Adaptor, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	ctx := c.Request.Context()
	logger := utils.Logf(ctx, "DoRequest")

	fullRequestURL, err := a.GetRequestURL(meta)
	if err != nil {
		return nil, fmt.Errorf("get request url failed: %w", err)
	}

	logger.Info("request params",
		xlog.String("fullRequestURL", fullRequestURL),
		xlog.String("method", c.Request.Method))

	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("new request failed: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	err = a.SetupRequestHeader(c, req, meta)
	if err != nil {
		return nil, fmt.Errorf("setup request failed: %w", err)
	}

	client := xhttp.NewClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("resp is nil")
	}

	req.Body.Close()
	c.Request.Body.Close()

	return resp, nil
}
