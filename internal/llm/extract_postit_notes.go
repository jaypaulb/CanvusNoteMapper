package llm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// ExtractPostitNotesInput represents the input for extracting Post-it notes from an image.
type ExtractPostitNotesInput struct {
	PhotoDataURI string `json:"photoDataUri"`
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

// ExtractPostitNotes extracts notes from a photoDataUri using Google Gemini.
func ExtractPostitNotes(input ExtractPostitNotesInput) ([]ExtractPostitNotesOutput, error) {
	apiKey := os.Getenv("GOOGLE_GENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GOOGLE_GENAI_API_KEY not set in environment")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	imageData, err := decodeDataURI(input.PhotoDataURI)
	if err != nil {
		return nil, err
	}

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

	model := client.GenerativeModel("gemini-2.0-flash")
	model.GenerationConfig = *config
	parts := []genai.Part{
		genai.Text(prompt),
		genai.Blob{MIMEType: "image/png", Data: imageData},
	}
	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, err
	}

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
					if err := json.Unmarshal([]byte(jsonStr), &outputs); err == nil && len(outputs) > 0 {
						return outputs, nil
					}
				}
			}
		}
	}
	return nil, errors.New("No valid JSON array found in LLM response")
}

// decodeDataURI decodes a data URI and returns the raw image bytes.
func decodeDataURI(dataURI string) ([]byte, error) {
	// Regex to extract base64 data from data URI
	re := regexp.MustCompile(`^data:[^;]+;base64,(.*)$`)
	matches := re.FindStringSubmatch(dataURI)
	if len(matches) != 2 {
		return nil, errors.New("invalid data URI format")
	}
	return base64.StdEncoding.DecodeString(matches[1])
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
