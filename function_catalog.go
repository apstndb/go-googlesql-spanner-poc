package spanalyzer

import (
	"fmt"
	"sort"
	"strings"

	googlesql "github.com/goccy/go-googlesql"
)

func (c *GoogleSQLCatalog) FunctionCatalogDebugString(verbose bool) (string, error) {
	functions, err := c.SimpleCatalog.Functions()
	if err != nil {
		return "", err
	}
	sort.Slice(functions, func(i, j int) bool {
		left := functionSortKey(functions[i])
		right := functionSortKey(functions[j])
		return left < right
	})

	var b strings.Builder
	for _, fn := range functions {
		debug, err := fn.DebugString(verbose)
		if err != nil {
			name, _ := fn.FullName(false)
			return "", fmt.Errorf("function %s: %w", name, err)
		}
		b.WriteString(debug)
		if !strings.HasSuffix(debug, "\n") {
			b.WriteByte('\n')
		}
	}
	return b.String(), nil
}

func functionSortKey(fn *googlesql.Function) string {
	name, err := fn.FullName(false)
	if err != nil {
		name, err = fn.Name()
	}
	if err != nil {
		return ""
	}
	return strings.ToLower(name)
}
