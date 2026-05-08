package main

import "strings"

func stripGoogleSQLHints(sql string) string {
	var b strings.Builder
	b.Grow(len(sql))
	for i := 0; i < len(sql); {
		switch {
		case strings.HasPrefix(sql[i:], "--"):
			next := strings.IndexByte(sql[i:], '\n')
			if next < 0 {
				b.WriteString(sql[i:])
				i = len(sql)
				continue
			}
			b.WriteString(sql[i : i+next+1])
			i += next + 1
		case strings.HasPrefix(sql[i:], "/*"):
			next := strings.Index(sql[i+2:], "*/")
			if next < 0 {
				b.WriteString(sql[i:])
				i = len(sql)
				continue
			}
			end := i + 2 + next + 2
			b.WriteString(sql[i:end])
			i = end
		case sql[i] == '\'' || sql[i] == '"' || sql[i] == '`':
			end := scanQuotedSQL(sql, i)
			b.WriteString(sql[i:end])
			i = end
		case strings.HasPrefix(sql[i:], "@{"):
			end := scanGoogleSQLHint(sql, i)
			if end < 0 {
				b.WriteByte(sql[i])
				i++
				continue
			}
			i = end
		default:
			b.WriteByte(sql[i])
			i++
		}
	}
	return strings.TrimSpace(strings.Join(strings.Fields(b.String()), " "))
}

func scanQuotedSQL(sql string, start int) int {
	quote := sql[start]
	for i := start + 1; i < len(sql); i++ {
		if sql[i] != quote {
			continue
		}
		if i+1 < len(sql) && sql[i+1] == quote {
			i++
			continue
		}
		return i + 1
	}
	return len(sql)
}

func scanGoogleSQLHint(sql string, start int) int {
	depth := 0
	for i := start; i < len(sql); i++ {
		switch {
		case sql[i] == '\'' || sql[i] == '"' || sql[i] == '`':
			i = scanQuotedSQL(sql, i) - 1
		case strings.HasPrefix(sql[i:], "@{"):
			depth++
			i++
		case sql[i] == '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}
