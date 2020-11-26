package sqlbuilder


import (
	"bytes"
	"strings"
)

type sqlBuilder struct {
	sql    bytes.Buffer
	values []interface{}
	where  bool
	link   bool
}

func NewSqlBuilder() *sqlBuilder {
	t := new(sqlBuilder)
	return t
}

func (m *sqlBuilder) Select(cols ...string) *sqlBuilder {
	m.sql.WriteString("SELECT ")
	m.sql.WriteString(strings.Join(cols, ", "))
	return m
}

func (m *sqlBuilder) From(table string) *sqlBuilder {
	m.sql.WriteString(" FORM ")
	m.sql.WriteString(table)
	m.sql.WriteString(" a ")

	return m
}

func (m *sqlBuilder) Where(query string, value interface{}) *sqlBuilder {
	if value == nil || value == "" {
		return m
	}
	if !m.where {
		m.sql.WriteString("WHERE ")
		m.where = true
	}
	m.And()
	m.sql.WriteString(query)
	m.values = append(m.values, value)
	m.link = true
	return m
}

func (m *sqlBuilder) Join(join string, value ...interface{}) *sqlBuilder {
	m.sql.WriteString(join)
	if len(value) > 0 {
		m.values = append(m.values, value...)
	}
	return m
}

func (m *sqlBuilder) Append(query string) *sqlBuilder {
	m.sql.WriteString(query)
	return m
}

func (m *sqlBuilder) Build() (string, []interface{}) {
	return m.sql.String(), m.values
}

func (m *sqlBuilder) And() *sqlBuilder {
	if m.link {
		m.sql.WriteString("AND ")
		m.link = false
	}
	return m
}

func (m *sqlBuilder) Or() *sqlBuilder {
	m.sql.WriteString("AND ")
	return m
}
