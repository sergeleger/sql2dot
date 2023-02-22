package main

import (
	"io"

	"golang.org/x/exp/slices"
)

func md(w io.Writer, tables Tables, section int) error {
	t, err := newTemplate(mdTemplate)
	if err != nil {
		return err
	}

	slices.SortFunc(tables, func(a, b Table) bool {
		return a.Name < b.Name
	})

	return t.Execute(w, struct {
		Tables  Tables
		Section int
	}{
		Tables:  tables,
		Section: section,
	})
}

var mdTemplate = `
{{range $i, $t := .Tables }}
{{- if eq $.Section 1 }}# {{end -}}
{{- if eq $.Section 2 }}## {{end -}}
{{- if eq $.Section 3 }}### {{end -}}
{{- if eq $.Section 4 }}#### {{end -}}
{{- if eq $.Section 5 }}##### {{end -}}
{{- if eq $.Section 6 }}###### {{end -}}
{{$t.Name}}

| Column | Type | Default | Refers |
| ------ | ---- | ------- | ------ |
{{range $j, $c := $t.Columns -}}
| {{$c.Name}}{{if $c.PrimaryKey}} (pk){{end}} | {{$c.Type}} | {{$c.Default}}{{if not $c.NotNull}} (nullable){{end}} | {{fk $t $c}} |
{{end}}

{{end}}
`
