package cli_test

import (
	"cmdic/internal/cli"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func newTestModel() cli.Model {
	return cli.NewModel("TESTID", "https://examplecmdic.com/detailed", "https://examplecmdic.com/example", "new")
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

func TestModelUpdate_ExampleCommand_Highlight(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html>
			<body>
				<ul>
					<li>
						<span class="txt_example">I bought this book by myself.</span>
						<span class="mean_example">나는 내 스스로 이 책을 샀다.</span>
					</li>
					<li>
						<span class="txt_example">I like this book.</span>
						<span class="mean_example">나는 이 책이 좋다.</span>
					</li>
				</ul>
			</body>
		</html>`)
		}))
	defer testServer.Close()

	m := cli.NewModel("TESTID", testServer.URL, testServer.URL, "book")
	var updated tea.Model
	updated, _ = m.Update(tea.KeyMsg{Runes: []rune("e")})
	m = updated.(cli.Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(cli.Model)
	output := m.View()
	assert.Contains(t, output, "book", "`book` must be included")
	assert.Contains(t, output, "\x1b[", "ANSI escape sequence must be included")
}
