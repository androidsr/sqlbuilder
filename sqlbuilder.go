package sqlbuilder

import (
	"bytes"
	"strings"
)
var _VERSION_ = "v0.0.1"
type builder struct {
	sql    bytes.Buffer
	values []interface{}
	where  bool
	link   bool
}

type selectBuilder struct {
	builder
}

type insertBuilder struct {
	builder
}

type updateBuilder struct {
	builder
}

type deleteBuilder struct {
	builder
}

func NewSelect() *selectBuilder {
	t := new(selectBuilder)
	return t
}

func NewInsert() *insertBuilder {
	t := new(insertBuilder)
	return t
}

func NewUpdate() *updateBuilder {
	t := new(updateBuilder)
	return t
}

func NewDelete() *deleteBuilder {
	t := new(deleteBuilder)
	return t
}

//===================================插入语句==============================================
//插入SQL生成器
func (m *insertBuilder) Insert(table string) *insertBuilder {
	m.sql.WriteString("INSERT INTO ")
	m.sql.WriteString(table)
	m.sql.WriteString(" ")
	return m
}

//插入SQL生成器
func (m *insertBuilder) Columns(cols ...string) *insertBuilder {
	m.sql.WriteString("(")
	m.sql.WriteString(strings.Join(cols, ", "))
	m.sql.WriteString(") ")
	return m
}

//插入SQL生成器
func (m *insertBuilder) Values(cols ...interface{}) *insertBuilder {
	m.sql.WriteString("VALUES (")
	for i, _ := range cols {
		m.sql.WriteString("?")
		if len(cols)-1 != i {
			m.sql.WriteString(", ")
		}
	}
	m.sql.WriteString(") ")
	m.values = append(m.values, cols...)
	return m
}

//======================================更新语句===============================================
//更新SQL生成器
func (m *updateBuilder) Update(table string) *updateBuilder {
	m.sql.WriteString("UPDATE ")
	m.sql.WriteString(table)
	m.sql.WriteString(" ")
	return m
}

//更新SQL生成器
func (m *updateBuilder) Set(cols ...string) *updateBuilder {
	m.sql.WriteString("SET ")
	for i, c := range cols {
		m.sql.WriteString(c)
		m.sql.WriteString(" = ?")
		if len(cols)-1 != i {
			m.sql.WriteString(", ")
		}
	}
	return m
}

//======================================删除语句==============================================
//更新SQL生成器
func (m *deleteBuilder) Delete() *deleteBuilder {
	m.sql.WriteString("DELETE ")
	return m
}

//=========================================查询语句============================================
//查询SQL生成器
func (m *selectBuilder) Select(cols ...string) *selectBuilder {
	m.sql.WriteString("SELECT ")
	m.sql.WriteString(strings.Join(cols, ", "))
	return m
}

//查询SQL生成器
func (m *selectBuilder) From(table string) *selectBuilder {
	m.sql.WriteString(" FORM ")
	m.sql.WriteString(table)
	m.sql.WriteString(" a ")

	return m
}

//=======================================公共条件================================================
//SQL通用条件生成器
func (m *builder) Where(query string, value interface{}) *builder {
	if value == nil || value == "" {
		return m
	}
	if !m.where {
		m.sql.WriteString("WHERE ")
		m.where = true
	}
	m.And()
	m.sql.WriteString(query)
	m.sql.WriteString(" ")
	m.values = append(m.values, value)
	m.link = true
	return m
}

//SQL IN条件生成器
func (m *builder) In(query string, value []interface{}) *builder {
	if len(value) == 0 {
		return m
	}
	m.And()
	for i, _ := range value {
		m.sql.WriteString("?")
		if len(value)-1 != i {
			m.sql.WriteString(", ")
		}
	}
	m.sql.WriteString(query)
	m.values = append(m.values, value...)
	m.link = true
	return m
}

//SQL LIKE条件生成器
func (m *builder) Like(query string, value interface{}) *builder {
	if value == nil || value == "" {
		return m
	}
	m.And()
	m.sql.WriteString(strings.Replace(query, "?", "CONCAT('%',?,'%')", -1))
	m.values = append(m.values, value)
	m.link = true
	return m
}

//SQL表连接生成器
func (m *builder) Join(join string, value ...interface{}) *builder {
	m.sql.WriteString(join)
	if len(value) > 0 {
		m.values = append(m.values, value...)
	}
	m.sql.WriteString(" ")
	return m
}

//追加任意SQL
func (m *builder) Append(query string) *builder {
	m.sql.WriteString(query)
	m.sql.WriteString(" ")
	return m
}

//返回最终SQL及参数
func (m *builder) Build() (string, []interface{}) {
	return m.sql.String(), m.values
}

//手动增加AND
func (m *builder) And() *builder {
	if m.link {
		m.sql.WriteString("AND ")
		m.link = false
	}
	return m
}

//手动增加OR
func (m *builder) Or() *builder {
	m.sql.WriteString("OR ")
	return m
}
