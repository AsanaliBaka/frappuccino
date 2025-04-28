package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Customer struct {
	ID          string  `json:"customer_id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Phone       string  `json:"phone"`
	Preferences PrefMap `json:"preferences"`
}

type PrefMap map[string]interface{}

func (p PrefMap) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *PrefMap) Scan(src interface{}) error {
	var rawData []byte

	switch v := src.(type) {
	case []byte:
		rawData = v
	case string:
		rawData = []byte(v)
	case nil:
		*p = make(PrefMap)
		return nil
	default:
		return fmt.Errorf("incompatible type for PrefMap: %T", src)
	}

	if len(rawData) == 0 {
		*p = make(PrefMap)
		return nil
	}

	parsed := make(map[string]interface{})
	if err := json.Unmarshal(rawData, &parsed); err != nil {
		return fmt.Errorf("failed to unmarshal PrefMap: %w", err)
	}

	*p = parsed
	return nil
}
