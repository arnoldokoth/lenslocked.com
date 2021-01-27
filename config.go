package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// PostgresConfig ...
type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Dialect ...
func (c PostgresConfig) Dialect() string {
	return "postgres"
}

// ConnString ...
func (c PostgresConfig) ConnString() string {
	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
			c.Host, c.Port, c.Username, c.Name)
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, c.Name)
}

// DefaultPostgresConfig ...
func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "arnoldokoth",
		Password: "Password123!",
		Name:     "lenslocked_dev",
	}
}

// Config ...
type Config struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Pepper   string         `json:"pepper"`
	HMACKey  string         `json:"hmac_key"`
	Database PostgresConfig `json:"database"`
	Mailgun  MailgunConfig  `json:"mailgun"`
	Dropbox  OAuthConfig    `json:"dropbox"`
}

// IsProd ...
func (c Config) IsProd() bool {
	return c.Env == "production"
}

// DefaultConfig ...
func DefaultConfig() Config {
	return Config{
		Port:     3000,
		Env:      "development",
		Pepper:   "5881f867b9078bd1d3ce164cc2466b13c4028ea12df14dfee9a6465e8c0b39ee",
		HMACKey:  "4ed10e653ae1c61f0d842491c00eba6bd0f34fa5702f75abb5a12aaba721c2a9",
		Database: DefaultPostgresConfig(),
	}
}

// MailgunConfig ...
type MailgunConfig struct {
	APIKey       string `json:"api_key"`
	PublicAPIKey string `json:"public_api_key"`
	Domain       string `json:"domain"`
}

// OAuthConfig ...
type OAuthConfig struct {
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	AuthURL  string `json:"auth_url"`
	TokenURL string `json:"token_url"`
}

// LoadConfig ...
func LoadConfig() Config {
	file, err := os.Open(".config.json")
	if err != nil {
		log.Println("Using Default Config")
		return DefaultConfig()
	}

	var c Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		log.Fatalln("LoadConfig() ERROR:", err)
	}

	log.Println("Successfully Loaded .config File")
	return c
}
