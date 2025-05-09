package dbconfig_postgres_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/nttlong/regorm/dbconfig/dbconfig_postgres"
	"github.com/nttlong/regorm/dberrors"

	assert "github.com/stretchr/testify/assert"
)

var yamlFile = "E:/Docker/go/quicky-go/be/gormex/postgres.yaml"
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

func TestNew(t *testing.T) {
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, cnnNoDb, cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)

}
func TestCreateDatabaseIfNotEx(t *testing.T) {
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, cnnNoDb, cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)
	err = cfg.CreateDbIfNotExist("test")
	assert.NoError(t, err)

}
func TestEntities(t *testing.T) {
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	e := cfg.GetAllModelsInEntity(&Emp{})
	assert.Equal(t, 3, len(e))
	e = cfg.GetAllModelsInEntity(&Dept{})
	assert.Equal(t, 2, len(e))

}
func TestGetStorage(t *testing.T) {
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, cnnNoDb, cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)
	s, err := cfg.GetStorage("test")
	assert.NoError(t, err)
	fmt.Println(s)
}

func TestGetStorageAutoMigrate(t *testing.T) {
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, cnnNoDb, cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)
	s, err := cfg.GetStorage("test")
	assert.NoError(t, err)
	err = s.AutoMigrate(&Emp{})
	if err != nil {
		panic(err)
	}
	db := s.GetDb()
	db.Save(&User{
		ID:       "123456",
		Username: "admin",
		Password: "123456",
	})
	// check if table emp is created in db including user table
	rd := s.GetDb().Exec("SELECT * FROM emps")

	assert.NoError(t, rd.Error)

	var user Emp
	err = db.Raw("SELECT * FROM emps WHERE FirstName = 'admin'").
		Scan(&user).Error
	assert.Error(t, err)
	rd = s.GetDb().Exec("SELECT * FROM users where username = ?", "admin")
	assert.NoError(t, rd.Error)
	rd = s.GetDb().Exec("SELECT * FROM users where username like ?", "%%ad%%")
	assert.NoError(t, rd.Error)
	assert.Equal(t, int64(1), rd.RowsAffected)

	getU := &Emp{}

	u1 := &User{}
	r1 := db.Model(&User{}).Where("username = ?", "admin").First(u1)
	assert.NoError(t, r1.Error)
	u2 := &User{}
	r2 := db.Model(&User{}).Where("username like ?", "%%ad%%").First(u2)
	assert.NoError(t, r2.Error)
	assert.Equal(t, u1.ID, u2.ID)

	rd2 := s.GetDb().Model(&Emp{}).
		Where(`first_name = ?`, "username").
		First(&getU).Error

	assert.NoError(t, rd2)
	fmt.Println(s)
	pInfo := &PersonalInfo{}
	rd2 = s.GetDb().Model(&PersonalInfo{}).
		Where(`date_part('year', birth_day) = ?`, 2025).
		First(&pInfo).Error

	assert.NoError(t, rd2)
}
func TestDeleteData(t *testing.T) {
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, cnnNoDb, cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)
	s, err := cfg.GetStorage("test")
	assert.NoError(t, err)
	err = s.Delete(&User{}, "id = ?", "123456")
	assert.NoError(t, err)
	err = s.Delete(&User{ID: "123456"})
	assert.NoError(t, err)

}
func TestSaveData(t *testing.T) {
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, cnnNoDb, cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)
	s, err := cfg.GetStorage("test")
	assert.NoError(t, err)
	err = s.Delete(&User{ID: "123456"})
	assert.NoError(t, err)

	err = s.Save(&User{
		ID:       "123456",
		Username: "admin",
		Password: "123456",
	})
	assert.NoError(t, err)
	err = s.Save(&User{
		ID:       "123456",
		Username: "admin",
		Password: "123456",
	})
	assert.Error(t, err)

}
func TestTranslateError(t *testing.T) {
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, cnnNoDb, cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)
	s, err := cfg.GetStorage("test")
	assert.NoError(t, err)
	err = s.Delete(&User{ID: "123456"})
	assert.NoError(t, err)

	err = s.Save(&User{
		ID:       "123456",
		Username: "admin",
		Password: "123456",
	})
	assert.NoError(t, err)
	err = s.Save(&User{
		ID:       "123456",
		Username: "admin",
		Password: "123456",
	})
	ft := s.GetDbConfig().TranslateError(err, &User{}, "save")
	assert.Equal(t, dberrors.Duplicate, ft.Code)
	assert.Equal(t, "save", ft.Action)
	assert.Equal(t, 1, len(ft.RefColumns))
	assert.Equal(t, "id", ft.RefColumns[0])

}
func randomTimeBetween(start, end time.Time) time.Time {
	// Calculate the duration between start and end
	duration := end.Sub(start)

	// Generate a random duration within the range
	randomDuration := time.Duration(rand.Int63n(int64(duration)))

	// Add the random duration to start time
	return start.Add(randomDuration)
}

func TestPposgresStorageDatabaseInteract(t *testing.T) {
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, cnnNoDb, cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)
	s, _ := cfg.GetStorage("test")
	err = s.Delete(&User{}, "Username = ?", "admin")

	assert.NoError(t, err)
	err = s.Save(&User{
		ID:       "123456",
		Username: "admin",
		Password: "123456",
	})
	assert.NoError(t, err)

	var u User
	s.First(&u, "Userame =?", "admin")

	assert.Equal(t, "admin", u.Username)
}
func TestInsertMultiUsers(t *testing.T) {
	start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 23, 59, 59, 999999999, time.UTC)
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, cnnNoDb, cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)
	s, _ := cfg.GetStorage("test")
	err = s.Delete(&User{}, "Username = ?", "admin")

	assert.NoError(t, err)
	//var users []*User
	users := []User{}
	for i := 0; i < 1000; i++ {
		// create random time
		randomTime := randomTimeBetween(start, end)
		users = append(users, User{
			BaseInfo: BaseInfo{
				CreatedBy: "admin",
				CreatedOn: randomTime,
			},
			ID:       fmt.Sprintf("%d", i),
			Username: fmt.Sprintf("someuser%d", i),
			Password: "123456",
		})
	}
	s.GetDb().Begin()
	err = s.CreateInBatches(users, 100)

	t.Log(err)

	var u User

	s.First(&u, "Username =?", "someuser1")

	assert.Equal(t, "someuser1", u.Username)

}
func TestCondinalParse(t *testing.T) {
	cfg := dbconfig_postgres.New()
	cfg.LoadFromYamlFile(yamlFile)
	fmt.Println(cfg.GetConectionStringNoDatabase())
	assert.Equal(t, cnnNoDb, cfg.GetConectionStringNoDatabase())
	err := cfg.PingDb()
	assert.NoError(t, err)
	s, _ := cfg.GetStorage("test")
	users := []User{}
	err = s.Find(&users, "(year(CreatedOn)==?)&&(month(CreatedOn))==?", 2025, 5)
	if err != nil {
		fmt.Println(err)
	}
	var count int64
	if c, e := s.Count(&User{}, "(year(CreatedOn) == ?) && (month(CreatedOn) == ?)", 2025, 5); e == nil {
		count = c
	} else {
		assert.Equal(t, len(users), count)
	}
	// assert.True(t, count > int64(0))
	fmt.Println(count)
	t.Log(err)

	err = s.First(&User{}, "(year(CreatedOn) == ?) && (month(CreatedOn) == ?)", 2025, 5)
	assert.Nil(t, err)

}
