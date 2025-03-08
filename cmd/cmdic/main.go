package main

import (
	"cmdic/internal/cli"
	"cmdic/internal/dic"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: cmdic <keyword>")
		os.Exit(1)
	}

	keyword := os.Args[1]

	searchURL := fmt.Sprintf("%ssearch.do?q=%s&dic=%s", dic.DAUM_DICT_HOST, url.QueryEscape(keyword), dic.LANG)

	body, err := dic.FetchURL(searchURL)
	if err != nil {
		log.Fatalf("Error fetching search result: %v", err)
	}

	meanings, wordId, err := dic.Parse(body)
	if err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}
	fmt.Println(meanings)
	if meanings == "No results found." && wordId == "" {
		return
	}

	detailedURL := fmt.Sprintf("https://dic.daum.net/word/view.do?wordid=%s", wordId)

	p := tea.NewProgram(cli.NewModel(wordId, detailedURL))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
