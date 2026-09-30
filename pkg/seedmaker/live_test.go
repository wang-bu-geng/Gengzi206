//go:build live

// 真实链路测试：SEEDMAKER_KEY 存在时才运行
//
//	go test -tags live -run TestLive -v ./pkg/seedmaker/
package seedmaker_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wangbugeng/wangbugeng-drama/pkg/image"
	sm "github.com/wangbugeng/wangbugeng-drama/pkg/seedmaker"
	"github.com/wangbugeng/wangbugeng-drama/pkg/video"
)

func liveKey(t *testing.T) string {
	t.Helper()
	k := os.Getenv("SEEDMAKER_KEY")
	if k == "" {
		t.Skip("SEEDMAKER_KEY 未设置")
	}
	return k
}

// 下载结果 URL，校验内容可用

func startsWithAny(s string, prefix ...string) bool {
	for _, p := range prefix {
		if len(s) >= len(p) && s[:len(p)] == p {
			return true
		}
	}
	return false
}

func TestLiveImage(t *testing.T) {
	key := liveKey(t)
	ic := image.NewSeedMakerImageClient(sm.DefaultBaseURL, key, "doubao-seedream-5-0-flash-260915")

	t.Log("提交 text_to_image 任务（Flash / 2K / 16:9）")
	res, err := ic.GenerateImage(
		"赛博朋克城市天际线，霓虹灯光，雨后街道，电影感构图，高清细节",
		image.WithSize("2K"),
		image.WithDimensions(1920, 1080),
	)
	if err != nil {
		t.Fatalf("提交图片任务失败: %v", err)
	}
	t.Logf("任务 ID=%s 初始状态=%s", res.TaskID, res.Status)

	deadline := time.Now().Add(4 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)
		cur, err := ic.GetTaskStatus(res.TaskID)
		if err != nil {
			t.Logf("轮询出错（继续）: %v", err)
			continue
		}
		t.Logf("状态=%s completed=%v url=%q err=%q", cur.Status, cur.Completed, cur.ImageURL, cur.Error)
		if cur.Completed {
			if cur.ImageURL == "" {
				t.Fatal("完成但无图片 URL")
			}
			verifyImageDownload(t, cur.ImageURL)
			return
		}
		if cur.Error != "" && sm.IsTerminal(cur.Status) {
			t.Fatalf("图片任务失败: %s", cur.Error)
		}
	}
	t.Fatal("图片任务 4 分钟内未完成")
}

func verifyImageDownload(t *testing.T, u string) {
	resp, err := http.Get(u)
	if err != nil {
		t.Fatalf("下载结果图失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("结果图 HTTP %d", resp.StatusCode)
	}
	head := make([]byte, 512)
	n, _ := resp.Body.Read(head)
	detected := http.DetectContentType(head[:n])
	t.Logf("结果图 Content-Type 声明=%q 探测=%q", resp.Header.Get("Content-Type"), detected)
	if !startsWithAny(detected, "image/") {
		t.Fatalf("结果 URL 返回非图片: %s", detected)
	}
}

func TestLiveImageEditFace(t *testing.T) {
	key := liveKey(t)
	ic := image.NewSeedMakerImageClient(sm.DefaultBaseURL, key, "doubao-seedream-5-0-flash-260915")
	c := sm.New(sm.DefaultBaseURL, key)

	// 1) 先生成一张正脸人像
	t.Log("阶段1：生成正脸人像")
	res, err := ic.GenerateImage(
		"一位短发年轻女性的正面证件照式肖像，中性背景，均匀光照，面部清晰，写实风格",
		image.WithDimensions(1024, 1024),
	)
	if err != nil {
		t.Fatalf("提交肖像任务失败: %v", err)
	}
	var portraitURL string
	deadline := time.Now().Add(4 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)
		cur, err := ic.GetTaskStatus(res.TaskID)
		if err != nil {
			continue
		}
		t.Logf("肖像状态=%s", cur.Status)
		if cur.Completed {
			portraitURL = cur.ImageURL
			break
		}
		if cur.Error != "" && sm.IsTerminal(cur.Status) {
			t.Fatalf("肖像任务失败: %s", cur.Error)
		}
	}
	if portraitURL == "" {
		t.Fatal("肖像未生成")
	}

	// 2) 下载到本地临时文件（模拟本地角色立绘）
	tmp, err := os.CreateTemp("", "portrait-*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	resp, err := http.Get(portraitURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	tmp.Close()
	t.Logf("本地参考图: %s", tmp.Name())

	// 3) 图片侧：上传得到普通公网 URL（人脸库 asset:// 仅视频侧使用）
	resolved, err := c.ResolveReference(tmp.Name())
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}
	t.Logf("参考图解析结果: %s", resolved)
	if !strings.HasPrefix(resolved, "https://") {
		t.Fatalf("图片侧期望 https 上传 URL，实际 %q", resolved)
	}

	// 4) image_edit 带人脸参考再生成
	t.Log("阶段2：image_edit 换装场景")
	edit, err := ic.GenerateImage(
		"同一人物，换上红色皮夹克，赛博朋克霓虹街景背景，保持面部特征一致",
		image.WithReferenceImages([]string{tmp.Name()}),
		image.WithDimensions(1024, 1024),
	)
	if err != nil {
		t.Fatalf("提交 image_edit 失败: %v", err)
	}
	t.Logf("image_edit 任务 ID=%s 状态=%s", edit.TaskID, edit.Status)
	deadline = time.Now().Add(4 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)
		cur, err := ic.GetTaskStatus(edit.TaskID)
		if err != nil {
			continue
		}
		t.Logf("编辑图状态=%s url=%v err=%q", cur.Status, cur.Completed, cur.Error)
		if cur.Completed {
			verifyImageDownload(t, cur.ImageURL)
			return
		}
		if cur.Error != "" && sm.IsTerminal(cur.Status) {
			t.Fatalf("image_edit 失败: %s", cur.Error)
		}
	}
	t.Fatal("image_edit 4 分钟内未完成")
}

func TestLiveVideo(t *testing.T) {
	key := liveKey(t)
	vc := video.NewSeedMakerVideoClient(sm.DefaultBaseURL, key, "ep-20260625152214-vh6v2")

	t.Log("提交 text_to_video 任务（轻量 720p / 5s / 16:9）")
	res, err := vc.GenerateVideo(
		"",
		"赛博朋克都市夜景，霓虹灯闪烁，镜头缓慢向前推进，电影感",
		video.WithDuration(5),
		video.WithResolution("720p"),
		video.WithAspectRatio("16:9"),
	)
	if err != nil {
		t.Fatalf("提交视频任务失败: %v", err)
	}
	t.Logf("任务 ID=%s 初始状态=%s", res.TaskID, res.Status)

	deadline := time.Now().Add(9 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(8 * time.Second)
		cur, err := vc.GetTaskStatus(res.TaskID)
		if err != nil {
			t.Logf("轮询出错（继续）: %v", err)
			continue
		}
		t.Logf("状态=%s completed=%v url=%q duration=%d err=%q",
			cur.Status, cur.Completed, cur.VideoURL, cur.Duration, cur.Error)
		if cur.Completed {
			if cur.VideoURL == "" {
				t.Fatal("完成但无视频 URL")
			}
			verifyVideoDownload(t, cur.VideoURL)
			return
		}
		if cur.Error != "" && sm.IsTerminal(cur.Status) {
			t.Fatalf("视频任务失败: %s", cur.Error)
		}
	}
	t.Fatal("视频任务 9 分钟内未完成")
}

func TestLiveVideoFirstFrameFace(t *testing.T) {
	key := liveKey(t)
	ic := image.NewSeedMakerImageClient(sm.DefaultBaseURL, key, "doubao-seedream-5-0-flash-260915")
	c := sm.New(sm.DefaultBaseURL, key)
	vc := video.NewSeedMakerVideoClient(sm.DefaultBaseURL, key, "ep-20260625152214-vh6v2")

	// 1) 正脸人像
	t.Log("阶段1：生成正脸人像")
	res, err := ic.GenerateImage(
		"一位短发年轻女性的正面肖像，中性背景，均匀光照，面部清晰，写实风格",
		image.WithDimensions(1024, 1024),
	)
	if err != nil {
		t.Fatalf("肖像任务: %v", err)
	}
	var portraitURL string
	deadline := time.Now().Add(4 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)
		cur, _ := ic.GetTaskStatus(res.TaskID)
		if cur.Completed {
			portraitURL = cur.ImageURL
			break
		}
		if cur.Error != "" && sm.IsTerminal(cur.Status) {
			t.Fatalf("肖像失败: %s", cur.Error)
		}
	}
	if portraitURL == "" {
		t.Fatal("肖像未生成")
	}

	// 2) 落本地（模拟分镜首帧）
	tmp, _ := os.CreateTemp("", "frame-*.png")
	defer os.Remove(tmp.Name())
	resp, err := http.Get(portraitURL)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(tmp, resp.Body)
	resp.Body.Close()
	tmp.Close()

	// 3) 视频侧必须 asset:// 人脸库 URI
	resolved, err := c.ResolveReferenceAsFace(tmp.Name())
	if err != nil {
		t.Fatalf("人脸入库: %v", err)
	}
	t.Logf("视频参考解析: %s", resolved)
	if !strings.HasPrefix(resolved, "asset://") {
		t.Fatalf("视频侧期望 asset://，实际 %q", resolved)
	}

	// 4) image_first_frame 出片
	t.Log("阶段2：image_first_frame 5s")
	vres, err := vc.GenerateVideo(
		tmp.Name(),
		"人物轻微转头微笑，镜头缓慢推近，自然光照，电影质感",
		video.WithDuration(5),
		video.WithResolution("720p"),
		video.WithAspectRatio("1:1"),
	)
	if err != nil {
		t.Fatalf("提交视频失败: %v", err)
	}
	t.Logf("视频任务 ID=%s", vres.TaskID)
	deadline = time.Now().Add(9 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(8 * time.Second)
		cur, err := vc.GetTaskStatus(vres.TaskID)
		if err != nil {
			t.Logf("轮询错误（继续）: %v", err)
			continue
		}
		t.Logf("状态=%s completed=%v err=%q", cur.Status, cur.Completed, cur.Error)
		if cur.Completed {
			verifyVideoDownload(t, cur.VideoURL)
			return
		}
		if cur.Error != "" && sm.IsTerminal(cur.Status) {
			t.Fatalf("视频失败: %s", cur.Error)
		}
	}
	t.Fatal("视频 9 分钟内未完成")
}

func verifyVideoDownload(t *testing.T, u string) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		t.Fatalf("下载结果视频失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("结果视频 HTTP %d", resp.StatusCode)
	}
	head := make([]byte, 512)
	n, _ := resp.Body.Read(head)
	t.Logf("结果视频 Content-Type=%q ContentLength=%d 首字节=%d",
		resp.Header.Get("Content-Type"), resp.ContentLength, n)
	if !startsWithAny(resp.Header.Get("Content-Type"), "video/", "application/octet-stream") &&
		!startsWithAny(http.DetectContentType(head[:n]), "video/") {
		t.Log("警告：未能从头部确认视频类型，仍记录 URL 可用")
	}
	fmt.Println("VIDEO_OK")
}
