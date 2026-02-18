package cmd

import (
	"testing"
	"time"

	"github.com/naokiiida/t42-cli/internal/api"
)

func TestParseInscriptionRules(t *testing.T) {
	rules := []api.ProjectSessionRule{
		{
			ID:       1,
			Required: true,
			Rule: api.RuleDefinition{
				Kind:         "inscription",
				InternalName: "QuestsValidated",
			},
			Params: []api.ProjectSessionRuleParam{
				{Value: "common-core-rank-05"},
			},
		},
		{
			ID:       2,
			Required: true,
			Rule: api.RuleDefinition{
				Kind:         "inscription",
				InternalName: "QuestsNotValidated",
			},
			Params: []api.ProjectSessionRuleParam{
				{Value: "common-core"},
			},
		},
		{
			ID:       3,
			Required: true,
			Rule: api.RuleDefinition{
				Kind:         "inscription",
				InternalName: "NeitherOngoingOrValidated",
			},
			Params: []api.ProjectSessionRuleParam{
				{Value: "ft_transcendence"},
			},
		},
		{
			// Non-inscription rule should be ignored
			ID:       4,
			Required: true,
			Rule: api.RuleDefinition{
				Kind:         "correction",
				InternalName: "OnSameCampus",
			},
		},
	}

	reqs := parseInscriptionRules(rules)

	if len(reqs.requiredQuests) != 1 || reqs.requiredQuests[0] != "common-core-rank-05" {
		t.Errorf("requiredQuests = %v, want [common-core-rank-05]", reqs.requiredQuests)
	}
	if len(reqs.forbiddenQuests) != 1 || reqs.forbiddenQuests[0] != "common-core" {
		t.Errorf("forbiddenQuests = %v, want [common-core]", reqs.forbiddenQuests)
	}
	if len(reqs.forbiddenProjects) != 1 || reqs.forbiddenProjects[0] != "ft_transcendence" {
		t.Errorf("forbiddenProjects = %v, want [ft_transcendence]", reqs.forbiddenProjects)
	}
}

func TestParseInscriptionRulesEmpty(t *testing.T) {
	reqs := parseInscriptionRules(nil)
	if len(reqs.requiredQuests) != 0 || len(reqs.forbiddenQuests) != 0 || len(reqs.forbiddenProjects) != 0 {
		t.Errorf("expected empty requirements for nil rules, got %+v", reqs)
	}
}

func TestCheckRequiredQuests(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		questUsers []api.QuestUser
		required   []string
		want       bool
	}{
		{
			name:     "no requirements always passes",
			required: nil,
			want:     true,
		},
		{
			name: "all required quests validated",
			questUsers: []api.QuestUser{
				{Quest: api.Quest{Slug: "common-core-rank-05"}, ValidatedAt: &now},
				{Quest: api.Quest{Slug: "exam-rank-05"}, ValidatedAt: &now},
			},
			required: []string{"common-core-rank-05"},
			want:     true,
		},
		{
			name: "required quest not validated",
			questUsers: []api.QuestUser{
				{Quest: api.Quest{Slug: "common-core-rank-03"}, ValidatedAt: &now},
			},
			required: []string{"common-core-rank-05"},
			want:     false,
		},
		{
			name:       "empty quest users fails when requirements exist",
			questUsers: nil,
			required:   []string{"common-core-rank-05"},
			want:       false,
		},
		{
			name: "quest present but not validated (nil ValidatedAt)",
			questUsers: []api.QuestUser{
				{Quest: api.Quest{Slug: "common-core-rank-05"}, ValidatedAt: nil},
			},
			required: []string{"common-core-rank-05"},
			want:     false,
		},
		{
			name: "multiple required quests all validated",
			questUsers: []api.QuestUser{
				{Quest: api.Quest{Slug: "common-core-rank-05"}, ValidatedAt: &now},
				{Quest: api.Quest{Slug: "exam-rank-05"}, ValidatedAt: &now},
			},
			required: []string{"common-core-rank-05", "exam-rank-05"},
			want:     true,
		},
		{
			name: "multiple required quests one missing",
			questUsers: []api.QuestUser{
				{Quest: api.Quest{Slug: "common-core-rank-05"}, ValidatedAt: &now},
			},
			required: []string{"common-core-rank-05", "exam-rank-05"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkRequiredQuests(tt.questUsers, tt.required)
			if got != tt.want {
				t.Errorf("checkRequiredQuests() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckForbiddenQuests(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		questUsers []api.QuestUser
		forbidden  []string
		want       bool
	}{
		{
			name:      "no forbidden quests always passes",
			forbidden: nil,
			want:      true,
		},
		{
			name: "forbidden quest not validated passes",
			questUsers: []api.QuestUser{
				{Quest: api.Quest{Slug: "common-core"}, ValidatedAt: nil},
			},
			forbidden: []string{"common-core"},
			want:      true,
		},
		{
			name: "forbidden quest validated fails",
			questUsers: []api.QuestUser{
				{Quest: api.Quest{Slug: "common-core"}, ValidatedAt: &now},
			},
			forbidden: []string{"common-core"},
			want:      false,
		},
		{
			name:       "empty quest users passes",
			questUsers: nil,
			forbidden:  []string{"common-core"},
			want:       true,
		},
		{
			name: "unrelated quest validated passes",
			questUsers: []api.QuestUser{
				{Quest: api.Quest{Slug: "common-core-rank-05"}, ValidatedAt: &now},
			},
			forbidden: []string{"common-core"},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkForbiddenQuests(tt.questUsers, tt.forbidden)
			if got != tt.want {
				t.Errorf("checkForbiddenQuests() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckForbiddenProjects(t *testing.T) {
	trueVal := true

	tests := []struct {
		name         string
		projectUsers []api.ProjectUser
		forbidden    []string
		want         bool
	}{
		{
			name:      "no forbidden projects always passes",
			forbidden: nil,
			want:      true,
		},
		{
			name:         "empty projects passes",
			projectUsers: nil,
			forbidden:    []string{"ft_transcendence"},
			want:         true,
		},
		{
			name: "forbidden project validated fails",
			projectUsers: []api.ProjectUser{
				{
					Project:   api.Project{Slug: "ft_transcendence"},
					Status:    "finished",
					Validated: &trueVal,
				},
			},
			forbidden: []string{"ft_transcendence"},
			want:      false,
		},
		{
			name: "forbidden project in progress fails",
			projectUsers: []api.ProjectUser{
				{
					Project: api.Project{Slug: "ft_transcendence"},
					Status:  "in_progress",
				},
			},
			forbidden: []string{"ft_transcendence"},
			want:      false,
		},
		{
			name: "forbidden project creating_group fails",
			projectUsers: []api.ProjectUser{
				{
					Project: api.Project{Slug: "ft_transcendence"},
					Status:  "creating_group",
				},
			},
			forbidden: []string{"ft_transcendence"},
			want:      false,
		},
		{
			name: "forbidden project searching_a_group fails",
			projectUsers: []api.ProjectUser{
				{
					Project: api.Project{Slug: "ft_transcendence"},
					Status:  "searching_a_group",
				},
			},
			forbidden: []string{"ft_transcendence"},
			want:      false,
		},
		{
			name: "forbidden project waiting_for_correction fails",
			projectUsers: []api.ProjectUser{
				{
					Project: api.Project{Slug: "ft_transcendence"},
					Status:  "waiting_for_correction",
				},
			},
			forbidden: []string{"ft_transcendence"},
			want:      false,
		},
		{
			name: "unrelated project does not affect result",
			projectUsers: []api.ProjectUser{
				{
					Project:   api.Project{Slug: "libft"},
					Status:    "finished",
					Validated: &trueVal,
				},
			},
			forbidden: []string{"ft_transcendence"},
			want:      true,
		},
		{
			name: "forbidden project failed (not validated, not ongoing) passes",
			projectUsers: []api.ProjectUser{
				{
					Project: api.Project{Slug: "ft_transcendence"},
					Status:  "finished",
					// Validated is nil (failed)
				},
			},
			forbidden: []string{"ft_transcendence"},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkForbiddenProjects(tt.projectUsers, tt.forbidden)
			if got != tt.want {
				t.Errorf("checkForbiddenProjects() = %v, want %v", got, tt.want)
			}
		})
	}
}
