package cmd

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/hmans/beans/internal/config"
)

func TestPromptTemplateIsValid(t *testing.T) {
	_, err := template.New("prompt").Parse(agentPromptTemplate)
	if err != nil {
		t.Fatalf("failed to parse prompt template: %v", err)
	}
}

func TestPromptMinimalTemplateIsValid(t *testing.T) {
	_, err := template.New("prompt_minimal").Parse(agentPromptMinimalTemplate)
	if err != nil {
		t.Fatalf("failed to parse minimal prompt template: %v", err)
	}
}

func TestPromptTemplateRenders(t *testing.T) {
	tmpl, err := template.New("prompt").Parse(agentPromptTemplate)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := promptData{
		Types:      config.DefaultTypes,
		Statuses:   config.DefaultStatuses,
		Priorities: config.DefaultPriorities,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()

	// Check for key sections
	requiredSections := []string{
		"Core Rules",
		"Session Close Protocol",
		"Finding Work",
		"Creating & Updating",
		"Git Integration",
		"Common Workflows",
	}

	for _, section := range requiredSections {
		if !strings.Contains(output, section) {
			t.Errorf("output missing required section: %s", section)
		}
	}

	// Check that types are rendered
	for _, typ := range config.DefaultTypes {
		if !strings.Contains(output, typ.Name) {
			t.Errorf("output missing type: %s", typ.Name)
		}
	}

	// Check that statuses are rendered
	for _, status := range config.DefaultStatuses {
		if !strings.Contains(output, status.Name) {
			t.Errorf("output missing status: %s", status.Name)
		}
	}
}

func TestPromptMinimalTemplateRenders(t *testing.T) {
	tmpl, err := template.New("prompt_minimal").Parse(agentPromptMinimalTemplate)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := promptData{
		Types:      config.DefaultTypes,
		Statuses:   config.DefaultStatuses,
		Priorities: config.DefaultPriorities,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()

	// Check for key content
	if !strings.Contains(output, "Core Rules") {
		t.Error("minimal output missing Core Rules")
	}
	if !strings.Contains(output, "beans list --json --ready") {
		t.Error("minimal output missing ready command")
	}
	if !strings.Contains(output, "beans prime") {
		t.Error("minimal output missing reference to full prime")
	}
}

func TestPromptMinimalIsShorterThanFull(t *testing.T) {
	fullTmpl, _ := template.New("full").Parse(agentPromptTemplate)
	minTmpl, _ := template.New("min").Parse(agentPromptMinimalTemplate)

	data := promptData{
		Types:      config.DefaultTypes,
		Statuses:   config.DefaultStatuses,
		Priorities: config.DefaultPriorities,
	}

	var fullBuf, minBuf bytes.Buffer
	fullTmpl.Execute(&fullBuf, data)
	minTmpl.Execute(&minBuf, data)

	if minBuf.Len() >= fullBuf.Len() {
		t.Errorf("minimal template (%d bytes) should be shorter than full template (%d bytes)",
			minBuf.Len(), fullBuf.Len())
	}

	// Minimal should be significantly shorter (at least 5x shorter)
	if minBuf.Len() > fullBuf.Len()/5 {
		t.Errorf("minimal template (%d bytes) should be at least 5x shorter than full template (%d bytes)",
			minBuf.Len(), fullBuf.Len())
	}
}
