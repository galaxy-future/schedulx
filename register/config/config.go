package config

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

var GlobalConfig *Config

type Config struct {
	DebugMode           bool                `yaml:"DebugMode"`
	ServerPort          int                 `yaml:"ServerPort"`
	LogFile             string              `yaml:"LogFile"`
	LogLevel            string              `yaml:"LogLevel"`
	WriteDB             DBConfig            `yaml:"WriteDB"`
	ReadDB              DBConfig            `yaml:"ReadDB"`
	BridgXHost          string              `yaml:"BridgXHost"`
	JwtToken            JwtTokenConfig      `yaml:"JwtToken"`
	AlibabaCloudAccount AlibabaCloudAccount `yaml:"AlibabaCloudAccount"`
}

type DBConfig struct {
	Name         string `yaml:"Name"`
	Host         string `yaml:"Host"`
	Port         string `yaml:"Port"`
	User         string `yaml:"User"`
	Password     string `yaml:"Password"`
	Timeout      string `yaml:"Timeout"`
	ReadTimeout  string `yaml:"ReadTimeout"`
	WriteTimeout string `yaml:"WriteTimeout"`
	MaxIdleConns int    `yaml:"MaxIdleConns"`
	MaxOpenConns int    `yaml:"MaxOpenConns"`
}

type JwtTokenConfig struct {
	JwtTokenSignKey        string `yaml:"JwtTokenSignKey"`
	JwtTokenCreatedExpires int64  `yaml:"JwtTokenCreatedExpires"`
	JwtTokenRefreshExpires int64  `yaml:"JwtTokenRefreshExpires"`
	BindContextKeyName     string `yaml:"BindContextKeyName"`
}

type AlibabaCloudAccount struct {
	Region    string `yaml:"Region"`
	AccessKey string `yaml:"AccessKey"`
	Secret    string `yaml:"Secret"`
}

func Init(configPath string) {
	GlobalConfig = loadConfig(configPath)
}

// loadConfig configPath = register/conf/config.yml
func loadConfig(configPath string) *Config {
	filepath := composeConfigFileName(configPath, os.Getenv("env"))
	log.Printf("config filepath:%s", filepath)
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	var config Config
	if err = yaml.Unmarshal(f, &config); err != nil {
		panic(err)
	}
	return &config
}

func composeConfigFileName(basePath string, suffix string) string {
	var filepath = basePath

	if suffix != "" {
		filepath = strings.Join([]string{filepath, suffix}, ".")
	}

	return filepath
}
