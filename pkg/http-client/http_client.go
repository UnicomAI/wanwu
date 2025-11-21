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
)

var _default *HttpClient = CreateDefault()

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

func Create(client *http.Client) *HttpClient {
	return &HttpClient{client}
}

func CreateDefault() *HttpClient {
	return &HttpClient{newHttpClient()}
}

func Default() *HttpClient {
	return _default
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

// Delete Delete data
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

// SendRequest This method is implemented as a general http calling method and is also the core http calling.
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
			// 6.Print log
			logRequest(ctx, httpRequestParams, requestType, start, -1, nil, err)
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

func ReadHttpResp(result *http.Response) (body []byte, err error) {
	defer func(Body io.ReadCloser) {
		err1 := Body.Close()
		if err1 != nil {
			//todo general log file
			err = err1
		}
	}(result.Body) // Make sure to close the response body

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
