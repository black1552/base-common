package task

// Task 通用任务结构体
type Task struct {
	Processed int        `json:"processed" dc:"已处理数量"`
	Total     int        `json:"total" dc:"总计数量"`
	Status    TaskStatus `json:"status" dc:"状态：processing, completed, failed"`
	Message   string     `json:"message" dc:"消息"`
	Path      string     `json:"path" dc:"文件路径"`
}
