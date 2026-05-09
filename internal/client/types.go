package client

type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type Repository struct {
	ID                     string  `json:"id"`
	Key                    string  `json:"key"`
	Name                   string  `json:"name"`
	Format                 string  `json:"format"`
	RepoType               string  `json:"repo_type"`
	Description            *string `json:"description"`
	IsPublic               bool    `json:"is_public"`
	QuotaBytes             *int64  `json:"quota_bytes"`
	StorageUsedBytes       int64   `json:"storage_used_bytes"`
	UpstreamAuthConfigured bool    `json:"upstream_auth_configured"`
	UpstreamAuthType       *string `json:"upstream_auth_type"`
	UpstreamURL            *string `json:"upstream_url"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

type User struct {
	ID                 string  `json:"id"`
	Username           string  `json:"username"`
	Email              string  `json:"email"`
	DisplayName        *string `json:"display_name"`
	AuthProvider       string  `json:"auth_provider"`
	IsActive           bool    `json:"is_active"`
	IsAdmin            bool    `json:"is_admin"`
	MustChangePassword bool    `json:"must_change_password"`
	CreatedAt          string  `json:"created_at"`
	LastLoginAt        *string `json:"last_login_at"`
}

type Group struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	MemberCount int64   `json:"member_count"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	Members     []GroupMember
}

type GroupMember struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	JoinedAt    string  `json:"joined_at"`
}

type Permission struct {
	ID            string   `json:"id"`
	PrincipalType string   `json:"principal_type"`
	PrincipalID   string   `json:"principal_id"`
	PrincipalName *string  `json:"principal_name"`
	TargetType    string   `json:"target_type"`
	TargetID      string   `json:"target_id"`
	TargetName    *string  `json:"target_name"`
	Actions       []string `json:"actions"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type Webhook struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	URL             string   `json:"url"`
	Events          []string `json:"events"`
	RepositoryID    *string  `json:"repository_id"`
	IsEnabled       bool     `json:"is_enabled"`
	CreatedAt       string   `json:"created_at"`
	LastTriggeredAt *string  `json:"last_triggered_at"`
}

type Package struct {
	ID            string         `json:"id"`
	RepositoryKey string         `json:"repository_key"`
	Name          string         `json:"name"`
	Version       string         `json:"version"`
	Format        string         `json:"format"`
	SizeBytes     int64          `json:"size_bytes"`
	DownloadCount int64          `json:"download_count"`
	Description   *string        `json:"description"`
	Metadata      map[string]any `json:"metadata"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

type PackageVersion struct {
	Version       string `json:"version"`
	SizeBytes     int64  `json:"size_bytes"`
	DownloadCount int64  `json:"download_count"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}
