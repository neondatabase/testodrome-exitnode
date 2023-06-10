package models

import "gorm.io/gorm"

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

	// TODO:
	// Comment about a policy of creation.
	// CreationComment string

	// Comment about a policy of deletion.
	// DeletionComment string
}
