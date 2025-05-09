package dbconfig_mysql_test

import (
	"fmt"
	"testing"

	"github.com/nttlong/regorm/dbconfig/dbconfig_mysql"

	assert "github.com/stretchr/testify/assert"
)

var yamlFile = "E:/Docker/go/quicky-go/be/gormex/config.yaml"

func TestNew(t *testing.T) {
	cfg := dbconfig_mysql.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, "root:123456@tcp(localhost:3306)/", cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)

}
