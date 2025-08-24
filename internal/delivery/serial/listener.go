package serial

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/physicist2018/gomodserial-v1/internal/infrastructure/serial"
	"github.com/physicist2018/gomodserial-v1/internal/usecase"
)

type SerialListener struct {
	portListener  *serial.PortListener
	measurementUC *usecase.MeasurementUseCase
	currentExpID  int
	mu            sync.Mutex
	stopChan      chan struct{}
	isRunning     bool

	cancelFunc context.CancelFunc
	ctx        context.Context
}

func NewSerialListener(portName string, baudRate int, measurementUC *usecase.MeasurementUseCase) *SerialListener {
	return &SerialListener{
		portListener:  serial.NewPortListener(portName, baudRate),
		measurementUC: measurementUC,
		stopChan:      make(chan struct{}),
	}
}

// func (sl *SerialListener) Start(ctx context.Context, experimentID int) error {
// 	sl.mu.Lock()
// 	defer sl.mu.Unlock()

// 	// Stop previous experiment if running
// 	if sl.isRunning {
// 		close(sl.stopChan)
// 		sl.stopChan = make(chan struct{})
// 	}

// 	sl.currentExpID = experimentID
// 	sl.isRunning = true

// 	if err := sl.portListener.Open(); err != nil {
// 		return err
// 	}

// 	dataChan := make(chan string)
// 	errorChan := make(chan error)

// 	go sl.portListener.Listen(ctx, dataChan, errorChan)

// 	go func() {
// 		defer func() {
// 			sl.mu.Lock()
// 			sl.isRunning = false
// 			sl.mu.Unlock()
// 		}()

// 		for {
// 			select {
// 			case data := <-dataChan:
// 				lines := strings.Split(data, "\n")
// 				for _, line := range lines {
// 					line = strings.TrimSpace(line)
// 					if line != "" {
// 						if err := sl.measurementUC.CreateMeasurement(ctx, sl.currentExpID, line); err != nil {
// 							log.Printf("Failed to save measurement: %v", err)
// 						} else {
// 							log.Printf("Saved measurement to experiment %d: %s", sl.currentExpID, line)
// 						}
// 					}
// 				}
// 			case err := <-errorChan:
// 				log.Printf("Serial port error: %v", err)
// 				return
// 			case <-sl.stopChan:
// 				log.Printf("Stopping data collection for experiment %d", sl.currentExpID)
// 				return
// 			case <-ctx.Done():
// 				return
// 			}
// 		}
// 	}()

// 	return nil
// }

func (sl *SerialListener) Start(experimentID int) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Останавливаем предыдущий сбор данных, если запущен
	if sl.isRunning {
		sl.stop()
	}

	sl.currentExpID = experimentID

	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	sl.ctx = ctx
	sl.cancelFunc = cancel

	// Запускаем сбор данных в отдельной горутине
	go sl.collectData(ctx)

	sl.isRunning = true
	log.Printf("Started data collection for experiment %d", experimentID)
	return nil
}

func (sl *SerialListener) Stop() error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	if !sl.isRunning {
		return nil
	}

	sl.stop()
	log.Printf("Stopped data collection for experiment %d", sl.currentExpID)
	return nil
}

func (sl *SerialListener) stop() {
	if sl.cancelFunc != nil {
		sl.cancelFunc()
	}
	close(sl.stopChan)
	sl.stopChan = make(chan struct{})
	sl.isRunning = false
	sl.currentExpID = 0
}

func (sl *SerialListener) collectData(ctx context.Context) {
	// Инициализация порта
	port := serial.NewPortListener(sl.portListener.Name(), sl.portListener.BaudRate())
	if err := port.Open(); err != nil {
		log.Printf("Failed to open port: %v", err)
		return
	}
	defer port.Close()

	dataChan := make(chan string)
	errorChan := make(chan error)

	// Запускаем прослушивание порта
	go port.Listen(ctx, dataChan, errorChan)

	for {
		select {
		case data := <-dataChan:
			lines := strings.Split(data, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					if err := sl.measurementUC.CreateMeasurement(ctx, sl.currentExpID, line); err != nil {
						log.Printf("Failed to save measurement: %v", err)
					} else {
						log.Printf("Saved measurement to experiment %d: %s", sl.currentExpID, line)
					}
				}
			}
		case err := <-errorChan:
			log.Printf("Serial port error: %v", err)
			return
		case <-sl.stopChan:
			log.Printf("Data collection stopped by user")
			return
		case <-ctx.Done():
			log.Printf("Data collection context cancelled")
			return
		}
	}
}

func (sl *SerialListener) IsRunning() bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.isRunning
}

func (sl *SerialListener) CurrentExperimentID() int {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.currentExpID
}

func (sl *SerialListener) Status() map[string]any {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	return map[string]any{
		"is_running":         sl.isRunning,
		"current_experiment": sl.currentExpID,
		"port":               sl.portListener.Name(),
		"baud_rate":          sl.portListener.BaudRate(),
	}
}
