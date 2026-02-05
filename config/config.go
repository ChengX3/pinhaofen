package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Match    MatchConfig    `yaml:"match"`
	QRCode   QRCodeConfig   `yaml:"qrcode"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type MatchConfig struct {
	TargetScore int `yaml:"target_score"`
	FuzzyMin    int `yaml:"fuzzy_min"`
	FuzzyMax    int `yaml:"fuzzy_max"`
}

type QRCodeConfig struct {
	ValidURLPrefix string `yaml:"valid_url_prefix"`
	UploadDir      string `yaml:"upload_dir"`
	MaxPerDayIP    int    `yaml:"max_per_day_ip"`
}

var AppConfig *Config

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	AppConfig = &Config{}
	return yaml.Unmarshal(data, AppConfig)
}

func Get() *Config {
	return AppConfig
}
