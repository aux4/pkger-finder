package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: missing command (expected \"find\" or \"reindex\")")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "find":
		query := ""
		if len(os.Args) > 2 {
			query = os.Args[2]
		}
		limit := 10
		if len(os.Args) > 3 && strings.TrimSpace(os.Args[3]) != "" {
			if n, err := strconv.Atoi(strings.TrimSpace(os.Args[3])); err == nil && n > 0 {
				limit = n
			}
		}
		if err := runFind(query, limit); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "reindex":
		if err := runReindex(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q (expected \"find\" or \"reindex\")\n", os.Args[1])
		os.Exit(1)
	}
}

// runReindex rebuilds the BM25 index from scratch over every installed
// package and persists it. It is invoked directly and from the
// aux4:pkger/install and aux4:pkger/uninstall hooks.
func runReindex() error {
	idx, err := buildIndex()
	if err != nil {
		return err
	}
	if err := saveIndex(idx); err != nil {
		return err
	}
	fmt.Printf("Indexed %d documents from installed packages\n", len(idx.Documents))
	return nil
}

// runFind searches the index for the given query. If no index exists yet
// (e.g. right after this package was installed, since a package's own
// install does not trigger its hooks), it is built lazily on first use.
func runFind(query string, limit int) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return fmt.Errorf("missing search query")
	}

	idx, err := loadOrBuildIndex()
	if err != nil {
		return err
	}

	results := idx.Search(query, limit)
	if len(results) == 0 {
		fmt.Printf("No commands or packages found for %q\n", query)
		return nil
	}

	printResults(results)
	return nil
}

func loadOrBuildIndex() (*Index, error) {
	if indexExists() && indexIsFresh() {
		idx, err := loadIndex()
		if err == nil {
			return idx, nil
		}
		// Fall through and rebuild on a corrupt/unreadable index.
	}
	idx, err := buildIndex()
	if err != nil {
		return nil, err
	}
	if err := saveIndex(idx); err != nil {
		return nil, err
	}
	return idx, nil
}
