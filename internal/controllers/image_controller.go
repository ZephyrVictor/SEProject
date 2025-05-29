package controllers

import (
	"awesomeProject/config"
	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
	"context"
	"fmt"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImageController struct {
	Queue *services.TaskQueue
	DB    *gorm.DB
	Cfg   *config.Config
}

func NewImageController(q *services.TaskQueue, db *gorm.DB, cfg *config.Config) *ImageController {
	return &ImageController{Queue: q, DB: db, Cfg: cfg}
}

// GET /edit.html
func (ic *ImageController) ShowEdit(c *gin.Context) {
	// 英文 preset 到中文描述的映射
	presets := map[string]string{
		"stylization_all":         "整体风格化",
		"stylization_local":       "局部风格化",
		"description_edit":        "描述编辑",
		"remove_watermark":        "去水印",
		"expand":                  "扩图",
		"super_resolution":        "超分辨率",
		"colorization":            "上色",
		"doodle":                  "涂鸦",
		"control_cartoon_feature": "卡通特征控制",
	}
	c.HTML(http.StatusOK, "edit.html", gin.H{"presets": presets})
}

// POST /api/image/edit
func (ic *ImageController) SubmitEdit(c *gin.Context) {
	email := c.GetString("userEmail")
	f, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传图片"})
		return
	}

	src, _ := f.Open()
	defer src.Close()
	img, _, err := image.Decode(src)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片解码失败"})
		return
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	if w > ic.Cfg.MaxImageWidth || h > ic.Cfg.MaxImageWidth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片尺寸超过限制,最大为 " + fmt.Sprintf("%dx%d", ic.Cfg.MaxImageWidth, ic.Cfg.MaxImageWidth)})
		return
	}

	if w < ic.Cfg.MinImageWidth || h < ic.Cfg.MinImageWidth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片尺寸过小,最小为 " + fmt.Sprintf("%dx%d", ic.Cfg.MaxImageWidth, ic.Cfg.MaxImageWidth)})
		// 自动裁切
		img = imaging.Fit(img, ic.Cfg.MaxImageWidth, ic.Cfg.MaxImageWidth, imaging.Lanczos)
	}

	// 保存图片到本地并构造对外可访问 URL
	id := uuid.NewString()
	os.MkdirAll("static/uploads", 0755)
	savePath := filepath.Join("static", "uploads", id+".png")
	imaging.Save(img, savePath)
	// 使用配置中 BaseUrl（请确保 .env: BASE_URL 使用可被 DashScope 访问的公网地址）
	baseURL := strings.TrimRight(ic.Cfg.BaseUrl, "/") + "/static/uploads/" + id + ".png"

	fn := c.PostForm("function")
	prompt := c.PostForm("prompt")
	// 业务逻辑允许同时选择不同的功能
	extras := map[string]interface{}{"n": 1}
	if v := c.PostForm("watermark"); v == "1" {
		extras["watermark"] = true
	}

	job := services.ImageJob{
		ID:           uuid.NewString(),
		ParentID:     nil,
		UserEmail:    email,
		Function:     fn,
		Prompt:       prompt,
		BaseImageURL: baseURL,
		Params:       extras,
		CreatedAt:    time.Now().Unix(),
	}
	// 放入数据库并保存原图 ID
	ic.DB.Create(&models.Task{
		ID:           job.ID,
		ParentID:     nil,
		UserEmail:    email,
		Function:     fn,
		Prompt:       prompt,
		BaseImageURL: baseURL,
		ImageID:      &id,
		Status:       "QUEUED",
		CreatedAt:    time.Now(),
	})
	ic.Queue.Enqueue(context.Background(), job)
	c.JSON(http.StatusAccepted, gin.H{"job_id": job.ID})
}

// GET /history.html
func (ic *ImageController) ShowHistory(c *gin.Context) {
	email := c.GetString("userEmail")
	var tasks []models.Task
	ic.DB.Where("user_email = ?", email).Order("created_at desc").Find(&tasks)
	c.HTML(http.StatusOK, "history.html", gin.H{"tasks": tasks})
}

// GET /history/:task_id
func (ic *ImageController) ShowBranch(c *gin.Context) {
	tid := c.Param("task_id")
	var root models.Task
	ic.DB.First(&root, "id=?", tid)
	var branches []models.Task
	ic.DB.Where("parent_id = ?", tid).Find(&branches)
	c.HTML(http.StatusOK, "branch.html", gin.H{
		"root": root, "branches": branches,
	})
}

// ListJSON 返回当前用户的图片历史（JSON）
func (ic *ImageController) ListJSON(c *gin.Context) {
	email := c.GetString("userEmail")
	var tasks []models.Task
	ic.DB.Where("user_email = ?", email).Order("created_at desc").Find(&tasks)
	var images []gin.H
	for _, t := range tasks {
		var url *string
		if t.ResultURL != nil && *t.ResultURL != "" {
			url = t.ResultURL
		} else {
			url = &t.BaseImageURL
		}
		images = append(images, gin.H{
			"id":     t.ID,
			"url":    url,
			"status": t.Status,
		})
	}
	c.JSON(http.StatusOK, gin.H{"images": images})
}
