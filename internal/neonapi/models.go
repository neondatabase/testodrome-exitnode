package neonapi

import "time"

type CreateProjectRequest struct {
	Project *CreateProject `json:"project"`
}

type CreateProjectResponse struct {
	Project        Project         `json:"project"`
	ConnectionUris []ConnectionURI `json:"connection_uris"`
	Roles          []Role          `json:"roles"`
	Databases      []Database      `json:"databases"`
	Operations     []Operation     `json:"operations"`
	Branch         Branch          `json:"branch"`
	Endpoints      []Endpoint      `json:"endpoints"`
}

type Project struct {
	DataStorageBytesHour        int       `json:"data_storage_bytes_hour"`
	DataTransferBytes           int       `json:"data_transfer_bytes"`
	WrittenDataBytes            int       `json:"written_data_bytes"`
	ComputeTimeSeconds          int       `json:"compute_time_seconds"`
	ActiveTimeSeconds           int       `json:"active_time_seconds"`
	CPUUsedSec                  int       `json:"cpu_used_sec"`
	ID                          string    `json:"id"`
	PlatformID                  string    `json:"platform_id"`
	RegionID                    string    `json:"region_id"`
	Name                        string    `json:"name"`
	Provisioner                 string    `json:"provisioner"`
	PgVersion                   int       `json:"pg_version"`
	ProxyHost                   string    `json:"proxy_host"`
	BranchLogicalSizeLimit      int       `json:"branch_logical_size_limit"`
	BranchLogicalSizeLimitBytes int64     `json:"branch_logical_size_limit_bytes"`
	StorePasswords              bool      `json:"store_passwords"`
	CreationSource              string    `json:"creation_source"`
	HistoryRetentionSeconds     int       `json:"history_retention_seconds"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
	ConsumptionPeriodStart      time.Time `json:"consumption_period_start"`
	ConsumptionPeriodEnd        time.Time `json:"consumption_period_end"`
	OwnerID                     string    `json:"owner_id"`
}

type ConnectionParameters struct {
	Database   string `json:"database"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	Host       string `json:"host"`
	PoolerHost string `json:"pooler_host"`
}

type ConnectionURI struct {
	ConnectionURI        string               `json:"connection_uri"`
	ConnectionParameters ConnectionParameters `json:"connection_parameters"`
}

type Role struct {
	BranchID  string    `json:"branch_id"`
	Name      string    `json:"name"`
	Password  string    `json:"password"`
	Protected bool      `json:"protected"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Database struct {
	ID        int       `json:"id"`
	BranchID  string    `json:"branch_id"`
	Name      string    `json:"name"`
	OwnerName string    `json:"owner_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Operation struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	BranchID      string    `json:"branch_id"`
	Action        string    `json:"action"`
	Status        string    `json:"status"`
	FailuresCount int       `json:"failures_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	EndpointID    string    `json:"endpoint_id,omitempty"`
}

type Branch struct {
	ID                 string    `json:"id"`
	ProjectID          string    `json:"project_id"`
	Name               string    `json:"name"`
	CurrentState       string    `json:"current_state"`
	PendingState       string    `json:"pending_state"`
	CreationSource     string    `json:"creation_source"`
	Primary            bool      `json:"primary"`
	CPUUsedSec         int       `json:"cpu_used_sec"`
	ComputeTimeSeconds int       `json:"compute_time_seconds"`
	ActiveTimeSeconds  int       `json:"active_time_seconds"`
	WrittenDataBytes   int       `json:"written_data_bytes"`
	DataTransferBytes  int       `json:"data_transfer_bytes"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Settings struct {
}

type Endpoint struct {
	Host                  string    `json:"host"`
	ID                    string    `json:"id"`
	ProjectID             string    `json:"project_id"`
	BranchID              string    `json:"branch_id"`
	AutoscalingLimitMinCu float64   `json:"autoscaling_limit_min_cu"`
	AutoscalingLimitMaxCu float64   `json:"autoscaling_limit_max_cu"`
	RegionID              string    `json:"region_id"`
	Type                  string    `json:"type"`
	CurrentState          string    `json:"current_state"`
	PendingState          string    `json:"pending_state"`
	Settings              Settings  `json:"settings"`
	PoolerEnabled         bool      `json:"pooler_enabled"`
	PoolerMode            string    `json:"pooler_mode"`
	Disabled              bool      `json:"disabled"`
	PasswordlessAccess    bool      `json:"passwordless_access"`
	CreationSource        string    `json:"creation_source"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	ProxyHost             string    `json:"proxy_host"`
	SuspendTimeoutSeconds int       `json:"suspend_timeout_seconds"`
	Provisioner           string    `json:"provisioner"`
}

type CreateProject struct {
	Name     string `json:"name"`
	RegionID string `json:"region_id"`

	PgVersion   int    `json:"pg_version"`
	Provisioner string `json:"provisioner"`
}

type DeleteProjectResponse struct {
	Project struct {
		DataStorageBytesHour        int       `json:"data_storage_bytes_hour"`
		DataTransferBytes           int       `json:"data_transfer_bytes"`
		WrittenDataBytes            int       `json:"written_data_bytes"`
		ComputeTimeSeconds          int       `json:"compute_time_seconds"`
		ActiveTimeSeconds           int       `json:"active_time_seconds"`
		CPUUsedSec                  int       `json:"cpu_used_sec"`
		ID                          string    `json:"id"`
		PlatformID                  string    `json:"platform_id"`
		RegionID                    string    `json:"region_id"`
		Name                        string    `json:"name"`
		Provisioner                 string    `json:"provisioner"`
		PgVersion                   int       `json:"pg_version"`
		ProxyHost                   string    `json:"proxy_host"`
		BranchLogicalSizeLimit      int       `json:"branch_logical_size_limit"`
		BranchLogicalSizeLimitBytes int64     `json:"branch_logical_size_limit_bytes"`
		StorePasswords              bool      `json:"store_passwords"`
		CreationSource              string    `json:"creation_source"`
		HistoryRetentionSeconds     int       `json:"history_retention_seconds"`
		CreatedAt                   time.Time `json:"created_at"`
		UpdatedAt                   time.Time `json:"updated_at"`
		SyntheticStorageSize        int       `json:"synthetic_storage_size"`
		ConsumptionPeriodStart      time.Time `json:"consumption_period_start"`
		ConsumptionPeriodEnd        time.Time `json:"consumption_period_end"`
		OwnerID                     string    `json:"owner_id"`
	} `json:"project"`
}

type UpdateEndpointRequest struct {
	Endpoint *UpdateEndpoint `json:"endpoint"`
}

type UpdateEndpoint struct {
	SuspendTimeoutSeconds *int `json:"suspend_timeout_seconds"`
}

type UpdateEndpointResponse struct {
	Endpoint   *Endpoint   `json:"endpoint"`
	Operations []Operation `json:"operations"`
}

type GetOperationsResponse struct {
	Operations []Operation `json:"operations"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Cursor time.Time `json:"cursor"`
}
