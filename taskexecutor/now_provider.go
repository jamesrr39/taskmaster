package taskexecutor

import (
	"time"
)

type NowProvider func() time.Time
