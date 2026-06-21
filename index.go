package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Document is a single searchable unit: either a package README or one
// command man page. Term frequencies and length are stored so that
// searching never needs to re-read the source files.
type Document struct {
	ID          string         `json:"id"`
	Scope       string         `json:"scope"`
	Name        string         `json:"name"`
	Type        string         `json:"type"` // "command" or "readme"
	Command     string         `json:"command"`
	Description string         `json:"description"`
	Path        string         `json:"path"`
	Length      int            `json:"length"`
	TermFreqs   map[string]int `json:"termFreqs"`
}

// Index is the persisted BM25 corpus.
type Index struct {
	Documents []Document     `json:"documents"`
	DocFreq   map[string]int `json:"docFreq"`
	AvgLength float64        `json:"avgLength"`
}

// aux4HomeDir resolves ~/.aux4.config. AUX4_PKGER_FINDER_HOME overrides the base
// directory under which packages/ is scanned and .pkg-index/ is written;
// it is package-specific so it never interferes with aux4 core resolution.
func aux4HomeDir() (string, error) {
	if dir := os.Getenv("AUX4_PKGER_FINDER_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".aux4.config"), nil
}

func packagesDir() (string, error) {
	home, err := aux4HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "packages"), nil
}

func indexPath() (string, error) {
	home, err := aux4HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pkg-index", "index.json"), nil
}

// buildIndex walks every installed package and indexes its README and man
// pages. Each man page becomes one document keyed by the command it
// documents, so a search resolves to a specific command to run.
func buildIndex() (*Index, error) {
	pkgDir, err := packagesDir()
	if err != nil {
		return nil, err
	}

	idx := &Index{
		Documents: []Document{},
		DocFreq:   map[string]int{},
	}

	scopes, err := os.ReadDir(pkgDir)
	if err != nil {
		if os.IsNotExist(err) {
			return idx, nil
		}
		return nil, err
	}

	for _, scope := range scopes {
		if !scope.IsDir() {
			continue // skip all.json and other files
		}
		scopePath := filepath.Join(pkgDir, scope.Name())
		names, err := os.ReadDir(scopePath)
		if err != nil {
			continue
		}
		for _, name := range names {
			if !name.IsDir() {
				continue
			}
			collectPackageDocs(idx, scope.Name(), name.Name(), filepath.Join(scopePath, name.Name()))
		}
	}

	finalizeIndex(idx)
	return idx, nil
}

func collectPackageDocs(idx *Index, scope, name, pkgPath string) {
	if content, err := os.ReadFile(filepath.Join(pkgPath, "README.md")); err == nil {
		doc := Document{
			ID:          scope + "/" + name + ":readme",
			Scope:       scope,
			Name:        name,
			Type:        "readme",
			Command:     "aux4 aux4 pkger install " + scope + "/" + name,
			Description: firstParagraph(string(content)),
			Path:        filepath.Join(pkgPath, "README.md"),
		}
		indexDocument(idx, &doc, string(content), scope+" "+name)
		idx.Documents = append(idx.Documents, doc)
	}

	manDir := filepath.Join(pkgPath, "man")
	entries, err := os.ReadDir(manDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		manPath := filepath.Join(manDir, e.Name())
		content, err := os.ReadFile(manPath)
		if err != nil {
			continue
		}
		invocation := commandFromManFile(e.Name())
		doc := Document{
			ID:          scope + "/" + name + ":" + e.Name(),
			Scope:       scope,
			Name:        name,
			Type:        "command",
			Command:     invocation,
			Description: manDescription(string(content)),
			Path:        manPath,
		}
		// Fold the invocation, scope and name into the searchable text so
		// queries that name the command or package rank it highly.
		indexDocument(idx, &doc, string(content), invocation+" "+scope+" "+name)
		idx.Documents = append(idx.Documents, doc)
	}
}

// commandFromManFile reconstructs the CLI invocation from a man page
// filename. "__" separates hierarchy levels and "_" stands in for ":" in
// profile names. e.g. "aux4_pkger__find.md" -> "aux4 aux4 pkger find".
func commandFromManFile(filename string) string {
	parts := strings.Split(strings.TrimSuffix(filename, ".md"), "__")
	for i, p := range parts {
		parts[i] = strings.ReplaceAll(p, "_", " ")
	}
	return "aux4 " + strings.Join(parts, " ")
}

func indexDocument(idx *Index, doc *Document, content, extra string) {
	tokens := tokenize(content + " " + extra)
	tf := map[string]int{}
	for _, t := range tokens {
		tf[t]++
	}
	doc.TermFreqs = tf
	doc.Length = len(tokens)
	for term := range tf {
		idx.DocFreq[term]++
	}
}

func finalizeIndex(idx *Index) {
	total := 0
	for _, d := range idx.Documents {
		total += d.Length
	}
	if len(idx.Documents) > 0 {
		idx.AvgLength = float64(total) / float64(len(idx.Documents))
	}
}

func tokenize(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	tokens := make([]string, 0, len(fields))
	for _, f := range fields {
		if len(f) < 2 || stopWords[f] {
			continue
		}
		tokens = append(tokens, f)
	}
	return tokens
}

var stopWords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "that": true,
	"this": true, "from": true, "are": true, "was": true, "but": true,
	"not": true, "you": true, "your": true, "can": true, "has": true,
	"have": true, "will": true, "via": true, "its": true, "any": true,
	"when": true, "where": true, "which": true, "what": true,
}

func saveIndex(idx *Index) error {
	path, err := indexPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func loadIndex() (*Index, error) {
	path, err := indexPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

func indexExists() bool {
	path, err := indexPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// indexIsFresh reports whether the saved index is at least as new as the
// package ledger (packages/all.json), which pkger rewrites on every
// install and uninstall. This keeps `find` self-refreshing even where the
// install/uninstall hooks are unavailable. When no ledger exists (e.g. a
// test fixture), an existing index is treated as fresh.
func indexIsFresh() bool {
	ipath, err := indexPath()
	if err != nil {
		return false
	}
	istat, err := os.Stat(ipath)
	if err != nil {
		return false
	}
	home, err := aux4HomeDir()
	if err != nil {
		return false
	}
	lstat, err := os.Stat(filepath.Join(home, "packages", "all.json"))
	if err != nil {
		return true
	}
	return !istat.ModTime().Before(lstat.ModTime())
}
