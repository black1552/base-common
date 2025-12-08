package task

import "sync"

// TaskManager 任务管理器
type sTaskManager struct {
	tasks map[string]*Task
	mutex sync.RWMutex
}

var Manager = &sTaskManager{}

func init() {
	Manager = &sTaskManager{
		tasks: make(map[string]*Task),
	}
}

// CreateTask 创建任务
func (m *sTaskManager) CreateTask(token string, total int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.tasks[token] = &Task{
		Processed: 0,
		Total:     total,
		Status:    TaskStatusProcessing,
		Message:   "开始处理...",
		Path:      "",
	}
}

// UpdateProgress 更新任务进度
func (m *sTaskManager) UpdateProgress(token string, processed int, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if task, exists := m.tasks[token]; exists {
		task.Processed = processed
		task.Message = message
	}
}

// CompleteTask 完成任务
func (m *sTaskManager) CompleteTask(token string, message string, path string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if task, exists := m.tasks[token]; exists {
		task.Status = TaskStatusCompleted
		task.Message = message
		task.Path = path
	}
}

// FailTask 失败任务
func (m *sTaskManager) FailTask(token string, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if task, exists := m.tasks[token]; exists {
		task.Status = TaskStatusFailed
		task.Message = message
	}
}

// GetTask 获取任务
func (m *sTaskManager) GetTask(token string) (*Task, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	task, exists := m.tasks[token]
	return task, exists
}

// RemoveTask 移除任务
func (m *sTaskManager) RemoveTask(token string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.tasks, token)
}
