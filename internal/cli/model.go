package cli

import (
	"cmdic/internal/dic"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var CMD_SET = map[string]string{
	"a": "antonym",
	"e": "example sentences",
	"s": "synonym",
	"q": "quit",
}

var COMMANDS string = "more: " + strings.Join([]string{
	"antonym(a)",
	"example sentences(e)",
	"synonym(s)",
	"quit(q)",
}, " | ")

type Model struct {
	textInput    textinput.Model
	output       string
	wordId       string
	detailedURL  string
	exampleURL   string
	detailedText string
	searchTerm   string
}

func NewModel(wordId, detailedURL, exampleURL, searchTerm string) Model {
	ti := textinput.New()
	ti.Placeholder = "명령어 입력 (a/e/s/q)"
	ti.Focus()
	ti.CharLimit = 10
	ti.Width = 30

	return Model{
		textInput:   ti,
		output:      "",
		wordId:      wordId,
		detailedURL: detailedURL,
		exampleURL:  exampleURL,
		searchTerm:  searchTerm,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			input := strings.TrimSpace(m.textInput.Value())
			cmd, ok := CMD_SET[input]
			if !ok {
				m.output = "Error Occured"
			} else if input == "q" {
				return m, tea.Quit
			} else if input == "e" {
				var exampleURL string
				if m.exampleURL != "" {
					exampleURL = m.exampleURL
				} else {
					exampleURL = dic.ExampleURL(m.wordId, 1)
				}
				result, err := dic.ParseExample(exampleURL)
				if err != nil {
					m.output = fmt.Sprintf("Error parsing example: %v", err)
				} else {

					paragraphs := strings.Split(result, "\n\n")
					highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
					for i, para := range paragraphs {
						paragraphs[i] = strings.ReplaceAll(para, m.searchTerm, highlightStyle.Render(m.searchTerm))
					}
					m.output = lipgloss.JoinVertical(lipgloss.Top, paragraphs...)
				}
			} else if input == "a" || input == "s" {
				if m.detailedText == "" {
					resp, err := http.Get(m.detailedURL)
					if err != nil {
						m.output = fmt.Sprintf("error detailed url: %v", err)
						m.textInput.Reset()
						return m, nil
					}
					bytes, err := io.ReadAll(resp.Body)
					resp.Body.Close()
					if err != nil {
						m.output = fmt.Sprintf("error detailed response: %v", err)
						m.textInput.Reset()
						return m, nil
					}
					m.detailedText = string(bytes)
				}
				result, err := dic.ParseDetail(m.detailedText, m.wordId, cmd)
				if err != nil {
					m.output = fmt.Sprintf("Error parsing detail: %v", err)
				} else {
					m.output = fmt.Sprintf("%s:\n%s", cmd, result)
				}
			}
			m.textInput.Reset()
		}
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return fmt.Sprintf("%s\n\n%s\n\n%s", COMMANDS, m.textInput.View(), m.output)
}
