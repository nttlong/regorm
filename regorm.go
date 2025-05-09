package regorm

import (
	"sync"

	"github.com/nttlong/regorm/dbconfig"
	_ "github.com/nttlong/regorm/dbconfig"
	"github.com/nttlong/regorm/dbconfig/dbconfig_mysql"
	_ "github.com/nttlong/regorm/dbconfig/dbconfig_mysql"
	"github.com/nttlong/regorm/dbconfig/dbconfig_postgres"
	_ "github.com/nttlong/regorm/dbconfig/dbconfig_postgres"
)

var (
	cachdBconfig      map[string]dbconfig.IDbConfig = make(map[string]dbconfig.IDbConfig)
	lockCacheDbconfig                               = new(sync.RWMutex)
)

func New(driverName string) dbconfig.IDbConfig {
	//check cache
	lockCacheDbconfig.RLock()
	ret, ok := cachdBconfig[driverName]
	lockCacheDbconfig.RUnlock()
	if ok {
		return ret
	}

	//create new dbconfig
	lockCacheDbconfig.Lock()
	defer lockCacheDbconfig.Unlock()
	if ret, ok = cachdBconfig[driverName]; ok {
		return ret
	}
	if driverName == "postgres" {
		ret = &dbconfig_postgres.PostgresDbConfig{}
	} else if driverName == "mysql" {
		ret = &dbconfig_mysql.MySqlDbConfig{}
	} else {
		panic("not support driver")
	}
	cachdBconfig[driverName] = ret
	return ret

}
