package models

import (
	"bytes"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Project is a single DBMS created in a region.
type Project struct {
	gorm.Model

	// RegionID is a foreign key to the region.
	RegionID uint

	// Name is given by the user.
	Name string

	// ProjectID is given by the provider.
	ProjectID string

	// ConnectionString to the main branch.
	ConnectionString string

	// Taken from `EXITNODE` environment variable.
	CreatedByExitnode string

	// Specified at the creation time.
	PgVersion int

	// Specified at the creation time.
	Provisioner string

	// Default endpoint will be shut down after this timeout.
	SuspendTimeoutSeconds int

	// Mode name, which is used by rules to define query strategy.
	CurrentMode string

	// TODO:
	// Comment about a policy of creation.
	// CreationComment string

	// Comment about a policy of deletion.
	// DeletionComment string
}

func (p *Project) SuspendTimeout() time.Duration {
	if p.SuspendTimeoutSeconds == 0 {
		const defaultTimeout = 5 * 60
		return time.Second * defaultTimeout
	}
	return time.Second * time.Duration(p.SuspendTimeoutSeconds)
}

func CommonProjectFeatures(projects []Project) map[string]json.RawMessage {
	var features map[string]json.RawMessage
	for _, project := range projects {
		j, _ := json.Marshal(project)
		var f map[string]json.RawMessage
		_ = json.Unmarshal(j, &f)

		if features == nil {
			features = f
			continue
		}

		for k, v := range features {
			if !bytes.Equal(v, f[k]) {
				delete(features, k)
			}
		}
	}
	return features
}
