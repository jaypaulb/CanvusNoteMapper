package api

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"

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
		ImageData       []byte  `json:"imageData"`
		ImageDimensions [2]int  `json:"imageDimensions"`
		ZoneDimensions  [2]int  `json:"zoneDimensions"`
		ZoneLocation    [2]int  `json:"zoneLocation"`
		ZoneScale       float64 `json:"zoneScale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid JSON"}`))
		return
	}
	// Call LLM stub (now returns notes, imgW, imgH)
	notes, imgW, imgH := llm.AnalyzeImage(req.ImageData)
	// Call mapping stub (now takes imageWidth, imageHeight, zoneDimensions, zoneLocation as args)
	mapped := mapping.MapNotesToZone(notes, imgW, imgH, req.ZoneDimensions, req.ZoneLocation)
	// Set the scale of the mapped notes to the anchor's scale
	for i := range mapped {
		mapped[i].Scale = req.ZoneScale
	}
	// Transform mapped notes to MCS API format
	mcsNotes := mapping.MapNotesToMCSFormat(mapped)
	w.Header().Set("Content-Type", "application/json")
	// Respond with mapped notes (in MCS format) and original image size for downstream scaling
	resp := map[string]interface{}{
		"notes":       mcsNotes,
		"imageWidth":  imgW,
		"imageHeight": imgH,
	}
	json.NewEncoder(w).Encode(resp)
	// Log relational data and mapping process
	log.Printf("[ScanNotesHandler] Image size: width=%d, height=%d", imgW, imgH)
	log.Printf("[ScanNotesHandler] Zone size: width=%d, height=%d", req.ZoneDimensions[0], req.ZoneDimensions[1])
	log.Printf("[ScanNotesHandler] Zone location: x=%d, y=%d", req.ZoneLocation[0], req.ZoneLocation[1])
	log.Printf("[ScanNotesHandler] Zone scale: %.4f", req.ZoneScale)
	// Calculate scale and offsets for logging (same as mapping logic)
	scale := 0.0
	offsetX := 0.0
	offsetY := 0.0
	if imgW != 0 && imgH != 0 && req.ZoneDimensions[0] != 0 && req.ZoneDimensions[1] != 0 {
		scale = min(float64(req.ZoneDimensions[0])/float64(imgW), float64(req.ZoneDimensions[1])/float64(imgH))
		offsetX = (float64(req.ZoneDimensions[0]) - float64(imgW)*scale) / 2
		offsetY = (float64(req.ZoneDimensions[1]) - float64(imgH)*scale) / 2
	}
	log.Printf("[ScanNotesHandler] Scaling factor: %.4f", scale)
	log.Printf("[ScanNotesHandler] Offset: x=%.2f, y=%.2f", offsetX, offsetY)
	log.Printf("[ScanNotesHandler] Source notes (raw): %v", notes)
	log.Printf("[ScanNotesHandler] Mapped notes: %v", mapped)
	log.Println("Scanned notes and returned mock mapped data (with image size)")
}

// GET /api/get-anchors
func GetAnchorsHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	log.Println("[GetAnchorsHandler] Called /api/get-anchors")
	if cfg.MCSServer == "" || cfg.APIKey == "" {
		log.Println("[GetAnchorsHandler] MCS credentials not set")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"MCS credentials not set"}`))
		return
	}
	var req struct {
		CanvasID string `json:"canvasID"`
	}
	decErr := json.NewDecoder(r.Body).Decode(&req)
	if decErr != nil && decErr.Error() != "EOF" {
		log.Printf("[GetAnchorsHandler] Error decoding request body: %v\n", decErr)
	}
	log.Printf("[GetAnchorsHandler] Request: %+v\n", req)
	client := mcs.NewClient(cfg.MCSServer, cfg.APIKey, req.CanvasID)
	canvases, err := client.GetCanvases()
	if err != nil {
		log.Printf("[GetAnchorsHandler] Failed to fetch canvases: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to fetch canvases: "` + err.Error() + `}`))
		return
	}
	anchors := []mcs.AnchorInfo{}
	if req.CanvasID != "" {
		anchors, err = client.GetAnchors(req.CanvasID)
		if err != nil {
			log.Printf("[GetAnchorsHandler] Failed to fetch anchors for canvas %s: %v\n", req.CanvasID, err)
		}
	}
	resp := map[string]interface{}{
		"canvases": canvases,
		"anchors":  anchors,
	}
	log.Printf("[GetAnchorsHandler] Response: %+v\n", resp)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	log.Println("[GetAnchorsHandler] Returned canvases and anchors from MCS API")
}

// POST /api/create-notes
func CreateNotesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[CreateNotesHandler] Called /api/create-notes")
	if r.Method != http.MethodPost {
		log.Println("[CreateNotesHandler] Invalid method")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// --- Scaling Logic ---
	// Require imageWidth and imageHeight in the request
	var req struct {
		CanvasID    string        `json:"canvasID"`
		Notes       []interface{} `json:"notes"`
		ZoneID      string        `json:"zoneID"`
		ImageWidth  float64       `json:"imageWidth"`
		ImageHeight float64       `json:"imageHeight"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CreateNotesHandler] Error decoding request body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid JSON: "` + err.Error() + `}`))
		return
	}
	if req.ImageWidth == 0.0 || req.ImageHeight == 0.0 {
		log.Printf("[CreateNotesHandler] imageWidth or imageHeight missing or zero, cannot scale notes correctly.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"imageWidth and imageHeight required"}`))
		return
	}

	cfg := config.GetConfig()
	client := mcs.NewClient(cfg.MCSServer, cfg.APIKey, req.CanvasID)

	// Fetch anchor info for the selected zone
	anchor, err := client.GetAnchorInfo(req.CanvasID, req.ZoneID)
	if err != nil {
		log.Printf("[CreateNotesHandler] Failed to fetch anchor info: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to fetch anchor info: "` + err.Error() + `}`))
		return
	}
	anchorJson, _ := json.MarshalIndent(anchor, "", "  ")
	log.Printf("[CreateNotesHandler] Anchor zone details: %s", string(anchorJson))

	// Calculate scaleFactor (difference between image and anchor zone size)
	scaleFactor := 1.0
	if anchor.Width > 0 && anchor.Height > 0 {
		scaleFactor = min(anchor.Width/req.ImageWidth, anchor.Height/req.ImageHeight)
	}
	log.Printf("[CreateNotesHandler] Calculated scaleFactor: %.4f (imageWidth=%.2f, imageHeight=%.2f, anchor.Width=%.2f, anchor.Height=%.2f)", scaleFactor, req.ImageWidth, req.ImageHeight, anchor.Width, anchor.Height)
	// --- End Scaling Logic ---

	finalScale := scaleFactor * anchor.Scale

	for i, note := range req.Notes {
		noteMap, _ := note.(map[string]interface{})
		noteJson, _ := json.MarshalIndent(noteMap, "", "  ")
		log.Printf("[CreateNotesHandler][Note %d] Source: %s", i+1, string(noteJson))

		// 1. Scale location and size by scaleFactor * anchor.Scale
		// 2. Offset location by anchor.X and anchor.Y
		// 3. Set note.scale = 1
		if noteMap["location"] != nil {
			if loc, ok := noteMap["location"].(map[string]interface{}); ok {
				noteMap["location"] = map[string]interface{}{
					"x": anchor.X + loc["x"].(float64)*finalScale,
					"y": anchor.Y + loc["y"].(float64)*finalScale,
				}
			}
		}
		if noteMap["size"] != nil {
			if size, ok := noteMap["size"].(map[string]interface{}); ok {
				noteMap["size"] = map[string]interface{}{
					"width":  size["width"].(float64) * finalScale,
					"height": size["height"].(float64) * finalScale,
				}
			}
		}
		// Set the note's scale to 1 (all scaling handled in math above)
		noteMap["scale"] = 1
		// Optionally, set parent_id or other anchor fields if needed

		noteJson, _ = json.MarshalIndent(noteMap, "", "  ")
		log.Printf("[CreateNotesHandler][Note %d] Target (to MCS): %s", i+1, string(noteJson))
		resp, err := client.CreateNote(req.CanvasID, noteMap)
		respJson, _ := json.MarshalIndent(resp, "", "  ")
		log.Printf("[CreateNotesHandler][Note %d] Response from MCS: %s", i+1, string(respJson))
		// Validation: check if response matches target (ignoring id, parent_id, etc.)
		match := true
		for k, v := range noteMap {
			if k == "id" || k == "parent_id" || k == "ParentID" || k == "ID" {
				continue
			}
			if !reflect.DeepEqual(resp[k], v) {
				match = false
				log.Printf("[CreateNotesHandler][Note %d] Validation mismatch: key '%s' target=%v response=%v", i+1, k, v, resp[k])
			}
		}
		if match {
			log.Printf("[CreateNotesHandler][Note %d] Validation: PASS", i+1)
		} else {
			log.Printf("[CreateNotesHandler][Note %d] Validation: FAIL", i+1)
		}
		if err != nil {
			log.Printf("[CreateNotesHandler] Failed to create note: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to create note: "` + err.Error() + `}`))
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"notes created"}`))
	log.Println("[CreateNotesHandler] Created notes via MCS API")
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

// GET /api/get-canvas-size
func GetCanvasSizeHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	if cfg.MCSServer == "" || cfg.APIKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"MCS credentials not set"}`))
		return
	}
	var req struct {
		CanvasID string `json:"canvasID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid JSON"}`))
		return
	}
	client := mcs.NewClient(cfg.MCSServer, cfg.APIKey, req.CanvasID)
	size, err := client.GetCanvasSize(req.CanvasID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to fetch canvas size: "` + err.Error() + `}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(size)
}

// GET /api/get-canvases
func GetCanvasesHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	if cfg.MCSServer == "" || cfg.APIKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"MCS credentials not set"}`))
		return
	}
	client := mcs.NewClient(cfg.MCSServer, cfg.APIKey, "")
	canvases, err := client.GetCanvases()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to fetch canvases: "` + err.Error() + `}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"canvases": canvases})
}

// GET /api/get-anchors?canvasID=...
func GetAnchorsOnlyHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	canvasID := r.URL.Query().Get("canvasID")
	if cfg.MCSServer == "" || cfg.APIKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"MCS credentials not set"}`))
		return
	}
	if canvasID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"canvasID required"}`))
		return
	}
	client := mcs.NewClient(cfg.MCSServer, cfg.APIKey, canvasID)
	anchors, err := client.GetAnchors(canvasID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to fetch anchors: "` + err.Error() + `}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"anchors": anchors})
}

// GET /api/get-anchor-info?canvasID=...&anchorID=...
func GetAnchorInfoHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	canvasID := r.URL.Query().Get("canvasID")
	anchorID := r.URL.Query().Get("anchorID")
	if cfg.MCSServer == "" || cfg.APIKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"MCS credentials not set"}`))
		return
	}
	if canvasID == "" || anchorID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"canvasID and anchorID required"}`))
		return
	}
	client := mcs.NewClient(cfg.MCSServer, cfg.APIKey, canvasID)
	anchor, err := client.GetAnchorInfo(canvasID, anchorID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to fetch anchor info: "` + err.Error() + `}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(anchor)
}

// Helper for min (copy from mapping)
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
