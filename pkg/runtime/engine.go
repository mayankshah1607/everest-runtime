package runtime

type ComponentStatus string

var (
	ComponentStatusRecommended ComponentStatus = "recommended"
)

type ComponentType string

type Component struct {
	Version string
	Image   string
	Status  ComponentStatus
}

type VersionInfo struct {
	OperatorVersion string
	Components      map[ComponentType][]Component
}

// DatabaseEngine is an interface that defines how to get relevant metadata information about the Database Engine.
type DatabaseEngine interface {
	GetName() string
	GetInstalledOperatorVersion() (string, error)
	GetVersionInfo() (*VersionInfo, error)
	// TODO: operator upgrades
}
