package base

import "fmt"

type DatabaseParams interface {
	Reset()
	Next() string
}

type ParamsMaker func() DatabaseParams

type SqliteParams struct {
}

func (p *SqliteParams) Reset() {
}

func (p *SqliteParams) Next() string {
	return "?"
}

type PostgreSqlParams struct {
	count int
}

func (p *PostgreSqlParams) Reset() {
	p.count = 0
}

func (p *PostgreSqlParams) Next() string {
	p.count++
	return fmt.Sprintf("$%d", p.count)
}
