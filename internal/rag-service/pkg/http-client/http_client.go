package http_client

import (
	"bytes"
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
)

const (
	timeout        = 120 * time.Second
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
	fileParams []*HttpRequestFileParams
	Url        string
	Timeout    time.Duration
	MonitorKey string
	LogLevel   LogLevel
}

type HttpRequestFileParams struct {
	FileName string
	FileData io.Reader
}

var httpClient = HttpClient{}

type HttpClient struct {
	Client *http.Client
}

func init() {
	httpClient.Client = newHttpClient()
}

func (c HttpClient) LoadType() string {
	return "http-client"
}

func (c HttpClient) Stop() error {
	return nil
}

func GetClient() HttpClient {
	return httpClient
}

// newHttpClient initializes httpclient. httpclient is a relatively heavy resource.
// In order to reuse the http connection, initialization is done at startup, but please note that if you need absolute isolation of http requests, you can create other httpclients.
func newHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   connectTimeout, // Connection timeout
				KeepAlive: timeout,        // How long the connection remains active
			}).DialContext,
			MaxIdleConnsPerHost:   100,
			ResponseHeaderTimeout: timeout,
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
			if !strings.HasSuffix(url, "?") && !strings.HasSuffix(url, "&") {
				buffer.WriteString("&")
			}

			for k, v := range httpRequestParams.Params {
				buffer.WriteString(k)
				buffer.WriteString("=")
				buffer.WriteString(v)
				buffer.WriteString("&")
			}
			url = buffer.String()
		}

		request, err2 := http.NewRequest("GET", url, nil)
		return request, "", err2
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

// PostJsonOriResp This method needs to set the content timeout externally and perform defer cancel.
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

// PostFile If you want to pass other parameters, pass them through params
func (c HttpClient) PostFile(ctx context.Context, httpRequestParams *HttpRequestParams) (result []byte, err error) {
	return SendRequest(ctx, c.Client, httpRequestParams, "POST-FILE", func(params *HttpRequestParams, ctx context.Context) (*http.Request, string, error) {
		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)
		if len(httpRequestParams.Params) > 0 {
			for k, v := range httpRequestParams.Params {
				err2 := writer.WriteField(k, v)
				if err2 != nil {
					return nil, "", err2
				}
			}
		}
		if len(httpRequestParams.fileParams) == 0 {
			return nil, "", errors.New("no file params")
		}
		for _, fileParam := range httpRequestParams.fileParams {
			fileWriter, errW := writer.CreateFormFile("file", fileParam.FileName)
			if errW != nil {
				return nil, "", errW
			}
			_, errC := io.Copy(fileWriter, fileParam.FileData)
			if errC != nil {
				return nil, "", errC
			}
		}

		if err := writer.Close(); err != nil {
			return nil, "", err
		}

		request, err2 := http.NewRequest("POST", httpRequestParams.Url, payload)
		return request, writer.FormDataContentType(), err2
	})
}

// Delete Delete data
func (c HttpClient) Delete(ctx context.Context, httpRequestParams *HttpRequestParams) (result []byte, err error) {
	return SendRequest(ctx, c.Client, httpRequestParams, "DELETE", func(params *HttpRequestParams, ctx context.Context) (*http.Request, string, error) {
		request, err2 := http.NewRequest("DELETE", httpRequestParams.Url, nil)
		return request, "", err2
	})
}

// SendRequest This method is implemented as a general http calling method and is also the core http calling.
func SendRequest(ctx context.Context, client *http.Client, httpRequestParams *HttpRequestParams, requestType string, buildRequest func(*HttpRequestParams, context.Context) (*http.Request, string, error)) (result []byte, err error) {
	start := time.Now()
	if httpRequestParams == nil {
		return nil, errors.New("httpRequestParams is nil")
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered in f, r=%v\n", r)
			log.Errorf("SendRequest panic %v", r)
			err = errors.New("sendHttpRequest panic err")
		}
	}()

	//1. Turn on timeout monitoring
	if httpRequestParams.Timeout == 0 {
		httpRequestParams.Timeout = time.Minute * 1
	}
	ctx, cancel := context.WithTimeout(ctx, httpRequestParams.Timeout)
	defer cancel()

	//2. Construct a request
	req, contentType, err := buildRequest(httpRequestParams, ctx)
	if err != nil {
		return nil, err
	}
	//3. Set request headers
	setHeader(req, httpRequestParams.Headers, contentType)
	req = req.WithContext(ctx)
	//4. Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	//5. Process the returned results
	defer func(Body io.ReadCloser) {
		err1 := Body.Close()
		if err1 != nil {
			//todo general log file
			err = err1
		}
	}(resp.Body) // Make sure to close the response body

	// Read response body
	body, err := io.ReadAll(resp.Body)

	// 6.Print log
	logRequest(ctx, httpRequestParams, requestType, start, resp.StatusCode, body, err)
	return body, err
}

// SendRequestOriResp This method is implemented as a general http calling method and is also the core http calling.
func SendRequestOriResp(ctx context.Context, client *http.Client, httpRequestParams *HttpRequestParams, requestType string, buildRequest func(*HttpRequestParams, context.Context) (*http.Request, string, error)) (result *http.Response, err error) {
	start := time.Now()
	if httpRequestParams == nil {
		return nil, errors.New("httpRequestParams is nil")
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered in f, r=%v\n", r)
			log.Errorf("SendRequest panic %v", r)
			err = errors.New("sendHttpRequest panic err")
		}
	}()

	//2. Construct a request
	req, contentType, err := buildRequest(httpRequestParams, ctx)
	if err != nil {
		return nil, err
	}
	//3. Set request headers
	setHeader(req, httpRequestParams.Headers, contentType)
	req = req.WithContext(ctx)
	//4. Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// 6.Print log
	logRequest(ctx, httpRequestParams, requestType, start, resp.StatusCode, nil, err)
	return resp, err
}

// setHeader sets the request header
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

// logRequest prints http request log without throwing panic
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

func Get(ctx context.Context, params *HttpRequestParams) ([]byte, error) {
	return httpClient.Get(ctx, params)
}
