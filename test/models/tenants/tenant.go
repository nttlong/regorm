package tenants

import (
	"github.com/nttlong/regorm/test/models/bases"
)

type Tenant struct {
	bases.BaseModel
	Name string `gorm:"varchar(50);index;unique"`
	Code string `gorm:"varchar(50);index;unique"`
}
