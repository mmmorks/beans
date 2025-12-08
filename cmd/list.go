package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"hmans.dev/beans/internal/bean"
	"hmans.dev/beans/internal/config"
	"hmans.dev/beans/internal/output"
	"hmans.dev/beans/internal/ui"
)

var (
	listJSON       bool
	listStatus     []string
	listNoStatus   []string
	listType       []string
	listNoType     []string
	listPriority   []string
	listNoPriority []string
	listLinks      []string
	listLinkedAs   []string
	listNoLinks    []string
	listNoLinkedAs []string
	listTag        []string
	listNoTag      []string
	listQuiet      bool
	listSort       string
	listFull       bool
)

// linkFilter represents a pre-parsed link filter criterion.
type linkFilter struct {
	linkType string
	targetID string // empty means "any target"
}

// parseLinkFilters parses filter strings like "blocks" or "blocks:id" into linkFilter structs.
func parseLinkFilters(filters []string) []linkFilter {
	result := make([]linkFilter, len(filters))
	for i, f := range filters {
		parts := strings.SplitN(f, ":", 2)
		result[i].linkType = parts[0]
		if len(parts) == 2 {
			result[i].targetID = parts[1]
		}
	}
	return result
}

// linkIndex holds precomputed data structures for efficient link filtering.
type linkIndex struct {
	byID       map[string]*bean.Bean      // ID -> Bean lookup
	targetedBy map[string]map[string]bool // linkType -> set of target IDs
}

// buildLinkIndex creates a linkIndex from a slice of beans.
// This should be called once before any filtering to capture all relationships.
func buildLinkIndex(beans []*bean.Bean) *linkIndex {
	idx := &linkIndex{
		byID:       make(map[string]*bean.Bean),
		targetedBy: make(map[string]map[string]bool),
	}
	for _, b := range beans {
		idx.byID[b.ID] = b
		for _, link := range b.Links {
			if idx.targetedBy[link.Type] == nil {
				idx.targetedBy[link.Type] = make(map[string]bool)
			}
			idx.targetedBy[link.Type][link.Target] = true
		}
	}
	return idx
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all beans",
	Long:    `Lists all beans in the .beans directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		beans := core.All()

		// Parse filter criteria once (avoid repeated string splitting)
		linksFilters := parseLinkFilters(listLinks)
		linkedAsFilters := parseLinkFilters(listLinkedAs)
		noLinksFilters := parseLinkFilters(listNoLinks)
		noLinkedAsFilters := parseLinkFilters(listNoLinkedAs)

		// Build link index once from all beans (before status filtering)
		// This ensures relationships are captured even if source bean is filtered out
		idx := buildLinkIndex(beans)

		// Apply filters (positive first, then exclusions)
		beans = filterBeans(beans, listStatus)
		beans = excludeByStatus(beans, listNoStatus)
		beans = filterByType(beans, listType)
		beans = excludeByType(beans, listNoType)
		beans = filterByPriority(beans, listPriority)
		beans = excludeByPriority(beans, listNoPriority)
		beans = filterByLinks(beans, linksFilters)
		beans = filterByLinkedAs(beans, linkedAsFilters, idx)
		beans = excludeByLinks(beans, noLinksFilters)
		beans = excludeByLinkedAs(beans, noLinkedAsFilters, idx)
		beans = filterByTags(beans, listTag)
		beans = excludeByTags(beans, listNoTag)

		// Sort beans
		sortBeans(beans, listSort, cfg)

		// JSON output
		if listJSON {
			if !listFull {
				for _, b := range beans {
					b.Body = ""
				}
			}
			return output.SuccessMultiple(beans)
		}

		// Quiet mode: just IDs
		if listQuiet {
			for _, b := range beans {
				fmt.Println(b.ID)
			}
			return nil
		}

		// Human-friendly output
		if len(beans) == 0 {
			fmt.Println(ui.Muted.Render("No beans found. Create one with: beans new <title>"))
			return nil
		}

		// Calculate max ID width
		maxIDWidth := 2 // minimum for "ID" header
		for _, b := range beans {
			if len(b.ID) > maxIDWidth {
				maxIDWidth = len(b.ID)
			}
		}
		maxIDWidth += 2 // padding

		// Check if any beans have tags
		hasTags := false
		for _, b := range beans {
			if len(b.Tags) > 0 {
				hasTags = true
				break
			}
		}

		// Column styles with widths for alignment (order: ID, Type, Status, Tags, Title - matches TUI)
		idStyle := lipgloss.NewStyle().Width(maxIDWidth)
		typeStyle := lipgloss.NewStyle().Width(12)
		statusStyle := lipgloss.NewStyle().Width(14)
		tagsStyle := lipgloss.NewStyle().Width(24)
		titleStyle := lipgloss.NewStyle()

		// Header style
		headerCol := lipgloss.NewStyle().Foreground(ui.ColorMuted)

		// Header
		var header string
		var dividerWidth int
		if hasTags {
			header = lipgloss.JoinHorizontal(lipgloss.Top,
				idStyle.Render(headerCol.Render("ID")),
				typeStyle.Render(headerCol.Render("TYPE")),
				statusStyle.Render(headerCol.Render("STATUS")),
				tagsStyle.Render(headerCol.Render("TAGS")),
				titleStyle.Render(headerCol.Render("TITLE")),
			)
			dividerWidth = maxIDWidth + 12 + 14 + 20 + 30
		} else {
			header = lipgloss.JoinHorizontal(lipgloss.Top,
				idStyle.Render(headerCol.Render("ID")),
				typeStyle.Render(headerCol.Render("TYPE")),
				statusStyle.Render(headerCol.Render("STATUS")),
				titleStyle.Render(headerCol.Render("TITLE")),
			)
			dividerWidth = maxIDWidth + 12 + 14 + 30
		}
		fmt.Println(header)
		fmt.Println(ui.Muted.Render(strings.Repeat("─", dividerWidth)))

		for _, b := range beans {
			// Get status color from config
			statusCfg := cfg.GetStatus(b.Status)
			statusColor := "gray"
			if statusCfg != nil {
				statusColor = statusCfg.Color
			}
			isArchive := cfg.IsArchiveStatus(b.Status)

			// Get type color from config
			typeColor := ""
			if typeCfg := cfg.GetType(b.Type); typeCfg != nil {
				typeColor = typeCfg.Color
			}

			// Get priority color and render symbol
			priorityColor := ""
			if priorityCfg := cfg.GetPriority(b.Priority); priorityCfg != nil {
				priorityColor = priorityCfg.Color
			}
			prioritySymbol := ui.RenderPrioritySymbol(b.Priority, priorityColor)
			if prioritySymbol != "" {
				prioritySymbol += " "
			}

			var row string
			if hasTags {
				tagsStr := ui.RenderTagsCompact(b.Tags, 1)
				row = lipgloss.JoinHorizontal(lipgloss.Top,
					idStyle.Render(ui.ID.Render(b.ID)),
					typeStyle.Render(ui.RenderTypeText(b.Type, typeColor)),
					statusStyle.Render(ui.RenderStatusTextWithColor(b.Status, statusColor, isArchive)),
					tagsStyle.Render(tagsStr),
					titleStyle.Render(prioritySymbol+truncate(b.Title, 50)),
				)
			} else {
				row = lipgloss.JoinHorizontal(lipgloss.Top,
					idStyle.Render(ui.ID.Render(b.ID)),
					typeStyle.Render(ui.RenderTypeText(b.Type, typeColor)),
					statusStyle.Render(ui.RenderStatusTextWithColor(b.Status, statusColor, isArchive)),
					titleStyle.Render(prioritySymbol+truncate(b.Title, 50)),
				)
			}
			fmt.Println(row)
		}

		return nil
	},
}

func sortBeans(beans []*bean.Bean, sortBy string, cfg *config.Config) {
	statusNames := cfg.StatusNames()
	priorityNames := cfg.PriorityNames()
	typeNames := cfg.TypeNames()

	switch sortBy {
	case "created":
		sort.Slice(beans, func(i, j int) bool {
			if beans[i].CreatedAt == nil && beans[j].CreatedAt == nil {
				return beans[i].ID < beans[j].ID
			}
			if beans[i].CreatedAt == nil {
				return false
			}
			if beans[j].CreatedAt == nil {
				return true
			}
			return beans[i].CreatedAt.After(*beans[j].CreatedAt)
		})
	case "updated":
		sort.Slice(beans, func(i, j int) bool {
			if beans[i].UpdatedAt == nil && beans[j].UpdatedAt == nil {
				return beans[i].ID < beans[j].ID
			}
			if beans[i].UpdatedAt == nil {
				return false
			}
			if beans[j].UpdatedAt == nil {
				return true
			}
			return beans[i].UpdatedAt.After(*beans[j].UpdatedAt)
		})
	case "status":
		// Build status order from configured statuses
		statusOrder := make(map[string]int)
		for i, s := range statusNames {
			statusOrder[s] = i
		}
		sort.Slice(beans, func(i, j int) bool {
			oi, oj := statusOrder[beans[i].Status], statusOrder[beans[j].Status]
			if oi != oj {
				return oi < oj
			}
			return beans[i].ID < beans[j].ID
		})
	case "priority":
		// Build priority order from configured priorities
		priorityOrder := make(map[string]int)
		for i, p := range priorityNames {
			priorityOrder[p] = i
		}
		// Find normal priority index for beans without priority
		normalIdx := len(priorityNames)
		for i, p := range priorityNames {
			if p == "normal" {
				normalIdx = i
				break
			}
		}
		sort.Slice(beans, func(i, j int) bool {
			pi := normalIdx
			if beans[i].Priority != "" {
				if order, ok := priorityOrder[beans[i].Priority]; ok {
					pi = order
				}
			}
			pj := normalIdx
			if beans[j].Priority != "" {
				if order, ok := priorityOrder[beans[j].Priority]; ok {
					pj = order
				}
			}
			if pi != pj {
				return pi < pj
			}
			return beans[i].ID < beans[j].ID
		})
	case "id":
		sort.Slice(beans, func(i, j int) bool {
			return beans[i].ID < beans[j].ID
		})
	default:
		// Default: sort by status order, then priority, then type order, then title (same as TUI)
		bean.SortByStatusPriorityAndType(beans, statusNames, priorityNames, typeNames)
	}
}

func filterBeans(beans []*bean.Bean, statuses []string) []*bean.Bean {
	if len(statuses) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		// Filter by status
		matched := false
		for _, s := range statuses {
			if b.Status == s {
				matched = true
				break
			}
		}
		if matched {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

// excludeByStatus excludes beans that match any of the given statuses.
// Inverse of filterBeans: returns beans that DON'T match the criteria.
//
// Examples:
//   - --no-status done returns beans that are not done
//   - --no-status done --no-status archived returns beans that are neither done nor archived
func excludeByStatus(beans []*bean.Bean, statuses []string) []*bean.Bean {
	if len(statuses) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		excluded := false
		for _, s := range statuses {
			if b.Status == s {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

// filterByType filters beans that match any of the given types (OR logic).
func filterByType(beans []*bean.Bean, types []string) []*bean.Bean {
	if len(types) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		for _, t := range types {
			if b.Type == t {
				filtered = append(filtered, b)
				break
			}
		}
	}
	return filtered
}

// excludeByType excludes beans that match any of the given types.
func excludeByType(beans []*bean.Bean, types []string) []*bean.Bean {
	if len(types) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		excluded := false
		for _, t := range types {
			if b.Type == t {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

// filterByPriority filters beans that match any of the given priorities (OR logic).
// Empty priority in the bean matches "normal" filter.
func filterByPriority(beans []*bean.Bean, priorities []string) []*bean.Bean {
	if len(priorities) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		for _, p := range priorities {
			// Match if priority equals filter, or if bean has no priority and filter is "normal"
			if b.Priority == p || (b.Priority == "" && p == "normal") {
				filtered = append(filtered, b)
				break
			}
		}
	}
	return filtered
}

// excludeByPriority excludes beans that match any of the given priorities.
// Empty priority in the bean is treated as "normal" for exclusion purposes.
func excludeByPriority(beans []*bean.Bean, priorities []string) []*bean.Bean {
	if len(priorities) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		excluded := false
		for _, p := range priorities {
			// Exclude if priority equals filter, or if bean has no priority and filter is "normal"
			if b.Priority == p || (b.Priority == "" && p == "normal") {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

// filterByLinks filters beans by outgoing relationship.
// Supports two formats:
//   - "type:id" - Returns beans that have id in their links[type]
//   - "type" - Returns beans that have ANY link of this type
//
// Use repeated flags for multiple values (OR logic).
//
// Examples:
//   - --links blocks:A returns beans that block A
//   - --links blocks returns all beans that block something
//   - --links blocks --links parent returns beans that block something OR have a parent link
func filterByLinks(beans []*bean.Bean, filters []linkFilter) []*bean.Bean {
	if len(filters) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		matched := false
		for _, f := range filters {
			if f.targetID == "" {
				// Type-only: check if this bean has ANY link of this type
				if b.Links.HasType(f.linkType) {
					matched = true
				}
			} else {
				// Type:ID: check if this bean links to the specific target
				if b.Links.HasLink(f.linkType, f.targetID) {
					matched = true
				}
			}

			if matched {
				break
			}
		}
		if matched {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

// filterByLinkedAs filters beans by their outgoing links to a specific target.
// Supports two formats:
//   - "type:id" - Returns beans that have a link of this type pointing to id
//   - "type" - Returns beans that are targeted by ANY bean with this link type
//
// Use repeated flags for multiple values (OR logic).
//
// Examples:
//   - --linked-as parent:milestone-id returns beans that have milestone-id as parent
//   - --linked-as blocks returns all beans that are blocked by something
//   - --linked-as blocks --linked-as parent returns beans that are blocked OR have a parent
func filterByLinkedAs(beans []*bean.Bean, filters []linkFilter, idx *linkIndex) []*bean.Bean {
	if len(filters) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		matched := false
		for _, f := range filters {
			if f.targetID == "" {
				// Type-only: check if this bean is targeted by ANY bean with this link type
				if targets, ok := idx.targetedBy[f.linkType]; ok && targets[b.ID] {
					matched = true
				}
			} else {
				// Type:ID: check if this bean has a link of this type pointing to the target
				// e.g., --linked-as parent:milestone-id finds beans with parent: milestone-id
				if b.Links.HasLink(f.linkType, f.targetID) {
					matched = true
				}
			}

			if matched {
				break
			}
		}
		if matched {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

// excludeByLinks excludes beans by outgoing relationship.
// Inverse of filterByLinks: returns beans that DON'T match the criteria.
//
// Examples:
//   - --no-links blocks returns beans that don't block anything
//   - --no-links parent returns beans without a parent link
func excludeByLinks(beans []*bean.Bean, filters []linkFilter) []*bean.Bean {
	if len(filters) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		excluded := false
		for _, f := range filters {
			if f.targetID == "" {
				// Type-only: exclude if this bean has ANY link of this type
				if b.Links.HasType(f.linkType) {
					excluded = true
				}
			} else {
				// Type:ID: exclude if this bean links to the specific target
				if b.Links.HasLink(f.linkType, f.targetID) {
					excluded = true
				}
			}

			if excluded {
				break
			}
		}
		if !excluded {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

// excludeByLinkedAs excludes beans by their outgoing links to a specific target.
// Inverse of filterByLinkedAs: returns beans that DON'T match the criteria.
//
// Examples:
//   - --no-linked-as blocks returns beans not blocked by anything (actionable work)
//   - --no-linked-as parent:milestone-id returns beans that don't have milestone-id as parent
func excludeByLinkedAs(beans []*bean.Bean, filters []linkFilter, idx *linkIndex) []*bean.Bean {
	if len(filters) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		excluded := false
		for _, f := range filters {
			if f.targetID == "" {
				// Type-only: exclude if this bean is targeted by ANY bean with this link type
				if targets, ok := idx.targetedBy[f.linkType]; ok && targets[b.ID] {
					excluded = true
				}
			} else {
				// Type:ID: exclude if this bean has a link of this type pointing to the target
				// e.g., --no-linked-as parent:milestone-id excludes beans with parent: milestone-id
				if b.Links.HasLink(f.linkType, f.targetID) {
					excluded = true
				}
			}

			if excluded {
				break
			}
		}
		if !excluded {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// filterByTags filters beans that have ANY of the specified tags (OR logic).
func filterByTags(beans []*bean.Bean, tags []string) []*bean.Bean {
	if len(tags) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		for _, tag := range tags {
			if b.HasTag(tag) {
				filtered = append(filtered, b)
				break
			}
		}
	}
	return filtered
}

// excludeByTags excludes beans that have ANY of the specified tags.
func excludeByTags(beans []*bean.Bean, tags []string) []*bean.Bean {
	if len(tags) == 0 {
		return beans
	}

	var filtered []*bean.Bean
	for _, b := range beans {
		excluded := false
		for _, tag := range tags {
			if b.HasTag(tag) {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	listCmd.Flags().StringArrayVarP(&listStatus, "status", "s", nil, "Filter by status (can be repeated)")
	listCmd.Flags().StringArrayVar(&listNoStatus, "no-status", nil, "Exclude by status (can be repeated)")
	listCmd.Flags().StringArrayVarP(&listType, "type", "t", nil, "Filter by type (can be repeated)")
	listCmd.Flags().StringArrayVar(&listNoType, "no-type", nil, "Exclude by type (can be repeated)")
	listCmd.Flags().StringArrayVarP(&listPriority, "priority", "p", nil, "Filter by priority (can be repeated)")
	listCmd.Flags().StringArrayVar(&listNoPriority, "no-priority", nil, "Exclude by priority (can be repeated)")
	listCmd.Flags().StringArrayVar(&listLinks, "links", nil, "Filter by outgoing relationship (format: type or type:id)")
	listCmd.Flags().StringArrayVar(&listLinkedAs, "linked-as", nil, "Filter by incoming relationship (format: type or type:id)")
	listCmd.Flags().StringArrayVar(&listNoLinks, "no-links", nil, "Exclude beans with outgoing relationship (format: type or type:id)")
	listCmd.Flags().StringArrayVar(&listNoLinkedAs, "no-linked-as", nil, "Exclude beans with incoming relationship (format: type or type:id)")
	listCmd.Flags().StringArrayVar(&listTag, "tag", nil, "Filter by tag (can be repeated, OR logic)")
	listCmd.Flags().StringArrayVar(&listNoTag, "no-tag", nil, "Exclude beans with tag (can be repeated)")
	listCmd.Flags().BoolVarP(&listQuiet, "quiet", "q", false, "Only output IDs (one per line)")
	listCmd.Flags().StringVar(&listSort, "sort", "", "Sort by: created, updated, status, priority, id (default: status, priority, type, title)")
	listCmd.Flags().BoolVar(&listFull, "full", false, "Include bean body in JSON output")
	listCmd.MarkFlagsMutuallyExclusive("json", "quiet")
	rootCmd.AddCommand(listCmd)
}
