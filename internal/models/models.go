package models

import (
	"database/sql"
	"time"
)

// DBModel is the type for database connection values
type DBModel struct {
	DB *sql.DB
}

// Models is the wrapper for all models
type Models struct {
	DB DBModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{
			DB: db,
		},
	}
}

type Widget struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	Inventorylevel int `json:"inventory_level"`
	Price int `json:"price"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}