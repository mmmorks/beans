package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hmans/beans/internal/ui"
)

// openHelpMsg requests opening the help overlay
type openHelpMsg struct{}

// closeHelpMsg is sent when the help overlay is closed
type closeHelpMsg struct{}

// helpOverlayModel displays keyboard shortcuts organized by context
type helpOverlayModel struct {
	width  int
	height int
}

func newHelpOverlayModel(width, height int) helpOverlayModel {
	return helpOverlayModel{
		width:  width,
		height: height,
	}
}

func (m helpOverlayModel) Init() tea.Cmd {
	return nil
}

func (m helpOverlayModel) Update(msg tea.Msg) (helpOverlayModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "?", "esc":
			return m, func() tea.Msg {
				return closeHelpMsg{}
			}
		}
	}

	return m, nil
}

func (m helpOverlayModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Calculate modal dimensions - make it wider than pickers
	modalWidth := max(70, min(90, m.width*70/100))

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.ColorPrimary).
		Render("Keyboard Shortcuts")

	// Helper to create a shortcut line
	shortcut := func(key, desc string) string {
		keyStyle := lipgloss.NewStyle().
			Foreground(ui.ColorPrimary).
			Bold(true).
			Width(20).
			Render(key)
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fff")).
			Render(desc)
		return keyStyle + descStyle
	}

	// Helper to create a section header
	sectionHeader := func(name string) string {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(ui.ColorBlue).
			Render(name)
	}

	var content strings.Builder
	content.WriteString(title + "\n\n")

	// List view section
	content.WriteString(sectionHeader("List View") + "\n")
	content.WriteString(shortcut("j/k, ↓/↑", "Navigate up/down") + "\n")
	content.WriteString(shortcut("enter", "View bean details") + "\n")
	content.WriteString(shortcut("c", "Create new bean") + "\n")
	content.WriteString(shortcut("e", "Edit bean in $EDITOR") + "\n")
	content.WriteString(shortcut("s", "Change status") + "\n")
	content.WriteString(shortcut("t", "Change type") + "\n")
	content.WriteString(shortcut("P", "Change priority") + "\n")
	content.WriteString(shortcut("p", "Set parent") + "\n")
	content.WriteString(shortcut("b", "Manage blocking relationships") + "\n")
	content.WriteString(shortcut("/", "Filter list") + "\n")
	content.WriteString(shortcut("g t", "Go to tags (filter by tag)") + "\n")
	content.WriteString(shortcut("esc", "Clear filter") + "\n")
	content.WriteString(shortcut("q", "Quit") + "\n")
	content.WriteString("\n")

	// Detail view section
	content.WriteString(sectionHeader("Detail View") + "\n")
	content.WriteString(shortcut("j/k, ↓/↑", "Scroll up/down") + "\n")
	content.WriteString(shortcut("tab", "Switch focus (links/body)") + "\n")
	content.WriteString(shortcut("enter", "Navigate to linked bean") + "\n")
	content.WriteString(shortcut("e", "Edit bean in $EDITOR") + "\n")
	content.WriteString(shortcut("s", "Change status") + "\n")
	content.WriteString(shortcut("t", "Change type") + "\n")
	content.WriteString(shortcut("P", "Change priority") + "\n")
	content.WriteString(shortcut("p", "Set parent") + "\n")
	content.WriteString(shortcut("b", "Manage blocking relationships") + "\n")
	content.WriteString(shortcut("esc/backspace", "Back to list/previous bean") + "\n")
	content.WriteString(shortcut("q", "Quit") + "\n")
	content.WriteString("\n")

	// Picker/Dialog section
	content.WriteString(sectionHeader("Pickers & Dialogs") + "\n")
	content.WriteString(shortcut("j/k, ↓/↑", "Navigate up/down") + "\n")
	content.WriteString(shortcut("enter", "Select/confirm") + "\n")
	content.WriteString(shortcut("/", "Filter items") + "\n")
	content.WriteString(shortcut("esc", "Cancel") + "\n")
	content.WriteString("\n")

	// Footer
	footer := helpKeyStyle.Render("?/esc") + " " + helpStyle.Render("close")

	// Border style
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary).
		Padding(1, 2).
		Width(modalWidth)

	return border.Render(content.String() + footer)
}

// ModalView returns the help overlay as a centered modal on top of the background
func (m helpOverlayModel) ModalView(bgView string, fullWidth, fullHeight int) string {
	modal := m.View()
	return overlayModal(bgView, modal, fullWidth, fullHeight)
}
