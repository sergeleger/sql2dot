package main

type Tables []Table

type Table struct {
	Type       string
	Name       string
	Columns    []Column
	References []ForeignKey
}

type Column struct {
	Name       string
	Type       string
	NotNull    bool
	Default    interface{}
	PrimaryKey bool
}

type ForeignKey struct {
	Sequence   int64
	FromColumn string
	ToTable    string
	ToColumn   string
}

// Table performs a table lookup using the table's name
func (t Tables) Table(name string) (int, Table) {
	for i := range t {
		if t[i].Name == name {
			return i, t[i]
		}
	}

	return -1, Table{}
}

// Column performs a column lookup by name.
func (t Table) Column(name string) (int, Column) {
	for i, c := range t.Columns {
		if c.Name == name {
			return i, c
		}
	}

	return -1, Column{}
}

func (t Table) Refers(column string) (int, ForeignKey) {
	for i, r := range t.References {
		if r.FromColumn == column {
			return i, r
		}
	}

	return -1, ForeignKey{}
}
