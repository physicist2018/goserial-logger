package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/physicist2018/gomodserial-v1/internal/domain"
	"github.com/physicist2018/gomodserial-v1/internal/domain/entity"
	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Check if tables exist
	if err := checkAndCreateTables(db); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

func checkAndCreateTables(db *sql.DB) error {
	// Check if experiments table exists
	var tableExists int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='experiments'").Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check for experiments table: %w", err)
	}

	if tableExists == 0 {
		// Create tables
		_, err := db.Exec(`
			CREATE TABLE experiments (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				description TEXT NOT NULL,
				created_at DATETIME NOT NULL
			);

			CREATE TABLE measurements (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				experiment_id INTEGER NOT NULL,
				value TEXT NOT NULL,
				timestamp DATETIME NOT NULL,
				FOREIGN KEY (experiment_id) REFERENCES experiments (id) ON DELETE CASCADE
			);

			CREATE INDEX idx_measurements_experiment_id ON measurements (experiment_id);
		`)
		if err != nil {
			return fmt.Errorf("failed to create tables: %w", err)
		}
		domain.DomainLogger.Info("Database tables created successfully")
	}

	return nil
}

func (r *SQLiteRepository) CreateExperiment(ctx context.Context, experiment *entity.Experiment) (int, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO experiments (name, description, created_at) VALUES (?, ?, ?)",
		experiment.Name, experiment.Description, experiment.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (r *SQLiteRepository) GetAllExperiments(ctx context.Context) ([]entity.Experiment, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, description, created_at FROM experiments ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var experiments []entity.Experiment
	for rows.Next() {
		var exp entity.Experiment
		if err := rows.Scan(&exp.ID, &exp.Name, &exp.Description, &exp.CreatedAt); err != nil {
			return nil, err
		}
		experiments = append(experiments, exp)
	}

	return experiments, nil
}

func (r *SQLiteRepository) GetExperimentByID(ctx context.Context, id int) (*entity.Experiment, error) {
	var exp entity.Experiment
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, description, created_at FROM experiments WHERE id = ?",
		id,
	).Scan(&exp.ID, &exp.Name, &exp.Description, &exp.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &exp, nil
}

func (r *SQLiteRepository) CreateMeasurement(ctx context.Context, measurement *entity.Measurement) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO measurements (experiment_id, value, timestamp) VALUES (?, ?, ?)",
		measurement.ExperimentID, measurement.Value, measurement.Timestamp,
	)
	return err
}

func (r *SQLiteRepository) GetMeasurementsByExperimentID(ctx context.Context, experimentID int) ([]entity.Measurement, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, experiment_id, value, timestamp FROM measurements WHERE experiment_id = ? ORDER BY timestamp",
		experimentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var measurements []entity.Measurement
	for rows.Next() {
		var m entity.Measurement
		if err := rows.Scan(&m.ID, &m.ExperimentID, &m.Value, &m.Timestamp); err != nil {
			return nil, err
		}
		measurements = append(measurements, m)
	}

	return measurements, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
