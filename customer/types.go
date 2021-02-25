package customer

import (
	"time"
)

// BaseModel is base for database struct
type BaseModel struct {
	ID        uint64    `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
