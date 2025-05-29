package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type DashScopeClient struct {
	ApiKey  string
	BaseURL string
	Client  *http.Client
}

func NewDashScopeClient(apiKey, baseURL string) *DashScopeClient {
	return &DashScopeClient{
		ApiKey:  apiKey,
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type createReq struct {
	Model      string                 `json:"model"`
	Input      map[string]interface{} `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type createResp struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
	} `json:"output"`
	RequestID string `json:"request_id"`
}

type pollResp struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
		Results    []struct {
			URL string `json:"url"`
		} `json:"results"`
	} `json:"output"`
}

// SubmitTask 提交图像编辑任务到 DashScope
// function: 图像编辑功能名称
// prompt: 图像编辑的描述性提示
// baseImageURL: 基础图像的 URL
// extraParams: 额外的参数，如分辨率、样式等
// 返回任务 ID 或错误
// 注意：此函数假设 DashScope API 支持异步任务提交
func (d *DashScopeClient) SubmitTask(ctx context.Context, function, prompt, baseImageURL string, extraParams map[string]interface{}) (string, error) {
	reqBody := createReq{
		Model: "wanx2.1-imageedit",
		Input: map[string]interface{}{
			"function":       function,
			"prompt":         prompt,
			"base_image_url": baseImageURL,
		},
		Parameters: extraParams,
	}
	bts, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST",
		d.BaseURL+"/api/v1/services/aigc/image2image/image-synthesis",
		bytes.NewReader(bts))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.ApiKey)
	req.Header.Set("X-DashScope-Async", "enable")

	resp, err := d.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// 读取并日志输出原始响应
	data, _ := ioutil.ReadAll(resp.Body)
	log.Printf("DashScope SubmitTask raw response: %s", string(data))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dashscope submit error: %s", string(data))
	}
	var cr createResp
	if err := json.Unmarshal(data, &cr); err != nil {
		return "", err
	}
	return cr.Output.TaskID, nil
}

func (d *DashScopeClient) PollTask(ctx context.Context, taskID string) (string, []string, error) {
	url := fmt.Sprintf("%s/api/v1/tasks/%s", d.BaseURL, taskID)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+d.ApiKey)

	resp, err := d.Client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	// 读取并日志输出原始响应
	data, _ := io.ReadAll(resp.Body)
	log.Printf("DashScope PollTask raw response: %s", string(data))
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("dashscope poll error: %s", string(data))
	}
	var pr pollResp
	if err := json.Unmarshal(data, &pr); err != nil {
		return "", nil, err
	}
	var urls []string
	for _, result := range pr.Output.Results {
		urls = append(urls, result.URL)
	}
	return pr.Output.TaskID, urls, nil
}
