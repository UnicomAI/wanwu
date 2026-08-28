package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-resty/resty/v2"
)

type client struct {
	endpoint string
	client   *resty.Client
}

func newClient(endpoint string) *client {
	return &client{
		endpoint: endpoint,
		client:   resty.New().SetTimeout(5 * time.Minute),
	}
}

type execResult struct {
	SessionID string `json:"session_id"`
	Command   string `json:"command"`
	Status    string `json:"status"`
	Output    string `json:"output"`
	ExitCode  int    `json:"exit_code"`
}

type viewResult struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Output    string `json:"output"`
	ExitCode  int    `json:"exit_code"`
}

type apiResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func (c *client) exec(ctx context.Context, cmd, workDir string) (*execResult, error) {
	var resp apiResponse[execResult]
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"command":  cmd,
			"exec_dir": workDir,
		}).
		SetResult(&resp).
		Post(c.endpoint + "/v1/shell/exec")
	if err != nil {
		return nil, fmt.Errorf("exec request failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("exec failed: %s", resp.Message)
	}
	return &resp.Data, nil
}

func (c *client) execAsync(ctx context.Context, cmd, workDir string) (*execResult, error) {
	var resp apiResponse[execResult]
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"command":    cmd,
			"exec_dir":   workDir,
			"async_mode": true,
		}).
		SetResult(&resp).
		Post(c.endpoint + "/v1/shell/exec")
	if err != nil {
		return nil, fmt.Errorf("exec async request failed: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("exec async failed: %s", resp.Message)
	}
	return &resp.Data, nil
}

func (c *client) view(ctx context.Context, sessionID string) (*viewResult, error) {
	var resp apiResponse[viewResult]
	_, err := c.client.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"id": sessionID,
		}).
		SetResult(&resp).
		Post(c.endpoint + "/v1/shell/view")
	if err != nil {
		return nil, fmt.Errorf("view request failed: %w", err)
	}
	return &resp.Data, nil
}

// func (c *client) wait(ctx context.Context, sessionID string, timeoutSeconds int) error {
// 	var resp apiResponse[struct {
// 		Status string `json:"status"`
// 	}]
// 	_, err := c.client.R().
// 		SetContext(ctx).
// 		SetBody(map[string]interface{}{
// 			"id":      sessionID,
// 			"seconds": timeoutSeconds,
// 		}).
// 		SetResult(&resp).
// 		Post(c.endpoint + "/v1/shell/wait")
// 	if err != nil {
// 		return fmt.Errorf("wait request failed: %w", err)
// 	}
// 	if !resp.Success {
// 		return fmt.Errorf("wait failed: %s", resp.Message)
// 	}
// 	return nil
// }

func (c *client) upload(ctx context.Context, localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer func() { _ = file.Close() }()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(localPath))
	if err != nil {
		return fmt.Errorf("create form file failed: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy file content failed: %w", err)
	}

	_ = writer.WriteField("path", remotePath)
	_ = writer.Close()

	var resp apiResponse[struct {
		FilePath string `json:"file_path"`
		FileSize int64  `json:"file_size"`
		Success  bool   `json:"success"`
	}]
	httpResp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", writer.FormDataContentType()).
		SetBody(buf.Bytes()).
		SetResult(&resp).
		Post(c.endpoint + "/v1/file/upload")
	if err != nil {
		return fmt.Errorf("upload request failed: %w", err)
	}
	if httpResp.StatusCode() >= 300 {
		return fmt.Errorf("upload failed: status=%d, body=%s", httpResp.StatusCode(), httpResp.String())
	}
	if !resp.Success {
		return fmt.Errorf("upload failed: %s", resp.Message)
	}
	return nil
}

func (c *client) uploadData(ctx context.Context, data []byte, remotePath string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(remotePath))
	if err != nil {
		return fmt.Errorf("create form file failed: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("write data failed: %w", err)
	}

	_ = writer.WriteField("path", remotePath)
	_ = writer.Close()

	var resp apiResponse[struct {
		FilePath string `json:"file_path"`
		FileSize int64  `json:"file_size"`
		Success  bool   `json:"success"`
	}]
	httpResp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", writer.FormDataContentType()).
		SetBody(buf.Bytes()).
		SetResult(&resp).
		Post(c.endpoint + "/v1/file/upload")
	if err != nil {
		return fmt.Errorf("upload data request failed (size=%d, path=%s): %w", len(data), remotePath, err)
	}
	if httpResp.StatusCode() >= 300 {
		return fmt.Errorf("upload data failed (size=%d, path=%s): status=%d, body=%s", len(data), remotePath, httpResp.StatusCode(), httpResp.String())
	}
	if !resp.Success {
		return fmt.Errorf("upload data failed (size=%d, path=%s): %s", len(data), remotePath, resp.Message)
	}
	return nil
}

func (c *client) download(ctx context.Context, remotePath, localPath string) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"path": remotePath,
		}).
		SetDoNotParseResponse(true).
		Get(c.endpoint + "/v1/file/download")
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}

	rawResp := resp.RawResponse
	if rawResp == nil {
		return fmt.Errorf("no response")
	}
	defer func() { _ = rawResp.Body.Close() }()

	if rawResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(rawResp.Body)
		return fmt.Errorf("download failed: status=%d, body=%s", rawResp.StatusCode, string(body))
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("create local dir failed: %w", err)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create local file failed: %w", err)
	}
	defer func() { _ = file.Close() }()

	if _, err := io.Copy(file, rawResp.Body); err != nil {
		return fmt.Errorf("write local file failed: %w", err)
	}
	return nil
}

func (c *client) downloadData(ctx context.Context, remotePath string) ([]byte, error) {
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"path": remotePath,
		}).
		Get(c.endpoint + "/v1/file/download")
	if err != nil {
		return nil, fmt.Errorf("download data request failed: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("download data failed: status=%d, body=%s", resp.StatusCode(), string(resp.Body()))
	}
	return resp.Body(), nil
}

// func (c *client) writeFile(ctx context.Context, remotePath, content string) error {
// 	var resp apiResponse[struct {
// 		File         string `json:"file"`
// 		BytesWritten int    `json:"bytes_written"`
// 	}]
// 	_, err := c.client.R().
// 		SetContext(ctx).
// 		SetBody(map[string]interface{}{
// 			"file":    remotePath,
// 			"content": content,
// 		}).
// 		SetResult(&resp).
// 		Post(c.endpoint + "/v1/file/write")
// 	if err != nil {
// 		return fmt.Errorf("write file request failed: %w", err)
// 	}
// 	if !resp.Success {
// 		return fmt.Errorf("write file failed: %s", resp.Message)
// 	}
// 	return nil
// }

// func (c *client) readFile(ctx context.Context, remotePath string) (string, error) {
// 	var resp apiResponse[struct {
// 		Content string `json:"content"`
// 		File    string `json:"file"`
// 	}]
// 	_, err := c.client.R().
// 		SetContext(ctx).
// 		SetBody(map[string]interface{}{
// 			"file": remotePath,
// 		}).
// 		SetResult(&resp).
// 		Post(c.endpoint + "/v1/file/read")
// 	if err != nil {
// 		return "", fmt.Errorf("read file request failed: %w", err)
// 	}
// 	if !resp.Success {
// 		return "", fmt.Errorf("read file failed: %s", resp.Message)
// 	}
// 	return resp.Data.Content, nil
// }

func (c *client) delete(ctx context.Context, remotePath string) error {
	_, err := c.exec(ctx, fmt.Sprintf("rm -rf \"%s\"", remotePath), "/")
	return err
}

func (c *client) execWithOutput(ctx context.Context, cmd, workDir string, outputCh chan<- string) error {
	result, err := c.execAsync(ctx, cmd, workDir)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastOutput string
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			view, err := c.view(ctx, result.SessionID)
			if err != nil {
				return err
			}
			if view.Output != "" && view.Output != lastOutput {
				delta := view.Output[len(lastOutput):]
				if delta != "" {
					select {
					case outputCh <- delta:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				lastOutput = view.Output
			}
			if view.Status == "completed" || view.Status == "error" {
				return nil
			}
		}
	}
}
