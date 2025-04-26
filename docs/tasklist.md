MT-Canvus NoteScan Local Development Plan
This plan outlines the steps required to build and run the MT-Canvus NoteScan PWA locally, starting with only the /libs/canvusapi.go and /docs/tasklist.md files in place.

Starting Point:

A project directory containing:

/libs/canvusapi.go (Contains functions for interacting with the MCS API)

/docs/tasklist.md (Your existing task list)

Objective: Get a functional local version of the PWA running, capable of capturing images, processing notes via a stubbed or integrated LLM, applying spatial mapping, and interacting with a target MCS server (assuming one is accessible from your local network for testing).

Tasks:

1. Project Setup & Initial Structure
[x] Create Project Directory Structure: Establish the necessary folders for your Go backend and PWA frontend. A typical structure might look like:

/
├── cmd/
│   └── main.go         # Entry point for the Go application
├── internal/           # Internal packages (application logic)
│   ├── api/            # HTTP handlers and API routing
│   ├── config/         # Configuration loading and handling
│   ├── llm/            # LLM integration logic (stub initially)
│   ├── mapping/        # Spatial mapping logic (stub initially)
│   └── mcs/            # MCS interaction logic (wraps canvusapi.go)
├── web/                # PWA Frontend files
│   ├── css/
│   ├── js/
│   ├── index.html      # Main PWA file
│   └── manifest.json
├── libs/               # External or shared libraries (canvusapi.go is here)
│   └── canvusapi.go    # Provided MCS API client
└── go.mod              # Go module file
└── go.sum

[x] Initialize Go Module: Navigate to the project root in your terminal and run go mod init <your_module_name> (e.g., go mod init github.com/yourname/notescan).

[x] Move canvusapi.go: Ensure canvusapi.go is correctly placed in the /libs directory and update its package declaration if necessary to match the libs package name.

2. Backend Development (Go)
[x] Implement Configuration Handling: Create logic in /internal/config to load application settings, including the MCS server address and API key. Initially, use environment variables or a local config file (.env, config.json) for these credentials.

[x] Develop MCS Interaction Wrapper: Create a package in /internal/mcs that wraps the functions provided in /libs/canvusapi.go. This wrapper should handle authentication details from the configuration and provide cleaner functions for the API handlers to call (e.g., mcsClient.GetCanvases(), mcsClient.GetAnchorZones(canvasID), mcsClient.CreateNote(noteData)).

[x] Implement LLM Stub: Create a package in /internal/llm with a function (e.g., llm.AnalyzeImage(imageData)) that initially returns mock data for detected notes (content, color, position, size). This allows frontend and mapping development to proceed.

[x] Implement Spatial Mapping Stub: Create a package in /internal/mapping with a function (e.g., mapping.MapNotesToZone(notes, imageDimensions, zoneDimensions)) that initially returns the input note data or mock mapped data.

[x] Implement API Endpoints (/internal/api):

[x] POST /api/upload-image: Receives the image file from the frontend and saves it temporarily or passes it directly for processing.

[x] POST /api/scan-notes: Triggers the note analysis process. This handler will call the LLM stub, then the spatial mapping stub, and return the detected/mapped note data to the frontend for review.

[x] GET /api/get-zones: Calls the MCS wrapper (/internal/mcs) to fetch available canvases and anchor zones using canvusapi.go and returns them to the frontend. Requires MCS credentials.

[x] POST /api/create-notes: Receives the final note data (potentially adjusted by the user on the frontend) and calls the MCS wrapper (/internal/mcs) to create the notes using canvusapi.go. Requires MCS credentials.

[x] POST /api/set-credentials: Receives MCS server address and API key from the frontend and updates the application's configuration/credential storage.

[x] Set up HTTP Server: In /cmd/main.go, initialize the HTTP server, define routes using a router (like net/http or a framework), and link routes to the handlers in /internal/api. Configure the server to serve static files from the /web directory.

[x] Add Basic Logging: Integrate a logging library (like the standard log package or a more structured one) to output information about requests, processing steps, and errors.

[x] Implement Basic Error Handling: Add initial error handling in handlers to catch obvious issues (e.g., missing request data) and return appropriate HTTP status codes.

3. Frontend Development (PWA)
[x] Basic HTML Structure: Create /web/index.html with the basic structure for a PWA, including links to CSS and JS files, and placeholders for UI elements.

[x] Implement UI for Credentials: Add HTML forms and JavaScript to capture the MCS server address and API key. Implement JavaScript to send this data to the POST /api/set-credentials backend endpoint.

[x] Implement UI for Canvas/Zone Selection: Add HTML elements and JavaScript to display the list of canvases and zones fetched from the GET /api/get-zones endpoint and allow the user to select one.

[x] Implement UI for Image Input: Add HTML input elements (<input type="file" accept="image/*" capture="camera">) and JavaScript to allow users to select or capture an image. Implement JavaScript to send the image data to the POST /api/upload-image endpoint.

[x] Implement UI to Trigger Scan & Display Results: Add a button to trigger the scan (POST /api/scan-notes). Implement JavaScript to receive the detected/mapped note data from the backend and display it visually on the frontend (e.g., overlaying outlines on the image, listing note details).

[x] Implement UI to Create Notes: Add a button to trigger note creation (POST /api/create-notes). Implement JavaScript to send the final note data to the backend.

[x] Add PWA Manifest: Create /web/manifest.json to define the PWA's properties (name, icons, start URL, display mode).

[x] Add Service Worker: Create /web/service-worker.js to handle caching of static assets, enabling basic offline functionality (loading the UI shell). Register the service worker in index.html.

[x] Add Basic CSS: Create /web/css/style.css to add basic styling to the UI elements.

4. Integration
[x] Verify Data Flow: Test that data is correctly sent from the frontend to the backend, processed, and returned to the frontend at each step of the workflow (Credentials -> Fetch Zones -> Upload Image -> Scan -> Create Notes).

5. Local Testing & Refinement
[ ] Manual UI/UX Walkthrough: Interact with the PWA in your browser, following the user journey: enter credentials, fetch/select zones, upload an image, trigger scan, review results, create notes. Test on both desktop and mobile browsers (using browser developer tools or connecting a device).
[ ] Test Backend Endpoints: Use curl or Postman to directly test each backend API endpoint (/api/set-credentials, /api/get-zones, /api/upload-image, /api/scan-notes, /api/create-notes) with various inputs to ensure they handle requests and return responses as expected.
[ ] Validate Error Handling (Local): Intentionally provide incorrect input (e.g., wrong MCS address, invalid API key, non-image file) to trigger errors and verify that the backend returns appropriate responses and the frontend displays helpful error messages.
[ ] Refine UI/UX: Based on manual testing, improve the user interface and experience for clarity and ease of use.
[ ] Code Review and Cleanup: Review both Go and frontend code for quality, consistency, and removal of any temporary code. Run go fmt ./... and go mod tidy.

[Redundant] Test Offline Functionality (Local): Offline mode is not supported, as the app requires internet access to communicate with the LLM and MCS services. The PWA shell may load, but core features will not function without connectivity.

[Resolved] File Input/Camera UX: The file input now opens the file picker by default on Android, and the camera capture button is handled separately.

# Note: This application is intended for online use only, as it depends on external LLM and MCS services.

6. Implement Actual LLM and Spatial Mapping
[ ] Integrate Actual LLM: Replace the LLM stub (/internal/llm) with code that interacts with your chosen LLM service API. This requires implementing the API client logic and parsing the LLM's response to extract note data.

[ ] Implement Actual Spatial Mapping: Replace the spatial mapping stub (/internal/mapping) with the full implementation of the "longest side method" algorithm in Go. This logic will use the real note data from the LLM and the dimensions of the image and selected zone to calculate the final digital note positions and sizes.

[ ] Test LLM and Mapping: Create specific tests for the LLM integration and spatial mapping logic using sample images and expected outputs.

7. Final Local Testing
[ ] End-to-End Testing with Real Logic: Perform thorough end-to-end testing using real images and the integrated LLM and spatial mapping. Verify that notes are accurately detected, mapped, and created in the target MCS server.

[ ] Comprehensive Error Testing: Re-run error testing scenarios with the full backend logic to ensure robust handling of issues related to LLM communication, mapping failures, and MCS interactions.

[ ] Performance Testing (Basic): Get a sense of how long the scanning process takes with the integrated LLM and mapping.