package dic

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	DAUM_DICT_HOST = "https://dic.daum.net/"
	LANG           = "eng"
)

func Parse(htmlStr string) (string, string, error) {
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

func ParseDetail(htmlStr, wordId, category string) (string, error) {
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

func ParseExample(exampleURL string) (string, error) {
	resp, err := http.Get(exampleURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch example URL: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to create document: %w", err)
	}

	var sentences []string
	doc.Find("li").Each(
		func(i int, s *goquery.Selection) {
			engPhrase := s.Find("span.txt_example").Text()
			if idx := strings.Index(engPhrase, "\n"); idx != -1 {
				engPhrase = engPhrase[:idx]
			}
			meanPhrase := s.Find("span.mean_example").Text()
			phraseSet := fmt.Sprintf("%s\n -> %s\n\n", engPhrase, meanPhrase)
			sentences = append(sentences, phraseSet)
		})
	return strings.Join(sentences, ""), nil
}

func FetchURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL %s: %w", url, err)
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	return string(bytes), nil
}

func ExampleURL(wordId string, page int) string {
	exampleHost := DAUM_DICT_HOST + "word/view_example_more.do"
	q := url.Values{}
	q.Set("wordid", wordId)
	q.Set("summaryid", "etc")
	q.Set("page", fmt.Sprintf("%d", page))
	return exampleHost + "?" + q.Encode()
}
