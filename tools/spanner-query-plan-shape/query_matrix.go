package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type queryMatrixAxis struct {
	Name   string
	Values []queryMatrixAxisValue
}

type queryMatrixAxisValue struct {
	Label  string
	Fields map[string]string
}

func buildQueryMatrixCases(prefix, sqlTemplate string, axes ...queryMatrixAxis) []queryCase {
	tmpl := template.Must(template.New(prefix).Option("missingkey=error").Parse(sqlTemplate))
	var out []queryCase
	walkQueryMatrix(tmpl, prefix, axes, nil, map[string]map[string]string{}, &out)
	return out
}

func walkQueryMatrix(tmpl *template.Template, prefix string, axes []queryMatrixAxis, labels []string, data map[string]map[string]string, out *[]queryCase) {
	if len(axes) == 0 {
		var b bytes.Buffer
		if err := tmpl.Execute(&b, data); err != nil {
			panic(fmt.Sprintf("execute query matrix template %q: %v", prefix, err))
		}
		*out = append(*out, queryCase{
			Label: prefix + "/" + strings.Join(labels, "/"),
			SQL:   strings.TrimSpace(b.String()),
		})
		return
	}

	axis := axes[0]
	for _, value := range axis.Values {
		nextData := cloneQueryMatrixData(data)
		nextData[axis.Name] = value.Fields
		nextLabels := append(append([]string{}, labels...), value.Label)
		walkQueryMatrix(tmpl, prefix, axes[1:], nextLabels, nextData, out)
	}
}

func cloneQueryMatrixData(in map[string]map[string]string) map[string]map[string]string {
	out := make(map[string]map[string]string, len(in))
	for name, fields := range in {
		nextFields := make(map[string]string, len(fields))
		for key, value := range fields {
			nextFields[key] = value
		}
		out[name] = nextFields
	}
	return out
}
