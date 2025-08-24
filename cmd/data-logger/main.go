package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	http2 "github.com/physicist2018/gomodserial-v1/internal/delivery/http"
	"github.com/physicist2018/gomodserial-v1/internal/delivery/serial"
	"github.com/physicist2018/gomodserial-v1/internal/infrastructure/database"
	"github.com/physicist2018/gomodserial-v1/internal/usecase"
	"github.com/physicist2018/gomodserial-v1/pkg/config"
)

const (
	dbPath         = "data/experiments.db"
	portName       = "/dev/random" // Change this to your actual COM port
	baudRate       = 9600
	templatesDir   = "./web/templates"
	staticFilesDir = "./web/static"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	dbRepo, err := database.NewSQLiteRepository(cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbRepo.Close()

	// Create use cases
	experimentUC := usecase.NewExperimentUseCase(dbRepo)
	measurementUC := usecase.NewMeasurementUseCase(dbRepo)

	// Create serial listener
	serialListener := serial.NewSerialListener(cfg.PortName, baudRate, measurementUC)

	// Create HTTP handler
	webHandler := http2.NewWebHandler(experimentUC, measurementUC, serialListener, templatesDir)

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", webHandler.Home)
	mux.HandleFunc("/experiments/new", webHandler.NewExperiment)
	mux.HandleFunc("/experiments", webHandler.ListExperiments)
	mux.HandleFunc("/experiment", webHandler.ShowExperiment)
	// В функции main() после создания обработчиков:
	mux.HandleFunc("/api/stop", webHandler.StopDataCollection)
	mux.HandleFunc("/api/status", webHandler.DataCollectionStatus)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticFilesDir))))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: mux,
	}

	// Context for graceful shutdown
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// Start HTTP server
	go func() {
		log.Printf("Starting server on port %d", cfg.ServerPort)
		log.Printf("Database path: %s", cfg.DBName)
		log.Printf("COM port: %s (%d baud)", cfg.PortName, baudRate)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Channel for experiment IDs from HTTP requests
	expChan := make(chan int)

	// Start serial listener when a new experiment is created
	go func() {
		for expID := range expChan {
			if err := serialListener.Start(expID); err != nil {
				log.Fatalf("Failed to start serial listener: %v", err)
			}
		}
	}()

	// Simulate receiving experiment IDs (in a real app, this would come from HTTP requests)
	// For demonstration, we'll create a test experiment
	// testExp, err := experimentUC.CreateExperiment(context.Background(), "Initial Test", "")
	// if err != nil {
	// 	log.Printf("Failed to create test experiment: %v", err)
	// } else {
	// 	expChan <- testExp.ID
	// }

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
