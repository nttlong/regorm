// package repositoryposgres_test
package repositorypostgres_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/nttlong/regorm/dbconfig/dbconfig_postgres"

	"github.com/nttlong/regorm/repository/repositorypostgres"
	_ "github.com/nttlong/regorm/repository/repositorypostgres"

	"github.com/stretchr/testify/assert"
)

var yamlFile = "E:/Docker/go/gormdb/test-config.yml"
var cnnNoDb = "postgres://postgres:123456@localhost:5432/"

type BaseInfo struct {
	CreatedOn time.Time  `gorm:"type:timestamp;default:current_timestamp"`
	UpdatedOn *time.Time `gorm:"type:timestamp;default:current_timestamp"`
	CreatedBy string     `gorm:"type:varchar(50);default:admin"`
	UpdatedBy *string    `gorm:"type:varchar(50);"`
}
type User struct {
	BaseInfo
	ID       string `gorm:"type:varchar(36);primary_key"`
	Username string `gorm:"type:varchar(50);uniqueIndex:idx_name_username"`
	Password string `gorm:"type:varchar(256);"`
}
type PersonalInfo struct {
	ID        string `gorm:"type:varchar(36);primary_key"`
	FirstName string `gorm:"type:varchar(50);"`
	LastName  string `gorm:"type:varchar(50);"`
	BirthDay  string `gorm:"type:date;index:idx_birthday"`
}
type Working struct {
	ID        string `gorm:"type:varchar(36);primary_key"`
	StartDate string `gorm:"type:date;"`
	EndDate   string `gorm:"type:date;"`
	User      *User  `gorm:"foreignKey:ID"`
}
type Emp struct {
	ID           string `gorm:"type:varchar(36);primary_key"`
	DepartmentID string `gorm:"type:varchar(36);index:idx_department_id"`

	User  *User      `gorm:"foreignKey:ID"`
	Works []*Working `gorm:"foreignKey:ID"`

	Info *PersonalInfo `gorm:"foreignKey:ID"`
}
type Dept struct {
	ID   string `gorm:"type:varchar(36);primary_key"`
	Name string `gorm:"type:varchar(50);uniqueIndex:idx_name_name"`
	Emps []*Emp `gorm:"foreignKey:DepartmentID"`
}

func TestEntitiesPostgres(t *testing.T) {
	t.Log("Testing entities postgres")
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, cnnNoDb, cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)
	s, err := cfg.GetStorage("test")
	assert.NoError(t, err)
	// err = s.Delete(&User{ID: "123456"})
	// assert.NoError(t, err)
	userRepo := repositorypostgres.New[User](s)
	u, err := userRepo.Create(User{ID: "123456", Username: "admin", Password: "123456"})
	if err != nil {
		t.Error(err)
	} else {
		t.Log(u)
		assert.Equal(t, "admin", u.Username)
		assert.Equal(t, "123456", u.Password)
	}

	u, err = userRepo.First(User{ID: "123456"})
	if err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "admin", u.Username)
		assert.Equal(t, "123456", u.Password)
	}

}
