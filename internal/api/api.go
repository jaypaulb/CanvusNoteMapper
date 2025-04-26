package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jaypaulb/CanvusNoteMapper/internal/config"
	"github.com/jaypaulb/CanvusNoteMapper/internal/llm"
	"github.com/jaypaulb/CanvusNoteMapper/internal/mapping"
	"github.com/jaypaulb/CanvusNoteMapper/internal/mcs"
)

// POST /api/upload-image
func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Image file required"}`))
		return
	}
	defer file.Close()
	imageData := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			imageData = append(imageData, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	// For now, just return success (mock)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"uploaded"}`))
	log.Println("Image uploaded (mock)")
}

// POST /api/scan-notes
func ScanNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ImageData       []byte `json:"imageData"`
		ImageDimensions [2]int `json:"imageDimensions"`
		ZoneDimensions  [2]int `json:"zoneDimensions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid JSON"}`))
		return
	}
	// Call LLM stub
	notes := llm.AnalyzeImage(req.ImageData)
	// Call mapping stub
	mapped := mapping.MapNotesToZone(notes, req.ImageDimensions, req.ZoneDimensions)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapped)
	log.Println("Scanned notes and returned mock mapped data")
}

// GET /api/get-zones
func GetZonesHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	if cfg.MCSServer == "" || cfg.APIKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"MCS credentials not set"}`))
		return
	}
	client := mcs.NewClient(cfg.MCSServer, cfg.APIKey)
	canvases := client.GetCanvases()
	zones := client.GetAnchorZones("") // For mock, no canvasID needed
	resp := map[string]interface{}{
		"canvases": canvases,
		"zones":    zones,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	log.Println("Returned mock canvases and zones")
}

// POST /api/create-notes
func CreateNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var notes []interface{}
	if err := json.NewDecoder(r.Body).Decode(&notes); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid JSON"}`))
		return
	}
	cfg := config.GetConfig()
	client := mcs.NewClient(cfg.MCSServer, cfg.APIKey)
	for _, note := range notes {
		_ = client.CreateNote(note) // Mock, ignore error
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"notes created (mock)"}`))
	log.Println("Created notes (mock)")
}

// POST /api/set-credentials
func SetCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MCSServer string `json:"mcsServer"`
		APIKey    string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid JSON"}`))
		return
	}
	config.SetConfig(&config.Config{
		MCSServer: req.MCSServer,
		APIKey:    req.APIKey,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
	log.Printf("Set credentials: server=%s\n", req.MCSServer)
}
