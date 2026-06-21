package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

const (
	bm25K1 = 1.5
	bm25B  = 0.75
)

// Result is a scored document.
type Result struct {
	Doc   Document
	Score float64
}

// Search ranks documents against the query using Okapi BM25.
func (idx *Index) Search(query string, limit int) []Result {
	queryTerms := tokenize(query)
	if len(queryTerms) == 0 || len(idx.Documents) == 0 || idx.AvgLength == 0 {
		return nil
	}

	n := float64(len(idx.Documents))
	results := make([]Result, 0, len(idx.Documents))

	for _, doc := range idx.Documents {
		score := 0.0
		for _, term := range queryTerms {
			tf, ok := doc.TermFreqs[term]
			if !ok {
				continue
			}
			df := idx.DocFreq[term]
			if df == 0 {
				continue
			}
			idf := math.Log(1 + (n-float64(df)+0.5)/(float64(df)+0.5))
			denom := float64(tf) + bm25K1*(1-bm25B+bm25B*float64(doc.Length)/idx.AvgLength)
			score += idf * (float64(tf) * (bm25K1 + 1)) / denom
		}
		if score > 0 {
			results = append(results, Result{Doc: doc, Score: score})
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

func printResults(results []Result) {
	cmd, pkg, reset := colors()
	for i, r := range results {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("%s%s%s  %s(%s/%s)%s\n",
			cmd, r.Doc.Command, reset,
			pkg, r.Doc.Scope, r.Doc.Name, reset)
		if r.Doc.Description != "" {
			fmt.Printf("%s\n", r.Doc.Description)
		}
	}
}

// colors returns ANSI codes for the command (yellow), package (cyan) and a
// reset; the description uses the default color. aux4 captures a command's
// stdout through a pipe and forwards it to the terminal, so a TTY check on
// our own stdout would always fail — instead emit color unconditionally and
// honor the NO_COLOR convention (also used by the tests).
func colors() (cmd, pkg, reset string) {
	if os.Getenv("NO_COLOR") != "" {
		return "", "", ""
	}
	return "\033[33m", "\033[36m", "\033[0m"
}

// firstParagraph returns the first meaningful prose line of a markdown
// document, skipping headings, code fences, badges and list markers.
func firstParagraph(content string) string {
	for _, line := range strings.Split(content, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "```") ||
			strings.HasPrefix(t, "-") || strings.HasPrefix(t, "[") || strings.HasPrefix(t, "|") {
			continue
		}
		return truncate(cleanMarkdown(t), 160)
	}
	return ""
}

// manDescription extracts the prose under a man page's "#### Description"
// section, falling back to the first paragraph.
func manDescription(content string) string {
	inDesc := false
	var buf []string
	for _, line := range strings.Split(content, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "####") {
			if inDesc {
				break
			}
			if strings.Contains(strings.ToLower(t), "description") {
				inDesc = true
			}
			continue
		}
		if !inDesc {
			continue
		}
		if t == "" {
			if len(buf) > 0 {
				break
			}
			continue
		}
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "-") {
			break
		}
		buf = append(buf, t)
	}
	if len(buf) == 0 {
		return firstParagraph(content)
	}
	return truncate(cleanMarkdown(strings.Join(buf, " ")), 160)
}

func cleanMarkdown(s string) string {
	return strings.TrimSpace(strings.NewReplacer("`", "", "**", "", "*", "", "__", "").Replace(s))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return strings.TrimSpace(s[:max]) + "…"
}
