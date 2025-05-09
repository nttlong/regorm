package repositorypostgres

import (
	"reflect"

	"github.com/nttlong/regorm/dbconfig"
	_ "github.com/nttlong/regorm/dbconfig"
	"github.com/nttlong/regorm/dbconfig/dbconfig_postgres"
	_ "github.com/nttlong/regorm/dbconfig/dbconfig_postgres"
	"github.com/nttlong/regorm/repository"
	_ "github.com/nttlong/regorm/repository"
)

type RepositoryPostgres[T any] struct {
	storage dbconfig.IStorage
}

func (e *RepositoryPostgres[T]) First(cond T) (*T, error) {

	err := e.storage.First(&cond)
	if err != nil {
		return nil, err
	}
	return &cond, nil
}
func (e *RepositoryPostgres[T]) Find(conds ...[]interface{}) ([]T, error) {
	var entities []T
	err := e.storage.Find(&entities, conds)
	return entities, err
}
func (e *RepositoryPostgres[T]) Create(entity T) (*T, error) {
	err := e.storage.Create(&entity)
	if err != nil {
		return nil, err
	} else {
		return &entity, nil
	}
}
func (e *RepositoryPostgres[T]) Update(entity T) error {
	return e.storage.Update(&entity)
}
func (e *RepositoryPostgres[T]) Delete(entity T) error {
	return e.storage.Delete(&entity)
}
func (e *RepositoryPostgres[T]) Count(conds ...[]interface{}) (int64, error) {
	var zero T
	t := reflect.TypeOf(zero)
	return e.storage.Count(t, conds)

}
func (e *RepositoryPostgres[T]) Save(entity T) error {
	return e.storage.Save(&entity)
}

func New[T any](Storage dbconfig.IStorage) repository.IRepository[T] {
	// get type of Storage
	typ := reflect.TypeOf(Storage)

	if typ != reflect.TypeOf(new(dbconfig_postgres.PostgresStorage)) {
		panic("Storage must be a pointer to dbconfig_postgres.PostgresStorage")
	}
	var zero T

	Storage.AutoMigrate(&zero)

	return &RepositoryPostgres[T]{
		storage: Storage,
	}
}
