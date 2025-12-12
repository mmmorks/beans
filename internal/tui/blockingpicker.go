package tui

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hmans/beans/internal/bean"
	"github.com/hmans/beans/internal/config"
	"github.com/hmans/beans/internal/graph"
	"github.com/hmans/beans/internal/ui"
)

// blockingToggledMsg is sent when a blocking relationship is toggled
type blockingToggledMsg struct {
	beanID   string // the bean we're modifying
	targetID string // the bean being blocked/unblocked
	added    bool   // true if blocking was added, false if removed
}

// closeBlockingPickerMsg is sent when the blocking picker is cancelled
type closeBlockingPickerMsg struct{}

// openBlockingPickerMsg requests opening the blocking picker for a bean
type openBlockingPickerMsg struct {
	beanID          string
	beanTitle       string
	currentBlocking []string // IDs of beans currently being blocked
}

// blockingItem wraps a bean to implement list.Item for the blocking picker
type blockingItem struct {
	bean       *bean.Bean
	cfg        *config.Config
	isBlocking bool // true if current bean is blocking this one
}

func (i blockingItem) Title() string       { return i.bean.Title }
func (i blockingItem) Description() string { return i.bean.ID }
func (i blockingItem) FilterValue() string { return i.bean.Title + " " + i.bean.ID }

// blockingItemDelegate handles rendering of blocking picker items
type blockingItemDelegate struct {
	cfg *config.Config
}

func (d blockingItemDelegate) Height() int                             { return 1 }
func (d blockingItemDelegate) Spacing() int                            { return 0 }
func (d blockingItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d blockingItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(blockingItem)
	if !ok {
		return
	}

	var cursor string
	if index == m.Index() {
		cursor = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Render("▌") + " "
	} else {
		cursor = "  "
	}

	// Show blocking indicator
	var blockingIndicator string
	if item.isBlocking {
		blockingIndicator = lipgloss.NewStyle().Foreground(ui.ColorDanger).Bold(true).Render("● ") // Red dot for blocking
	} else {
		blockingIndicator = lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("○ ") // Empty circle for not blocking
	}

	// Get colors from config
	colors := d.cfg.GetBeanColors(item.bean.Status, item.bean.Type, item.bean.Priority)

	// Format: [indicator] [type] title (id)
	typeBadge := ui.RenderTypeText(item.bean.Type, colors.TypeColor)
	title := item.bean.Title
	if colors.IsArchive {
		title = ui.Muted.Render(title)
	}
	id := ui.Muted.Render(" (" + item.bean.ID + ")")

	fmt.Fprint(w, cursor+blockingIndicator+typeBadge+" "+title+id)
}

// blockingPickerModel is the model for the blocking picker view
type blockingPickerModel struct {
	list            list.Model
	beanID          string   // the bean we're setting blocking for
	beanTitle       string   // the bean's title
	currentBlocking []string // IDs currently being blocked
	width           int
	height          int
}

func newBlockingPickerModel(beanID, beanTitle string, currentBlocking []string, resolver *graph.Resolver, cfg *config.Config, width, height int) blockingPickerModel {
	// Fetch all beans
	allBeans, _ := resolver.Query().Beans(context.Background(), nil)

	// Create a set of currently blocked IDs for quick lookup
	blockingSet := make(map[string]bool)
	for _, id := range currentBlocking {
		blockingSet[id] = true
	}

	// Filter out the current bean and build items
	var eligibleBeans []*bean.Bean
	for _, b := range allBeans {
		if b.ID != beanID {
			eligibleBeans = append(eligibleBeans, b)
		}
	}

	// Sort by type order, then by title
	typeNames := cfg.TypeNames()
	typeOrder := make(map[string]int)
	for i, t := range typeNames {
		typeOrder[t] = i
	}
	sort.Slice(eligibleBeans, func(i, j int) bool {
		ti, tj := typeOrder[eligibleBeans[i].Type], typeOrder[eligibleBeans[j].Type]
		if ti != tj {
			return ti < tj
		}
		return strings.ToLower(eligibleBeans[i].Title) < strings.ToLower(eligibleBeans[j].Title)
	})

	delegate := blockingItemDelegate{cfg: cfg}

	// Build items list
	items := make([]list.Item, 0, len(eligibleBeans))
	for _, b := range eligibleBeans {
		items = append(items, blockingItem{
			bean:       b,
			cfg:        cfg,
			isBlocking: blockingSet[b.ID],
		})
	}

	// Calculate modal dimensions (60% width, 60% height, with min/max constraints)
	modalWidth := max(40, min(80, width*60/100))
	modalHeight := max(10, min(20, height*60/100))
	listWidth := modalWidth - 6
	listHeight := modalHeight - 7

	l := list.New(items, delegate, listWidth, listHeight)
	l.Title = "Manage Blocking"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles.Title = listTitleStyle
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 0, 0, 0)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(ui.ColorPrimary)

	return blockingPickerModel{
		list:            l,
		beanID:          beanID,
		beanTitle:       beanTitle,
		currentBlocking: currentBlocking,
		width:           width,
		height:          height,
	}
}

func (m blockingPickerModel) Init() tea.Cmd {
	return nil
}

func (m blockingPickerModel) Update(msg tea.Msg) (blockingPickerModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		modalWidth := max(40, min(80, msg.Width*60/100))
		modalHeight := max(10, min(20, msg.Height*60/100))
		listWidth := modalWidth - 6
		listHeight := modalHeight - 7
		m.list.SetSize(listWidth, listHeight)

	case tea.KeyMsg:
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "enter":
				if item, ok := m.list.SelectedItem().(blockingItem); ok {
					// Toggle the blocking relationship
					return m, func() tea.Msg {
						return blockingToggledMsg{
							beanID:   m.beanID,
							targetID: item.bean.ID,
							added:    !item.isBlocking, // toggle
						}
					}
				}
			case "esc", "backspace":
				return m, func() tea.Msg {
					return closeBlockingPickerMsg{}
				}
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m blockingPickerModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	return renderPickerModal(pickerModalConfig{
		Title:       "Manage Blocking",
		BeanTitle:   m.beanTitle,
		BeanID:      m.beanID,
		ListContent: m.list.View(),
		Description: "● = blocking, ○ = not blocking",
		Width:       m.width,
		WidthPct:    60,
		MaxWidth:    80,
	})
}

// ModalView returns the picker rendered as a centered modal overlay on top of the background
func (m blockingPickerModel) ModalView(bgView string, fullWidth, fullHeight int) string {
	modal := m.View()
	return overlayModal(bgView, modal, fullWidth, fullHeight)
}
