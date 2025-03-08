package dic_test

import (
	"cmdic/internal/dic"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse_NoResults(t *testing.T) {
	htmlStr := `<html><head><meta property="og:description" content="" /></head><body></body></html>`

	content, wordId, err := dic.Parse(htmlStr)
	assert.NoError(t, err)
	assert.Equal(t, "No results found.", content)
	assert.Equal(t, "", wordId)
}

func TestParse_ValidContent(t *testing.T) {
	cntParam := "Test Description"
	wordIdParam := "TEST123"

	htmlStr := fmt.Sprintf(`
	<html>
	  <head>
	    <meta property="og:description" content="%s" />
	    <meta http-equiv="Refresh" content="5;URL=https://dic.daum.net/word/view.do?wordid=%s" />
	  </head>
	  <body></body>
	</html>
	`, cntParam, wordIdParam)

	content, wordId, err := dic.Parse(htmlStr)
	assert.NoError(t, err)
	assert.Equal(t, cntParam, content)
	assert.Equal(t, wordIdParam, wordId)
}
