package taskexecutor

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Timestamp time.Time

func (t Timestamp) Value() (driver.Value, error) {
	tt := time.Time(t)
	return tt.UnixMilli(), nil
}
func (t *Timestamp) Scan(val interface{}) error {
	v := val.(int64)
	timeMills := time.UnixMilli(v)
	timeTimestamp := Timestamp(timeMills)
	*t = timeTimestamp
	return nil
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", time.Time(t).UnixMilli())), nil
}

func (t *Timestamp) UnmarshalJSON(v interface{}) ([]byte, error) {
	panic("not implemented")
}
