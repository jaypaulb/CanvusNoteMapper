package llm

// Note represents a detected note in raw image pixel coordinates
// The image size must be provided separately for scaling.
type Note struct {
	Content string
	Color   string
	X, Y    int     // pixel location in image
	Width   int     // pixel width in image
	Height  int     // pixel height in image
	Scale   float64 // scale of the note (to match anchor scale)
}

// AnalyzeImage returns mock note data and the image size (width, height).
func AnalyzeImage(imageData []byte) ([]Note, int, int) {
	imgW, imgH := 1280, 720 // mock image size
	return []Note{
		// Top-left
		{Content: "Top Left", Color: "#FF0000", X: 0, Y: 0, Width: 200, Height: 200},
		// Top-right
		{Content: "Top Right", Color: "#00FF00", X: 1080, Y: 0, Width: 200, Height: 200},
		// Bottom-left
		{Content: "Bottom Left", Color: "#0000FF", X: 0, Y: 520, Width: 200, Height: 200},
		// Bottom-right
		{Content: "Bottom Right", Color: "#FFFF00", X: 1080, Y: 520, Width: 200, Height: 200},
		// Center
		{Content: "Center", Color: "#800080", X: 540, Y: 260, Width: 200, Height: 200},
	}, imgW, imgH
}
