package cmd

import (
	"testing"
	"time"

	"github.com/naokiiida/t42-cli/internal/api"
)

func TestMatchesBlackholeStatus(t *testing.T) {
	now := time.Now()
	pastDate := now.AddDate(0, 0, -10)    // 10 days ago
	futureDate := now.AddDate(0, 0, 10)   // 10 days from now
	farFuture := now.AddDate(0, 0, 60)    // 60 days from now

	tests := []struct {
		name       string
		cursusUser *api.CursusUser
		status     string
		days       int
		want       bool
	}{
		// "none" status tests
		{
			name:       "none: nil cursusUser returns true",
			cursusUser: nil,
			status:     "none",
			days:       30,
			want:       true,
		},
		{
			name: "none: no blackhole date returns true",
			cursusUser: &api.CursusUser{
				BlackholedAt: nil,
			},
			status: "none",
			days:   30,
			want:   true,
		},
		{
			name: "none: has blackhole date returns false",
			cursusUser: &api.CursusUser{
				BlackholedAt: &futureDate,
			},
			status: "none",
			days:   30,
			want:   false,
		},

		// "active" status tests
		{
			name:       "active: nil cursusUser returns false",
			cursusUser: nil,
			status:     "active",
			days:       30,
			want:       false,
		},
		{
			name: "active: has blackhole but no end date returns true",
			cursusUser: &api.CursusUser{
				BlackholedAt: &futureDate,
				EndAt:        nil,
			},
			status: "active",
			days:   30,
			want:   true,
		},
		{
			name: "active: has blackhole and end date returns false",
			cursusUser: &api.CursusUser{
				BlackholedAt: &pastDate,
				EndAt:        &pastDate,
			},
			status: "active",
			days:   30,
			want:   false,
		},

		// "past" status tests - this was the bug we fixed
		{
			name:       "past: nil cursusUser returns false",
			cursusUser: nil,
			status:     "past",
			days:       30,
			want:       false,
		},
		{
			name: "past: blackhole in past with EndAt set returns true",
			cursusUser: &api.CursusUser{
				BlackholedAt: &pastDate,
				EndAt:        &pastDate,
			},
			status: "past",
			days:   30,
			want:   true,
		},
		{
			name: "past: blackhole in past WITHOUT EndAt returns false (not actually blackholed)",
			cursusUser: &api.CursusUser{
				BlackholedAt: &pastDate,
				EndAt:        nil,
			},
			status: "past",
			days:   30,
			want:   false,
		},
		{
			name: "past: blackhole in future returns false",
			cursusUser: &api.CursusUser{
				BlackholedAt: &futureDate,
				EndAt:        nil,
			},
			status: "past",
			days:   30,
			want:   false,
		},

		// "upcoming" status tests
		{
			name:       "upcoming: nil cursusUser returns false",
			cursusUser: nil,
			status:     "upcoming",
			days:       30,
			want:       false,
		},
		{
			name: "upcoming: no blackhole returns false",
			cursusUser: &api.CursusUser{
				BlackholedAt: nil,
			},
			status:     "upcoming",
			days:       30,
			want:       false,
		},
		{
			name: "upcoming: blackhole within threshold returns true",
			cursusUser: &api.CursusUser{
				BlackholedAt: &futureDate, // 10 days from now
			},
			status: "upcoming",
			days:   30, // threshold is 30 days
			want:   true,
		},
		{
			name: "upcoming: blackhole beyond threshold returns false",
			cursusUser: &api.CursusUser{
				BlackholedAt: &farFuture, // 60 days from now
			},
			status: "upcoming",
			days:   30, // threshold is 30 days
			want:   false,
		},
		{
			name: "upcoming: blackhole in past returns false",
			cursusUser: &api.CursusUser{
				BlackholedAt: &pastDate,
			},
			status: "upcoming",
			days:   30,
			want:   false,
		},

		// Unknown status
		{
			name: "unknown status returns true",
			cursusUser: &api.CursusUser{
				BlackholedAt: &futureDate,
			},
			status: "unknown_status",
			days:   30,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesBlackholeStatus(tt.cursusUser, tt.status, tt.days, now)
			if got != tt.want {
				t.Errorf("matchesBlackholeStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindCursusUser(t *testing.T) {
	now := time.Now()
	futureDate := now.AddDate(0, 0, 10)
	pastDate := now.AddDate(0, 0, -10)

	cursusUsers := []api.CursusUser{
		{
			ID:    1,
			Level: 5.0,
			Cursus: api.Cursus{
				ID:   9,
				Name: "Piscine C",
			},
			EndAt: &pastDate, // ended
		},
		{
			ID:    2,
			Level: 10.5,
			Cursus: api.Cursus{
				ID:   21,
				Name: "42cursus",
			},
			EndAt: nil, // active
		},
		{
			ID:    3,
			Level: 3.0,
			Cursus: api.Cursus{
				ID:   67,
				Name: "Discovery Piscine",
			},
			EndAt: &futureDate, // will end in future
		},
	}

	tests := []struct {
		name        string
		cursusUsers []api.CursusUser
		cursusID    int
		wantID      int
		wantNil     bool
	}{
		{
			name:        "find specific cursus by ID",
			cursusUsers: cursusUsers,
			cursusID:    21,
			wantID:      2,
		},
		{
			name:        "find piscine cursus by ID",
			cursusUsers: cursusUsers,
			cursusID:    9,
			wantID:      1,
		},
		{
			name:        "cursus not found returns nil",
			cursusUsers: cursusUsers,
			cursusID:    999,
			wantNil:     true,
		},
		{
			name:        "cursusID 0 returns most recent active cursus",
			cursusUsers: cursusUsers,
			cursusID:    0,
			wantID:      3, // Discovery Piscine is last and EndAt is in future (still active)
		},
		{
			name:        "empty cursusUsers with cursusID 0 returns nil",
			cursusUsers: []api.CursusUser{},
			cursusID:    0,
			wantNil:     true,
		},
		{
			name: "all ended cursusUsers with cursusID 0 returns last one",
			cursusUsers: []api.CursusUser{
				{ID: 1, Cursus: api.Cursus{ID: 9}, EndAt: &pastDate},
				{ID: 2, Cursus: api.Cursus{ID: 21}, EndAt: &pastDate},
			},
			cursusID: 0,
			wantID:   2, // returns last one when all are ended
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findCursusUser(tt.cursusUsers, tt.cursusID)
			if tt.wantNil {
				if got != nil {
					t.Errorf("findCursusUser() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("findCursusUser() = nil, want non-nil with ID %d", tt.wantID)
				return
			}
			if got.ID != tt.wantID {
				t.Errorf("findCursusUser().ID = %d, want %d", got.ID, tt.wantID)
			}
		})
	}
}

func TestCountCompletedProjects(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name         string
		projectUsers []api.ProjectUser
		want         int
	}{
		{
			name:         "empty list returns 0",
			projectUsers: []api.ProjectUser{},
			want:         0,
		},
		{
			name: "all validated returns count",
			projectUsers: []api.ProjectUser{
				{Validated: &trueVal},
				{Validated: &trueVal},
				{Validated: &trueVal},
			},
			want: 3,
		},
		{
			name: "none validated returns 0",
			projectUsers: []api.ProjectUser{
				{Validated: &falseVal},
				{Validated: &falseVal},
			},
			want: 0,
		},
		{
			name: "mixed validation returns correct count",
			projectUsers: []api.ProjectUser{
				{Validated: &trueVal},
				{Validated: &falseVal},
				{Validated: &trueVal},
				{Validated: nil}, // in progress
			},
			want: 2,
		},
		{
			name: "nil validated values not counted",
			projectUsers: []api.ProjectUser{
				{Validated: nil},
				{Validated: nil},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countCompletedProjects(tt.projectUsers)
			if got != tt.want {
				t.Errorf("countCompletedProjects() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFilterUsers(t *testing.T) {
	now := time.Now()
	pastDate := now.AddDate(0, 0, -10)
	futureDate := now.AddDate(0, 0, 10)
	trueVal := true

	users := []api.User{
		{
			ID:    1,
			Login: "user1",
			CursusUsers: []api.CursusUser{
				{Cursus: api.Cursus{ID: 21}, Level: 5.0, BlackholedAt: nil},
			},
			ProjectsUsers: []api.ProjectUser{
				{Validated: &trueVal},
				{Validated: &trueVal},
			},
		},
		{
			ID:    2,
			Login: "user2",
			CursusUsers: []api.CursusUser{
				{Cursus: api.Cursus{ID: 21}, Level: 10.0, BlackholedAt: &futureDate},
			},
			ProjectsUsers: []api.ProjectUser{
				{Validated: &trueVal},
			},
		},
		{
			ID:    3,
			Login: "user3",
			CursusUsers: []api.CursusUser{
				{Cursus: api.Cursus{ID: 21}, Level: 15.0, BlackholedAt: &pastDate, EndAt: &pastDate},
			},
			ProjectsUsers: []api.ProjectUser{
				{Validated: &trueVal},
				{Validated: &trueVal},
				{Validated: &trueVal},
				{Validated: &trueVal},
				{Validated: &trueVal},
			},
		},
	}

	tests := []struct {
		name     string
		criteria filterCriteria
		wantLen  int
		wantIDs  []int
	}{
		{
			name:     "no filters returns all",
			criteria: filterCriteria{},
			wantLen:  3,
		},
		{
			name: "filter by min projects",
			criteria: filterCriteria{
				minProjects: 3,
			},
			wantLen: 1,
			wantIDs: []int{3},
		},
		{
			name: "filter by min level",
			criteria: filterCriteria{
				minLevel: 8.0,
				cursusID: 21,
			},
			wantLen: 2,
			wantIDs: []int{2, 3},
		},
		{
			name: "filter by max level",
			criteria: filterCriteria{
				maxLevel: 6.0,
				cursusID: 21,
			},
			wantLen: 1,
			wantIDs: []int{1},
		},
		{
			name: "filter by level range",
			criteria: filterCriteria{
				minLevel: 5.0,
				maxLevel: 12.0,
				cursusID: 21,
			},
			wantLen: 2,
			wantIDs: []int{1, 2},
		},
		{
			name: "filter by blackhole status 'none'",
			criteria: filterCriteria{
				blackholeStatus: "none",
				cursusID:        21,
			},
			wantLen: 1,
			wantIDs: []int{1},
		},
		{
			name: "filter by blackhole status 'past'",
			criteria: filterCriteria{
				blackholeStatus: "past",
				cursusID:        21,
			},
			wantLen: 1,
			wantIDs: []int{3},
		},
		{
			name: "filter by blackhole status 'upcoming'",
			criteria: filterCriteria{
				blackholeStatus: "upcoming",
				blackholeDays:   30,
				cursusID:        21,
			},
			wantLen: 1,
			wantIDs: []int{2},
		},
		{
			name: "combined filters",
			criteria: filterCriteria{
				minProjects: 2,
				minLevel:    4.0,
				cursusID:    21,
			},
			wantLen: 2,
			wantIDs: []int{1, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterUsers(users, tt.criteria)
			if len(got) != tt.wantLen {
				t.Errorf("filterUsers() returned %d users, want %d", len(got), tt.wantLen)
			}
			if tt.wantIDs != nil {
				gotIDs := make([]int, len(got))
				for i, u := range got {
					gotIDs[i] = u.ID
				}
				for _, wantID := range tt.wantIDs {
					found := false
					for _, gotID := range gotIDs {
						if gotID == wantID {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("filterUsers() missing expected user ID %d, got IDs: %v", wantID, gotIDs)
					}
				}
			}
		})
	}
}
