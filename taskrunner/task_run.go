package taskrunner

type JobRunState int

const (
	JOB_RUN_STATE_UNKNOWN JobRunState = iota
	JOB_RUN_STATE_FAILED
	JOB_RUN_STATE_SUCCESS
	JOB_RUN_STATE_IN_PROGRESS
	JOB_RUN_STATE_NOT_STARTED
)

var jobRunStates = [...]string{
	"Unknown",
	"Failed",
	"Success",
	"In Progress",
	"Not Started",
}

func (e JobRunState) String() string {
	return jobRunStates[e]
}

func (e JobRunState) IsFinished() bool {
	switch e {
	case JOB_RUN_STATE_SUCCESS, JOB_RUN_STATE_FAILED:
		return true
	default:
		return false
	}
}

type TriggerType string

type TaskRun struct {
	Id             uint64      `json:"id"`
	State          JobRunState `json:"status"`
	StartTimestamp int64       `json:"startTimestamp"`
	EndTimestamp   int64       `json:"endTimestamp,omitempty"`
	Trigger        TriggerType `json:"trigger"`
	Task           *Task       `json:"-"`
	Pid            *int        `json:"pid"`      // nil for not started
	ExitCode       *int        `json:"exitCode"` // nil for not started
	Logs           JobRunLogs  `json:"logs"`
}

type JobRunLogs struct {
	LogConfig LogConfig `json:"logConfig"`
	Stderr    LogFile   `json:"stderr"`
	Stdout    LogFile   `json:"stdout"`
}

type LogFile struct {
	RawSize        uint64 `json:"rawSize"`
	CompressedSize uint64 `json:"compressedSize"`
}

func (task *Task) NewTaskRun(trigger TriggerType) *TaskRun {
	return &TaskRun{
		Id:             0,
		State:          JOB_RUN_STATE_NOT_STARTED,
		StartTimestamp: 0,
		EndTimestamp:   0,
		Trigger:        trigger,
		Task:           task,
		Pid:            nil,
		ExitCode:       nil,
	}
}
