package expr

import (
	"github.com/nttlong/regorm/expr/compiler"
)

type IBaseExpr interface {
	Compile(cond string) (*compiler.SimpleExprTree, error)
	GetStrExpr(node *compiler.SimpleExprTree) (string, error)
	SetResolver(resolver func(node *compiler.SimpleExprTree) error)
}

type BaseExpr struct {
	resolver func(node *compiler.SimpleExprTree) error
}

func (b *BaseExpr) Compile(cond string) (*compiler.SimpleExprTree, error) {
	return compiler.ParseExpr(cond)
}
func (b *BaseExpr) SetResolver(resolver func(node *compiler.SimpleExprTree) error) {
	b.resolver = resolver
}

func (b *BaseExpr) GetStrExpr(node *compiler.SimpleExprTree) (string, error) {
	if b.resolver == nil {
		panic("resolver is not set, please call SetResolver() first")
	}
	return compiler.Resolve(node, b.resolver)
}

var bBaseExpr IBaseExpr = &BaseExpr{}

func NewBaseExpr() IBaseExpr {
	return bBaseExpr
}

type IExpr interface {
	IBaseExpr
	// compiler to sqldb driver
	CompileExpr(expr string) (string, error)
}
