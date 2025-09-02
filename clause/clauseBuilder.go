package clause

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/debyten/database"
	"gorm.io/gorm"
)

var isSafeIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)?$`)

// Builder provides a fluent API to compose ordered query clauses with logical connectors.
// Example:
//
//	b := clause.New().
//	    Eq("mycol", v1).And().
//	    In("mycol2", v2a, v2b).Or().
//	    Between("date", d1, d2)
//	clauses := b.Build()
//
// The resulting clauses preserve insertion order and can be passed to Crud methods using varargs expansion.
// The gorm adapter's queryFromMap understands the special operator markers added by And()/Or().
type Builder struct {
	clauses []database.QueryClauses
}

// New creates a new Builder.
func New() *Builder { return &Builder{clauses: make([]database.QueryClauses, 0)} }

// Build returns the ordered list of QueryClauses to be passed to CRUD methods.
func (b *Builder) Build() []database.QueryClauses {
	return b.clauses
}

// Eq adds an equality condition for the given column.
func (b *Builder) Eq(column string, value any) *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{column: Eq(value)})
	return b
}

// Neq adds a not-equal condition for the given column.
func (b *Builder) Neq(column string, value any) *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{column: Neq(value)})
	return b
}

// In adds an IN condition for the given column.
func (b *Builder) In(column string, values ...any) *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{column: In(values...)})
	return b
}

// Like adds a LIKE condition for the given column.
func (b *Builder) Like(column string, value string) *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{column: Like(value)})
	return b
}

// NotLike adds a NOT LIKE condition for the given column.
func (b *Builder) NotLike(column string, value string) *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{column: NotLike(value)})
	return b
}

// Between adds a BETWEEN condition for the given column.
func (b *Builder) Between(column string, min, max any) *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{column: Between(min, max)})
	return b
}

// Gt adds a greater than condition for the given column.
func (b *Builder) Gt(column string, value any) *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{column: Gt(value)})
	return b
}

// Gte adds a greater than or equal condition for the given column.
func (b *Builder) Gte(column string, value any) *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{column: Gte(value)})
	return b
}

// Lt adds a less than condition for the given column.
func (b *Builder) Lt(column string, value any) *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{column: Lt(value)})
	return b
}

// Lte adds a less than or equal condition for the given column.
func (b *Builder) Lte(column string, value any) *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{column: Lte(value)})
	return b
}

// And adds a logical AND connector between conditions.
func (b *Builder) And() *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{"$and": operatorArg("AND")})
	return b
}

// Or adds a logical OR connector between conditions.
func (b *Builder) Or() *Builder {
	b.clauses = append(b.clauses, database.QueryClauses{"$or": operatorArg("OR")})
	return b
}

// Build builds a SQL WHERE clause from the provided query clauses.
// It supports two modes:
//  1. Legacy mode: a single map[string]ConditionArg, combined with AND (order not guaranteed).
//  2. Ordered mode: multiple QueryClauses provided as a slice, preserving order and supporting
//     logical connectors via special maps: {"$and": ...} and {"$or": ...}.
func Build(conn *gorm.DB, query []database.QueryClauses) (string, []any, bool) {
	if len(query) == 0 {
		return "", nil, false
	}

	var tokens []string
	args := make([]any, 0, 4)
	connector := "AND" // default connector between clauses

	for _, m := range query {
		if len(m) == 0 {
			continue
		}
		// Detect operator markers ("$and"/"$or"). They come as a single-entry map.
		if op, ok := extractOperator(m); ok {
			// Only set connector if there is already at least one condition emitted; otherwise ignore.
			if len(tokens) > 0 {
				connector = op
			}
			continue
		}

		// Normal condition map: can contain one or more field conditions; combine them with AND locally.
		localParts := make([]string, 0, len(m))
		for k, v := range m {
			col := conn.NamingStrategy.ColumnName("", k)
			if !isSafeIdentifier.MatchString(col) {
				return "", nil, false
			}
			localParts = append(localParts, fmt.Sprintf("%s %s", col, v.Condition()))
			args = v.JoinArgs(args)
		}
		local := strings.Join(localParts, " AND ")

		if len(tokens) > 0 {
			tokens = append(tokens, connector)
		}
		tokens = append(tokens, local)
		// Reset the connector to default after it's used, so that missing connectors imply AND
		connector = "AND"
	}

	if len(tokens) == 0 {
		return "", nil, false
	}

	return strings.Join(tokens, " "), args, true
}

func hasOperatorMarker(m database.QueryClauses) bool {
	if len(m) != 1 {
		return false
	}
	for k := range m {
		return k == "$and" || k == "$or"
	}
	return false
}

func extractOperator(m database.QueryClauses) (string, bool) {
	if len(m) != 1 {
		return "", false
	}
	for k := range m {
		switch k {
		case "$and":
			return "AND", true
		case "$or":
			return "OR", true
		}
	}
	return "", false
}
