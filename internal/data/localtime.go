package data

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type LocalTime struct {
	time.Time
}

func (lt *LocalTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		return err
	}
	lt.Time = t
	return nil
}

func (lt LocalTime) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "\"%s\"", lt.Format("15:04:05")), nil
}

func (lt LocalTime) Before(other LocalTime) bool {
	return lt.Time.Before(other.Time)
}

func (lt *LocalTime) Value() (driver.Value, error) {
	if lt == nil {
		return nil, nil
	}
	return lt.Format("15:04:05"), nil
}

func (lt *LocalTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		lt.Time = v
		return nil
	case []byte:
		t, err := time.Parse("15:04:05", string(v))
		if err != nil {
			return err
		}
		lt.Time = t
		return nil
	case string:
		t, err := time.Parse("15:04:05", v)
		if err != nil {
			return err
		}
		lt.Time = t
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into LocalTime", value)
	}
}
