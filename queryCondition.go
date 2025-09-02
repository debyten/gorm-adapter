package gormadapter

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type queryCondition struct {
	condition   string
	arg         any
	flattenArgs bool
}

func (q queryCondition) applyArgs(args []any) []any {
	if q.flattenArgs {
		return append(args, q.arg.([]any)...)
	}
	return append(args, q.arg)
}

// queryEq is the fallback condition for when no other condition is provided.
func queryEq(value any) queryCondition {
	return queryCondition{
		condition: "= ?",
		arg:       value,
	}
}

// QueryIn creates a query condition for checking if a field's value matches any value in the provided list.
func QueryIn(values ...any) queryCondition {
	return queryCondition{
		condition: "IN (?)",
		arg:       values,
	}
}

// QueryNeq creates a query condition for checking if a field's value does not match the provided value.
func QueryNeq(value any) queryCondition {
	return queryCondition{
		condition: "!= ?",
		arg:       value,
	}
}

func QueryLike(value string) queryCondition {
	return queryCondition{
		condition: "LIKE ?",
		arg:       value,
	}
}

func QueryBetween(min, max any) queryCondition {
	return queryCondition{
		condition:   "BETWEEN ? AND ?",
		arg:         []any{min, max},
		flattenArgs: true,
	}
}

func QueryNotLike(value string) queryCondition {
	return queryCondition{
		condition: "NOT LIKE ?",
		arg:       value,
	}
}

func queryFromMap(conn *gorm.DB, query []map[string]any) (string, []any, bool) {
	if len(query) != 1 || len(query[0]) == 0 {
		return "", nil, false
	}
	m := query[0]
	queries := make([]string, 0, len(m))
	args := make([]any, 0, len(m))
	for k, v := range m {
		col := conn.NamingStrategy.ColumnName("", k)
		if !isSafeIdentifier.MatchString(col) {
			return "", nil, false
		}
		queryCond, ok := v.(queryCondition)
		if !ok {
			queryCond = queryEq(v)
		}

		queries = append(queries, fmt.Sprintf("%s %s", col, queryCond.condition))
		args = queryCond.applyArgs(args)
	}
	output := strings.Join(queries, " AND ")
	return output, args, true
}
