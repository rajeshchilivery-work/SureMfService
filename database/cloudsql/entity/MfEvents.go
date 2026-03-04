package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONB is a helper type for Postgres jsonb columns
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, j)
}

type MfEvent struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"not null"`
	EventType   string    `json:"event_type" gorm:"not null"`
	FpEntityID  *string   `json:"fp_entity_id"`
	ISIN        *string   `json:"isin"`
	Amount      *float64  `json:"amount"`
	Units       *float64  `json:"units"`
	RawPayload  JSONB     `json:"raw_payload" gorm:"type:jsonb"`
	EventAt     time.Time `json:"event_at"`
}

func (MfEvent) TableName() string {
	return "sure_mf.mf_events"
}
