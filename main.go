package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
)

const (
	DAUM_DICT_HOST = "https://dic.daum.net/"
	LANG           = "eng"
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

func parse(htmlStr string) (string, string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		return "", "", err
	}

	content, exists := doc.Find("meta[property='og:description']").Attr("content")
	if !exists || content == "" {
		return "No results found.", "", nil
	}

	var wordId string

	if metaRefresh := doc.Find("meta[http-equiv='Refresh']"); metaRefresh.Length() > 0 {
		refreshContent, _ := metaRefresh.Attr("content")
		parts := strings.Split(refreshContent, "URL=")
		if len(parts) > 1 {
			redirURL := parts[1]
			u, err := url.Parse(redirURL)
			if err == nil {
				q := u.Query()
				wordId = q.Get("wordid")
			}
		}
	} else {
		if href, exists := doc.Find("a[txt_cleansch]").First().Attr("href"); exists {
			u, err := url.Parse(href)
			if err == nil {
				q := u.Query()
				wordId = q.Get("wordid")
			}
		}
	}
	return content, wordId, nil
}

func parseDetail(htmlStr, wordId, category string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		return "", err
	}

	idSet := map[string]string{
		"antonym": "OPPOSITE_WORD",
		"synonym": "SIMILAR_WORD",
	}

	if id, ok := idSet[category]; ok {
		words := doc.Find("#" + id)
		if words.Length() == 0 {
			return "No results found.", nil
		}

		var results []string
		words.Find("li").Each(
			func(i int, s *goquery.Selection) {
				aText := s.Find("a").Text()
				spanText := s.Find("span").Text()
				results = append(results, fmt.Sprintf("%s: %s", aText, spanText))
			})
		return strings.Join(results, "\n"), nil
	}
	return "", nil
}

type model struct {
	textInput    textinput.Model
	output       string
	wordId       string
	detailedURL  string
	detailedText string
}

func initialModel(wordId, detailedURL string) model {
	ti := textinput.New()
	ti.Placeholder = "명령어 입력 (a/e/s/q)"
	ti.Focus()
	ti.CharLimit = 10
	ti.Width = 30

	return model{
		textInput:   ti,
		output:      "",
		wordId:      wordId,
		detailedURL: detailedURL,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				result, err := parseDetail(m.detailedText, m.wordId, cmd)
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

func (m model) View() string {
	return fmt.Sprintf("%s\n\n%s\n\n%s", COMMANDS, m.textInput.View(), m.output)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: cmdic <keyword>")
		os.Exit(1)
	}

	keyword := os.Args[1]

	searchURL := fmt.Sprintf("%ssearch.do?q=%s&dic=%s", DAUM_DICT_HOST, url.QueryEscape(keyword), LANG)
	resp, err := http.Get(searchURL)
	if err != nil {
		log.Fatalf("Error fetching search result: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}
	bodyStr := string(bodyBytes)
	meanings, wordId, err := parse(bodyStr)
	if err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}
	fmt.Println(meanings)
	if meanings == "No results found." && wordId == "" {
		return
	}

	detailedURL := fmt.Sprintf("https://dic.daum.net/word/view.do?wordid=%s", wordId)

	p := tea.NewProgram(initialModel(wordId, detailedURL))
	if err := p.Start(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
