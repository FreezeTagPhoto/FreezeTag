package queries

import "fmt"

type DatabaseQuery interface {
	StatementWithArgs() (string, []any)
}

func ImageIdPreparable(dq DatabaseQuery) (string, []any) {
	stmt, args := dq.StatementWithArgs()
	return fmt.Sprintf(`SELECT id FROM Images WHERE %s`, stmt), args
}
