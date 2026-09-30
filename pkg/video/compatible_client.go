package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CompatibleVideoClient struct {
	BaseURL       string
	APIKey        string
	Model         string
	Endpoint      string
	QueryEndpoint string
	Provider      string
	HTTPClient    *http.Client
}

type CompatibleVideoRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	ImageURL    string   `json:"image_url,omitempty"`
	Duration    int      `json:"duration,omitempty"`
	Size        string   `json:"size,omitempty"`
	AspectRatio string   `json:"aspect_ratio,omitempty"`
	Seconds     int      `json:"seconds,omitempty"`
	Seed        int64    `json:"seed,omitempty"`
}

type CompatibleVideoResponse struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Model       string `json:"model"`
	Status      string `json:"status"`
	Progress    int    `json:"progress"`
	CreatedAt   int64  `json:"created_at"`
	CompletedAt int64  `json:"completed_at"`
	VideoURL    string `json:"video_url"`
	Video       struct {
		URL string `json:"url"`
	} `json:"video"`
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func NewCompatibleVideoClient(baseURL, apiKey, model, endpoint, queryEndpoint, provider string) *CompatibleVideoClient {
	if endpoint == "" {
		endpoint = "/videos"
	}
	if queryEndpoint == "" {
		queryEndpoint = "/videos/{taskId}"
	}
	if provider == "" {
		provider = "compatible"
	}
	return &CompatibleVideoClient{
		BaseURL:       baseURL,
		APIKey:        apiKey,
		Model:         model,
		Endpoint:      endpoint,
		QueryEndpoint: queryEndpoint,
		Provider:      provider,
		HTTPClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

func (c *CompatibleVideoClient) logPrefix() string {
	return fmt.Sprintf("[%s Video]", c.Provider)
}

func (c *CompatibleVideoClient) GenerateVideo(imageURL, prompt string, opts ...VideoOption) (*VideoResult, error) {
	options := &VideoOptions{
		Duration:    5,
		AspectRatio: "16:9",
	}

	for _, opt := range opts {
		opt(options)
	}

	model := c.Model
	if options.Model != "" {
		model = options.Model
	}

	reqBody := CompatibleVideoRequest{
		Model:       model,
		Prompt:      prompt,
		ImageURL:    imageURL,
		Duration:    options.Duration,
		Seconds:     options.Duration,
		AspectRatio: options.AspectRatio,
		Seed:        options.Seed,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s marshal request: %w", c.logPrefix(), err)
	}

	endpoint := c.BaseURL + c.Endpoint
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("%s create request: %w", c.logPrefix(), err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s send request: %w", c.logPrefix(), err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s read response: %w", c.logPrefix(), err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("%s API error (status %d): %s", c.logPrefix(), resp.StatusCode, string(body))
	}

	var result CompatibleVideoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("%s parse response: %w", c.logPrefix(), err)
	}

	if result.Error.Message != "" {
		return nil, fmt.Errorf("%s error: %s", c.logPrefix(), result.Error.Message)
	}

	videoResult := &VideoResult{
		TaskID:    result.ID,
		Status:    result.Status,
		Completed: result.Status == "completed" || result.Status == "succeeded",
	}

	if result.VideoURL != "" {
		videoResult.VideoURL = result.VideoURL
	} else if result.Video.URL != "" {
		videoResult.VideoURL = result.Video.URL
	}

	return videoResult, nil
}

func (c *CompatibleVideoClient) GetTaskStatus(taskID string) (*VideoResult, error) {
	queryPath := c.QueryEndpoint
	if strings.Contains(queryPath, "{taskId}") {
		queryPath = strings.ReplaceAll(queryPath, "{taskId}", taskID)
	} else if strings.Contains(queryPath, "{task_id}") {
		queryPath = strings.ReplaceAll(queryPath, "{task_id}", taskID)
	} else {
		queryPath = queryPath + "/" + taskID
	}

	endpoint := c.BaseURL + queryPath
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%s create request: %w", c.logPrefix(), err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s send request: %w", c.logPrefix(), err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s read response: %w", c.logPrefix(), err)
	}

	var result CompatibleVideoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("%s parse response: %w", c.logPrefix(), err)
	}

	videoResult := &VideoResult{
		TaskID:    result.ID,
		Status:    result.Status,
		Completed: result.Status == "completed" || result.Status == "succeeded",
	}

	if result.Error.Message != "" {
		videoResult.Error = result.Error.Message
	}

	if result.VideoURL != "" {
		videoResult.VideoURL = result.VideoURL
	} else if result.Video.URL != "" {
		videoResult.VideoURL = result.Video.URL
	}

	return videoResult, nil
}
