package mapping

import "github.com/jaypaulb/CanvusNoteMapper/internal/llm"

// MapNotesToZone maps notes from image pixel coordinates to the target zone, scaling and offsetting as needed.
func MapNotesToZone(notes []llm.Note, imageWidth, imageHeight int, zoneDimensions [2]int, zoneLocation [2]int) []llm.Note {
	zoneW, zoneH := zoneDimensions[0], zoneDimensions[1]
	zoneX, zoneY := zoneLocation[0]+10, zoneLocation[1]+10
	if imageWidth == 0 || imageHeight == 0 || zoneW == 0 || zoneH == 0 {
		return notes // fallback: no mapping
	}

	// Calculate scale to fit the image into the zone, maintaining aspect ratio
	scale := min(float64(zoneW)/float64(imageWidth), float64(zoneH)/float64(imageHeight))
	offsetX := (float64(zoneW) - float64(imageWidth)*scale) / 2
	offsetY := (float64(zoneH) - float64(imageHeight)*scale) / 2

	mapped := make([]llm.Note, len(notes))
	for i, n := range notes {
		mapped[i] = n
		mapped[i].X = int(float64(n.X)*scale+offsetX) + zoneX
		mapped[i].Y = int(float64(n.Y)*scale+offsetY) + zoneY
		mapped[i].Width = int(float64(n.Width) * scale)
		mapped[i].Height = int(float64(n.Height) * scale)
	}
	return mapped
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// MapNotesToMCSFormat transforms mapped notes to the MCS API note creation format.
func MapNotesToMCSFormat(notes []llm.Note) []map[string]interface{} {
	mcsNotes := make([]map[string]interface{}, len(notes))
	for i, n := range notes {
		mcsNotes[i] = map[string]interface{}{
			"background_color": n.Color,
			"text":             n.Content,
			"location": map[string]interface{}{
				"x": n.X,
				"y": n.Y,
			},
			"size": map[string]interface{}{
				"width":  n.Width,
				"height": n.Height,
			},
			"scale":       n.Scale,
			"widget_type": "Note",
			"state":       "normal",
			// Add more fields as needed (e.g., pinned, depth, etc.)
		}
	}
	return mcsNotes
}
