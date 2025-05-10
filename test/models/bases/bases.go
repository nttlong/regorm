package bases

import (
	"time"

	"github.com/google/uuid"
)

type BaseModel struct {
	ID          uuid.UUID  `gorm:"primary_key;char(35)"`
	Description string     `gorm:"varchar(255)"`
	CreatedOn   time.Time  `gorm:"index"`
	CreatedBy   string     `gorm:"varchar(50);index"`
	UpdatedOn   *time.Time `gorm:"index"`
	UpdatedBy   *string    `gorm:"varchar(50);index"`
}
