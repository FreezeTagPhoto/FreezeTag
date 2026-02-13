package queries

import "fmt"

type SortOrder uint8

const (
	Descending SortOrder = iota
	Ascending
)

func (o SortOrder) String() string {
	switch o {
	case Ascending:
		return "ASC"
	case Descending:
		return "DESC"
	}
	return "DESC"
}

type SortField uint8

const (
	DateAdded SortField = iota
	DateCreated
)

func (f SortField) String() string {
	switch f {
	case DateAdded:
		return "dateUploaded"
	case DateCreated:
		return "COALESCE(dateTaken,dateUploaded)"
	}
	return "dateUploaded"
}

type DatabaseQuery interface {
	StatementWithArgs() (string, []any)
}

func ImageIdPreparable(dq DatabaseQuery, sf SortField, so SortOrder) (string, []any) {
	stmt, args := dq.StatementWithArgs()
	field := sf.String()
	order := so.String()
	return fmt.Sprintf(`SELECT id FROM Images WHERE %s ORDER BY %s %s`, stmt, field, order), args
}
