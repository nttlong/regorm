package exprpostgres

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/nttlong/regorm/expr"

	"github.com/nttlong/regorm/expr/compiler"
)

type ExprPostgres struct {
	expr.IBaseExpr
}

func (e *ExprPostgres) CompileExpr(expr string) (string, error) {
	n, err := e.Compile(expr)
	if err != nil {
		return "", errors.New(fmt.Sprintf("\nerror compiling expression: %s\t %s", err.Error(), expr))
	}

	r, err := e.GetStrExpr(n)
	if err != nil {
		return "", errors.New(fmt.Sprintf("\nerror compiling expression: %s\t %s", err.Error(), expr))
	}
	return r, nil
}

var exprPostgres = &ExprPostgres{}
var once sync.Once

func New() expr.IExpr {
	once.Do(func() {
		exprPostgres = &ExprPostgres{}
		exprPostgres.IBaseExpr = expr.NewBaseExpr()
		exprPostgres.SetResolver(resolvePostgres)
	})

	return exprPostgres

}

var compilerOp = map[string]string{
	"&&": "AND",
	"||": "OR",
	"!":  "NOT",
	"==": "=",
}

func resolvePostgres(n *compiler.SimpleExprTree) error {
	if p, ok := compilerOp[n.Op]; ok {
		n.Op = p
	}
	if n.Nt == "field" {
		if !compiler.IsValidColumnName(n.V) {
			return fmt.Errorf("invalid column name: %s", n.V)
		}
		n.V = compiler.ToSnakeCase(n.V)
	}
	if n.Nt == "func" {
		fncName := strings.ToLower(n.V)
		switch fncName {
		case "year", "month", "day", "hour", "minute", "second":
			err := compileTimeFunc(n)
			if err != nil {
				return err
			}

		}

		// TODO: implement resolver for postgres
	}
	return nil
}
func compileTimeFunc(n *compiler.SimpleExprTree) error {
	/**
	"year(ID)->date_part('year',id)",
	"month(ID)->date_part('month',id)",
	"day(ID)->date_part('day',id)",
	"hour(ID)->date_part('hour',id)",
	"minute(ID)->date_part('minute',id)",
	"second(ID)->date_part('second',id)",
	*/
	//fP := fmt.Sprint("date_part('%s',id)")
	if n.Ns == nil || len(n.Ns) != 1 {
		return errors.New(fmt.Sprintf("\nnvalid function call. Function %s requires only one argument", n.V))
	}
	oldFnName := n.V
	n.V = "date_part"
	oldNs := n.Ns
	n.Ns = []*compiler.SimpleExprTree{
		{
			Nt: "const",
			V:  "'" + oldFnName + "'",
		},
	}
	// append oldNs to n.Ns

	n.Ns = append(n.Ns, oldNs...)

	return nil

}
