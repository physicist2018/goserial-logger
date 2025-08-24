package entity

import (
	"context"
	"time"
)

type Measurement struct {
	ID           int       `json:"id"`
	ExperimentID int       `json:"experiment_id"`
	Value        string    `json:"value"`
	Timestamp    time.Time `json:"timestamp"`
}

type MeasurementRepository interface {
	CreateMeasurement(ctx context.Context, measurement *Measurement) error
	GetMeasurementsByExperimentID(ctx context.Context, experimentID int) ([]Measurement, error)
}
