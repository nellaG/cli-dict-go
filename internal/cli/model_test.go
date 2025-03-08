package cli_test

import (
	"cmdic/internal/cli"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func newTestModel() cli.Model {
	return cli.NewModel("TESTID", "https://examplecmdic.com/detailed")
}

func TestModelView(t *testing.T) {
	m := newTestModel()
	view := m.View()

	assert.Contains(t, view, "antonym(a)")
	assert.Contains(t, view, "example sentences(e)")
	assert.Contains(t, view, "synonym(s)")
	assert.Contains(t, view, "quit(q)")
}

func TestModelUpdate_UnknownCommand(t *testing.T) {
	m := newTestModel()

	var updated tea.Model
	for _, r := range "unknown" {
		updated, _ = m.Update(tea.KeyMsg{Runes: []rune{r}})
		m = updated.(cli.Model)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(cli.Model)
	viewOutput := m.View()
	assert.Contains(t, viewOutput, "Error Occured")
}
