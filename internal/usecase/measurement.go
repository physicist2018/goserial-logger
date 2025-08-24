package usecase

import (
	"context"
	"time"

	"github.com/physicist2018/gomodserial-v1/internal/domain/entity"
)

type MeasurementUseCase struct {
	measurementRepo entity.MeasurementRepository
}

func NewMeasurementUseCase(repo entity.MeasurementRepository) *MeasurementUseCase {
	return &MeasurementUseCase{measurementRepo: repo}
}

func (uc *MeasurementUseCase) CreateMeasurement(ctx context.Context, experimentID int, value string) error {
	measurement := &entity.Measurement{
		ExperimentID: experimentID,
		Value:        value,
		Timestamp:    time.Now(),
	}

	err := uc.measurementRepo.CreateMeasurement(ctx, measurement)

	return err
}

func (uc *MeasurementUseCase) GetMeasurementsByExperimentID(ctx context.Context, experimentID int) ([]entity.Measurement, error) {
	return uc.measurementRepo.GetMeasurementsByExperimentID(ctx, experimentID)
}
