package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := run(os.Args)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var exclude = StringArrayFromFile{}
	var include = StringArrayFromFile{}
	var flag = flag.NewFlagSet(args[0], flag.ExitOnError)
	sqlite := flag.Bool("sqlite", false, "Access a SQLite database.")
	postgres := flag.Bool("postgres", false, "Access a PostgreSQL database defined by PGDATABASE_URL environment variable")
	listTables := flag.Bool("list", false, "List table names to standard output.")
	includeFK := flag.Bool("fk", false, "Add FK information to table output.")
	truncateDot := flag.Int("truncate", -1, "In Graphviz output, truncate large tables to `n` columns.")
	mdOutput := flag.Bool("md", false, "Output using Markdown.")
	htmlOutput := flag.Bool("html", false, "Output using HTML.")
	templateFile := flag.String("template", "", "Alternative Go template file.")
	section := flag.Int("section", 1, "Generate <h`N` /> elements (or equivalent Markdown)")

	flag.Var(&exclude, "exclude-file", "Tables or views to exclude, read from a file.")
	flag.Var(&include, "include-file", "Tables or views to include, read from a file.")
	flag.Var(&exclude.List, "exclude", "Tables or views to exclude.")
	flag.Var(&include.List, "include", "Tables or views to include.")
	flag.Parse(args[1:])

	if len(exclude.List) > 0 && len(include.List) > 0 {
		return errors.New("error: provide only one of --include or --exclude")
	}

	// validate user provided template now.
	var userTemplate *template.Template
	if *templateFile != "" {
		var err error
		if userTemplate, err = newTemplateFromFile(*templateFile); err != nil {
			return err
		}
	}

	// validate section: 1 <= section <= 6
	if *section < 1 || *section > 6 {
		*section = 1
	}

	// Read tables from the specified source
	var tables Tables
	switch {
	case *sqlite:
		db, err := sqlx.Open("sqlite3", flag.Arg(0))
		if err != nil {
			return err
		}
		defer db.Close()

		tables, err = parseSqlite(db)
		if err != nil {
			return err
		}

	case *postgres:
		db, err := sqlx.Open("pgx", os.Getenv("PGDATABASE_URL"))
		if err != nil {
			return err
		}
		defer db.Close()

		tables, err = parsePostgres(db)
		if err != nil {
			return err
		}

	default:
		return errors.New("error: provide -sqlite or -postgres")
	}

	// Filter tables
	tables = filter(tables, exclude.List, include.List)

	// Output requested results
	switch {
	case *listTables:
		for _, t := range tables {
			fmt.Println(t.Name)
		}
		return nil

	case userTemplate != nil:
		if err := userTemplate.Execute(os.Stdout, tables); err != nil {
			return err
		}
		return nil

	case *mdOutput:
		return md(os.Stdout, tables, *section)

	case *htmlOutput:
		return html(os.Stdout, tables, *section)

	default:
		return graph(os.Stdout, tables, *includeFK, *truncateDot)
	}
}

func filter(tables []Table, exclude, include []string) []Table {
	if len(exclude) == 0 && len(include) == 0 {
		return tables
	}

	var newTables []Table
	fn := newFilter(exclude, include)

	// remove tables
	for _, t := range tables {
		if fn(t.Name) {
			newTables = append(newTables, t)
		}
	}

	return newTables
}

func newFilter(exclude, include []string) func(n string) bool {
	var m = make(map[string]struct{})
	var wildcard = make([]string, 0)

	var fnNeg = func(n string) bool {
		n = strings.ToLower(n)
		if _, ok := m[n]; ok {
			return false
		}

		return !searchWildcard(n, wildcard)
	}

	var fnPos = func(n string) bool {
		n = strings.ToLower(n)
		if _, ok := m[n]; ok {
			return true
		}

		return searchWildcard(n, wildcard)
	}

	var fn = fnPos
	var src = include
	if len(exclude) > 0 {
		src = exclude
		fn = fnNeg
	}

	for _, k := range src {
		k = strings.ToLower(k)
		if i := strings.LastIndexByte(k, '*'); i >= 0 {
			wildcard = append(wildcard, k[:i])
			continue
		}

		m[strings.ToLower(k)] = struct{}{}
	}

	return fn

}

func searchWildcard(n string, wildcards []string) bool {
	for i := range wildcards {
		if strings.HasPrefix(n, wildcards[i]) {
			return true
		}
	}

	return false
}
