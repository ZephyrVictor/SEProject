package internal

import (
	"context"
	"log"
	"time"

	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

// StartImageWorker 从 Redis 队列消费 ImageJob，限并发 2，创建 DashScope 任务并轮询结果。
// 成功后写入 MySQL 并通过 WebSocket 推送给用户。
func StartImageWorker(ctx context.Context, dashClient *services.DashScopeClient, queue *services.TaskQueue, db *gorm.DB) {
	sem := make(chan struct{}, 2)
	for {
		job, err := queue.Dequeue(ctx)
		if err != nil {
			log.Println("Error dequeuing job:", err)
			continue
		}

		sem <- struct{}{} // Acquire slot
		go func(job *services.ImageJob) {
			defer func() { <-sem }() // Release slot

			// 更新任务状态为 PENDING
			db.Model(&models.Task{}).
				Where("id = ?", job.ID).
				Updates(map[string]interface{}{
					"status":     "PENDING",
					"updated_at": time.Now(),
				})

			log.Println("Submitting task:", job.ID)
			taskID, err := dashClient.SubmitTask(ctx,
				job.Function, job.Prompt,
				job.BaseImageURL, job.Params)
			if err != nil {
				log.Println("Error submitting task:", err)
				db.Model(&models.Task{}).Where("id = ?", job.ID).
					Updates(map[string]interface{}{"status": "FAILED", "updated_at": time.Now()})
				NotifyUser(job.UserEmail, map[string]interface{}{
					"job_id": job.ID, "status": "FAILED",
				})
				return
			}

			// Polling loop
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					status, urls, err := dashClient.PollTask(ctx, taskID)
					if err != nil {
						log.Println("Error polling task:", err)
						continue
					}
					log.Printf("Job[%s] status: %s\n", job.ID, status)

					switch status {
					case "SUCCEEDED":
						// 取第一个结果 URL
						var remoteURL string
						if len(urls) > 0 {
							remoteURL = urls[0]
						}
						// 下载远程结果，保存到本地上传目录
						localDir := filepath.Join("static", "uploads")
						os.MkdirAll(localDir, 0755)
						localName := job.ID + "_res.png"
						localPath := filepath.Join(localDir, localName)
						if remoteURL != "" {
							if resp, err := http.Get(remoteURL); err == nil {
								defer resp.Body.Close()
								out, _ := os.Create(localPath)
								defer out.Close()
								io.Copy(out, resp.Body)
							}
						}
						// 构造可访问 URL
						base := strings.TrimRight(models.Cfg.BaseUrl, "/")
						localURL := base + "/static/uploads/" + localName
						// 更新数据库状态和结果 URL
						db.Model(&models.Task{}).Where("id = ?", job.ID).
							Updates(map[string]interface{}{
								"status":     "SUCCEEDED",
								"result_url": localURL,
								"updated_at": time.Now(),
							})
						// WebSocket 通知客户端
						NotifyUser(job.UserEmail, map[string]interface{}{
							"job_id": job.ID, "status": "SUCCEEDED", "url": localURL,
						})
						return

					case "FAILED", "CANCELED":
						db.Model(&models.Task{}).Where("id = ?", job.ID).
							Updates(map[string]interface{}{
								"status":     status,
								"updated_at": time.Now(),
							})
						NotifyUser(job.UserEmail, map[string]interface{}{
							"job_id": job.ID, "status": status,
						})
						return

					}
				}
			}
		}(job)
	}
}
