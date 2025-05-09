package repository

type IRepository[T any] interface {
	First(cond T) (*T, error)
	Find(conds ...[]interface{}) ([]T, error)
	Create(entity T) (*T, error)
	Update(entity T) error
	Delete(entity T) error
	Count(conds ...[]interface{}) (int64, error)
	Save(entity T) error
}
