package image

import (
	"fmt"
	"strings"

	sm "github.com/wangbugeng/wangbugeng-drama/pkg/seedmaker"
)

// SeedMakerImageClient SeedMaker 异步图片客户端（Seedream / Nano Banana / GPT-image / Flux 等）
// 文档：POST /images 创建，GET /images/{id} 轮询，结果取 persistedImageUrls。
type SeedMakerImageClient struct {
	api   *sm.Client
	model string
}

func NewSeedMakerImageClient(baseURL, apiKey, model string) *SeedMakerImageClient {
	return &SeedMakerImageClient{
		api:   sm.New(baseURL, apiKey),
		model: model,
	}
}

// ── 请求/响应结构 ─────────────────────────────────────────

type smImageContentPart struct {
	Type        string         `json:"type"`
	Text        string         `json:"text,omitempty"`
	Size        string         `json:"size,omitempty"`
	Ratio       string         `json:"ratio,omitempty"`
	ForceSingle bool           `json:"forceSingle,omitempty"`
	ImageURL    *smImageURLRef `json:"image_url,omitempty"`
}

type smImageURLRef struct {
	URL string `json:"url"`
}

type smImageRequest struct {
	Mode        string               `json:"mode"`
	Model       string               `json:"model"`
	Prompt      string               `json:"prompt"`
	Content     []smImageContentPart `json:"content"`
	Ratio       string               `json:"ratio,omitempty"`
	Transparent bool                 `json:"transparent,omitempty"`
}

type smImageTask struct {
	ID                 string   `json:"id"`
	Status             string   `json:"status"`
	PersistedImageURLs []string `json:"persistedImageUrls"`
	ErrorMessage       string   `json:"errorMessage"`
}

// GenerateImage 提交图片任务。有参考图时走 image_edit（人物参考由共享层尝试人脸库）。
func (c *SeedMakerImageClient) GenerateImage(prompt string, opts ...ImageOption) (*ImageResult, error) {
	options := &ImageOptions{}
	for _, opt := range opts {
		opt(options)
	}
	model := c.model
	if options.Model != "" {
		model = options.Model
	}
	if model == "" {
		return nil, fmt.Errorf("seedmaker image: model is required")
	}

	ratio := smRatioFromOptions(options)
	fullPrompt := prompt
	if strings.TrimSpace(options.NegativePrompt) != "" {
		fullPrompt = strings.TrimRight(prompt, ".。 ") +
			". Avoid / 不希望出现: " + strings.TrimSpace(options.NegativePrompt)
	}

	content := []smImageContentPart{
		{Type: "text", Text: fullPrompt},
		{Type: "image_settings", Size: "2K", Ratio: ratio, ForceSingle: true},
	}

	mode := "text_to_image"
	for _, ref := range options.ReferenceImages {
		// 图片 image_edit 直接使用上传后的公网 URL（asset:// 人脸库仅视频侧要求）
		resolved, err := c.api.ResolveReference(ref)
		if err != nil {
			return nil, fmt.Errorf("seedmaker image: resolve reference: %w", err)
		}
		if resolved == "" {
			continue
		}
		mode = "image_edit"
		content = append(content, smImageContentPart{
			Type:     "image_url",
			ImageURL: &smImageURLRef{URL: resolved},
		})
	}

	reqBody := smImageRequest{
		Mode:    mode,
		Model:   model,
		Prompt:  fullPrompt,
		Content: content,
		Ratio:   ratio,
	}

	var task smImageTask
	if err := c.apiDoPost("/images", reqBody, &task); err != nil {
		return nil, err
	}
	if task.ID == "" {
		return nil, fmt.Errorf("seedmaker image: empty task id")
	}

	return c.mapResult(&task), nil
}

// GetTaskStatus 轮询任务状态
func (c *SeedMakerImageClient) GetTaskStatus(taskID string) (*ImageResult, error) {
	if taskID == "" {
		return nil, fmt.Errorf("seedmaker image: empty task id")
	}
	var task smImageTask
	if err := c.apiDoGet("/images/"+taskID, &task); err != nil {
		return nil, err
	}
	return c.mapResult(&task), nil
}

func (c *SeedMakerImageClient) mapResult(task *smImageTask) *ImageResult {
	r := &ImageResult{
		TaskID: task.ID,
		Status: task.Status,
	}
	switch {
	case task.Status == sm.StatusSucceeded:
		r.Completed = true
		if len(task.PersistedImageURLs) > 0 {
			r.ImageURL = task.PersistedImageURLs[0]
		}
	case task.Status == sm.StatusFailed:
		r.Error = task.ErrorMessage
		if r.Error == "" {
			r.Error = "seedmaker image task failed"
		}
	}
	return r
}

// apiDoPost / apiDoGet 透传到共享客户端的内部方法
func (c *SeedMakerImageClient) apiDoPost(path string, body any, out any) error {
	return c.api.Do("POST", path, body, out)
}

func (c *SeedMakerImageClient) apiDoGet(path string, out any) error {
	return c.api.Do("GET", path, nil, out)
}

// smRatioFromOptions 从显式尺寸推断 SeedMaker ratio 标识
func smRatioFromOptions(o *ImageOptions) string {
	w, h := o.Width, o.Height
	if (w == 0 || h == 0) && o.Size != "" {
		w, h = parseWH(o.Size)
	}
	if w <= 0 || h <= 0 {
		return "16:9" // 分镜默认横图
	}
	r := float64(w) / float64(h)
	switch {
	case r >= 1.6:
		return "16:9"
	case r > 1.15:
		return "4:3"
	case r > 0.85:
		return "1:1"
	case r > 0.62:
		return "3:4"
	default:
		return "9:16"
	}
}

func parseWH(size string) (int, int) {
	var w, h int
	for i, ch := range size {
		if ch == 'x' || ch == 'X' {
			fmt.Sscanf(size[:i], "%d", &w)
			fmt.Sscanf(size[i+1:], "%d", &h)
			return w, h
		}
	}
	return 0, 0
}
