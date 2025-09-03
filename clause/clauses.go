package clause

import "github.com/debyten/database"

type conditionArg struct {
	condition   string
	arg         any
	flattenArgs bool
}

func (q conditionArg) Condition() string {
	return q.condition
}

func (q conditionArg) JoinArgs(args []any) []any {
	if q.flattenArgs {
		return append(args, q.arg.([]any)...)
	}
	return append(args, q.arg)
}

func Eq(column string, value any) database.QueryClauses {
	return database.QueryClauses{column: conditionArg{condition: "= ?", arg: value}}
}

func Lt(column string, value any) database.QueryClauses {
	return database.QueryClauses{column: conditionArg{condition: "< ?", arg: value}}
}

func Lte(column string, value any) database.QueryClauses {
	return database.QueryClauses{column: conditionArg{condition: "<= ?", arg: value}}
}

func Gt(column string, value any) database.QueryClauses {
	return database.QueryClauses{column: conditionArg{condition: "> ?", arg: value}}
}

func Gte(column string, value any) database.QueryClauses {
	return database.QueryClauses{column: conditionArg{condition: ">= ?", arg: value}}
}

// In creates a query condition for checking if a field's value matches any value in the provided list.
func In(column string, values ...any) database.QueryClauses {
	return database.QueryClauses{column: conditionArg{condition: "IN (?)", arg: values}}
}

// Neq creates a query condition for checking if a field's value does not match the provided value.
func Neq(column string, value any) database.QueryClauses {
	return database.QueryClauses{column: conditionArg{condition: "!= ?", arg: value}}
}

func Like(column string, value string) database.QueryClauses {
	return database.QueryClauses{column: conditionArg{condition: "LIKE ?", arg: value}}
}

func Between(column string, min, max any) database.QueryClauses {
	return database.QueryClauses{column: conditionArg{condition: "BETWEEN ? AND ?", arg: []any{min, max}, flattenArgs: true}}
}

func NotLike(column string, value string) database.QueryClauses {
	return database.QueryClauses{column: conditionArg{condition: "NOT LIKE ?", arg: value}}
}

// operatorArg is an internal ConditionArg used to represent logical operators in the builder.
type operatorArg string

func (o operatorArg) Condition() string         { return string(o) }
func (o operatorArg) JoinArgs(args []any) []any { return args }
