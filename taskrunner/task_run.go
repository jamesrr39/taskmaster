package taskrunner

type JobRunState int

const (
	JOB_RUN_STATE_UNKNOWN      JobRunState = 0
	JOB_RUN_STATE_FAILED       JobRunState = 1
	JOB_RUN_STATE_SUCCESS      JobRunState = 2
	JOB_RUN_STATE_IN_PROGRESS  JobRunState = 3
	JOB_RUN_STATE_NOT_STARTED  JobRunState = 4
	JOB_RUN_STATE_FAILED_SETUP JobRunState = 5 // failed before user script could be run (e.g. couldn't write script to disk to execute it)
)

var jobRunStates = [...]string{
	"Unknown",
	"Failed",
	"Success",
	"In Progress",
	"Not Started",
	"Setup Failed",
}

func (e JobRunState) String() string {
	return jobRunStates[e]
}

func (e JobRunState) IsFinished() bool {
	switch e {
	case JOB_RUN_STATE_SUCCESS, JOB_RUN_STATE_FAILED, JOB_RUN_STATE_FAILED_SETUP:
		return true
	default:
		return false
	}
}

type TaskRun struct {
	RunNumber      uint64      `json:"runNumber" db:"task_run_number"`
	State          JobRunState `json:"status"`
	StartTimestamp Timestamp   `json:"startTimestamp" db:"start_time"`
	EndTimestamp   *Timestamp  `json:"endTimestamp,omitempty"`
	Task           *Task       `json:"-"`
	Pid            *int        `json:"pid"`      // nil for not started
	ExitCode       *int        `json:"exitCode"` // nil for not started
	Logs           JobRunLogs  `json:"logs" required:"true"`
}

type JobRunLogs struct {
	LogConfig LogConfig `json:"logConfig"`
	Stderr    LogFile   `json:"stderr" required:"true"`
	Stdout    LogFile   `json:"stdout" required:"true"`
}

type LogFile struct {
	RawSize        uint64 `json:"rawSize"`
	CompressedSize uint64 `json:"compressedSize"`
}

func (task *Task) NewTaskRun() *TaskRun {
	return &TaskRun{
		RunNumber:      0,
		State:          JOB_RUN_STATE_NOT_STARTED,
		StartTimestamp: Timestamp{},
		EndTimestamp:   nil,
		Task:           task,
		Pid:            nil,
		ExitCode:       nil,
	}
}
