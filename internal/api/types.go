package api

import "time"

// Token represents the OAuth2 token response from 42 API
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
	SecretValidUntil int64 `json:"secret_valid_until,omitempty"`
}

// ErrorResponse represents an error response from the 42 API
type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Message          string `json:"message"`
	Status           int    `json:"status"`
}

// User represents a 42 user
type User struct {
	ID                int               `json:"id"`
	Email             string            `json:"email"`
	Login             string            `json:"login"`
	FirstName         string            `json:"first_name"`
	LastName          string            `json:"last_name"`
	UsualName         string            `json:"usual_name"`
	URL               string            `json:"url"`
	Phone             string            `json:"phone"`
	DisplayName       string            `json:"displayname"`
	Image             UserImage         `json:"image"`
	Staff             bool              `json:"staff"`
	CorrectionPoint   int               `json:"correction_point"`
	PoolMonth         string            `json:"pool_month"`
	PoolYear          string            `json:"pool_year"`
	Location          string            `json:"location"`
	Wallet            int               `json:"wallet"`
	AnonymizeDate     *time.Time        `json:"anonymize_date"`
	DataErasureDate   *time.Time        `json:"data_erasure_date"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	AlumnizedAt       *time.Time        `json:"alumnized_at"`
	Alumni            bool              `json:"alumni"`
	Active            bool              `json:"active"`
	Groups            []Group           `json:"groups"`
	CursusUsers       []CursusUser      `json:"cursus_users"`
	ProjectsUsers     []ProjectUser     `json:"projects_users"`
	LanguagesUsers    []LanguageUser    `json:"languages_users"`
	Achievements      []Achievement     `json:"achievements"`
	Titles            []Title           `json:"titles"`
	TitlesUsers       []TitleUser       `json:"titles_users"`
	Partnerships      []Partnership     `json:"partnerships"`
	Patroned          []User            `json:"patroned"`
	Patroning         []User            `json:"patroning"`
	ExpertisesUsers   []ExpertiseUser   `json:"expertises_users"`
	Roles             []Role            `json:"roles"`
	Campus            []Campus          `json:"campus"`
	CampusUsers       []CampusUser      `json:"campus_users"`
}

// UserImage represents a user's profile image
type UserImage struct {
	Link     string              `json:"link"`
	Versions UserImageVersions   `json:"versions"`
}

// UserImageVersions represents different sizes of user images
type UserImageVersions struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
	Small  string `json:"small"`
	Micro  string `json:"micro"`
}

// Project represents a 42 project
type Project struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	Slug         string        `json:"slug"`
	Description  string        `json:"description"`
	Parent       *Project      `json:"parent"`
	Children     []Project     `json:"children"`
	Objectives   []string      `json:"objectives"`
	Tier         int           `json:"tier"`
	Attachment   *Attachment   `json:"attachment"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Exam         bool          `json:"exam"`
	GitURL       string        `json:"git_url"`
	Repository   string        `json:"repository"`
	Recommendation string     `json:"recommendation"`
	Cursus       []Cursus      `json:"cursus"`
	Videos       []Video       `json:"videos"`
	ProjectSessions []ProjectSession `json:"project_sessions"`
}

// ProjectUser represents a user's project
type ProjectUser struct {
	ID           int             `json:"id"`
	Occurrence   int             `json:"occurrence"`
	FinalMark    *int            `json:"final_mark"`
	Status       string          `json:"status"`
	Validated    *bool           `json:"validated"`
	CurrentTeamID *int           `json:"current_team_id"`
	Project      Project         `json:"project"`
	CursusIds    []int           `json:"cursus_ids"`
	MarkedAt     *time.Time      `json:"marked_at"`
	Marked       bool            `json:"marked"`
	RetriableAt  *time.Time      `json:"retriable_at"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	User         User            `json:"user"`
	Teams        []Team          `json:"teams"`
}

// Team represents a project team
type Team struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	URL          string        `json:"url"`
	FinalMark    *int          `json:"final_mark"`
	ProjectID    int           `json:"project_id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Status       string        `json:"status"`
	TerminatingAt *time.Time   `json:"terminating_at"`
	Users        []User        `json:"users"`
	Locked       bool          `json:"locked"`
	Validated    *bool         `json:"validated"`
	Closed       bool          `json:"closed"`
	RepoURL      string        `json:"repo_url"`
	RepoUUID     string        `json:"repo_uuid"`
	LockedAt     *time.Time    `json:"locked_at"`
	ClosedAt     *time.Time    `json:"closed_at"`
	ProjectSessionID int       `json:"project_session_id"`
	ProjectGitlabPath string   `json:"project_gitlab_path"`
}

// Cursus represents a 42 cursus (curriculum)
type Cursus struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Kind      string    `json:"kind"`
}

// CursusUser represents a user's cursus information
type CursusUser struct {
	ID         int       `json:"id"`
	BeginAt    time.Time `json:"begin_at"`
	EndAt      *time.Time `json:"end_at"`
	Grade      *string   `json:"grade"`
	Level      float64   `json:"level"`
	Skills     []Skill   `json:"skills"`
	BlackholedAt *time.Time `json:"blackholed_at"`
	User       User      `json:"user"`
	Cursus     Cursus    `json:"cursus"`
	HasCoalition bool   `json:"has_coalition"`
}

// Skill represents a cursus skill
type Skill struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Level float64 `json:"level"`
}

// Campus represents a 42 campus
type Campus struct {
	ID           int         `json:"id"`
	Name         string      `json:"name"`
	TimeZone     string      `json:"time_zone"`
	Language     Language    `json:"language"`
	UsersCount   int         `json:"users_count"`
	VogsphereID  int         `json:"vogsphere_id"`
	Country      string      `json:"country"`
	Address      string      `json:"address"`
	Zip          string      `json:"zip"`
	City         string      `json:"city"`
	Website      string      `json:"website"`
	Facebook     string      `json:"facebook"`
	Twitter      string      `json:"twitter"`
	Active       bool        `json:"active"`
	Public       bool        `json:"public"`
	EmailExtension string    `json:"email_extension"`
	DefaultHiddenPhone bool  `json:"default_hidden_phone"`
}

// CampusUser represents a user's campus relationship
type CampusUser struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	CampusID  int       `json:"campus_id"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Language represents a programming language
type Language struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// LanguageUser represents a user's language proficiency
type LanguageUser struct {
	ID         int      `json:"id"`
	LanguageID int      `json:"language_id"`
	UserID     int      `json:"user_id"`
	Position   int      `json:"position"`
	CreatedAt  time.Time `json:"created_at"`
}

// Achievement represents a 42 achievement
type Achievement struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Tier         string    `json:"tier"`
	Kind         string    `json:"kind"`
	Visible      bool      `json:"visible"`
	Image        string    `json:"image"`
	NbrOfSuccess *int      `json:"nbr_of_success"`
	UsersURL     string    `json:"users_url"`
}

// Title represents a 42 title
type Title struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TitleUser represents a user's title
type TitleUser struct {
	ID       int   `json:"id"`
	UserID   int   `json:"user_id"`
	TitleID  int   `json:"title_id"`
	Selected bool  `json:"selected"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Group represents a 42 group
type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Partnership represents a partnership
type Partnership struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

// ExpertiseUser represents a user's expertise
type ExpertiseUser struct {
	ID           int       `json:"id"`
	ExpertiseID  int       `json:"expertise_id"`
	Interested   bool      `json:"interested"`
	Value        int       `json:"value"`
	ContactMe    bool      `json:"contact_me"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       int       `json:"user_id"`
}

// Role represents a user role
type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Attachment represents a project attachment
type Attachment struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Video represents a project video
type Video struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ProjectSession represents a project session
type ProjectSession struct {
	ID                int         `json:"id"`
	Solo              bool        `json:"solo"`
	BeginAt           *time.Time  `json:"begin_at"`
	EndAt             *time.Time  `json:"end_at"`
	EstimateTime      string      `json:"estimate_time"`
	DurationDays      *int        `json:"duration_days"`
	TerminatingAfter  *int        `json:"terminating_after"`
	ProjectID         int         `json:"project_id"`
	CampusID          int         `json:"campus_id"`
	CursusID          int         `json:"cursus_id"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	MaxPeople         *int        `json:"max_people"`
	IsSubscriptable   bool        `json:"is_subscriptable"`
	Scales            []Scale     `json:"scales"`
	Uploads           []Upload    `json:"uploads"`
}

// Scale represents a project scale (evaluation)
type Scale struct {
	ID              int       `json:"id"`
	EvaluationID    int       `json:"evaluation_id"`
	Name            string    `json:"name"`
	IsIntroduction  bool      `json:"is_introduction"`
	CorrectionNumber int      `json:"correction_number"`
	Duration        int       `json:"duration"`
}

// Upload represents a project upload
type Upload struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ProjectSessionDetail represents a full project session response including rules
type ProjectSessionDetail struct {
	ID                   int                   `json:"id"`
	Solo                 bool                  `json:"solo"`
	BeginAt              *time.Time            `json:"begin_at"`
	EndAt                *time.Time            `json:"end_at"`
	EstimateTime         string                `json:"estimate_time"`
	DurationDays         *int                  `json:"duration_days"`
	TerminatingAfter     *int                  `json:"terminating_after"`
	ProjectID            int                   `json:"project_id"`
	CampusID             int                   `json:"campus_id"`
	CursusID             int                   `json:"cursus_id"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
	MaxPeople            *int                  `json:"max_people"`
	IsSubscriptable      bool                  `json:"is_subscriptable"`
	Scales               []Scale               `json:"scales"`
	Uploads              []Upload              `json:"uploads"`
	ProjectSessionsRules []ProjectSessionRule  `json:"project_sessions_rules"`
}

// ProjectSessionRule represents a rule attached to a project session
type ProjectSessionRule struct {
	ID       int                      `json:"id"`
	Required bool                     `json:"required"`
	Position int                      `json:"position"`
	Params   []ProjectSessionRuleParam `json:"params"`
	Rule     RuleDefinition           `json:"rule"`
}

// ProjectSessionRuleParam represents a parameter of a session rule
type ProjectSessionRuleParam struct {
	ID                    int       `json:"id"`
	ParamID               int       `json:"param_id"`
	ProjectSessionsRuleID int       `json:"project_sessions_rule_id"`
	Value                 string    `json:"value"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// RuleDefinition represents the kind and metadata of a rule
type RuleDefinition struct {
	ID           int       `json:"id"`
	Kind         string    `json:"kind"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Slug         string    `json:"slug"`
	InternalName string    `json:"internal_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Quest represents a 42 quest (progression checkpoint)
type Quest struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Kind         string     `json:"kind"`
	InternalName string     `json:"internal_name"`
	Description  string     `json:"description"`
	CursusID     int        `json:"cursus_id"`
	CampusID     *int       `json:"campus_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	GradeID      *int       `json:"grade_id"`
	Position     int        `json:"position"`
}

// QuestUser represents a user's quest completion record
type QuestUser struct {
	ID          int        `json:"id"`
	EndAt       *time.Time `json:"end_at"`
	QuestID     int        `json:"quest_id"`
	ValidatedAt *time.Time `json:"validated_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Quest       Quest      `json:"quest"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Count      int    `json:"count"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
	Links      Links  `json:"links"`
}

// Links represents pagination links
type Links struct {
	Self  string  `json:"self"`
	First string  `json:"first"`
	Prev  *string `json:"prev"`
	Next  *string `json:"next"`
	Last  string  `json:"last"`
}

// APIResponse represents a generic API response with pagination
type APIResponse[T any] struct {
	Data []T             `json:"data,omitempty"`
	Meta *PaginationMeta `json:"meta,omitempty"`
}