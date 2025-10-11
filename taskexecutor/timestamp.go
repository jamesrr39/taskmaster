package taskexecutor

import (
	"database/sql/driver"
	"time"
)

type Timestamp time.Time

func (t Timestamp) Value() (driver.Value, error) {
	tt := time.Time(t)
	return tt.UnixMilli(), nil
}
func (t Timestamp) Scan(val interface{}) error {
	v := val.(int64)
	tt := time.UnixMilli(v)
	t = Timestamp(tt)
	return nil
}
