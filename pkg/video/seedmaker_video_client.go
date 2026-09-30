package video

import (
	"fmt"
	"regexp"
	"strings"

	sm "github.com/wangbugeng/wangbugeng-drama/pkg/seedmaker"
)

// SeedMakerVideoClient SeedMaker 异步视频客户端（Seedance 系列）
// 文档：POST /videos 创建，GET /videos/{id} 轮询，结果取 persistedVideoUrl。
type SeedMakerVideoClient struct {
	api   *sm.Client
	model string
}

func NewSeedMakerVideoClient(baseURL, apiKey, model string) *SeedMakerVideoClient {
	return &SeedMakerVideoClient{
		api:   sm.New(baseURL, apiKey),
		model: model,
	}
}

type smVideoContentPart struct {
	Type     string         `json:"type"`
	Text     string         `json:"text,omitempty"`
	ImageURL *smVideoURLRef `json:"image_url,omitempty"`
	Role     string         `json:"role,omitempty"`
}

type smVideoURLRef struct {
	URL string `json:"url"`
}

type smVideoRequest struct {
	Mode       string               `json:"mode"`
	Model      string               `json:"model"`
	Content    []smVideoContentPart `json:"content"`
	Duration   int                  `json:"duration,omitempty"`
	Resolution string               `json:"resolution,omitempty"`
	Ratio      string               `json:"ratio,omitempty"`
}

type smVideoTask struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	PersistedVideoURL string `json:"persistedVideoUrl"`
	VideoDuration     int    `json:"videoDuration"`
	ErrorMessage      string `json:"errorMessage"`
}

var smRatioRe = regexp.MustCompile(`^\d+:\d+$`)

// GenerateVideo 根据参考图形态自动选择模式：
// 首尾帧 → image_first_last_frame；多参考 → multimodal_reference；
// 单图 → image_first_frame；无图 → text_to_video。
func (c *SeedMakerVideoClient) GenerateVideo(imageURL, prompt string, opts ...VideoOption) (*VideoResult, error) {
	options := &VideoOptions{
		Duration:    5,
		Resolution:  "720p",
		AspectRatio: "16:9",
	}
	for _, opt := range opts {
		opt(options)
	}
	model := c.model
	if options.Model != "" {
		model = options.Model
	}
	if model == "" {
		return nil, fmt.Errorf("seedmaker video: model is required")
	}

	content := []smVideoContentPart{{Type: "text", Text: prompt}}

	mode := "text_to_video"
	addImage := func(rawURL, role string) error {
		// 分镜帧常含人物：先尝试人脸库，非人物自动回退普通 URL
		resolved, err := c.api.ResolveReferenceAsFace(rawURL)
		if err != nil {
			return err
		}
		if resolved == "" {
			return nil
		}
		content = append(content, smVideoContentPart{
			Type:     "image_url",
			ImageURL: &smVideoURLRef{URL: resolved},
			Role:     role,
		})
		return nil
	}

	switch {
	case options.FirstFrameURL != "" && options.LastFrameURL != "":
		mode = "image_first_last_frame"
		if err := addImage(options.FirstFrameURL, "first_frame"); err != nil {
			return nil, fmt.Errorf("seedmaker video: first frame: %w", err)
		}
		if err := addImage(options.LastFrameURL, "last_frame"); err != nil {
			return nil, fmt.Errorf("seedmaker video: last frame: %w", err)
		}
	case len(options.ReferenceImageURLs) > 0:
		mode = "multimodal_reference"
		for _, ref := range options.ReferenceImageURLs {
			if err := addImage(ref, "reference_image"); err != nil {
				return nil, fmt.Errorf("seedmaker video: reference image: %w", err)
			}
		}
	case imageURL != "":
		mode = "image_first_frame"
		if err := addImage(imageURL, "first_frame"); err != nil {
			return nil, fmt.Errorf("seedmaker video: first frame: %w", err)
		}
	case options.FirstFrameURL != "":
		mode = "image_first_frame"
		if err := addImage(options.FirstFrameURL, "first_frame"); err != nil {
			return nil, fmt.Errorf("seedmaker video: first frame: %w", err)
		}
	}

	reqBody := smVideoRequest{
		Mode:    mode,
		Model:   model,
		Content: content,
	}
	if options.Duration > 0 {
		reqBody.Duration = options.Duration
	}
	if res := normalizeResolution(options.Resolution); res != "" {
		reqBody.Resolution = res
	}
	if smRatioRe.MatchString(options.AspectRatio) {
		reqBody.Ratio = options.AspectRatio
	}

	var task smVideoTask
	if err := c.api.Do("POST", "/videos", reqBody, &task); err != nil {
		return nil, err
	}
	if task.ID == "" {
		return nil, fmt.Errorf("seedmaker video: empty task id")
	}

	return c.mapResult(&task), nil
}

// GetTaskStatus 轮询任务状态
func (c *SeedMakerVideoClient) GetTaskStatus(taskID string) (*VideoResult, error) {
	if taskID == "" {
		return nil, fmt.Errorf("seedmaker video: empty task id")
	}
	var task smVideoTask
	if err := c.api.Do("GET", "/videos/"+taskID, nil, &task); err != nil {
		return nil, err
	}
	return c.mapResult(&task), nil
}

func (c *SeedMakerVideoClient) mapResult(task *smVideoTask) *VideoResult {
	r := &VideoResult{
		TaskID: task.ID,
		Status: task.Status,
	}
	switch {
	case task.Status == sm.StatusSucceeded:
		r.Completed = true
		r.VideoURL = task.PersistedVideoURL
		r.Duration = task.VideoDuration
	case task.Status == sm.StatusFailed:
		r.Error = task.ErrorMessage
		if r.Error == "" {
			r.Error = "seedmaker video task failed"
		}
	}
	return r
}

// normalizeResolution 接受 720p/1080p 或纯数字高度，其它值回退 720p
func normalizeResolution(res string) string {
	res = strings.TrimSpace(strings.ToLower(res))
	switch res {
	case "480p", "720p", "1080p":
		return res
	case "480":
		return "480p"
	case "1080":
		return "1080p"
	case "720", "":
		return "720p"
	default:
		if strings.HasSuffix(res, "p") {
			return res
		}
		return "720p"
	}
}
