package scalar

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/squirrel"
)

type Querier interface {
	Query(query squirrel.SelectBuilder, name string) squirrel.SelectBuilder
}

func Query(query squirrel.SelectBuilder, args interface{}) squirrel.SelectBuilder {
	v := reflect.ValueOf(args)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		vField := v.Field(i)
		tField := t.Field(i)

		if tField.Name == "Search" {
			value, ok := vField.Interface().(*string)
			if !ok || value == nil {
				continue
			}

			exprs := []squirrel.Sqlizer{}
			for i := 0; i < v.NumField(); i++ {
				vField := v.Field(i)
				tField := t.Field(i)
				tag, ok := tField.Tag.Lookup("db")
				if !ok || tag == "-" {
					continue
				}
				switch vField.Interface().(type) {
				case Regex, *Regex:
					exprs = append(exprs, squirrel.Expr(fmt.Sprintf("%s like ?", tag), fmt.Sprint("%", *value, "%")))

				}
			}
			query = query.Where(squirrel.Or(exprs))

		} else if tField.Name == "Sort" {
			cols, ok := vField.Interface().(*string)
			if !ok || cols == nil {
				continue
			}
			for _, col := range strings.Split(*cols, ",") {
				col = strings.TrimSpace(col)
				desc := ""
				if col[0] == '!' {
					desc = " desc"
					col = col[1:]
				}

				query = query.OrderBy(col + desc)
			}
		}

		tag, ok := tField.Tag.Lookup("db")
		if !ok || tag == "-" || vField.IsNil() {
			continue
		}

		switch value := vField.Interface().(type) {
		case Querier:
			query = value.Query(query, tag)
		default:
			if value != nil {
				query = query.Where(squirrel.Eq{tag: value})
			}
		}

	}
	return query
}
