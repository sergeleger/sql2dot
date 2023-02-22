package main

import (
	"bufio"
	"fmt"
	"io"

	"golang.org/x/exp/slices"
)

func d2(w io.Writer, tables Tables) error {
	slices.SortFunc(tables, func(a, b Table) bool {
		return a.Name < b.Name
	})

	bufW := bufio.NewWriter(w)
	defer bufW.Flush()

	for _, t := range tables {
		bufW.WriteString(t.Name)
		bufW.WriteString(": {")
		bufW.WriteString("\n")
		bufW.WriteString("   shape: sql_table")
		bufW.WriteString("\n\n")
		for _, c := range t.Columns {
			bufW.WriteString("   ")
			bufW.WriteString(c.Name)
			bufW.WriteString(": ")
			bufW.WriteString(c.Type)

			if c.Default != nil {
				fmt.Fprintf(bufW, " %v", c.Default)
			}

			if !c.NotNull {
				bufW.WriteString(" (nullable)")
			}

			if c.PrimaryKey {
				bufW.WriteString(" { constraint: primary_key }")
			}

			if i, _ := t.Refers(c.Name); i >= 0 {
				bufW.WriteString(" { constraint: foreign_key }")

			}

			bufW.WriteString("\n")
		}

		bufW.WriteString("}")
		bufW.WriteString("\n\n")
	}

	// Draw foreign  keys
	for _, t := range tables {
		for _, r := range t.References {
			tableIndex, _ := tables.Table(r.ToTable)
			if tableIndex == -1 {
				continue
			}

			fmt.Fprintf(bufW, "%s.%s -> %s.%s\n", t.Name, r.FromColumn, r.ToTable, r.ToColumn)
		}
	}

	return nil
}
