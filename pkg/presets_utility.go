package cmexl_utils

import (
	"fmt"
)

type Preset_t int

const (
	None Preset_t = iota
	Configure
	Build
	Test
	Package
	Workflow
	All
)

var (
	ConfigureStr = "configure"
	BuildStr     = "build"
	TestStr      = "test"
	PackageStr   = "package"
	WorkflowStr  = "workflow"
	AllStr       = "all"
)

type PresetMap_t map[PresetInfoKey]PresetInfo

func PresetIsAllowed(pr string) bool {
	switch pr {
	case ConfigureStr,
		BuildStr,
		TestStr,
		PackageStr,
		WorkflowStr,
		AllStr:
		return true
	default:
		return false
	}
}

func MapPresetStrToType(pr string) (Preset_t, error) {
	switch pr {
	case "configure":
		return Configure, nil
	case "build":
		return Build, nil
	case "test":
		return Test, nil
	case "package":
		return Package, nil
	case "workflow":
		return Workflow, nil
	case "all":
		return All, nil
	default:
		return None, fmt.Errorf("got unexpected preset string: %s", pr)
	}
}

type PresetInfoKey struct {
	Name string
	Type Preset_t
}

type PresetInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Hidden      bool   `json:"hidden"`
	File        string
	Type        Preset_t
}

func (pr Preset_t) String() string {
	switch pr {
	case Configure:
		return "configure"
	case Build:
		return "build"
	case Test:
		return "test"
	case Package:
		return "package"
	case Workflow:
		return "workflow"
	case All:
		return "all"
	default:
		return "None"
	}
}

func (prInfo PresetInfo) String() string {
	var msg string
	msg += fmt.Sprintf("Name: %s, ", prInfo.Name)
	msg += fmt.Sprintf("DisplayName: %s, ", prInfo.DisplayName)
	msg += fmt.Sprintf("File: %s, ", prInfo.File)
	msg += fmt.Sprintf("Type: %s", prInfo.Type.String())
	return msg
}
