package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"

	"golang.org/x/exp/slices"
)

func newTemplate(source string) (*template.Template, error) {
	tmpl := template.New("")
	tmpl.Funcs(map[string]interface{}{
		"fk": func(t Table, column Column) string {
			i, k := t.Refers(column.Name)
			if i == -1 {
				return ""
			}

			return k.ToTable + "." + k.ToColumn
		},
		"add": func(x, y int) int {
			return x + y
		},
		"sub": func(x, y int) int {
			return x - y
		},
	})

	if _, err := tmpl.Parse(source); err != nil {
		return nil, err
	}

	return tmpl, nil
}

func newTemplateFromFile(filename string) (*template.Template, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return newTemplate(string(buf))
}

// Create GraphViz graph
func graph(w io.Writer, tables Tables, includeFK bool, truncate int) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	bw.WriteString("digraph g {\n")
	bw.WriteString("rankdir=LR;\n")
	bw.WriteString("edge [ arrowsize=0.5, arrowtail=empty, arrowhead=empty ];\n")
	bw.WriteString("node [ shape=plain, height=0.1 ];\n")
	bw.WriteString(`fontsize="102pt	"` + "\n")

	tmpl, err := newTemplate(graphVizTemplate)
	if err != nil {
		return err
	}

	slices.SortFunc(tables, func(a, b Table) bool {
		return a.Name < b.Name
	})

	// Draw tables and columns
	var ctx = struct {
		Table    Table
		FK       bool
		Truncate int
	}{FK: includeFK, Truncate: truncate}
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

			fmt.Fprintf(bw, "%s:s%d -> %s:e%d;\n", sourceName, srcIndex, destNode, destIndex)

		}
	}

	bw.WriteString("}")

	return nil
}

var graphVizTemplate = `
<font point-size="12">
<table {{if eq .Table.Type "view" }} style="rounded" {{end}} border="1" cellpadding="2" cellborder="0" cellspacing="0">
   <tr>
      <td {{if .FK}} colspan="3" {{end}}>{{.Table.Name}}</td>
   </tr>
   <HR/>
   {{- range $i, $d := .Table.Columns }}
   {{- if or (le $.Truncate 0) (lt $i $.Truncate) }}
   <tr>
      <td align="left" port="e{{$i}}">{{$d.Name}}</td>
      <td align="left" {{- if not $.FK }} port="s{{$i}}" {{- end}}>
         {{- if $d.Type -}}
         <font point-size="10">{{$d.Type}}</font>
         {{- end -}}
      </td>
      {{- if $.FK }}
         <td align="left" port="s{{$i}}"><font point-size="10">{{ $fk := fk $.Table $d}}
         {{- if eq $fk ""}} {{else}}{{$fk}}{{end -}}
         </font></td>
      {{end}}
   </tr>
   {{- else if eq $.Truncate $i }}
   <tr>
	<td colspan="3">
	<font point-size="10"><i>{{sub (len $.Table.Columns) $.Truncate }} other column(s) omitted</i></font>
	</td>
   </tr>
   {{- end }}
   {{- end}}
</table>
</font>
`
