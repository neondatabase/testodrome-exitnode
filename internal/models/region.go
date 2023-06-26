package models

import "gorm.io/gorm"

type Region struct {
	gorm.Model

	// Name of the provider, e.g. "neon.tech"
	Provider string

	// Name of the region, e.g. "aws-us-east-1"
	DatabaseRegion string

	// Whether the region supports autoscaling (neonvm provisioner)
	SupportsNeonVM bool
}
