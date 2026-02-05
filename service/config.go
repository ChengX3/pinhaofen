package service

import (
	"strconv"

	"zufen/database"
	"zufen/model"
)

func GetConfigValue(key string) string {
	db := database.Get()
	var cfg model.Config
	if err := db.Where("`key` = ?", key).First(&cfg).Error; err != nil {
		return ""
	}
	return cfg.Value
}

func GetConfigInt(key string, defaultVal int) int {
	val := GetConfigValue(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

func SetConfigValue(key, value string) error {
	db := database.Get()
	var cfg model.Config
	if err := db.Where("`key` = ?", key).First(&cfg).Error; err != nil {
		return db.Create(&model.Config{Key: key, Value: value}).Error
	}
	return db.Model(&cfg).Update("value", value).Error
}

func GetMatchConfig() (targetScore, fuzzyMin, fuzzyMax int) {
	targetScore = GetConfigInt("target_score", 2026)
	fuzzyMin = GetConfigInt("fuzzy_min", 2024)
	fuzzyMax = GetConfigInt("fuzzy_max", 2028)
	return
}

func GetValidURLPrefix() string {
	return GetConfigValue("valid_url_prefix")
}
