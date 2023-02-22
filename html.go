package main

import (
	"io"

	"golang.org/x/exp/slices"
)

func html(w io.Writer, tables Tables, section int) error {
	t, err := newTemplate(htmlTemplate)
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

var htmlTemplate = `
{{range $i, $t := .Tables }}
<h{{$.Section}}>{{$t.Name}}</h{{$.Section}}>
<table>
<thead>
    <tr>
        <th>Column</th>
        <th>Type</th>
		<th>Default</th>
        <th>Refers</th>
	</tr>
</thead>
<tbody>
{{range $j, $c := $t.Columns }}
<tr>
   <td>{{$c.Name}}{{if $c.PrimaryKey}} (pk){{end}}</td>
   <td>{{$c.Type}}</td>
   <td>{{$c.Default}}{{if not $c.NotNull}} (nullable){{end}}</td>
   <td>{{- fk $t $c -}}</td>
</tr>
{{end}}
</tbody>
</table>
{{end}}
`
