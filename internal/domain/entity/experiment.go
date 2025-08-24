package entity

import (
	"context"
	"time"
)

type Experiment struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
}

type ExperimentRepository interface {
	CreateExperiment(ctx context.Context, experiment *Experiment) (int, error)
	GetAllExperiments(ctx context.Context) ([]Experiment, error)
	GetExperimentByID(ctx context.Context, id int) (*Experiment, error)
}
