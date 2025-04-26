package llm

// Note represents a detected note
type Note struct {
	Content string
	Color   string
	X, Y    int
	Width   int
	Height  int
}

// AnalyzeImage is a stub that returns mock note data. Replace with actual LLM integration.
func AnalyzeImage(imageData []byte) []Note {
	return []Note{
		{Content: "Red Note", Color: "#FF0000", X: 100, Y: 30, Width: 60, Height: 40},
		{Content: "Green Note", Color: "#00FF00", X: 170, Y: 110, Width: 60, Height: 40},
		{Content: "Blue Note", Color: "#0000FF", X: 140, Y: 200, Width: 60, Height: 40},
		{Content: "Yellow Note", Color: "#FFFF00", X: 60, Y: 200, Width: 60, Height: 40},
		{Content: "Purple Note", Color: "#800080", X: 30, Y: 110, Width: 60, Height: 40},
	}
}
