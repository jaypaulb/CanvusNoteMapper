package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jaypaulb/CanvusNoteMapper/internal/api"
	"github.com/joho/godotenv"
)

// noCacheHandler wraps the file server to add anti-cache headers
func noCacheHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set aggressive anti-cache headers for all static files
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Last-Modified", "")
		w.Header().Set("ETag", "")

		// Log the request for debugging
		log.Printf("[static] Serving %s with anti-cache headers", r.URL.Path)

		h.ServeHTTP(w, r)
	})
}

func main() {
	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("[main] No .env file found or failed to load .env (this is fine if running in prod with env vars set)")
	} else {
		log.Println("[main] .env file loaded successfully")
	}
	log.Printf("[main] GOOGLE_GENAI_API_KEY loaded: %v", os.Getenv("GOOGLE_GENAI_API_KEY") != "")

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/upload-image", api.UploadImageHandler)
	mux.HandleFunc("/api/scan-notes", api.ScanNotesHandler)
	mux.HandleFunc("/api/create-notes", api.CreateNotesHandler)
	mux.HandleFunc("/api/set-credentials", api.SetCredentialsHandler)
	mux.HandleFunc("/api/get-canvas-size", api.GetCanvasSizeHandler)
	mux.HandleFunc("/api/get-canvases", api.GetCanvasesHandler)
	mux.HandleFunc("/api/get-anchors", api.GetAnchorsOnlyHandler)
	mux.HandleFunc("/api/get-anchor-info", api.GetAnchorInfoHandler)

	// Serve static files from web directory with anti-cache headers
	fileServer := http.FileServer(http.Dir("web"))
	mux.Handle("/", noCacheHandler(fileServer))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s with anti-cache headers enabled...", addr)
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
