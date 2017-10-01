package utils

import (
	"sort"
	"strings"

	"github.com/jdkato/prose/summarize"
	"github.com/jdkato/prose/tokenize"
)

// Custom type to sort keywords in a message based on frequence
type keyword struct {
	word string
	freq int
}
type keywordArray []keyword

func (s keywordArray) Len() int {
	return len(s)
}

func (s keywordArray) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s keywordArray) Less(i, j int) bool {
	return s[i].freq > s[j].freq
}

// Map of suffix and prefix since we break them at tokenize
var ignoreFilter = map[string]string{
	"search": "@",
	"ignore": "@",
	"silent": "@",
	"quiet":  "@",
}

func ExtractTags(msg string, pct float64) ([]string, []string) {
	doc := summarize.NewDocument(msg)
	words := tokenize.NewTreebankWordTokenizer().Tokenize(doc.Content)

	var tags, keywords []string
	var prevWord string
	var tagMap = make(map[string]byte)

	for _, w := range words {
		// Look for ignore filter and skip bookmarking them
		if p, ok := ignoreFilter[w]; ok && prevWord == p {
			return nil, nil
		}

		// Store the hash tagMap and ignore multiple hashes
		if prevWord == "#" && w != "#" {
			lw := strings.ToLower(w)
			tags = append(tags, lw)
			tagMap[lw] = 1
		}

		prevWord = w
	}

	// TODO: For now, let us index only messages with explicit tagMap words
	if len(tagMap) == 0 {
		return nil, nil
	}

	// Check if we have sufficient keywords with round-off to search this message or enrich it
	moreTags := int(0.5 + (pct*doc.NumWords - float64(len(tagMap))))
	if moreTags > 0 {
		var kw keywordArray
		for k, v := range doc.Keywords() {
			lw := strings.ToLower(k)
			if _, ok := tagMap[lw]; !ok {
				kw = append(kw, keyword{lw, v})
			}
		}
		sort.Sort(kw)

		for _, w := range kw {
			if moreTags == len(keywords) {
				break
			}

			keywords = append(keywords, w.word)
		}
	}

	return tags, keywords
}
