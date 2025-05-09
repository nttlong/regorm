package dbconfig_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"
	"vngom/gormex/dbconfig"

	"github.com/stretchr/testify/assert"
)

type bases struct {
	CreatedOn time.Time  `gorm:"type:datetime"`
	UpdatedOn *time.Time `gorm:"type:datetime"`
}
type TestChildStruct struct {
	Id   string `gorm:"type:char(36);primaryKey"`
	Name string `gorm:"type:varchar(255);not null"`
	Age  int    `gorm:"type:int;default:0"`
}
type testStruct struct {
	bases
	Name          string `gorm:"type:char(36);primaryKey"`
	Code          string `gorm:"type:char(36);uniqueIndex:idx_code_name"`
	Age           int    `gorm:"type:int;default:0"`
	NoTgField     string
	DepartmetCode string            `gorm:"type:varchar(255);not null"`
	BirdBirthday  *time.Time        `gorm:"type:date"`
	JoinYear      int               `gorm:"type:int;default:0;index:idx_join_at"`
	JoinMonth     int               `gorm:"type:int;default:0;index:idx_join_at"`
	JoinDay       int               `gorm:"type:int;default:0;index:idx_join_at"`
	Child         TestChildStruct   `gorm:"foreignKey:Id"`
	Children      []TestChildStruct `gorm:"foreignKey:Id"`
}

var yamlFile = "E:/Docker/go/quicky-go/be/gormex/config.yaml"

func TestLoadFromYamlFile(t *testing.T) {
	cfg := dbconfig.NewDbConfigBase()
	err := cfg.LoadFromYamlFile(yamlFile)
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg.ToJSON())

	assert.True(t, cfg.CheckIsLoaded())

}
func TestGetTagInfoOfField(t *testing.T) {

	typ := reflect.TypeOf(testStruct{})
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	cfg := dbconfig.NewDbConfigBase()
	f := typ.Field(0)
	ret := cfg.GetColumInfoOfField(f)
	// first field in testStruct is bases, which has no tag info
	assert.True(t, ret == nil)
	f = typ.Field(1)
	ret = cfg.GetColumInfoOfField(f)
	assert.Equal(t, ret.Name, "name")
	assert.Equal(t, ret.DbType, "char")

	assert.Equal(t, ret.Length, 36)
	assert.Equal(t, ret.IsPk, true)
	assert.Equal(t, ret.IndexName, "")
	assert.Equal(t, ret.IsUnique, false)
	assert.Equal(t, ret.Typ, f)
	assert.Equal(t, ret.Content, "type:char(36);primaryKey")
	f = typ.Field(2)
	ret = cfg.GetColumInfoOfField(f)
	assert.Equal(t, ret.Name, "code")
	assert.Equal(t, ret.DbType, "char")

	assert.Equal(t, ret.Length, 36)
	assert.Equal(t, ret.IsPk, false)
	assert.Equal(t, ret.IndexName, "idx_code_name")
	assert.Equal(t, ret.IsUnique, true)
	assert.Equal(t, ret.Typ, f)
	assert.Equal(t, ret.Content, "type:char(36);uniqueIndex:idx_code_name")

	ret7 := cfg.GetColumInfoOfField(typ.Field(7))
	ret8 := cfg.GetColumInfoOfField(typ.Field(8))
	ret9 := cfg.GetColumInfoOfField(typ.Field(9))
	assert.True(t, ret7.IndexName == ret8.IndexName && ret8.IndexName == ret9.IndexName && ret7.IndexName == "idx_join_at")
	ret10 := cfg.GetColumInfoOfField(typ.Field(10))
	assert.Nil(t, ret10)

}
func TestGetColumnsInfo(t *testing.T) {
	// typ := reflect.TypeOf(testStruct{})
	// if typ.Kind() == reflect.Ptr {
	// 	typ = typ.Elem()
	// }
	cols := dbconfig.NewDbConfigBase().GetAllColumnsInfoFromEntity(&testStruct{})
	cols2 := dbconfig.NewDbConfigBase().GetAllColumnsInfoFromEntity(&testStruct{})
	//make sure the result is the same instance

	assert.Equal(t, cols, cols2)

	assert.Equal(t, cols, cols2)
	assert.True(t, len(cols) == 9)

}
func TestGetAllModelsInEntity(t *testing.T) {
	cfg := dbconfig.NewDbConfigBase()
	models := cfg.GetAllModelsInEntity(&testStruct{})
	for _, m := range models {
		fmt.Println(m)
		cls := cfg.GetAllColumnsInfoFromEntity(m)
		assert.Greater(t, len(cls), 0)
	}
	assert.True(t, len(models) == 1)
	assert.Equal(t, models[0], &testStruct{})
}

// toSnakeCase chuyển đổi chuỗi thành định dạng snake_case

func TestToSnakeCase(t *testing.T) {
	cfg := dbconfig.NewDbConfigBase()
	testData := []string{"Name", "DepartmetCode", "BirdBirthday", "JoinYear", "JoinMonth", "JoinDay", "Child", "Children", "ID", "Emp"}
	expected := []string{"name", "departmet_code", "bird_birthday", "join_year", "join_month", "join_day", "child", "children", "id", "emp"}
	for i, s := range testData {
		assert.Equal(t, expected[i], cfg.ToSnakeCase(s))
	}

}
