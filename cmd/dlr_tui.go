package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cwind-cli/cwind/internal/conductors"
)

type dlrTUIModel struct {
	viewport viewport.Model
	report   string
	cond     conductors.Conductor
	trace    string
	adminMin float64
	adminMax float64
	width    int
	height   int
	cursor   int
	action   string
}

var dlrActions = []string{"Exportar reporte (CSV/JSON)", "Finalizar sesión"}

func newDLRTUIModel(report string, cond conductors.Conductor, trace string, adminMin, adminMax float64) dlrTUIModel {
	v := viewport.New(1, 1)
	v.SetContent(report)
	return dlrTUIModel{
		viewport: v,
		report:   report,
		cond:     cond,
		trace:    trace,
		adminMin: adminMin,
		adminMax: adminMax,
	}
}

func (m dlrTUIModel) Init() tea.Cmd {
	return nil
}

func (m dlrTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resizeViewport()
	case tea.KeyMsg:
		key := msg.String()
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			key = string(msg.Runes)
		}
		switch key {
		case "ctrl+c", "q", "esc":
			m.action = "finalizar"
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(dlrActions)-1 {
				m.cursor++
			}
		case "enter":
			m.action = []string{"exportar", "finalizar"}[m.cursor]
			return m, tea.Quit
		case "pgup":
			m.viewport.PageUp()
		case "pgdown":
			m.viewport.PageDown()
		case "home", "g":
			m.viewport.GotoTop()
		case "end", "G":
			m.viewport.GotoBottom()
		default:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.viewport.LineUp(3)
		case tea.MouseButtonWheelDown:
			m.viewport.LineDown(3)
		case tea.MouseButtonLeft:
			if msg.Action == tea.MouseActionPress && msg.X < m.leftWidth() {
				actionY := 9
				if msg.Y >= actionY && msg.Y < actionY+len(dlrActions) {
					m.cursor = msg.Y - actionY
				}
			}
		}
	}
	return m, nil
}

func (m *dlrTUIModel) resizeViewport() {
	rightWidth := m.width - m.leftWidth() - 1
	if rightWidth < 1 {
		rightWidth = 1
	}
	height := m.height - 4
	if height < 1 {
		height = 1
	}
	m.viewport.Width = rightWidth - 4
	if m.viewport.Width < 1 {
		m.viewport.Width = 1
	}
	m.viewport.Height = height
}

func (m dlrTUIModel) leftWidth() int {
	if m.width < 80 {
		return 40
	}
	width := m.width * 36 / 100
	if width < 44 {
		return 44
	}
	return width
}

func (m dlrTUIModel) View() string {
	if m.width == 0 {
		return ""
	}
	leftWidth := m.leftWidth()
	rightWidth := m.width - leftWidth - 1
	if rightWidth < 1 {
		rightWidth = 1
	}
	left := m.leftView()
	right := lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.height).
		Padding(1, 1).
		Render(m.viewport.View())
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#394552")).
		Render(strings.TrimSuffix(strings.Repeat("│\n", m.height), "\n"))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, divider, right)
}

func (m dlrTUIModel) leftView() string {
	lines := []string{
		styleBannerTitle.Render("cwind / DLR"),
		styleBannerMuted.Render("PANEL DE CONTROL"),
		"",
		truncate(fmt.Sprintf("Conductor: %s (%.0f A)", m.cond.Name, m.cond.StaticRating), m.leftWidth()-4),
		"",
		styleAddSub.Render("Traza"),
		truncate(m.trace, m.leftWidth()-6),
		"",
		styleAddSub.Render("Límites administrativos"),
		fmt.Sprintf("mín: %.0f A", m.adminMin),
		fmt.Sprintf("máx: %.0f A", m.adminMax),
		"",
		styleAddSub.Render("Acciones"),
	}
	for i, action := range dlrActions {
		prefix := "  "
		if i == m.cursor {
			prefix = "▸ "
			action = stylePrompt.Render(action)
		}
		lines = append(lines, prefix+action)
	}
	lines = append(lines, "",
		styleBannerMuted.Render("↑↓/j k seleccionar"),
		styleBannerMuted.Render("Enter confirmar"),
		styleBannerMuted.Render("rueda/PgUp/PgDn reporte"),
		styleBannerMuted.Render("Esc salir"))
	return lipgloss.NewStyle().
		Width(m.leftWidth()).
		Height(m.height).
		Padding(1, 2).
		Render(strings.Join(lines, "\n"))
}

func runDLRTUI(out io.Writer, report string, cond conductors.Conductor, trace string, adminMin, adminMax float64) (string, error) {
	file, ok := out.(*os.File)
	if !ok || !termIsInteractive(file) || !termIsInteractive(os.Stdin) || terminalWidth(file) < 80 {
		_, err := io.WriteString(out, report)
		return "finalizar", err
	}

	model := newDLRTUIModel(report, cond, trace, adminMin, adminMax)
	program := tea.NewProgram(model, tea.WithOutput(out), tea.WithInput(os.Stdin), tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := program.Run()
	if err != nil {
		return "", err
	}
	result, ok := finalModel.(dlrTUIModel)
	if !ok || result.action == "" {
		return "finalizar", nil
	}
	return result.action, nil
}

func termIsInteractive(file *os.File) bool {
	if file == nil {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
