package extension

// Extension contains the information about the extension.
type Extension struct {
	// Name of the extension.
	Name string `json:"name"`
	// Name of the engine handled by this extension.
	Engine string `json:"engine"`
	// Version of the extension.
	Version string `json:"version"`
	// Description of the extension.
	Description string `json:"description"`
}
