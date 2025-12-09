package beancore

import (
	"github.com/hmans/beans/internal/bean"
)

// IncomingLink represents a link from another bean to a target bean.
type IncomingLink struct {
	FromBean *bean.Bean
	LinkType string
}

// BrokenLink represents a link to a non-existent bean.
type BrokenLink struct {
	BeanID   string `json:"bean_id"`
	LinkType string `json:"link_type"`
	Target   string `json:"target"`
}

// SelfLink represents a bean linking to itself.
type SelfLink struct {
	BeanID   string `json:"bean_id"`
	LinkType string `json:"link_type"`
}

// Cycle represents a circular dependency in links.
type Cycle struct {
	LinkType string   `json:"link_type"`
	Path     []string `json:"path"`
}

// LinkCheckResult contains all link validation issues found.
type LinkCheckResult struct {
	BrokenLinks []BrokenLink `json:"broken_links"`
	SelfLinks   []SelfLink   `json:"self_links"`
	Cycles      []Cycle      `json:"cycles"`
}

// HasIssues returns true if any link issues were found.
func (r *LinkCheckResult) HasIssues() bool {
	return len(r.BrokenLinks) > 0 || len(r.SelfLinks) > 0 || len(r.Cycles) > 0
}

// TotalIssues returns the total count of all issues.
func (r *LinkCheckResult) TotalIssues() int {
	return len(r.BrokenLinks) + len(r.SelfLinks) + len(r.Cycles)
}

// FindIncomingLinks returns all beans that link TO the given bean ID.
func (c *Core) FindIncomingLinks(targetID string) []IncomingLink {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []IncomingLink
	for _, b := range c.beans {
		for _, link := range b.Links {
			if link.Target == targetID {
				result = append(result, IncomingLink{
					FromBean: b,
					LinkType: link.Type,
				})
			}
		}
	}
	return result
}

// DetectCycle checks if adding a link from fromID to toID would create a cycle.
// Only checks for blocks and parent link types.
// Returns the cycle path if a cycle would be created, nil otherwise.
func (c *Core) DetectCycle(fromID, linkType, toID string) []string {
	// Only check hierarchical link types
	if linkType != "blocks" && linkType != "parent" {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Build adjacency list for the specific link type
	// Adding edge: fromID -> toID
	// Check if there's already a path from toID back to fromID
	visited := make(map[string]bool)
	path := []string{fromID, toID}

	return c.findPathToTarget(toID, fromID, linkType, visited, path)
}

// findPathToTarget uses DFS to find if there's a path from current to target.
// Returns the path if found, nil otherwise.
func (c *Core) findPathToTarget(current, target, linkType string, visited map[string]bool, path []string) []string {
	if current == target {
		return path
	}

	if visited[current] {
		return nil
	}
	visited[current] = true

	b, ok := c.beans[current]
	if !ok {
		return nil
	}

	for _, link := range b.Links {
		if link.Type != linkType {
			continue
		}
		newPath := append(path, link.Target)
		if result := c.findPathToTarget(link.Target, target, linkType, visited, newPath); result != nil {
			return result
		}
	}

	return nil
}

// CheckAllLinks validates all links across all beans.
func (c *Core) CheckAllLinks() *LinkCheckResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := &LinkCheckResult{
		BrokenLinks: []BrokenLink{},
		SelfLinks:   []SelfLink{},
		Cycles:      []Cycle{},
	}

	// Check for broken links and self-references
	for _, b := range c.beans {
		for _, link := range b.Links {
			// Check for self-reference
			if link.Target == b.ID {
				result.SelfLinks = append(result.SelfLinks, SelfLink{
					BeanID:   b.ID,
					LinkType: link.Type,
				})
				continue
			}

			// Check if target exists
			if _, ok := c.beans[link.Target]; !ok {
				result.BrokenLinks = append(result.BrokenLinks, BrokenLink{
					BeanID:   b.ID,
					LinkType: link.Type,
					Target:   link.Target,
				})
			}
		}
	}

	// Check for cycles in blocks and parent links
	for _, linkType := range []string{"blocks", "parent"} {
		cycles := c.findCycles(linkType)
		result.Cycles = append(result.Cycles, cycles...)
	}

	return result
}

// findCycles detects all cycles for a specific link type using DFS.
func (c *Core) findCycles(linkType string) []Cycle {
	var cycles []Cycle
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	seenCycles := make(map[string]bool) // To avoid duplicate cycle reports

	var dfs func(id string, path []string)
	dfs = func(id string, path []string) {
		if inStack[id] {
			// Found a cycle - find where the cycle starts
			cycleStart := -1
			for i, p := range path {
				if p == id {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cyclePath := append(path[cycleStart:], id)
				// Create a canonical key to avoid duplicate cycles
				key := canonicalCycleKey(cyclePath)
				if !seenCycles[key] {
					seenCycles[key] = true
					cycles = append(cycles, Cycle{
						LinkType: linkType,
						Path:     cyclePath,
					})
				}
			}
			return
		}

		if visited[id] {
			return
		}

		visited[id] = true
		inStack[id] = true

		b, ok := c.beans[id]
		if ok {
			for _, link := range b.Links {
				if link.Type == linkType {
					dfs(link.Target, append(path, id))
				}
			}
		}

		inStack[id] = false
	}

	for id := range c.beans {
		if !visited[id] {
			dfs(id, nil)
		}
	}

	return cycles
}

// canonicalCycleKey creates a unique key for a cycle to detect duplicates.
// It normalizes the cycle by starting from the smallest ID.
func canonicalCycleKey(path []string) string {
	if len(path) <= 1 {
		return ""
	}

	// Remove the duplicate end element (cycle closes back)
	cycle := path[:len(path)-1]

	// Find the minimum element to use as start
	minIdx := 0
	for i, id := range cycle {
		if id < cycle[minIdx] {
			minIdx = i
		}
	}

	// Rotate to start from minimum
	key := ""
	for i := 0; i < len(cycle); i++ {
		idx := (minIdx + i) % len(cycle)
		if i > 0 {
			key += "->"
		}
		key += cycle[idx]
	}

	return key
}

// RemoveLinksTo removes all links pointing to the given target ID from all beans.
// Returns the number of links removed.
func (c *Core) RemoveLinksTo(targetID string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0
	for _, b := range c.beans {
		originalLen := len(b.Links)
		var newLinks bean.Links
		for _, link := range b.Links {
			if link.Target != targetID {
				newLinks = append(newLinks, link)
			}
		}

		if len(newLinks) < originalLen {
			b.Links = newLinks
			if err := c.saveToDisk(b); err != nil {
				return removed, err
			}
			removed += originalLen - len(newLinks)
		}
	}

	return removed, nil
}

// FixBrokenLinks removes all broken links (links to non-existent beans) and self-references.
// Returns the number of issues fixed.
func (c *Core) FixBrokenLinks() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fixed := 0
	for _, b := range c.beans {
		originalLen := len(b.Links)
		var newLinks bean.Links
		for _, link := range b.Links {
			// Skip self-references
			if link.Target == b.ID {
				continue
			}
			// Skip broken links (target doesn't exist)
			if _, ok := c.beans[link.Target]; !ok {
				continue
			}
			newLinks = append(newLinks, link)
		}

		if len(newLinks) < originalLen {
			b.Links = newLinks
			if err := c.saveToDisk(b); err != nil {
				return fixed, err
			}
			fixed += originalLen - len(newLinks)
		}
	}

	return fixed, nil
}
