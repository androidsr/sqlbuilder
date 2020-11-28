package sqlbuilder

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

const (
	iotaString = iota
	iotaInt64
	iotaFloat64
	iotaBool
	iotaOther
)

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

type mapping struct {
	fields       []string      //属性名称
	cols         []string      //属性对应表列表名称
	notEmptyCols []string      //非空值 对应表列表名称
	values       []interface{} //非空值
	pk           string        //主键列名称
	pkValue      interface{}   //主键值
	rowsClose    bool          //自动关闭
	dataType     []int         //数据类型
	tableName    string        //表名
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

func NewMapping() *mapping {
	t := new(mapping)
	t.rowsClose = true
	return t
}

func (m *mapping) ReadTarget(target interface{}) *mapping {
	t := reflect.TypeOf(target)
	v := reflect.ValueOf(target)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	m.tableName = m.nameTag(t.Name())
	fmt.Println(m.tableName)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		m.fields = append(m.fields, f.Name)
		value := v.Field(i).Interface()
		switch value.(type) {
		case string:
			m.dataType = append(m.dataType, iotaString)
		case int64:
			m.dataType = append(m.dataType, iotaInt64)
		case float64:
			m.dataType = append(m.dataType, iotaFloat64)
		case bool:
			m.dataType = append(m.dataType, iotaBool)
		default:
			m.dataType = append(m.dataType, iotaOther)
		}
		tg := f.Tag
		dbName := tg.Get("db")
		if dbName == "" {
			dbName = tg.Get("json")
		}
		if dbName == "" {
			dbName = tg.Get(m.nameTag(f.Name))
		}
		m.cols = append(m.cols, dbName)
		if value != nil {
			m.notEmptyCols = append(m.notEmptyCols, dbName)
			m.values = append(m.values, value)
		}
		if tg.Get("pk") == "true" {
			m.pk = dbName
			m.pkValue = value
		}
	}
	return m
}

func (m *mapping) ScanStruct(rows *sql.Rows, dest interface{}) interface{} {
	columns, _ := rows.Columns()
	if len(m.fields) == 0 {
		m.ReadTarget(dest)
	}
	t := reflect.TypeOf(dest)
	if t.Kind() == reflect.Ptr { //指针类型获取真正type需要调用Elem
		t = t.Elem()
	}
	cache := make([]interface{}, len(columns))
	for index, _ := range cache {
		var a interface{}
		cache[index] = &a
	}
	if rows.Next() {
		rows.Scan(cache...)
	}
	newStruc := reflect.New(t)
	for c, col := range columns {
		for i, tag := range m.cols {
			if col == tag {
				name := m.fields[i]
				switch m.dataType[i] {
				case iotaString:
					newStruc.Elem().FieldByName(name).SetString(*cache[c].(*string))
				case iotaInt64:
					newStruc.Elem().FieldByName(name).SetInt(*cache[c].(*int64))
				case iotaFloat64:
					newStruc.Elem().FieldByName(name).SetFloat(*cache[c].(*float64))
				case iotaBool:
					newStruc.Elem().FieldByName(name).SetBool(*cache[c].(*bool))
				default:
					newStruc.Elem().FieldByName(name).SetString(*cache[c].(*string))
				}
			}
		}
	}
	fmt.Println(newStruc)
	return newStruc
}

func (m *mapping) ScanArrayStruct(rows *sql.Rows, dest interface{}) {
	if m.rowsClose {
		defer rows.Close()
	}

}

func (m *mapping) ScanMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	if m.rowsClose {
		defer rows.Close()
	}
	columns, _ := rows.Columns()
	columnLength := len(columns)
	cache := make([]interface{}, columnLength)
	for index, _ := range cache {
		var a interface{}
		cache[index] = &a
	}
	var list []map[string]interface{}
	for rows.Next() {
		err := rows.Scan(cache...)
		if err != nil {
			return nil, err
		}
		item := make(map[string]interface{})
		for i, data := range cache {
			item[columns[i]] = *data.(*interface{})
		}
		list = append(list, item)
	}
	return list, nil
}

func (m *mapping) RowsClose(auto bool) *mapping {
	m.rowsClose = auto
	return m
}

func (m *mapping) nameTag(name string) string {
	vv := []rune(name)
	buf := bytes.Buffer{}
	for _, v := range vv {
		if v < 97 || v > 122 {
			buf.WriteRune('_')
		}
		buf.WriteRune(v)
	}
	return buf.String()
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
