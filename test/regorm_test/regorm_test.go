package regorm_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/nttlong/regorm"
	_ "github.com/nttlong/regorm"
	"github.com/stretchr/testify/assert"

	"github.com/nttlong/regorm/dbconfig"
	"github.com/nttlong/regorm/test"
	_ "github.com/nttlong/regorm/test"
	"github.com/nttlong/regorm/test/models"
	_ "github.com/nttlong/regorm/test/models"
	"github.com/nttlong/regorm/test/models/bases"
)

var cfg dbconfig.IDbConfig

func TestLoadConfig(t *testing.T) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)
	results := make(chan regorm.IDbConfig, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		semaphore <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()
			newcfg := regorm.New("postgres")
			results <- newcfg // Gửi kết quả vào channel
		}()
	}

	go func() {
		wg.Wait()
		close(results) // Đóng channel sau khi tất cả goroutines hoàn thành
	}()

	cfgs := []regorm.IDbConfig{}
	for cfg := range results {
		cfgs = append(cfgs, cfg)
	}
	isTheSame := true
	for _, c := range cfgs {
		if c != cfgs[0] {
			isTheSame = false
			break
		}
	}
	assert.True(t, isTheSame)
	cfg = cfgs[0]

	assert.NotNil(t, cfg)
	err := cfg.LoadFromYamlFile(test.GetConFigFile("/../"))
	if err != nil {
		assert.NoError(t, err)
	}
	assert.NotEmpty(t, cfg)

}
func Runner(dbNamePrefix string) error {
	cfg := regorm.New("postgres")
	cfg.LoadFromYamlFile(test.GetConFigFile("/../"))
	for i := 0; i < 5; i++ {
		dbName := fmt.Sprintf("%s%d", dbNamePrefix, i)
		storage, err := cfg.GetStorage(dbName)
		if err != nil {
			return err
		}
		err = storage.AutoMigrate(&models.Tenant{})
		if err != nil {
			return err
		}

		n, err := storage.Count(&models.Tenant{})
		if err != nil {
			fmt.Print(storage.GetDb().Dialector.Name())
			fmt.Print(storage.GetDbName())
			return err
		}
		if n == 0 {
			newTenant := models.Tenant{
				BaseModel: bases.BaseModel{
					Description: "test",
					ID:          uuid.New(),
				},
				Name: "test",
				Code: "test",
			}
			err = storage.Create(&newTenant)
			if err != nil {
				return err
			}

		}

	}
	return nil
}
func TestRunner(t *testing.T) {
	err := Runner("test_runner")
	if err != nil {
		t.Error(err)
	}
}
func BenchmarkTestLoadConfig(b *testing.B) {

	for i := 0; i < b.N; i++ {
		err := Runner("bm1")
		if err != nil {
			b.Error(err)
		}

	}
}
func TestRegorm(t *testing.T) {
	TestLoadConfig(t)
	for i := 0; i < 10; i++ {
		storage, err := cfg.GetStorage(fmt.Sprintf("test%d", i))
		assert.NoError(t, err)
		err = storage.AutoMigrate(&models.Tenant{})
		if err != nil {
			panic(err)
		}

		n, err := storage.Count(&models.Tenant{})
		if err != nil {
			fmt.Print(storage.GetDb().Dialector.Name())
			fmt.Print(storage.GetDbName())
			panic(err)
		}
		if n == 0 {
			newTenant := models.Tenant{
				BaseModel: bases.BaseModel{
					Description: "test",
					ID:          uuid.New(),
				},
				Name: "test",
				Code: "test",
			}
			err = storage.Create(&newTenant)
			if err != nil {
				t.Log(err)
			}
			assert.Equal(t, newTenant.ID, newTenant.ID)
		}
		assert.NoError(t, err)
	}

}
