package tui

import (
	"github.com/gentleman-programming/project-accelerator/internal/config"
	"github.com/gentleman-programming/project-accelerator/internal/core"
)

// backToMenuMsg signals the app to return to the main menu.
type backToMenuMsg struct{}

// configLoadedMsg carries the loaded configuration.
type configLoadedMsg struct {
	cfg *config.Config
}

// configErrorMsg carries a configuration loading error.
type configErrorMsg struct {
	err error
}

// registryLoadedMsg carries the loaded registry.
type registryLoadedMsg struct {
	registry *config.Registry
}

// registryErrorMsg carries a registry loading error.
type registryErrorMsg struct {
	err error
}

// scaffoldDoneMsg carries the result of a scaffold operation.
type scaffoldDoneMsg struct {
	result *core.ScaffoldResult
}

// scaffoldErrorMsg carries a scaffold operation error.
type scaffoldErrorMsg struct {
	err error
}

// syncDoneMsg carries the result of a sync operation.
type syncDoneMsg struct {
	result *core.SyncResult
}

// syncErrorMsg carries a sync operation error.
type syncErrorMsg struct {
	err error
}
