package llm

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// ExtractPostitNotesInput represents the input for extracting Post-it notes from an image.
type ExtractPostitNotesInput struct {
	ImageData []byte `json:"imageData"`
}

// ExtractPostitNotesOutput represents a single extracted Post-it note in the required format.
type ExtractPostitNotesOutput struct {
	BackgroundColor string         `json:"background_color"`
	Location        map[string]int `json:"location"`
	Scale           float64        `json:"scale"`
	Size            map[string]int `json:"size"`
	State           string         `json:"state"`
	Text            string         `json:"text"`
	WidgetType      string         `json:"widget_type"`
}

// ExtractPostitNotes extracts notes from an image using Google Gemini.
func ExtractPostitNotes(input ExtractPostitNotesInput) ([]ExtractPostitNotesOutput, error) {
	apiKey := os.Getenv("GOOGLE_GENAI_API_KEY")
	if apiKey == "" {
		log.Printf("[ExtractPostitNotes] GOOGLE_GENAI_API_KEY environment variable is not set")
		return nil, errors.New("GOOGLE_GENAI_API_KEY not set in environment")
	}
	log.Printf("[ExtractPostitNotes] API key found, length: %d", len(apiKey))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	log.Printf("[ExtractPostitNotes] Created context with 5-minute timeout")

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("[ExtractPostitNotes] Failed to create Gemini client: %v", err)
		return nil, err
	}
	defer client.Close()
	log.Printf("[ExtractPostitNotes] Successfully created Gemini client")

	log.Printf("[ExtractPostitNotes] Processing image data, size: %d bytes", len(input.ImageData))

	// Upload the image file to Gemini
	file, err := client.UploadFile(ctx, input.ImageData, "image/png")
	if err != nil {
		log.Printf("[ExtractPostitNotes] Failed to upload file to Gemini: %v", err)
		return nil, err
	}
	defer file.Close()
	log.Printf("[ExtractPostitNotes] Successfully uploaded file to Gemini")

	// Use controlled generation with responseSchema and responseMimeType
	config := &genai.GenerationConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"background_color": {Type: genai.TypeString},
					"location": {
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"x": {Type: genai.TypeInteger},
							"y": {Type: genai.TypeInteger},
						},
						Required: []string{"x", "y"},
					},
					"scale": {Type: genai.TypeNumber},
					"size": {
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"height": {Type: genai.TypeInteger},
							"width":  {Type: genai.TypeInteger},
						},
						Required: []string{"height", "width"},
					},
					"text":        {Type: genai.TypeString},
					"widget_type": {Type: genai.TypeString},
				},
				Required: []string{"background_color", "location", "scale", "size", "text", "widget_type"},
			},
		},
	}
	log.Printf("[ExtractPostitNotes] Created generation config with JSON schema")

	model := client.GenerativeModel("gemini-2.5-flash-preview-05-20")
	model.GenerationConfig = *config
	log.Printf("[ExtractPostitNotes] Created model with config")

	prompt := `Analyze <image> for post-it notes. Extract content, color, size, and precise top-left pixel location ('x','y'). Relative positioning and size matter, but location ('x','y') is key.

NB: The relative location of the notes within the image frame is important.

Return JSON array. Each object structure:
{
  "background_color": "<hex_code>",
  "location": {"x": <pixel>, "y": <pixel>}, // Top-left corner
  "scale": <float>,
  "size": {"height": <pixel>, "width": <pixel>},
  "state": "<string>",
  "text": "<extracted_text>",
  "widget_type": "Note"
}`

	parts := []genai.Part{
		genai.Text(prompt),
		file,
	}
	log.Printf("[ExtractPostitNotes] Created parts with prompt and uploaded file")

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		log.Printf("[ExtractPostitNotes] Failed to generate content: %v", err)
		return nil, err
	}
	log.Printf("[ExtractPostitNotes] Successfully generated content from model")

	// Log the full raw LLM response for debugging
	respJson, _ := json.MarshalIndent(resp, "", "  ")
	log.Printf("[ExtractPostitNotes] Raw LLM response: %s", string(respJson))

	// Parse the JSON array from the LLM response
	var outputs []ExtractPostitNotesOutput
	for _, c := range resp.Candidates {
		if c.Content != nil {
			for _, part := range c.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					jsonStr := extractJSONFromMarkdown(string(txt))
					log.Printf("[ExtractPostitNotes] Extracted JSON string: %s", jsonStr)
					if err := json.Unmarshal([]byte(jsonStr), &outputs); err == nil && len(outputs) > 0 {
						log.Printf("[ExtractPostitNotes] Successfully parsed %d notes from response", len(outputs))
						return outputs, nil
					} else {
						log.Printf("[ExtractPostitNotes] Failed to parse JSON or no notes found: %v", err)
					}
				}
			}
		}
	}
	log.Printf("[ExtractPostitNotes] No valid JSON array found in LLM response")
	return nil, errors.New("No valid JSON array found in LLM response")
}

// extractJSONFromMarkdown extracts JSON from a Markdown code block if present.
func extractJSONFromMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = s[len("```json"):]
	} else if strings.HasPrefix(s, "```") {
		s = s[len("```"):]
	}
	if strings.HasSuffix(s, "```") {
		s = s[:len(s)-len("```")]
	}
	return strings.TrimSpace(s)
}
