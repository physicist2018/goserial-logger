package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type AppConfig struct {
	PortName   string
	DBName     string
	ServerPort int
}

func Load() (*AppConfig, error) {
	cfg := &AppConfig{}

	// Установка значений по умолчанию
	defaultDBPath := filepath.Join("data", "experiments.db")

	// Парсинг флагов
	flag.StringVar(&cfg.DBName, "db", defaultDBPath, "SQLite database file path")
	flag.StringVar(&cfg.PortName, "com", "/dev/ttyUSB0", "COM port name")
	flag.IntVar(&cfg.ServerPort, "port", 5000, "Server port number")

	// Кастомное сообщение при использовании -h
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Валидация
	if cfg.ServerPort < 1 || cfg.ServerPort > 65535 {
		return nil, fmt.Errorf("invalid port number %d. Must be between 1 and 65535", cfg.ServerPort)
	}

	// Создаем директорию для БД если не существует
	if err := os.MkdirAll(filepath.Dir(cfg.DBName), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	return cfg, nil
}
