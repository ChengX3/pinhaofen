package database

import (
	"fmt"

	"zufen/config"
	"zufen/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	cfg := config.Get().Database

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接池失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	if err := DB.AutoMigrate(&model.Participant{}, &model.Config{}); err != nil {
		return fmt.Errorf("自动迁移表结构失败: %w", err)
	}

	if err := initDefaultConfig(); err != nil {
		return fmt.Errorf("初始化默认配置失败: %w", err)
	}

	return nil
}

func initDefaultConfig() error {
	defaults := map[string]string{
		"target_score":     "2026",
		"fuzzy_min":        "2024",
		"fuzzy_max":        "2028",
		"valid_url_prefix": "https://u.alipay.cn/",
		"upload_dir":       "./uploads",
		"max_per_day_ip":   "3",
	}

	for key, value := range defaults {
		var count int64
		DB.Model(&model.Config{}).Where("`key` = ?", key).Count(&count)
		if count == 0 {
			DB.Create(&model.Config{Key: key, Value: value})
		}
	}

	return nil
}

func Get() *gorm.DB {
	return DB
}
