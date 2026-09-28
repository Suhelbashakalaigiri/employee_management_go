package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const DateFormat = "2006-01-02"

// Date represents a calendar date in YYYY-MM-DD format.
type Date string

// String returns the string representation of the Date.
func (d Date) String() string {
	return string(d)
}

// Time parses the Date into a time.Time in UTC.
func (d Date) Time() (time.Time, error) {
	return time.Parse(DateFormat, string(d))
}

// IsValid checks if the Date is non-empty and in valid YYYY-MM-DD format.
func (d Date) IsValid() bool {
	if strings.TrimSpace(string(d)) == "" {
		return false
	}
	_, err := time.Parse(DateFormat, string(d))
	return err == nil
}

// MarshalJSON marshals Date to a JSON string in YYYY-MM-DD format.
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(d))
}

// UnmarshalJSON unmarshals a JSON string into Date.
func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*d = Date(strings.TrimSpace(s))
	return nil
}

// Scan implements the sql.Scanner interface for reading DATE columns from MySQL.
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		*d = ""
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*d = Date(v.Format(DateFormat))
		return nil
	case []byte:
		str := string(v)
		// Handle timestamp format if returned by driver
		if len(str) >= 10 {
			*d = Date(str[:10])
		} else {
			*d = Date(str)
		}
		return nil
	case string:
		if len(v) >= 10 {
			*d = Date(v[:10])
		} else {
			*d = Date(v)
		}
		return nil
	default:
		return fmt.Errorf("model.Date: cannot scan type %T into Date", value)
	}
}

// Value implements the driver.Valuer interface for writing to MySQL DATE columns.
func (d Date) Value() (driver.Value, error) {
	if d == "" {
		return nil, nil
	}
	t, err := time.Parse(DateFormat, string(d))
	if err != nil {
		return nil, fmt.Errorf("model.Date: invalid date format %q, expected YYYY-MM-DD: %w", d, err)
	}
	return t.Format(DateFormat), nil
}
