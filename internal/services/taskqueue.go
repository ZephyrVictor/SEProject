package services

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

const IMAGE_TASK_QUEUE = "image_edit_queue"

type ImageJob struct {
	ID           string                 `json:"id"`         // 唯一作业 ID
	ParentID     *string                `json:"parent_id"`  // 父作业 ID
	UserEmail    string                 `json:"user_email"` // 用户标识
	Function     string                 `json:"function"`   // 编辑功能
	Prompt       string                 `json:"prompt"`     // 提示词
	BaseImageURL string                 `json:"base_image_url"`
	Params       map[string]interface{} `json:"params"`
	CreatedAt    int64                  `json:"created_at"`
}

type TaskQueue struct {
	RDB *redis.Client
}

func NewTaskQueue(rdb *redis.Client) *TaskQueue {
	return &TaskQueue{RDB: rdb}
}

func (q *TaskQueue) Enqueue(ctx context.Context, job ImageJob) error {
	bs, _ := json.Marshal(job)
	return q.RDB.LPush(ctx, IMAGE_TASK_QUEUE, bs).Err()
}

func (q *TaskQueue) Dequeue(ctx context.Context) (*ImageJob, error) {
	res, err := q.RDB.BRPop(ctx, 0, IMAGE_TASK_QUEUE).Result()
	if err != nil {
		return nil, err
	}
	var job ImageJob
	if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
		return nil, err
	}
	return &job, nil
}
