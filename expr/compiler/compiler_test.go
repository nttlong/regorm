package compiler_test

import (
	"strings"
	"testing"

	"github.com/nttlong/regorm/expr/compiler"

	"github.com/stretchr/testify/assert"
)

var testList []string = []string{
	"(year(CreatedOn) == ?) && (month(CreatedOn) == ?)->(year(CreatedOn) == ?) && (month(CreatedOn) == ?)",
	"(Code==? and Price<=?) or len(name)==?->(Code == ? and Price <= ?) or len(name) == ?",
	"(Code==? and Price<=?) || len(name)==?->(Code == ? and Price <= ?) || len(name) == ?",
	"(a+b)/c*d-f+sum(1m2,3m4,5m6)->(a + b) / c * d - f + sum(1m2, 3m4, 5m6)",
	"(Code123==? and Price<=?) or len(name)==?->(Code123 == ? and Price <= ?) or len(name) == ?",
	"max(salary, bonus)^2<=BasicSalary->max(salary, bonus) ^ 2 <= BasicSalary",
	"(concat(firstName,' ', lastName)) like '%?%'->(concat(firstName, ' ', lastName)) like '%?%'",
}

func TestTree(t *testing.T) {
	resolver := func(n *compiler.SimpleExprTree) error {
		// if n.Nt == "func" {
		// 	n.V = "pg." + n.V

		// }

		return nil
	}
	for _, s := range testList {
		Input := strings.Split(s, "->")[0]
		Output := strings.Split(s, "->")[1]
		fx, err := compiler.ParseExpr(Input)

		if err != nil && Output != "error" {
			t.Error(err)
		}
		r, err := compiler.Resolve(fx, resolver)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, Output, r)
	}
}
