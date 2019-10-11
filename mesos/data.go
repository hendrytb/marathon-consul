package mesos

type Apps struct {
	Apps []App `json:"apps"`
}

type App struct {
	ID    string `json:"id"`
	Tasks []Task `json:"tasks"`
}

type Task struct {
	ID    string `json:"id"`
	Host  string `json:"host"`
	Ports []int  `json:"ports"`
	State State  `json:"state"`
}

type State string

const (
	StateUnknown  State = ""
	StateStaging  State = "TASK_STAGING"
	StateStarting State = "TASK_STARTING"
	StateRunning  State = "TASK_RUNNING"
	StateFinished State = "TASK_FINISHED"
	StateFailed   State = "TASK_FAILED"
	StateKilling  State = "TASK_KILLING" // (only when the task_killing feature is enabled)
	StateKilled   State = "TASK_KILLED"
	StateLost     State = "TASK_LOST"
)

// List of task related events (eventType)
// - status_update_event
// - app_terminated_event
// - instance_changed_event
// - unknown_instance_terminated_event
// - instance_health_changed_event

type EventStatusUpdate struct {
	Type      string `json:"eventType"`
	TimeStamp string `json:"timestamp"`
	AppID     string `json:"appId"`
	Host      string `json:"host"`
	Ports     []int  `json:"ports"`
	TaskID    string `json:"taskId"`
	TaskState State  `json:"taskStatus"`
}
