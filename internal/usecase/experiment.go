package usecase

import (
	"context"
	"time"

	"github.com/physicist2018/gomodserial-v1/internal/domain"
	"github.com/physicist2018/gomodserial-v1/internal/domain/entity"
)

type ExperimentUseCase struct {
	experimentRepository entity.ExperimentRepository
}

func NewExperimentUseCase(experimentRepository entity.ExperimentRepository) *ExperimentUseCase {
	return &ExperimentUseCase{
		experimentRepository: experimentRepository,
	}
}

func (uc *ExperimentUseCase) CreateExperiment(ctx context.Context, name, description string) (*entity.Experiment, error) {
	experiment := &entity.Experiment{
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}

	id, err := uc.experimentRepository.CreateExperiment(ctx, experiment)
	if err != nil {
		domain.DomainLogger.Println(err)
		return nil, err
	}

	experiment.ID = id
	return experiment, nil
}

func (uc *ExperimentUseCase) GetAllExperiments(ctx context.Context) ([]entity.Experiment, error) {
	return uc.experimentRepository.GetAllExperiments(ctx)
}

func (uc *ExperimentUseCase) GetExperimentByID(ctx context.Context, id int) (*entity.Experiment, error) {
	return uc.experimentRepository.GetExperimentByID(ctx, id)
}
