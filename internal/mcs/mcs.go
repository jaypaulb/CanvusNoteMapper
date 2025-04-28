package mcs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jaypaulb/CanvusNoteMapper/internal/canvusapi"
)

// MCSClient wraps MCS API interactions
type MCSClient struct {
	Server   string
	APIKey   string
	CanvasID string
}

type CanvasInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AnchorInfo struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	// Add more fields as needed
}

type CanvasSize struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func NewClient(server, apiKey, canvasID string) *MCSClient {
	return &MCSClient{Server: server, APIKey: apiKey, CanvasID: canvasID}
}

func (c *MCSClient) GetCanvases() ([]CanvasInfo, error) {
	client := canvusapi.NewClient(c.Server, c.CanvasID, c.APIKey)
	var canvasesRaw []map[string]interface{}
	url := client.Server + "/api/v1/canvases"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Private-Token", c.APIKey)
	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MCS error: %s", string(bodyBytes))
	}
	if err := json.NewDecoder(resp.Body).Decode(&canvasesRaw); err != nil {
		return nil, err
	}
	result := make([]CanvasInfo, 0, len(canvasesRaw))
	for _, canvas := range canvasesRaw {
		id, _ := canvas["id"].(string)
		name, _ := canvas["name"].(string)
		if id != "" && name != "" {
			result = append(result, CanvasInfo{ID: id, Name: name})
		}
	}
	return result, nil
}

func (c *MCSClient) GetAnchors(canvasID string) ([]AnchorInfo, error) {
	client := canvusapi.NewClient(c.Server, canvasID, c.APIKey)
	// Fetch anchors from the correct endpoint
	url := client.Server + "/api/v1/canvases/" + canvasID + "/anchors"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Private-Token", c.APIKey)
	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MCS error: %s", string(bodyBytes))
	}
	var anchorsRaw []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&anchorsRaw); err != nil {
		return nil, err
	}
	result := make([]AnchorInfo, 0, len(anchorsRaw))
	for i, a := range anchorsRaw {
		anchor := AnchorInfo{}
		if id, ok := a["id"].(string); ok {
			anchor.ID = id
		}
		if name, ok := a["anchor_name"].(string); ok {
			anchor.Name = name
		}
		if loc, ok := a["location"].(map[string]interface{}); ok {
			if x, ok := loc["x"].(float64); ok {
				anchor.X = x
			}
			if y, ok := loc["y"].(float64); ok {
				anchor.Y = y
			}
		}
		if size, ok := a["size"].(map[string]interface{}); ok {
			if w, ok := size["width"].(float64); ok {
				anchor.Width = w
			}
			if h, ok := size["height"].(float64); ok {
				anchor.Height = h
			}
		}
		// Add more fields as needed
		fmt.Printf("[GetAnchors] Parsed anchor %d: %+v\n", i, anchor)
		result = append(result, anchor)
	}
	fmt.Printf("[GetAnchors] Returning %d anchors\n", len(result))
	return result, nil
}

func (c *MCSClient) CreateNote(canvasID string, noteData map[string]interface{}) (map[string]interface{}, error) {
	client := canvusapi.NewClient(c.Server, canvasID, c.APIKey)
	return client.CreateNote(noteData)
}

func (c *MCSClient) GetCanvasSize(canvasID string) (*CanvasSize, error) {
	client := canvusapi.NewClient(c.Server, canvasID, c.APIKey)
	widgets, err := client.GetWidgets(false)
	if err != nil {
		return nil, err
	}
	for _, w := range widgets {
		if wtype, ok := w["widget_type"].(string); ok && wtype == "SharedCanvas" {
			width, _ := w["width"].(float64)
			height, _ := w["height"].(float64)
			return &CanvasSize{Width: width, Height: height}, nil
		}
	}
	return nil, fmt.Errorf("SharedCanvas widget not found")
}
