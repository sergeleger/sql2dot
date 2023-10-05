package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type sqliteTable struct {
	Type       string `db:"type"`
	Name       string `db:"name"`
	Columns    []sqliteColumn
	References []sqliteForeignKey
}

type sqliteColumn struct {
	ID         int64       `db:"cid"`
	Name       string      `db:"name"`
	Type       string      `db:"type"`
	NotNull    int         `db:"notnull"`
	Default    interface{} `db:"dflt_value"`
	PrimaryKey int         `db:"pk"`
}

type sqliteForeignKey struct {
	ID         int64  `db:"id"`
	Sequence   int64  `db:"seq"`
	FromColumn string `db:"from"`
	ToTable    string `db:"table"`
	ToColumn   string `db:"to"`
	OnUpdate   string `db:"on_update"`
	OnDelete   string `db:"on_delete"`
	Match      string `db:"match"`
}

func parseSqlite(db *sqlx.DB) (Tables, error) {
	var sqliteTables []sqliteTable
	err := db.Select(&sqliteTables, `select type, name from sqlite_master where type="table" or type="view"`)
	if err != nil {
		return nil, fmt.Errorf("error: getting list of tables: %w", err)
	}

	// Capture columns and Foreign Keys
	for i, t := range sqliteTables {
		err := db.Select(&sqliteTables[i].Columns, fmt.Sprintf(`PRAGMA table_info(%q)`, t.Name))
		if err != nil {
			return nil, fmt.Errorf("error: getting list of columns for %s: %w", t.Name, err)
		}

		err = db.Select(&sqliteTables[i].References, fmt.Sprintf(`PRAGMA foreign_key_list(%q)`, t.Name))
		if err != nil {
			return nil, fmt.Errorf("error: getting list of foreign keys for %s: %w", t.Name, err)
		}
	}

	// convert the sqlite schema to the expected table structure
	var tables []Table
	for _, t := range sqliteTables {
		var table = Table{Name: t.Name, Type: t.Type}

		for _, c := range t.Columns {
			table.Columns = append(table.Columns, Column{
				Name:       c.Name,
				Type:       c.Type,
				NotNull:    c.NotNull == 1,
				Default:    c.Default,
				PrimaryKey: c.PrimaryKey == 1,
			})
		}

		for _, r := range t.References {
			table.References = append(table.References, ForeignKey{
				Sequence:   r.Sequence,
				FromColumn: r.FromColumn,
				ToTable:    r.ToTable,
				ToColumn:   r.ToColumn,
			})
		}

		tables = append(tables, table)
	}

	return tables, nil
}
