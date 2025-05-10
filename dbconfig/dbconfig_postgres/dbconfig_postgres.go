package dbconfig_postgres

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/nttlong/regorm/dbconfig"
	"github.com/nttlong/regorm/dberrors"
	"github.com/nttlong/regorm/expr"

	"github.com/nttlong/regorm/expr/exprpostgres"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDbConfig struct {
	dbconfig.DbConfigBase
}
type PostgresStorage struct {
	db       *gorm.DB
	dbConfig dbconfig.IDbConfig
	parser   expr.IExpr
	dbName   string
}

func (c *PostgresDbConfig) GetConectionString(dbname string) string {
	strOps := ""
	for k, v := range c.Options {
		if k == "collation" {
			continue
		}
		strOps += k + "=" + v + "&"
	}
	if len(strOps) > 0 {
		strOps = strOps[:len(strOps)-1]
	}
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + dbname + "?" + strOps

}

func (c *PostgresDbConfig) GetConectionStringNoDatabase() string {
	//create postgres connection string without database name
	strOps := ""
	for k, v := range c.Options {
		strOps += k + "=" + v + "&"
	}
	if len(strOps) > 0 {
		strOps = strOps[:len(strOps)-1]
	}
	//"host=%s port=%d user=%s password=%s sslmode=disable"
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/"

}
func (c *PostgresDbConfig) PingDb() error {
	d := postgres.New(postgres.Config{
		DSN: c.GetConectionStringNoDatabase(),
	})
	_, err := gorm.Open(d, &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}

var (
	cacheAutoMigrate = make(map[string]bool)
	lockAutoMigrate  = sync.RWMutex{}
)

func (s *PostgresStorage) AutoMigrate(entity interface{}) error {
	key := s.GetDbName() + ":" + reflect.TypeOf(entity).Name()
	lockAutoMigrate.RLock()
	isAutoMigrated := cacheAutoMigrate[key]
	lockAutoMigrate.RUnlock()
	if isAutoMigrated {
		return nil
	}
	lockAutoMigrate.Lock()
	defer lockAutoMigrate.Unlock()
	if cacheAutoMigrate[key] {
		return nil
	}
	entities := s.dbConfig.GetAllModelsInEntity(entity)
	err := AutoMigrate(s.db, s.dbConfig, entities...)

	if err != nil {
		return err
	}
	cacheAutoMigrate[key] = true
	return nil
}
func (s *PostgresStorage) Save(entity interface{}) error {
	err := s.AutoMigrate(entity)
	if err != nil {
		return err
	}
	return s.db.Save(entity).Error
}
func (s *PostgresStorage) Create(entity interface{}) error {
	err := s.AutoMigrate(entity)
	if err != nil {
		return err
	}
	return s.db.Create(entity).Error
}
func (s *PostgresStorage) CreateInBatches(entities interface{}, batchSize int) error {
	typ := reflect.TypeOf(entities)
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
		err := s.AutoMigrate(reflect.New(typ).Interface())
		if err != nil {
			return err
		}

	}

	return s.db.CreateInBatches(entities, batchSize).Error
}
func (s *PostgresStorage) Exec(sql string, values ...interface{}) error {
	return s.db.Exec(sql, values...).Error
}
func (s *PostgresStorage) Find(dest interface{}, conds ...interface{}) error {
	typ := reflect.TypeOf(dest)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
		erMisMigrate := s.AutoMigrate(reflect.New(typ).Interface())
		if erMisMigrate != nil {
			return erMisMigrate
		}
	}

	if conds != nil || len(conds) > 0 {
		if reflect.TypeOf(conds[0]) == reflect.TypeOf("string") {
			strCon := conds[0].(string)
			node, err := s.parser.CompileExpr(strCon)
			if err == nil {
				conds[0] = node
				// //var newCnds []interface{} = conds[1:]
				// sliceType := reflect.SliceOf(typ)
				// ret := reflect.New(sliceType).Interface()
				// newCons := make([]interface{}, len(conds))
				// for i := 1; i < len(conds); i++ {
				// 	newCons[i] = conds[i]
				// }
				err := s.db.Find(dest, conds...).Error
				if err != nil {
					return err
				}
				// reflect.ValueOf(ret).Elem().Set(reflect.ValueOf(dest))

				return nil
			}
		}

	}
	return s.db.Find(dest, conds).Error
}

func (s *PostgresStorage) Update(entity interface{}, conds ...interface{}) error {
	erMigrate := s.AutoMigrate(entity)
	if erMigrate != nil {
		return erMigrate
	}

	if conds != nil || len(conds) > 0 {
		if reflect.TypeOf(conds[0]) == reflect.TypeOf("string") {
			strCon := conds[0].(string)
			node, err := s.parser.CompileExpr(strCon)
			if err == nil {
				conds[0] = node
				return s.db.Model(entity).Where(node, conds[1:]).Updates(entity).Error
			}
		}

	}
	return s.db.Model(entity).Updates(entity).Error
}

func (s *PostgresStorage) First(dest interface{}, conds ...interface{}) error {

	erMigrate := s.AutoMigrate(dest)
	if erMigrate != nil {
		return erMigrate
	}
	//parse condition
	if conds != nil || len(conds) > 0 {
		if reflect.TypeOf(conds[0]) == reflect.TypeOf("string") {
			strCon := conds[0].(string)
			node, err := s.parser.CompileExpr(strCon)
			if err == nil {
				conds[0] = node
				return s.db.First(dest, conds).Error

			}
		}

	}

	return s.db.First(dest, conds).Error
}
func (s *PostgresStorage) Delete(value interface{}, conds ...interface{}) error {
	erMigrate := s.AutoMigrate(value)
	if erMigrate != nil {
		return erMigrate
	}
	if conds != nil || len(conds) > 0 {
		if reflect.TypeOf(conds[0]) == reflect.TypeOf("string") {
			strCon := conds[0].(string)
			node, err := s.parser.CompileExpr(strCon)
			if err == nil {
				conds[0] = node
				return s.db.Delete(value, conds).Error

			}
		}

	}
	return s.db.Delete(value, conds).Error
}
func (s *PostgresStorage) Count(entity interface{}, conds ...interface{}) (int64, error) {
	erMigrate := s.AutoMigrate(entity)
	if erMigrate != nil {
		return 0, erMigrate
	}
	var ret int64
	if conds != nil || len(conds) > 0 {
		if reflect.TypeOf(conds[0]) == reflect.TypeOf("string") {
			strCon := conds[0].(string)
			node, err := s.parser.CompileExpr(strCon)
			if err == nil {
				conds[0] = node
				err := s.db.Model(entity).Where(node, conds[1:]...).Count(&ret).Error
				if err != nil {
					return 0, err
				}
				return ret, nil

			}
		}
	}

	errL := s.db.Model(&entity).Count(&ret).Error
	if errL != nil {
		return 0, errL
	}
	return ret, nil
}
func (s *PostgresStorage) SetDbConfig(config dbconfig.IDbConfig) {
	s.dbConfig = config
}
func (s *PostgresStorage) GetDbConfig() dbconfig.IDbConfig {
	return s.dbConfig
}
func (s *PostgresStorage) GetDb() *gorm.DB {
	return s.db
}
func (c *PostgresStorage) GetParser() expr.IExpr {
	return c.parser
}
func (c *PostgresStorage) SetParser(parser expr.IExpr) {
	c.parser = parser
}
func (c *PostgresStorage) GetDbName() string {
	return c.dbName
}

var (
	cacheGetStorage = make(map[string]dbconfig.IStorage)
	lockGetStorage  = sync.RWMutex{}
)

func (c *PostgresDbConfig) GetStorage(dbName string) (dbconfig.IStorage, error) {
	//check if storage is cached
	lockGetStorage.RLock()
	storage := cacheGetStorage[dbName]
	lockGetStorage.RUnlock()
	if storage != nil {
		return storage, nil
	}
	lockGetStorage.Lock()
	defer lockGetStorage.Unlock()
	if cacheGetStorage[dbName] != nil {
		return cacheGetStorage[dbName], nil
	}
	//create new storage
	storage, err := c.createStorage(dbName)
	if err != nil {
		return nil, err
	}
	cacheGetStorage[dbName] = storage
	return storage, nil
}

func (c *PostgresDbConfig) createStorage(dbName string) (dbconfig.IStorage, error) {

	err := c.PingDb()
	if err != nil {
		return nil, err
	}
	if err = c.createDbIfNotExist(dbName); err != nil {
		return nil, err
	}
	dns := c.GetConectionString(dbName)
	d, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{
		db:       d,
		dbConfig: c,
		parser:   exprpostgres.New(),
		dbName:   dbName,
	}, nil

}
func (c *PostgresDbConfig) TranslateError(err error, entity interface{}, action string) dberrors.DataActionError {
	//dupliate error translate
	//"duplicate key value violates unique constraint \"users_pkey\""
	errStr := err.Error()
	if strings.Contains(errStr, "duplicate key value violates unique constraint") {
		tableName := c.GetTableName(entity)

		if strings.Contains(errStr, tableName+"_pkey") {
			cols := c.GetAllColumnsInfoFromEntity(entity)
			refCols := make([]string, 0)
			for _, col := range cols {
				if col.IsPk {
					refCols = append(refCols, col.Name)
				}
			}
			return dberrors.DataActionError{
				Err:          err,
				Action:       action,
				Code:         dberrors.Duplicate,
				RefColumns:   refCols,
				RefTableName: tableName,
			}

		}

	}
	return dberrors.DataActionError{
		Err: err,
	}
}
func New() dbconfig.IDbConfig {
	return &PostgresDbConfig{}
}

var (
	cacheCreateDbIfNotExist = make(map[string]bool)
	lockCreateDbIfNotExist  = sync.RWMutex{}
)

func (c *PostgresDbConfig) CreateDbIfNotExist(dbname string) error {
	lockCreateDbIfNotExist.RLock()
	isCreated := cacheCreateDbIfNotExist[dbname]
	lockCreateDbIfNotExist.RUnlock()
	if isCreated {
		return nil
	}
	lockCreateDbIfNotExist.Lock()
	defer lockCreateDbIfNotExist.Unlock()
	if cacheCreateDbIfNotExist[dbname] {
		return nil
	}
	//create database if not exist
	err := c.createDbIfNotExist(dbname)
	if err != nil {
		return err
	}
	cacheCreateDbIfNotExist[dbname] = true
	return nil
}

// ======================================================================
func (c *PostgresDbConfig) createDbIfNotExist(dbname string) error {
	//create postgres connection string without database name
	dns := c.GetConectionStringNoDatabase()
	//create new connection
	d, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		return err
	}
	//create database if not exist
	/**
		CREATE DATABASE mydb
	WITH ENCODING 'UTF8'
	LC_COLLATE 'vi_VN.UTF-8'
	LC_CTYPE 'vi_VN.UTF-8';
	*/
	//check if collaction is set

	collate := c.DbConfigBase.Options["collation"]
	sql := fmt.Sprintf("CREATE DATABASE \"%s\" WITH ENCODING 'UTF8'", dbname)
	if collate != "" {
		collate = "vi_VN.UTF-8"
		sql = fmt.Sprintf("CREATE DATABASE \"%s\" WITH ENCODING 'UTF8' LC_COLLATE '%s' LC_CTYPE '%s'", dbname, collate, collate)
	}

	err = d.Exec(sql).Error
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		postgresSQLEnablecitextExtension := "CREATE EXTENSION IF NOT EXISTS citext;"
		newDbCnn := c.GetConectionString(dbname)
		newDb, err := gorm.Open(postgres.Open(newDbCnn), &gorm.Config{})
		if err != nil {
			return err
		}
		err = newDb.Exec(postgresSQLEnablecitextExtension).Error
		if err != nil {
			return err
		}
		return nil
	}
	postgresSQLEnablecitextExtension := "CREATE EXTENSION IF NOT EXISTS citext;"
	newDbCnn := c.GetConectionString(dbname)
	newDb, err := gorm.Open(postgres.Open(newDbCnn), &gorm.Config{})
	if err != nil {
		return err
	}
	err = newDb.Exec(postgresSQLEnablecitextExtension).Error
	if err != nil {
		return err
	}

	return nil
}
func AutoMigrate(db *gorm.DB, cfg dbconfig.IDbConfig, entities ...interface{}) error {
	for _, e := range entities {
		err := db.AutoMigrate(e)
		cols := cfg.GetAllColumnsInfoFromEntity(e)
		tablbName := cfg.GetTableName(e)
		for _, col := range cols {
			if col.DbType == "varchar" {
				//alert colum to citext
				dbColName := cfg.ToSnakeCase(col.Name)
				sqlAlterCol := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE citext", tablbName, dbColName)
				err := db.Exec(sqlAlterCol).Error
				if err != nil {
					fmt.Println(sqlAlterCol)
					fmt.Println(err)
				}

			}

			if err != nil {
				return err
			}
		}

	}
	return nil
}
