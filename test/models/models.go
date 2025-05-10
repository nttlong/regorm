package models

import (
	"github.com/nttlong/regorm/test/models/bases"
	"github.com/nttlong/regorm/test/models/tenants"
	_ "github.com/nttlong/regorm/test/models/tenants"
)

type BaseModel bases.BaseModel
type Tenant tenants.Tenant
