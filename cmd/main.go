package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jaypaulb/CanvusNoteMapper/internal/api"
	"github.com/joho/godotenv"
)

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

	// Serve static files from web directory
	fileServer := http.FileServer(http.Dir("web"))
	mux.Handle("/", fileServer)

	// Start server
	log.Printf("[main] Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("[main] Server failed to start: %v", err)
	}
}
