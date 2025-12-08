package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"hmans.dev/beans/internal/bean"
	"hmans.dev/beans/internal/beancore"
)

// resolveContent returns content from a direct value or file flag.
// If value is "-", reads from stdin.
func resolveContent(value, file string) (string, error) {
	if value != "" && file != "" {
		return "", fmt.Errorf("cannot use both --body and --body-file")
	}

	if value == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}

	if value != "" {
		return value, nil
	}

	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("reading file: %w", err)
		}
		return string(data), nil
	}

	return "", nil
}

// parseLink parses a link in the format "type:id".
func parseLink(s string) (linkType, targetID string, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid link format: %q (expected type:id)", s)
	}
	return parts[0], parts[1], nil
}

// isKnownLinkType checks if a link type is recognized.
func isKnownLinkType(linkType string) bool {
	for _, t := range beancore.KnownLinkTypes {
		if t == linkType {
			return true
		}
	}
	return false
}

// applyTags adds tags to a bean, returning an error if any tag is invalid.
func applyTags(b *bean.Bean, tags []string) error {
	for _, tag := range tags {
		if err := b.AddTag(tag); err != nil {
			return err
		}
	}
	return nil
}

// applyLinks adds links to a bean, validating link types and checking target existence.
// Returns warnings for non-existent targets.
func applyLinks(b *bean.Bean, links []string) (warnings []string, err error) {
	for _, link := range links {
		linkType, targetID, err := parseLink(link)
		if err != nil {
			return nil, err
		}
		if !isKnownLinkType(linkType) {
			return nil, fmt.Errorf("unknown link type: %s (must be %s)", linkType, strings.Join(beancore.KnownLinkTypes, ", "))
		}
		// Check for self-reference
		if targetID == b.ID {
			return nil, fmt.Errorf("bean cannot link to itself")
		}
		// Check for cycles in hierarchical link types
		if linkType == "blocks" || linkType == "parent" {
			if cycle := core.DetectCycle(b.ID, linkType, targetID); cycle != nil {
				return nil, fmt.Errorf("cannot add link: would create cycle %s", formatCycle(cycle))
			}
		}
		// Check if target bean exists
		if _, err := core.Get(targetID); err != nil {
			warnings = append(warnings, fmt.Sprintf("target bean '%s' does not exist", targetID))
		}
		b.Links = b.Links.Add(linkType, targetID)
	}
	return warnings, nil
}

// formatCycle formats a cycle path for display.
func formatCycle(path []string) string {
	return strings.Join(path, " → ")
}

// removeLinks removes links from a bean.
func removeLinks(b *bean.Bean, links []string) error {
	for _, link := range links {
		linkType, targetID, err := parseLink(link)
		if err != nil {
			return err
		}
		b.Links = b.Links.Remove(linkType, targetID)
	}
	return nil
}
