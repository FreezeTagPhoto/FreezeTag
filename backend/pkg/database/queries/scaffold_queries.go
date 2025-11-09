package queries

import "strings"

type AndQuery struct {
	queries []DatabaseQuery
}

type OrQuery struct {
	queries []DatabaseQuery
}

func And(queries ...DatabaseQuery) *AndQuery {
	return &AndQuery{queries}
}

func Or(queries ...DatabaseQuery) *OrQuery {
	return &OrQuery{queries}
}

func (q *AndQuery) StatementWithArgs() (string, []any) {
	var builder strings.Builder
	var args []any
	for i, sq := range q.queries {
		builder.WriteString("(")
		s, a := sq.StatementWithArgs()
		builder.WriteString(s)
		builder.WriteString(")")
		args = append(args, a...)
		if i != len(q.queries)-1 {
			builder.WriteString(" AND ")
		}
	}
	return builder.String(), args
}

func (q *OrQuery) StatementWithArgs() (string, []any) {
	var builder strings.Builder
	var args []any
	for i, sq := range q.queries {
		builder.WriteString("(")
		s, a := sq.StatementWithArgs()
		builder.WriteString(s)
		builder.WriteString(")")
		args = append(args, a...)
		if i != len(q.queries)-1 {
			builder.WriteString(" OR ")
		}
	}
	return builder.String(), args
}
