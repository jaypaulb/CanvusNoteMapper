package mcs

import (
	"github.com/jaypaulb/CanvusNoteMapper/libs/canvusapi"
)

// MCSClient wraps MCS API interactions
type MCSClient struct {
	Server   string
	APIKey   string
	CanvasID string
}

func NewClient(server, apiKey, canvasID string) *MCSClient {
	return &MCSClient{Server: server, APIKey: apiKey, CanvasID: canvasID}
}

func (c *MCSClient) GetCanvases() ([]string, error) {
	client := canvusapi.NewClient(c.Server, c.CanvasID, c.APIKey)
	info, err := client.GetCanvasInfo()
	if err != nil {
		return nil, err
	}
	canvases, ok := info["canvases"].([]interface{})
	if !ok {
		return nil, nil
	}
	result := make([]string, 0, len(canvases))
	for _, v := range canvases {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result, nil
}

func (c *MCSClient) GetAnchorZones(canvasID string) ([]string, error) {
	client := canvusapi.NewClient(c.Server, canvasID, c.APIKey)
	info, err := client.GetCanvasInfo()
	if err != nil {
		return nil, err
	}
	zones, ok := info["zones"].([]interface{})
	if !ok {
		return nil, nil
	}
	result := make([]string, 0, len(zones))
	for _, v := range zones {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result, nil
}

func (c *MCSClient) CreateNote(canvasID string, noteData map[string]interface{}) (map[string]interface{}, error) {
	client := canvusapi.NewClient(c.Server, canvasID, c.APIKey)
	return client.CreateNote(noteData)
}
