package main

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

const pgAllTables = `
select
	'table' as type,
	tablename as name
FROM
	pg_catalog.pg_tables
WHERE
	schemaname != 'pg_catalog' AND
	schemaname != 'information_schema'
UNION
select
	'view' as type,
	viewname as name
from
	pg_catalog.pg_views
WHERE
	schemaname != 'pg_catalog' AND
	schemaname != 'information_schema'
`

const pgColumns = `
select
	column_name,
	udt_name,
	column_default,
	is_nullable='YES' as is_nullable,
	is_identity='YES' as is_identity
from
	information_schema.columns
where
	table_name = $1
order by ordinal_position
`

const pgKeys = `
SELECT
    tc.table_name as from_table,
    kcu.column_name as from_column,
    ccu.table_name AS  to_table,
    ccu.column_name AS to_column
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' and tc.table_name = $1
`

type postgresTable struct {
	Type       string `db:"type"`
	Name       string `db:"name"`
	Columns    []postgresColumn
	References []postgresForeignKey
}

type postgresColumn struct {
	Name       string         `db:"column_name"`
	Type       string         `db:"udt_name"`
	Default    sql.NullString `db:"column_default"`
	IsNullable bool           `db:"is_nullable"`
	PrimaryKey bool           `db:"is_identity"`
}

type postgresForeignKey struct {
	FromTable  string `db:"from_table"`
	FromColumn string `db:"from_column"`
	ToTable    string `db:"to_table"`
	ToColumn   string `db:"to_column"`
}

func parsePostgres(db *sqlx.DB) (Tables, error) {
	var pgTables []postgresTable
	err := db.Select(&pgTables, pgAllTables)
	if err != nil {
		return nil, err
	}

	// capture columns and foreign keys
	for i, t := range pgTables {
		if err := db.Select(&pgTables[i].Columns, pgColumns, t.Name); err != nil {
			return nil, err
		}

		if err := db.Select(&pgTables[i].References, pgKeys, t.Name); err != nil {
			return nil, err
		}
	}

	var tables []Table
	for _, t := range pgTables {
		var table = Table{Name: t.Name, Type: t.Type}

		for _, c := range t.Columns {
			col := Column{
				Name:       c.Name,
				Type:       c.Type,
				NotNull:    !c.IsNullable,
				PrimaryKey: c.PrimaryKey,
			}
			if c.Default.Valid {
				col.Default = c.Default.String
			}

			table.Columns = append(table.Columns, col)
		}

		for i, fk := range t.References {
			table.References = append(table.References, ForeignKey{
				Sequence:   int64(i + 1),
				FromColumn: fk.FromColumn,
				ToTable:    fk.ToTable,
				ToColumn:   fk.ToColumn,
			})
		}

		tables = append(tables, table)
	}

	return tables, nil
}
