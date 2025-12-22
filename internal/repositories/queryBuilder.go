package repositories

import (
	"fmt"
	"reflect"
	"strings"
)

type Column struct {
	Table string
	Name  string
}

type Condition struct {
	Expr string
	Args []any
}

type QueryBuilder struct {
	table      string
	columns    []string
	setters    []string
	conditions []Condition
	args       []any

	isInsert bool
	isUpdate bool
}

func Col(model any, fieldName string) Column {
	t := reflect.TypeOf(model)
	f, ok := t.FieldByName(fieldName)
	if !ok {
		panic("field not found: " + fieldName)
	}

	return Column{
		Table: strings.ToLower(t.Name()) + "s",
		Name:  f.Tag.Get("db"),
	}
}

func (c Column) Eq(val any) Condition {
	return Condition{
		Expr: fmt.Sprintf("%s.%s = ?", c.Table, c.Name),
		Args: []any{val},
	}
}

func From(model any) *QueryBuilder {
	t := reflect.TypeOf(model)
	return &QueryBuilder{
		table: strings.ToLower(t.Name()) + "s",
	}
}

func (qb *QueryBuilder) SelectModel(model any) *QueryBuilder {
	t := reflect.TypeOf(model)
	var cols []string
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("db")
		if tag == "" {
			continue
		}
		cols = append(cols, fmt.Sprintf("%s.%s", qb.table, tag))
	}
	qb.columns = append(qb.columns, cols...)
	return qb
}

func (qb *QueryBuilder) SelectAll() *QueryBuilder {
	qb.columns = append(qb.columns, "*")
	return qb
}

func (qb *QueryBuilder) Where(cond Condition) *QueryBuilder {
	qb.conditions = append(qb.conditions, cond)
	return qb
}

func (qb *QueryBuilder) Build() (string, []any) {
	if qb.isInsert {
		placeholders := make([]string, len(qb.columns))
		for i := range placeholders {
			placeholders[i] = "?"
		}

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			qb.table,
			strings.Join(qb.columns, ", "),
			strings.Join(placeholders, ", "),
		)

		return replaceArgs(query), qb.args
	}

	if qb.isUpdate {
		query := fmt.Sprintf(
			"UPDATE %s SET %s",
			qb.table,
			strings.Join(qb.setters, ", "),
		)

		var args []any
		args = append(args, qb.args...)

		if len(qb.conditions) > 0 {
			var conds []string
			for _, c := range qb.conditions {
				conds = append(conds, c.Expr)
				args = append(args, c.Args...)
			}
			query += " WHERE " + strings.Join(conds, " AND ")
		}

		return replaceArgs(query), args
	}

	var args []any
	var conds []string
	for _, c := range qb.conditions {
		conds = append(conds, c.Expr)
		args = append(args, c.Args...)
	}
	query := fmt.Sprintf(
		"SELECT %s FROM %s",
		strings.Join(qb.columns, ", "),
		qb.table,
	)
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}

	return replaceArgs(query), args
}

func replaceArgs(query string) string {
	var sb strings.Builder
	sb.Grow(len(query))

	paramIndex := 1
	for _, ch := range query {
		if ch == '?' {
			sb.WriteString(fmt.Sprintf("$%d", paramIndex))
			paramIndex++
		} else {
			sb.WriteRune(ch)
		}
	}

	return sb.String()
}

func InsertModel(model any) *QueryBuilder {
	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	var cols []string
	var args []any

	table := strings.ToLower(t.Name()) + "s"

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("db")
		if tag == "" || tag == "id" {
			continue
		}

		cols = append(cols, tag)
		args = append(args, v.Field(i).Interface())
	}

	qb := &QueryBuilder{
		table:    table,
		columns:  cols,
		args:     args,
		isInsert: true,
	}

	return qb
}

func UpdateModel(model any) *QueryBuilder {
	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	table := strings.ToLower(t.Name()) + "s"

	var setters []string
	var args []any

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("db")

		if tag == "" || tag == "id" {
			continue
		}

		setters = append(setters, fmt.Sprintf("%s = ?", tag))
		args = append(args, v.Field(i).Interface())
	}

	qb := &QueryBuilder{
		table:    table,
		setters:  setters,
		args:     args,
		isUpdate: true,
	}

	return qb
}
