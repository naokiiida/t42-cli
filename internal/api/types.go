package api

// ErrorResponse represents a standard error response from the 42 API.
// It is used to unmarshal JSON error objects returned by various endpoints.
type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Message          string `json:"message"` // Some endpoints use "message"
}

// Token represents the successful JSON response from an OAuth token request.
// This struct is used when exchanging a code for an access token or when
// refreshing a token.
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
}

// TokenInfo represents the detailed information about an access token,
// as returned by the `/oauth/token/info` endpoint.
type TokenInfo struct {
	ResourceOwnerID  int      `json:"resource_owner_id"`
	Scope            []string `json:"scope"`
	ExpiresInSeconds int      `json:"expires_in_seconds"`
	Application      struct {
		UID string `json:"uid"`
	} `json:"application"`
	CreatedAt int64 `json:"created_at"`
}

// Project represents a single project entity in the 42 system (e.g., "Libft").
// It typically appears nested within other response types like ProjectsUser.
type Project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ProjectsUser represents a user's specific engagement with a project.
// It contains user-specific details like their final mark, status, and
// whether it has been validated. This is the primary object for the `project list` command.
type ProjectsUser struct {
	ID          int     `json:"id"`
	FinalMark   int     `json:"final_mark"`
	Status      string  `json:"status"`
	IsValidated bool    `json:"validated?"` // The API uses "validated?" with a question mark
	Project     Project `json:"project"`
	CursusIDs   []int   `json:"cursus_ids"`
	UpdatedAt   string  `json:"updated_at"`
}

// User represents a user's profile as returned by the `/v2/me` endpoint.
// It can include a comprehensive list of the user's projects.
type User struct {
	ID            int            `json:"id"`
	Email         string         `json:"email"`
	Login         string         `json:"login"`
	DisplayName   string         `json:"displayname"`
	ImageURL      string         `json:"image_url"`
	ProjectsUsers []ProjectsUser `json:"projects_users"`
}

// Cursus represents a cursus within the 42 system (e.g., "42cursus").
type Cursus struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}