package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
)

// Create GraphViz graph
func graph(w io.Writer, tables Tables, includeFK bool) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	bw.WriteString("digraph g {\n")
	bw.WriteString("node [shape=plain,style=rounded,height=0.1];\n")
	bw.WriteString(`fontsize="102pt	"` + "\n")

	tmpl := template.New("")
	tmpl.Funcs(map[string]interface{}{
		"fk": func(t Table, column Column) string {
			i, k := t.Refers(column.Name)
			if i == -1 {
				return " "
			}

			return k.ToTable + "." + k.ToColumn
		},
	})

	if _, err := tmpl.Parse(tableTemplate); err != nil {
		return err
	}

	// Draw tables and columns
	var ctx = struct {
		Table Table
		FK    bool
	}{FK: includeFK}
	for i, t := range tables {
		ctx.Table = t

		nodeName := fmt.Sprintf("table%d", i)

		bw.WriteString(nodeName)
		bw.WriteString(" [label=<")
		if err := tmpl.Execute(bw, ctx); err != nil {
			return err
		}
		bw.WriteString(">]\n")
	}

	// Draw foreign  keys
	for i, t := range tables {
		sourceName := fmt.Sprintf("table%d", i)

		for _, r := range t.References {
			tableIndex, destTable := tables.Table(r.ToTable)
			if tableIndex == -1 {
				continue
			}

			srcIndex, _ := t.Column(r.FromColumn)
			destNode := fmt.Sprintf("table%d", tableIndex)
			destIndex, _ := destTable.Column(r.ToColumn)

			if includeFK {
				fmt.Fprintf(bw, "%s:s%d -> %s:e%d;\n", sourceName, srcIndex, destNode, destIndex)
			} else {
				fmt.Fprintf(bw, "%s:%d -> %s:%d;\n", sourceName, srcIndex, destNode, destIndex)
			}
		}
	}

	bw.WriteString("}")

	return nil
}

var tableTemplate = `
<font point-size="12">
<table {{if eq .Table.Type "view" }} style="rounded" {{end}} border="1" cellpadding="2" cellborder="0" cellspacing="0">
	<tr>
		<td {{if .FK}} colspan="3" {{end}}>{{.Table.Name}}</td>
	</tr>
	<HR/>
	{{- range $i, $d := .Table.Columns }}
	<tr>
		<td align="left"
			{{- if not $.FK }} port="{{$i}}"
			{{- else}} port="e{{$i}}"
			{{- end}}>
			{{$d.Name}}</td>

		<td align="left">
			{{if $d.Type}}
			<font point-size="10">{{$d.Type}}</font>
			{{end}}
		</td>

		{{ if $.FK }}
			<td align="left" port="s{{$i}}"><font point-size="10">{{ fk $.Table $d}}</font></td>
		{{end}}
	</tr>
	{{- end}}
</table>
</font>
`
