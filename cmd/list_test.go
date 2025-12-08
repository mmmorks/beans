package cmd

import (
	"testing"
	"time"

	"hmans.dev/beans/internal/bean"
	"hmans.dev/beans/internal/config"
)

func TestFilterBeans(t *testing.T) {
	// Create test beans
	beans := []*bean.Bean{
		{ID: "a1", Status: "open"},
		{ID: "b2", Status: "in-progress"},
		{ID: "c3", Status: "done"},
		{ID: "d4", Status: "open"},
		{ID: "e5", Status: "in-progress"},
	}

	tests := []struct {
		name      string
		statuses  []string
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "no filter",
			statuses:  nil,
			wantCount: 5,
		},
		{
			name:      "empty filter",
			statuses:  []string{},
			wantCount: 5,
		},
		{
			name:      "filter open",
			statuses:  []string{"open"},
			wantCount: 2,
			wantIDs:   []string{"a1", "d4"},
		},
		{
			name:      "filter in-progress",
			statuses:  []string{"in-progress"},
			wantCount: 2,
			wantIDs:   []string{"b2", "e5"},
		},
		{
			name:      "filter done",
			statuses:  []string{"done"},
			wantCount: 1,
			wantIDs:   []string{"c3"},
		},
		{
			name:      "multiple statuses",
			statuses:  []string{"open", "done"},
			wantCount: 3,
			wantIDs:   []string{"a1", "c3", "d4"},
		},
		{
			name:      "non-existent status",
			statuses:  []string{"invalid"},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterBeans(beans, tt.statuses)

			if len(got) != tt.wantCount {
				t.Errorf("filterBeans() count = %d, want %d", len(got), tt.wantCount)
			}

			if tt.wantIDs != nil {
				gotIDs := make([]string, len(got))
				for i, b := range got {
					gotIDs[i] = b.ID
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
						t.Errorf("filterBeans() missing expected ID %q", wantID)
					}
				}
			}
		})
	}
}

func TestExcludeByStatus(t *testing.T) {
	// Create test beans
	beans := []*bean.Bean{
		{ID: "a1", Status: "open"},
		{ID: "b2", Status: "in-progress"},
		{ID: "c3", Status: "done"},
		{ID: "d4", Status: "open"},
		{ID: "e5", Status: "in-progress"},
	}

	tests := []struct {
		name      string
		statuses  []string
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "no exclusion",
			statuses:  nil,
			wantCount: 5,
		},
		{
			name:      "empty exclusion",
			statuses:  []string{},
			wantCount: 5,
		},
		{
			name:      "exclude done",
			statuses:  []string{"done"},
			wantCount: 4,
			wantIDs:   []string{"a1", "b2", "d4", "e5"},
		},
		{
			name:      "exclude in-progress",
			statuses:  []string{"in-progress"},
			wantCount: 3,
			wantIDs:   []string{"a1", "c3", "d4"},
		},
		{
			name:      "exclude multiple statuses",
			statuses:  []string{"done", "in-progress"},
			wantCount: 2,
			wantIDs:   []string{"a1", "d4"},
		},
		{
			name:      "exclude all",
			statuses:  []string{"open", "in-progress", "done"},
			wantCount: 0,
		},
		{
			name:      "exclude non-existent status",
			statuses:  []string{"invalid"},
			wantCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := excludeByStatus(beans, tt.statuses)

			if len(got) != tt.wantCount {
				t.Errorf("excludeByStatus() count = %d, want %d", len(got), tt.wantCount)
			}

			if tt.wantIDs != nil {
				gotIDs := make([]string, len(got))
				for i, b := range got {
					gotIDs[i] = b.ID
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
						t.Errorf("excludeByStatus() missing expected ID %q", wantID)
					}
				}
			}
		})
	}
}

func TestSortBeans(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)
	evenEarlier := now.Add(-2 * time.Hour)

	// Statuses are now hardcoded, so we just use default config
	testCfg := config.Default()

	t.Run("sort by id", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "c3"},
			{ID: "a1"},
			{ID: "b2"},
		}
		sortBeans(beans, "id", testCfg)

		if beans[0].ID != "a1" || beans[1].ID != "b2" || beans[2].ID != "c3" {
			t.Errorf("sort by id: got [%s, %s, %s], want [a1, b2, c3]",
				beans[0].ID, beans[1].ID, beans[2].ID)
		}
	})

	t.Run("sort by created", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "old", CreatedAt: &evenEarlier},
			{ID: "new", CreatedAt: &now},
			{ID: "mid", CreatedAt: &earlier},
		}
		sortBeans(beans, "created", testCfg)

		// Should be newest first
		if beans[0].ID != "new" || beans[1].ID != "mid" || beans[2].ID != "old" {
			t.Errorf("sort by created: got [%s, %s, %s], want [new, mid, old]",
				beans[0].ID, beans[1].ID, beans[2].ID)
		}
	})

	t.Run("sort by created with nil", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "nil1", CreatedAt: nil},
			{ID: "has", CreatedAt: &now},
			{ID: "nil2", CreatedAt: nil},
		}
		sortBeans(beans, "created", testCfg)

		// Non-nil should come first, then nil sorted by ID
		if beans[0].ID != "has" {
			t.Errorf("sort by created with nil: first should be \"has\", got %q", beans[0].ID)
		}
	})

	t.Run("sort by updated", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "old", UpdatedAt: &evenEarlier},
			{ID: "new", UpdatedAt: &now},
			{ID: "mid", UpdatedAt: &earlier},
		}
		sortBeans(beans, "updated", testCfg)

		// Should be newest first
		if beans[0].ID != "new" || beans[1].ID != "mid" || beans[2].ID != "old" {
			t.Errorf("sort by updated: got [%s, %s, %s], want [new, mid, old]",
				beans[0].ID, beans[1].ID, beans[2].ID)
		}
	})

	t.Run("sort by status", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "c1", Status: "completed"},
			{ID: "t1", Status: "todo"},
			{ID: "i1", Status: "in-progress"},
			{ID: "t2", Status: "todo"},
		}
		sortBeans(beans, "status", testCfg)

		// Should be ordered by status config order (not-ready, ready, in-progress, completed, scrapped), then by ID within same status
		expected := []string{"t1", "t2", "i1", "c1"}
		for i, want := range expected {
			if beans[i].ID != want {
				t.Errorf("sort by status[%d]: got %q, want %q", i, beans[i].ID, want)
			}
		}
	})

	t.Run("default sort (archive status then type)", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "completed-bug", Status: "completed", Type: "bug"},
			{ID: "todo-feature", Status: "todo", Type: "feature"},
			{ID: "todo-task", Status: "todo", Type: "task"},
			{ID: "completed-task", Status: "completed", Type: "task"},
			{ID: "todo-bug", Status: "todo", Type: "bug"},
		}
		sortBeans(beans, "", testCfg)

		// Should be: non-archive first (sorted by type order from DefaultTypes: milestone, epic, bug, feature, task),
		// then archive (sorted by type)
		// DefaultTypes order: milestone, epic, bug, feature, task
		expected := []string{"todo-bug", "todo-feature", "todo-task", "completed-bug", "completed-task"}
		for i, want := range expected {
			if beans[i].ID != want {
				t.Errorf("default sort[%d]: got %q, want %q", i, beans[i].ID, want)
			}
		}
	})
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"very short max", "hello", 4, "h..."},
		{"empty string", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFilterByLinks(t *testing.T) {
	// Create test beans with various link configurations
	beans := []*bean.Bean{
		{ID: "a1", Links: bean.Links{{Type: "blocks", Target: "b2"}}},
		{ID: "b2", Links: bean.Links{{Type: "parent", Target: "epic1"}}},
		{ID: "c3", Links: bean.Links{{Type: "blocks", Target: "a1"}, {Type: "blocks", Target: "b2"}}},
		{ID: "d4", Links: nil}, // no links
		{ID: "e5", Links: bean.Links{{Type: "blocks", Target: "b2"}, {Type: "parent", Target: "epic1"}}},
	}

	tests := []struct {
		name    string
		filter  []string
		wantIDs []string
	}{
		{
			name:    "no filter",
			filter:  nil,
			wantIDs: []string{"a1", "b2", "c3", "d4", "e5"},
		},
		{
			name:    "filter by type only - blocks",
			filter:  []string{"blocks"},
			wantIDs: []string{"a1", "c3", "e5"},
		},
		{
			name:    "filter by type only - parent",
			filter:  []string{"parent"},
			wantIDs: []string{"b2", "e5"},
		},
		{
			name:    "filter by type:id - blocks:b2",
			filter:  []string{"blocks:b2"},
			wantIDs: []string{"a1", "c3", "e5"},
		},
		{
			name:    "filter by type:id - blocks:a1",
			filter:  []string{"blocks:a1"},
			wantIDs: []string{"c3"},
		},
		{
			name:    "multiple filters (OR logic)",
			filter:  []string{"blocks", "parent"},
			wantIDs: []string{"a1", "b2", "c3", "e5"},
		},
		{
			name:    "non-existent link type",
			filter:  []string{"nonexistent"},
			wantIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterByLinks(beans, parseLinkFilters(tt.filter))
			gotIDs := extractIDs(got)

			if !equalStringSlices(gotIDs, tt.wantIDs) {
				t.Errorf("filterByLinks() = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestFilterByLinkedAs(t *testing.T) {
	// Create test beans where some beans link to others
	// - a1 blocks b2
	// - b2 blocks c3
	// - child1 and child2 have epic1 as their parent
	beans := []*bean.Bean{
		{ID: "a1", Links: bean.Links{{Type: "blocks", Target: "b2"}}},
		{ID: "b2", Links: bean.Links{{Type: "blocks", Target: "c3"}}},
		{ID: "c3", Links: nil},
		{ID: "epic1", Links: nil},
		{ID: "child1", Links: bean.Links{{Type: "parent", Target: "epic1"}}},
		{ID: "child2", Links: bean.Links{{Type: "parent", Target: "epic1"}}},
	}

	// Build index once for all tests
	idx := buildLinkIndex(beans)

	tests := []struct {
		name    string
		filter  []string
		wantIDs []string
	}{
		{
			name:    "no filter",
			filter:  nil,
			wantIDs: []string{"a1", "b2", "c3", "epic1", "child1", "child2"},
		},
		{
			name:    "filter by type only - blocks (beans that are blocked by something)",
			filter:  []string{"blocks"},
			wantIDs: []string{"b2", "c3"}, // b2 is blocked by a1, c3 is blocked by b2
		},
		{
			name:    "filter by type:id - parent:epic1 (children of epic1)",
			filter:  []string{"parent:epic1"},
			wantIDs: []string{"child1", "child2"}, // beans with parent: epic1
		},
		{
			name:    "filter by type only - parent (beans that are targets of parent links)",
			filter:  []string{"parent"},
			wantIDs: []string{"epic1"}, // epic1 is targeted by parent links
		},
		{
			name:    "non-existent target bean",
			filter:  []string{"parent:nonexistent"},
			wantIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterByLinkedAs(beans, parseLinkFilters(tt.filter), idx)
			gotIDs := extractIDs(got)

			if !equalStringSlices(gotIDs, tt.wantIDs) {
				t.Errorf("filterByLinkedAs() = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestExcludeByLinks(t *testing.T) {
	// Create test beans with various link configurations
	beans := []*bean.Bean{
		{ID: "a1", Links: bean.Links{{Type: "blocks", Target: "b2"}}},
		{ID: "b2", Links: bean.Links{{Type: "parent", Target: "epic1"}}},
		{ID: "c3", Links: bean.Links{{Type: "blocks", Target: "a1"}, {Type: "blocks", Target: "b2"}}},
		{ID: "d4", Links: nil}, // no links
		{ID: "e5", Links: bean.Links{{Type: "blocks", Target: "b2"}, {Type: "parent", Target: "epic1"}}},
	}

	tests := []struct {
		name    string
		exclude []string
		wantIDs []string
	}{
		{
			name:    "no exclusion",
			exclude: nil,
			wantIDs: []string{"a1", "b2", "c3", "d4", "e5"},
		},
		{
			name:    "exclude by type - blocks (exclude beans that block something)",
			exclude: []string{"blocks"},
			wantIDs: []string{"b2", "d4"}, // only b2 and d4 don't have blocks links
		},
		{
			name:    "exclude by type - parent",
			exclude: []string{"parent"},
			wantIDs: []string{"a1", "c3", "d4"}, // these don't have parent links
		},
		{
			name:    "exclude by type:id - blocks:b2",
			exclude: []string{"blocks:b2"},
			wantIDs: []string{"b2", "d4"}, // a1, c3, e5 all block b2
		},
		{
			name:    "multiple exclusions",
			exclude: []string{"blocks", "parent"},
			wantIDs: []string{"d4"}, // only d4 has neither blocks nor parent
		},
		{
			name:    "non-existent link type",
			exclude: []string{"nonexistent"},
			wantIDs: []string{"a1", "b2", "c3", "d4", "e5"}, // nothing excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := excludeByLinks(beans, parseLinkFilters(tt.exclude))
			gotIDs := extractIDs(got)

			if !equalStringSlices(gotIDs, tt.wantIDs) {
				t.Errorf("excludeByLinks() = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestExcludeByLinkedAs(t *testing.T) {
	// Create test beans where some beans link to others
	// - a1 blocks b2
	// - b2 blocks c3
	// - child1 and child2 have epic1 as their parent
	beans := []*bean.Bean{
		{ID: "a1", Links: bean.Links{{Type: "blocks", Target: "b2"}}},
		{ID: "b2", Links: bean.Links{{Type: "blocks", Target: "c3"}}},
		{ID: "c3", Links: nil},
		{ID: "d4", Links: nil},
		{ID: "epic1", Links: nil},
		{ID: "child1", Links: bean.Links{{Type: "parent", Target: "epic1"}}},
		{ID: "child2", Links: bean.Links{{Type: "parent", Target: "epic1"}}},
	}

	// Build index once for all tests
	idx := buildLinkIndex(beans)

	tests := []struct {
		name    string
		exclude []string
		wantIDs []string
	}{
		{
			name:    "no exclusion",
			exclude: nil,
			wantIDs: []string{"a1", "b2", "c3", "d4", "epic1", "child1", "child2"},
		},
		{
			name:    "exclude blocked beans (actionable work)",
			exclude: []string{"blocks"},
			wantIDs: []string{"a1", "d4", "epic1", "child1", "child2"}, // b2 and c3 are blocked
		},
		{
			name:    "exclude by type:id - parent:epic1 (exclude children of epic1)",
			exclude: []string{"parent:epic1"},
			wantIDs: []string{"a1", "b2", "c3", "d4", "epic1"}, // child1 and child2 are excluded
		},
		{
			name:    "exclude beans that are parent targets",
			exclude: []string{"parent"},
			wantIDs: []string{"a1", "b2", "c3", "d4", "child1", "child2"}, // epic1 is excluded (it's a parent target)
		},
		{
			name:    "non-existent target bean",
			exclude: []string{"parent:nonexistent"},
			wantIDs: []string{"a1", "b2", "c3", "d4", "epic1", "child1", "child2"}, // nothing excluded
		},
		{
			name:    "multiple exclusions",
			exclude: []string{"blocks", "parent"},
			wantIDs: []string{"a1", "d4", "child1", "child2"}, // a1, d4, child1, child2 are neither blocked nor parent targets
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := excludeByLinkedAs(beans, parseLinkFilters(tt.exclude), idx)
			gotIDs := extractIDs(got)

			if !equalStringSlices(gotIDs, tt.wantIDs) {
				t.Errorf("excludeByLinkedAs() = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

// Helper function to extract IDs from beans slice
func extractIDs(beans []*bean.Bean) []string {
	ids := make([]string, len(beans))
	for i, b := range beans {
		ids[i] = b.ID
	}
	return ids
}

// Helper function to compare string slices (order-independent)
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aMap := make(map[string]int)
	for _, s := range a {
		aMap[s]++
	}
	for _, s := range b {
		aMap[s]--
		if aMap[s] < 0 {
			return false
		}
	}
	return true
}
