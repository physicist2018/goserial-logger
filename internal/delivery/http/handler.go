package http

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/flosch/pongo2/v6"
	"github.com/physicist2018/gomodserial-v1/internal/delivery/serial"
	"github.com/physicist2018/gomodserial-v1/internal/usecase"
)

type WebHandler struct {
	experimentUC   *usecase.ExperimentUseCase
	measurementUC  *usecase.MeasurementUseCase
	serialListener *serial.SerialListener
	templateDir    string
}

func NewWebHandler(
	experimentUC *usecase.ExperimentUseCase,
	measurementUC *usecase.MeasurementUseCase,
	serialListener *serial.SerialListener,
	templateDir string,
) *WebHandler {
	return &WebHandler{
		experimentUC:   experimentUC,
		measurementUC:  measurementUC,
		serialListener: serialListener,
		templateDir:    templateDir,
	}
}

func (h *WebHandler) StopDataCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.serialListener.Stop(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем JSON ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Data collection stopped",
	})
}

func (h *WebHandler) DataCollectionStatus(w http.ResponseWriter, r *http.Request) {
	status := h.serialListener.Status()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Обновляем NewExperiment для использования нового метода Start
func (h *WebHandler) NewExperiment(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		currentExpID := 0
		if h.serialListener.IsRunning() {
			currentExpID = h.serialListener.CurrentExperimentID()
		}

		data := map[string]interface{}{
			"CurrentExperimentID": currentExpID,
		}

		err := h.renderTemplate(w, "new_experiment.html", pongo2.Context(data))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		description := r.FormValue("description")
		if description == "" {
			http.Error(w, "Description is required", http.StatusBadRequest)
			return
		}

		experiment, err := h.experimentUC.CreateExperiment(r.Context(), name, description)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Запускаем сбор данных для нового эксперимента
		if err := h.serialListener.Start(experiment.ID); err != nil {
			http.Error(w, "Failed to start data collection", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/experiments", http.StatusSeeOther)
	}
}

// func (h *WebHandler) NewExperiment(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == http.MethodGet {
// 		currentExpID := 0
// 		if h.serialListener.IsRunning() {
// 			currentExpID = h.serialListener.CurrentExperimentID()
// 		}

// 		data := struct {
// 			CurrentExperimentID int
// 		}{
// 			CurrentExperimentID: currentExpID,
// 		}

// 		err := h.renderTemplate(w, "new_experiment.html", pongo2.Context{
// 			"experiment": data,
// 		})
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 		}

// 		return
// 	}

// 	if r.Method == http.MethodPost {
// 		if err := r.ParseForm(); err != nil {
// 			http.Error(w, "Bad Request", http.StatusBadRequest)
// 			return
// 		}

// 		name := r.FormValue("name")
// 		if name == "" {
// 			http.Error(w, "Name is required", http.StatusBadRequest)
// 			return
// 		}

// 		description := r.FormValue("description")
// 		if name == "" {
// 			http.Error(w, "Description is required", http.StatusBadRequest)
// 			return
// 		}

// 		experiment, err := h.experimentUC.CreateExperiment(r.Context(), name, description)
// 		if err != nil {
// 			log.Printf("Failed to create experiment: %v", err)
// 			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 			return
// 		}

// 		// Start data collection for the new experiment
// 		if err := h.serialListener.Start(experiment.ID); err != nil {
// 			log.Printf("Failed to start serial listener: %v", err)
// 			http.Error(w, "Failed to start data collection", http.StatusInternalServerError)
// 			return
// 		}

// 		http.Redirect(w, r, "/experiments", http.StatusSeeOther)
// 	}
// }

func (h *WebHandler) renderTemplate(w http.ResponseWriter, templateName string, ctx pongo2.Context) error {
	tplPath := filepath.Join(h.templateDir, templateName)
	tpl, err := pongo2.FromFile(tplPath)
	if err != nil {
		return err
	}
	return tpl.ExecuteWriter(ctx, w)
}

func (h *WebHandler) ListExperiments(w http.ResponseWriter, r *http.Request) {
	experiments, err := h.experimentUC.GetAllExperiments(r.Context())

	if err != nil {
		log.Printf("Failed to get experiments: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = h.renderTemplate(w, "experiments.html", pongo2.Context{
		"experiments": experiments,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) ShowExperiment(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid experiment ID", http.StatusBadRequest)
		return
	}

	experiment, err := h.experimentUC.GetExperimentByID(r.Context(), id)
	if err != nil {
		log.Printf("Failed to get experiment: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	measurements, err := h.measurementUC.GetMeasurementsByExperimentID(r.Context(), id)
	if err != nil {
		log.Printf("Failed to get measurements: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// data := struct {
	// 	Experiment   *entity.Experiment
	// 	Measurements []entity.Measurement
	// }{
	// 	Experiment:   experiment,
	// 	Measurements: measurements,
	// }

	content := pongo2.Context{
		"experiment":   experiment,
		"measurements": measurements,
	}

	err = h.renderTemplate(w, "experiment.html", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) Home(w http.ResponseWriter, r *http.Request) {
	log.Println("HOME")
	http.Redirect(w, r, "/experiments", http.StatusSeeOther)
}
