package repository

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	file     string
	instance *sql.DB
}

func NewSQLite(file string) *SQLite {
	return &SQLite{
		file: file,
	}
}

func (s *SQLite) Connect() error {
	db, err := sql.Open("sqlite3", s.file)
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
		return err
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping SQLite: %v", err)
		return err
	}
	s.instance = db
	return nil
}

func (s *SQLite) GetConfigByID(id string) (RateLimitConfig, error) {
	query := "SELECT config_value, limit_type, max_request, block_time FROM configs WHERE id = ?"
	row := s.instance.QueryRow(query, id)

	var result RateLimitConfig
	err := row.Scan(&result.ConfigValue, &result.LimitType, &result.MaxRequest, &result.BlockTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return RateLimitConfig{}, fmt.Errorf("record not found")
		}
		log.Printf("Failed to execute query: %v", err)
		return RateLimitConfig{}, err
	}

	return result, nil
}

func (s *SQLite) GetConfigByConfigValue(confiigValue string) (RateLimitConfig, error) {
	query := "SELECT config_value, limit_type, max_request, block_time FROM configs WHERE config_value = ?"
	row := s.instance.QueryRow(query, confiigValue)

	var result RateLimitConfig
	err := row.Scan(&result.ConfigValue, &result.LimitType, &result.MaxRequest, &result.BlockTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return RateLimitConfig{}, fmt.Errorf("record not found")
		}
		log.Printf("Failed to execute query: %v", err)
		return RateLimitConfig{}, err
	}

	return result, nil
}

func (s *SQLite) GetAllConfigs() ([]RateLimitConfig, error) {
	query := "SELECT id, config_value, limit_type, max_request, block_time FROM configs"
	rows, err := s.instance.Query(query)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var result []RateLimitConfig
	for rows.Next() {
		var config RateLimitConfig
		err = rows.Scan(&config.Id, &config.ConfigValue, &config.LimitType, &config.MaxRequest, &config.BlockTime)
		if err != nil {
			log.Printf("Failed to execute query: %v", err)
			return nil, err
		}
		result = append(result, config)
	}

	return result, nil
}

func (s *SQLite) CreateConfig(config RateLimitConfig) error {
	query := "INSERT INTO configs (id, config_value, limit_type, max_request, block_time) VALUES (?, ?, ?, ?, ?)"
	_, err := s.instance.Exec(query, config.Id, config.ConfigValue, config.LimitType, config.MaxRequest, config.BlockTime)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return err
	}
	return nil
}

func (s *SQLite) UpdateConfig(config RateLimitConfig) error {
	query := "UPDATE configs SET config_value = ?, limit_type = ?, max_request = ?, block_time = ? WHERE id = ?"
	_, err := s.instance.Exec(query, config.ConfigValue, config.LimitType, config.MaxRequest, config.BlockTime, config.Id)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return err
	}
	return nil
}

func (s *SQLite) DeleteConfig(id string) error {
	query := "DELETE FROM configs WHERE id = ?"
	_, err := s.instance.Exec(query, id)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return err
	}
	return nil
}
