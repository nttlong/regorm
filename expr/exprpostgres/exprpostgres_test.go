package exprpostgres_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nttlong/regorm/expr/exprpostgres"

	"github.com/stretchr/testify/assert"
)

var testData = []string{
	"(year(CreatedOn) == ?) && (month(CreatedOn) == ?)->(date_part('year', created_on) = ?) AND (date_part('month', created_on) = ?)",
	"year(Id,code)->error",
	"year(Id)->date_part('year', id)",
	"UserName like '%%adm\\%in%%'->user_name like '%%adm\\%in%%'",
	"year()->date_part('year', id)",

	"month(ID)->date_part('month', id)",
	"day(ID)->date_part('day', id)",
	"hour(ID)->date_part('hour', id)",
	"minute(ID)->date_part('minute', id)",
	"second(ID)->date_part('second', id)",
	"ID->id",
}

func TestParseConditional(t *testing.T) {
	parser := exprpostgres.New()
	for _, test := range testData {
		input := strings.Split(test, "->")[0]
		ouput := strings.Split(test, "->")[1]
		expr, err := parser.CompileExpr(input)
		if err != nil {
			if ouput == "error" {
				t.Log(err)
				fmt.Print(err)

			}

		} else {
			assert.Equal(t, ouput, expr)
		}
	}

}
