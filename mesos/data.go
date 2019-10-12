package mesos

type Apps struct {
	Apps []struct {
		App
		Tasks []Task `json:"tasks"`
	} `json:"apps"`
}

type App struct {
	ID          string       `json:"id"`
	HealthCheck HealthChecks `json:"healthChecks"`
}

type Task struct {
	ID    string `json:"id"`
	Host  string `json:"host"`
	Ports []int  `json:"ports"`
	State State  `json:"state"`

	// Provided by client
	AppID       string      `json:"-"`
	HealthCheck HealthCheck `json:"-"`
}

type State string

const (
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

type Event struct {
	Type      string `json:"eventType"`
	TimeStamp string `json:"timestamp"`

	// Specific for "api_post_event"
	App App `json:"appDefinition"`

	// For "status_update_event"
	AppID     string `json:"appId"`
	Host      string `json:"host"`
	Ports     []int  `json:"ports"`
	TaskID    string `json:"taskId"`
	TaskState State  `json:"taskStatus"`
}

type HealthCheckProtocol string

const (
	HealthCheckHTTP      HealthCheckProtocol = "HTTP"
	HealthCheckMesosHTTP HealthCheckProtocol = "MESOS_HTTP"
)

type HealthChecks []HealthCheck

type HealthCheck struct {
	Protocol           HealthCheckProtocol `json:"protocol"`
	Interval           int                 `json:"intervalSeconds"`
	MaxConsecutiveFail int                 `json:"maxConsecutiveFailures"`
	TimeOut            int                 `json:"timeoutSeconds"`
	Delay              int                 `json:"delaySeconds"`
	Path               string              `json:"path"`
	PortIndex          int                 `json:"portIndex"`
	Command            struct {
		Value string `json:"value"`
	} `json:"command"`
}

func (hc HealthChecks) Get() *HealthCheck {
	for _, h := range hc {
		if h.Protocol == HealthCheckHTTP || h.Protocol == HealthCheckMesosHTTP {
			return &h
		}
	}
	return nil
}
