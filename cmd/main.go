package main

import (
	"log"
	"net/http"

	"github.com/jaypaulb/CanvusNoteMapper/internal/api"
)

func main() {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/upload-image", api.UploadImageHandler)
	mux.HandleFunc("/api/scan-notes", api.ScanNotesHandler)
	mux.HandleFunc("/api/get-zones", api.GetZonesHandler)
	mux.HandleFunc("/api/create-notes", api.CreateNotesHandler)
	mux.HandleFunc("/api/set-credentials", api.SetCredentialsHandler)

	// Serve static files from web directory
	mux.Handle("/", http.FileServer(http.Dir("web")))

	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
