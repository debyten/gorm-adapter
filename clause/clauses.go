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

func Eq(value any) database.ConditionArg {
	return conditionArg{
		condition: "= ?",
		arg:       value,
	}
}

func Lt(value any) database.ConditionArg {
	return conditionArg{
		condition: "< ?",
		arg:       value,
	}
}

func Lte(value any) database.ConditionArg {
	return conditionArg{
		condition: "<= ?",
		arg:       value,
	}
}

func Gt(value any) database.ConditionArg {
	return conditionArg{
		condition: "> ?",
		arg:       value,
	}
}

func Gte(value any) database.ConditionArg {
	return conditionArg{
		condition: ">= ?",
		arg:       value,
	}
}

// In creates a query condition for checking if a field's value matches any value in the provided list.
func In(values ...any) database.ConditionArg {
	return conditionArg{
		condition: "IN (?)",
		arg:       values,
	}
}

// Neq creates a query condition for checking if a field's value does not match the provided value.
func Neq(value any) database.ConditionArg {
	return conditionArg{
		condition: "!= ?",
		arg:       value,
	}
}

func Like(value string) database.ConditionArg {
	return conditionArg{
		condition: "LIKE ?",
		arg:       value,
	}
}

func Between(min, max any) database.ConditionArg {
	return conditionArg{
		condition:   "BETWEEN ? AND ?",
		arg:         []any{min, max},
		flattenArgs: true,
	}
}

func NotLike(value string) database.ConditionArg {
	return conditionArg{
		condition: "NOT LIKE ?",
		arg:       value,
	}
}

// operatorArg is an internal ConditionArg used to represent logical operators in the builder.
type operatorArg string

func (o operatorArg) Condition() string         { return string(o) }
func (o operatorArg) JoinArgs(args []any) []any { return args }
