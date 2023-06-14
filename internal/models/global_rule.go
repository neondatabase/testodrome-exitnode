package models

import "encoding/json"

// A database model for a global rules, that are periodically executed on all nodes.
type GlobalRule struct {
	Enabled  bool
	Priority int
	// rdesc.Rule
	Desc json.RawMessage `gorm:"type:jsonb"`
}
