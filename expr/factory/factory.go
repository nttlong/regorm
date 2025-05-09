package factory

import (
	"github.com/nttlong/regorm/expr"

	"github.com/nttlong/regorm/expr/exprpostgres"
)

func NewExpr(driver string) expr.IExpr {
	switch driver {
	case "postgres":
		return exprpostgres.New()
	default:
		panic("Unsupported driver: " + driver)
	}
}
