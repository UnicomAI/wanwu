package http_client

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var _default *HttpClient = CreateDefault()

const (
	timeout        = 10 * time.Minute
	connectTimeout = 60 * time.Second
)

type LogLevel int

const (
	LogNone         LogLevel = 0
	LogBasic        LogLevel = 1
	LogParams       LogLevel = 2
	LogAll          LogLevel = 3
	formContentType          = "application/x-www-form-urlencoded"
	jsonContentType          = "application/json"
)

type HttpRequestParams struct {
	Headers    map[string]string
	Params     map[string]string
	Body       []byte
	FileParams []*HttpRequestFileParams
	Url        string
	Timeout    time.Duration
	MonitorKey string
	LogLevel   LogLevel
}

type HttpRequestFileParams struct {
	FileName string
	FileData io.Reader
}

type HttpClient struct {
	Client *http.Client
}

// tracingTransport 是一个自定义的 RoundTripper，在发送请求前注入 trace header
type tracingTransport struct {
	base http.RoundTripper
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 从 request context 中提取 trace header 并注入到请求中
	otel.GetTextMapPropagator().Inject(req.Context(), propagation.HeaderCarrier(req.Header))
	return t.base.RoundTrip(req)
}

func Create(client *http.Client) *HttpClient {
	return &HttpClient{client}
}

func CreateDefault() *HttpClient {
	return &HttpClient{newHttpClient()}
}

func Default() *HttpClient {
	return _default
}

// newHttpClient 初始化httpclient,httpclient 是一个比较重的资源，
// 为了http连接的复用在启动时做一次初始化，但是请注意如果需要做http请求的绝对隔离可以再创建其他的httpclient
func newHttpClient() *http.Client {
	return &http.Client{
		Transport: &tracingTransport{
			base: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				DialContext: (&net.Dialer{
					Timeout:   connectTimeout, // 连接超时时间
					KeepAlive: timeout,        // 连接保持活跃的时间
				}).DialContext,
				MaxIdleConnsPerHost:   100,
				ResponseHeaderTimeout: timeout,
			},
		},
		Timeout: timeout,
	}
}

func (c HttpClient) Get(ctx context.Context, httpRequestParams *HttpRequestParams) (result []byte, err error) {
	return SendRequest(ctx, c.Client, httpRequestParams, "GET", func(params *HttpRequestParams, ctx context.Context) (*http.Request, string, error) {
		var url = httpRequestParams.Url
		if len(httpRequestParams.Params) > 0 {
			var buffer bytes.Buffer
			buffer.WriteString(url)
			if !strings.Contains(url, "?") {
				buffer.WriteString("?")
			}
			url = buffer.String()
			if !strings.HasSuffix(url, "?") && !strings.HasSuffix(url, "&") {
				buffer.WriteString("&")
			}
			idx := 0
			for k, v := range httpRequestParams.Params {
				buffer.WriteString(k)
				buffer.WriteString("=")
				buffer.WriteString(v)
				if idx < len(httpRequestParams.Params)-1 {
					buffer.WriteString("&")
				}
				idx++
			}
			url = buffer.String()
		}

		request, err2 := http.NewRequest("GET", url, nil)
		return request, "", err2
	})
}

// GetOriResp 此方法需要在外部设置content 超时，并进行defer cancel
func (c HttpClient) GetOriResp(ctx context.Context, httpRequestParams *HttpRequestParams) (result *http.Response, err error) {
	return SendRequestOriResp(ctx, c.Client, httpRequestParams, "GET", func(params *HttpRequestParams, ctx context.Context) (*http.Request, string, error) {
		request, err2 := http.NewRequest("GET", httpRequestParams.Url, nil)
		return request, jsonContentType, err2
	})
}

func (c HttpClient) PostJson(ctx context.Context, httpRequestParams *HttpRequestParams) (result []byte, err error) {
	return SendRequest(ctx, c.Client, httpRequestParams, "POST-JSON", func(params *HttpRequestParams, ctx context.Context) (*http.Request, string, error) {
		var requestBody *bytes.Buffer
		if len(httpRequestParams.Body) > 0 {
			requestBody = bytes.NewBuffer(httpRequestParams.Body)
		}
		request, err2 := http.NewRequest("POST", httpRequestParams.Url, requestBody)
		return request, jsonContentType, err2
	})
}

// PostJsonOriResp 此方法需要在外部设置content 超时，并进行defer cancel
func (c HttpClient) PostJsonOriResp(ctx context.Context, httpRequestParams *HttpRequestParams) (result *http.Response, err error) {
	return SendRequestOriResp(ctx, c.Client, httpRequestParams, "POST-JSON", func(params *HttpRequestParams, ctx context.Context) (*http.Request, string, error) {
		var requestBody *bytes.Buffer
		if len(httpRequestParams.Body) > 0 {
			requestBody = bytes.NewBuffer(httpRequestParams.Body)
		}
		request, err2 := http.NewRequest("POST", httpRequestParams.Url, requestBody)
		return request, jsonContentType, err2
	})
}

func (c HttpClient) PostForm(ctx context.Context, httpRequestParams *HttpRequestParams) (result []byte, err error) {
	return SendRequest(ctx, c.Client, httpRequestParams, "POST-FORM", func(params *HttpRequestParams, ctx context.Context) (*http.Request, string, error) {
		data := url.Values{}
		if len(params.Params) > 0 {
			for k, v := range params.Params {
				data.Set(k, v)
			}
		}
		request, err2 := http.NewRequest("POST", httpRequestParams.Url, strings.NewReader(data.Encode()))
		return request, formContentType, err2
	})
}

func (c HttpClient) PostFile(ctx context.Context, httpRequestParams *HttpRequestParams) (result []byte, err error) {
	return SendRequest(ctx, c.Client, httpRequestParams, "POST-FILE", func(params *HttpRequestParams, ctx context.Context) (*http.Request, string, error) {
		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)

		if len(httpRequestParams.FileParams) == 0 {
			return nil, "", errors.New("no file params")
		}
		for _, fileParam := range httpRequestParams.FileParams {
			fileWriter, errW := writer.CreateFormFile("file", fileParam.FileName)
			if errW != nil {
				return nil, "", errW
			}
			_, errC := io.Copy(fileWriter, fileParam.FileData)
			if errC != nil {
				return nil, "", errC
			}
		}

		if len(httpRequestParams.Params) > 0 {
			for k, v := range httpRequestParams.Params {
				err2 := writer.WriteField(k, v)
				if err2 != nil {
					return nil, "", err2
				}
			}
		}

		err1 := writer.Close()
		if err1 != nil {
			return nil, "", err1
		}

		request, err2 := http.NewRequest("POST", httpRequestParams.Url, payload)
		return request, writer.FormDataContentType(), err2
	})
}

// Delete 删除数据
func (c HttpClient) Delete(ctx context.Context, httpRequestParams *HttpRequestParams) (result []byte, err error) {
	return SendRequest(ctx, c.Client, httpRequestParams, "DELETE", func(params *HttpRequestParams, ctx context.Context) (*http.Request, string, error) {
		var requestBody *bytes.Buffer
		if len(httpRequestParams.Body) > 0 {
			requestBody = bytes.NewBuffer(httpRequestParams.Body)
		}
		request, err2 := http.NewRequest("DELETE", httpRequestParams.Url, requestBody)
		return request, "", err2
	})
}

// SendRequest 此方法实现的目的是作为一个通用的http调用方法，也是最核心http调用
func SendRequest(ctx context.Context, client *http.Client, httpRequestParams *HttpRequestParams, requestType string, buildRequest func(*HttpRequestParams, context.Context) (*http.Request, string, error)) (result []byte, err error) {
	start := time.Now()
	if httpRequestParams == nil {
		return nil, errors.New("httpRequestParams is nil")
	}
	var hasLog = false
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("SendRequest panic %v", r)
			err = errors.New("sendHttpRequest panic err")
		}
		if !hasLog && err != nil {
			// 6.打印日志
			logRequest(ctx, httpRequestParams, requestType, start, -1, nil, err)
		}
	}()

	//1.开启超时监控
	if httpRequestParams.Timeout == 0 {
		httpRequestParams.Timeout = time.Minute * 1
	}
	ctx, cancel := context.WithTimeout(ctx, httpRequestParams.Timeout)
	defer cancel()

	//2.构造请求
	req, contentType, err := buildRequest(httpRequestParams, ctx)
	if err != nil {
		return nil, err
	}
	//3.设置请求头
	setHeader(req, httpRequestParams.Headers, contentType)
	req = req.WithContext(ctx)
	//4.执行请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	//5.处理返回结果
	defer func(Body io.ReadCloser) {
		err1 := Body.Close()
		if err1 != nil {
			//todo 通用日志文件
			err = err1
		}
	}(resp.Body) // 确保关闭响应体

	var body []byte
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer func() {
			err1 := reader.Close()
			if err1 != nil {
				err = err1
			}
		}()
		body, err = io.ReadAll(reader)
	} else {
		body, err = io.ReadAll(resp.Body)
	}

	hasLog = true
	// 6.打印日志
	logRequest(ctx, httpRequestParams, requestType, start, resp.StatusCode, body, err)
	return body, err
}

// SendRequestOriResp 此方法实现的目的是作为一个通用的http调用方法，也是最核心http调用
func SendRequestOriResp(ctx context.Context, client *http.Client, httpRequestParams *HttpRequestParams, requestType string, buildRequest func(*HttpRequestParams, context.Context) (*http.Request, string, error)) (result *http.Response, err error) {
	start := time.Now()
	if httpRequestParams == nil {
		return nil, errors.New("httpRequestParams is nil")
	}
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("SendRequest panic %v", r)
			err = errors.New("sendHttpRequest panic err")
		}
	}()

	//2.构造请求
	req, contentType, err := buildRequest(httpRequestParams, ctx)
	if err != nil {
		return nil, err
	}
	//3.设置请求头
	setHeader(req, httpRequestParams.Headers, contentType)
	req = req.WithContext(ctx)
	//4.执行请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// 6.打印日志
	logRequest(ctx, httpRequestParams, requestType, start, resp.StatusCode, nil, err)
	return resp, err
}

func ReadHttpResp(result *http.Response) (body []byte, err error) {
	defer func(Body io.ReadCloser) {
		err1 := Body.Close()
		if err1 != nil {
			//todo 通用日志文件
			err = err1
		}
	}(result.Body) // 确保关闭响应体

	if result.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(result.Body)
		if err != nil {
			return nil, err
		}
		defer func() {
			err1 := reader.Close()
			if err1 != nil {
				err = err1
			}
		}()
		body, err = io.ReadAll(reader)
	} else {
		body, err = io.ReadAll(result.Body)
	}
	return body, err
}

// setHeader 设置请求头
func setHeader(request *http.Request, headerMap map[string]string, contentType string) {
	hasContentType := false
	if len(headerMap) > 0 {
		for k, v := range headerMap {
			if k == "Content-Type" {
				hasContentType = true
			}
			request.Header.Set(k, v)
		}
	}
	if !hasContentType && len(contentType) > 0 {
		request.Header.Set("Content-Type", contentType)
	}
}

// logRequest 打印http请求日志，不会抛出panic
func logRequest(ctx context.Context, httpRequestParams *HttpRequestParams, requestType string, start time.Time, statusCode int, response []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("logRequest panic, r=%v\n", r)
		}
	}()
	if httpRequestParams.LogLevel == LogNone {
		return
	}
	//success := 0
	//if err == nil && statusCode == 200 {
	//	success = 1
	//}
	requestBody := ""
	if (httpRequestParams.LogLevel == LogParams || httpRequestParams.LogLevel == LogAll) && len(httpRequestParams.Body) > 0 {
		requestBody = string(httpRequestParams.Body)
	}
	responseBody := ""
	if (httpRequestParams.LogLevel == LogAll) && len(response) > 0 {
		responseBody = string(response)
	}
	var paramsMap = make(map[string]interface{})
	paramsMap["url"] = httpRequestParams.Url
	paramsMap["requestBody"] = requestBody
	LogRpcJson(ctx, "HTTP-"+requestType, httpRequestParams.MonitorKey, paramsMap, responseBody, err, start.UnixMilli())
}

// LogHttpRequest 打印http请求日志，不会抛出panic
func LogHttpRequest(ctx context.Context, business string, method string, requestURL string, params interface{}, result interface{}, err error, starTimestamp int64) {
	var paramsMap = make(map[string]interface{})
	paramsMap["url"] = requestURL
	paramsMap["requestBody"] = params
	LogRpcJson(ctx, business, "HTTP-"+method, paramsMap, result, err, starTimestamp)
}

func LogRpcJson(ctx context.Context, business string, method string, params interface{}, result interface{}, err error, starTimestamp int64) {
	defer func() {
		if err1 := recover(); err1 != nil {
			fmt.Println(err1)
		}
	}()
	var success = 1
	if err != nil {
		success = 0
	}
	var paramsStr = Convert2LogString(params)
	var resultStr = Convert2LogString(result)
	var errMsg = "-"
	if err != nil {
		errMsg = err.Error()
	}
	log.Log().Infof("%s|%s|%d|%d|%+v|%+v|%s", business, method, success, time.Now().UnixMilli()-starTimestamp, paramsStr, resultStr, errMsg)
}

func Convert2LogString(object interface{}) string {
	if object == nil {
		return "-"
	}
	switch obj := object.(type) {
	case string:
		return obj
	case []byte:
		return string(obj)
	default:
		bytesData, err := json.Marshal(object)
		if err != nil {
			return "-"
		}
		return string(bytesData)
	}
}
