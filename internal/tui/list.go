package tui

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hmans.dev/beans/internal/bean"
	"hmans.dev/beans/internal/config"
	"hmans.dev/beans/internal/ui"
)

// beanItem wraps a Bean to implement list.Item
type beanItem struct {
	bean *bean.Bean
	cfg  *config.Config
}

func (i beanItem) Title() string       { return i.bean.Title }
func (i beanItem) Description() string { return i.bean.ID + " · " + i.bean.Status }
func (i beanItem) FilterValue() string { return i.bean.Title + " " + i.bean.ID }

// itemDelegate handles rendering of list items
type itemDelegate struct {
	cfg *config.Config
}

func newItemDelegate(cfg *config.Config) itemDelegate {
	return itemDelegate{cfg: cfg}
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(beanItem)
	if !ok {
		return
	}

	// Get status color from config
	statusCfg := d.cfg.GetStatus(item.bean.Status)
	statusColor := "gray"
	if statusCfg != nil {
		statusColor = statusCfg.Color
	}
	isArchive := d.cfg.IsArchiveStatus(item.bean.Status)

	// Column widths
	idWidth := 12
	statusWidth := 14

	// Build columns
	id := ui.ID.Render(item.bean.ID)
	idCol := lipgloss.NewStyle().Width(idWidth).Render(id)

	status := ui.RenderStatusTextWithColor(item.bean.Status, statusColor, isArchive)
	statusCol := lipgloss.NewStyle().Width(statusWidth).Render(status)

	// Title (truncate if needed)
	title := item.bean.Title
	maxTitleWidth := m.Width() - idWidth - statusWidth - 4
	if maxTitleWidth > 0 && len(title) > maxTitleWidth {
		title = title[:maxTitleWidth-3] + "..."
	}

	isSelected := index == m.Index()

	var str string
	if isSelected {
		cursor := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Render("▌")
		titleStyled := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary).Render(title)
		str = cursor + " " + idCol + statusCol + titleStyled
	} else {
		titleStyled := lipgloss.NewStyle().Render(title)
		str = "  " + idCol + statusCol + titleStyled
	}

	fmt.Fprint(w, str)
}

// listModel is the model for the bean list view
type listModel struct {
	list   list.Model
	store  *bean.Store
	config *config.Config
	width  int
	height int
	err    error
}

func newListModel(store *bean.Store, cfg *config.Config) listModel {
	delegate := newItemDelegate(cfg)

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Beans"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = listTitleStyle
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 2)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(ui.ColorPrimary)

	return listModel{
		list:   l,
		store:  store,
		config: cfg,
	}
}

// beansLoadedMsg is sent when beans are loaded
type beansLoadedMsg struct {
	beans []*bean.Bean
}

// errMsg is sent when an error occurs
type errMsg struct {
	err error
}

// selectBeanMsg is sent when a bean is selected
type selectBeanMsg struct {
	bean *bean.Bean
}

func (m listModel) Init() tea.Cmd {
	return m.loadBeans
}

func (m listModel) loadBeans() tea.Msg {
	beans, err := m.store.FindAll()
	if err != nil {
		return errMsg{err}
	}
	return beansLoadedMsg{beans}
}

func (m listModel) Update(msg tea.Msg) (listModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Reserve space for border and footer
		m.list.SetSize(msg.Width-2, msg.Height-4)

	case beansLoadedMsg:
		sortBeans(msg.beans, m.config.StatusNames())
		items := make([]list.Item, len(msg.beans))
		for i, b := range msg.beans {
			items[i] = beanItem{bean: b, cfg: m.config}
		}
		m.list.SetItems(items)
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "enter":
				if item, ok := m.list.SelectedItem().(beanItem); ok {
					return m, func() tea.Msg {
						return selectBeanMsg{bean: item.bean}
					}
				}
			}
		}
	}

	// Always forward to the list component
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	if m.width == 0 {
		return "Loading..."
	}

	// Simple bordered container
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorMuted).
		Width(m.width - 2).
		Height(m.height - 4)

	content := border.Render(m.list.View())

	// Footer
	help := helpKeyStyle.Render("enter") + " " + helpStyle.Render("view") + "  " +
		helpKeyStyle.Render("/") + " " + helpStyle.Render("filter") + "  " +
		helpKeyStyle.Render("q") + " " + helpStyle.Render("quit")

	return content + "\n" + help
}

// sortBeans sorts beans by status order
func sortBeans(beans []*bean.Bean, statusNames []string) {
	statusOrder := make(map[string]int)
	for i, s := range statusNames {
		statusOrder[s] = i
	}
	sort.Slice(beans, func(i, j int) bool {
		oi, oj := statusOrder[beans[i].Status], statusOrder[beans[j].Status]
		if oi != oj {
			return oi < oj
		}
		return strings.ToLower(beans[i].Title) < strings.ToLower(beans[j].Title)
	})
}
