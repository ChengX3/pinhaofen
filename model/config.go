package model

import "time"

type Config struct {
	Key       string    `gorm:"column:key;primaryKey;size:50" json:"key"`
	Value     string    `gorm:"column:value;type:text;not null" json:"value"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Config) TableName() string {
	return "config"
}
