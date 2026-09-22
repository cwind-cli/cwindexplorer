package cmd

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cwind-cli/cwind/internal/conductors"
)

func TestDLRTUIModelKeepsReportInRightViewport(t *testing.T) {
	report := "reporte único\nlínea 2"
	model := newDLRTUIModel(report, conductors.Conductor{Name: "Drake"}, "linea.csv", 580, 600)
	model.width, model.height = 120, 30
	model.resizeViewport()

	view := model.View()
	if !strings.Contains(view, "reporte único") || !strings.Contains(view, "línea 2") {
		t.Fatalf("viewport report missing from view: %q", view)
	}

	if strings.Contains(view[:strings.Index(view, "│")], report) {
		t.Fatalf("report was duplicated in controls panel")
	}
	if !strings.Contains(view, "Exportar reporte") || !strings.Contains(view, "Finalizar sesión") {
		t.Fatalf("actions missing from controls panel: %q", view)
	}
}

func TestDLRTUIModelRendersStableDividerAtMinimumWidth(t *testing.T) {
	model := newDLRTUIModel("reporte", conductors.Conductor{Name: "Drake"}, "linea.csv", 0, 0)
	model.width, model.height = 80, 20
	model.resizeViewport()

	view := model.View()
	if !strings.Contains(view, "reporte") {
		t.Fatalf("report missing at minimum dashboard width: %q", view)
	}
	if strings.Count(view, "│") < model.height {
		t.Fatalf("divider has fewer than %d rows: got %d", model.height, strings.Count(view, "│"))
	}
}

func TestDLRTUIModelSelectsActionAndScrolls(t *testing.T) {
	model := newDLRTUIModel(strings.Repeat("línea\n", 50), conductors.Conductor{Name: "Drake"}, "linea.csv", 0, 0)
	model.width, model.height = 100, 20
	model.resizeViewport()

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = next.(dlrTUIModel)
	if model.cursor != 1 {
		t.Fatalf("expected second action selected, got %d", model.cursor)
	}

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = next.(dlrTUIModel)
	if model.action != "finalizar" || cmd == nil {
		t.Fatalf("expected finalization action and quit command, got %q", model.action)
	}

	model.viewport.GotoTop()
	next, _ = model.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	model = next.(dlrTUIModel)
	if model.viewport.YOffset == 0 {
		t.Fatal("mouse wheel did not advance report viewport")
	}
}
