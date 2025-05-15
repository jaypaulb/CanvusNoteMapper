# CanvusNoteMapper
App to map IRL notes from an image to digital notes in an anchor in canvus using the API and LLMs.

## LLM Integration

The backend now uses the Google Gemini LLM (e.g., gemini-2.0-flash or gemini-1.5-flash) for extracting Post-it notes from images. The LLM integration is fully implemented and tested. You must provide a valid API key in your `.env` file as `GOOGLE_GENAI_API_KEY`.

- The model and generation config are set in `internal/llm/extract_postit_notes.go`.
- The backend parses the LLM's JSON response to extract note data.

## .env Requirements

Create a `.env` file in the project root with the following variables:

```
# Required
GOOGLE_GENAI_API_KEY=your_google_gemini_api_key_here

# Optional
PORT=8080  # Default is 8080 if not specified
```

## Status

- The backend and frontend are now fully integrated and tested end-to-end with real LLM and spatial mapping logic.

## Download Binaries

Prebuilt binaries for Windows and Linux are available from the [GitHub Releases page](https://github.com/jaypaulb/CanvusNoteMapper/releases/latest).

- [Download notescanner.exe (Windows)](https://github.com/jaypaulb/CanvusNoteMapper/releases/latest/download/notescanner.exe)
- [Download notescanner-linux-amd64 (Linux)](https://github.com/jaypaulb/CanvusNoteMapper/releases/latest/download/notescanner-linux-amd64)

> **Note:** Binaries are no longer included in the repository. Always download the latest version from the Releases page above.

## Deployment Instructions

### 1. Prepare the .env File

Create a `.env` file in the project root with your Google Gemini API key:

```
GOOGLE_GENAI_API_KEY=your_google_gemini_api_key_here
```

### 2. Running on Windows

1. Download the prebuilt binary from the [Releases page](https://github.com/jaypaulb/CanvusNoteMapper/releases/latest):
   - [notescanner.exe](https://github.com/jaypaulb/CanvusNoteMapper/releases/latest/download/notescanner.exe)
2. Open PowerShell, navigate to the download location, and run:
   ```powershell
   .\notescanner.exe
   ```
3. The server will start on `http://localhost:8080` by default.

### 3. Running on Ubuntu 18.04 (Linux)

1. Download the prebuilt binary from the [Releases page](https://github.com/jaypaulb/CanvusNoteMapper/releases/latest):
   - [notescanner-linux-amd64](https://github.com/jaypaulb/CanvusNoteMapper/releases/latest/download/notescanner-linux-amd64)
2. Copy the binary and `.env` file to your Ubuntu server:
   ```sh
   scp notescanner-linux-amd64 user@your-server:/path/to/app/
   scp .env user@your-server:/path/to/app/
   ```
3. SSH into your server and make the binary executable:
   ```sh
   chmod +x notescanner-linux-amd64
   ```
4. Run the server:
   ```sh
   ./notescanner-linux-amd64
   ```
5. The server will start on `http://localhost:8080` by default.

### 4. Customization
- **Port Configuration**: 
  - Set the `PORT` environment variable in your `.env` file or system environment
  - Example: `PORT=3000` to run on port 3000
  - If not specified, the server defaults to port 8080
- Ensure all dependencies (such as the Gemini API key) are set in the environment or `.env` file.
