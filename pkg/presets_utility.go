package cmexl_utils

import (
	"errors"
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

type PresetStr_t string

const (
	ConfigureStr PresetStr_t = "configure"
	BuildStr     PresetStr_t = "build"
	TestStr      PresetStr_t = "test"
	PackageStr   PresetStr_t = "package"
	WorkflowStr  PresetStr_t = "workflow"
	AllStr       PresetStr_t = "all"
)

type PresetMap_t map[PresetInfoKey]PresetInfo

func PresetIsAllowed(pr string) bool {
	switch pr {
	case string(ConfigureStr),
		string(BuildStr),
		string(TestStr),
		string(PackageStr),
		string(WorkflowStr),
		string(AllStr):
		return true
	default:
		return false
	}
}

func MapPresetToType(pr PresetStr_t) (Preset_t, error) {
	switch pr {
	case ConfigureStr:
		return Configure, nil
	case BuildStr:
		return Build, nil
	case TestStr:
		return Test, nil
	case PackageStr:
		return Package, nil
	case WorkflowStr:
		return Workflow, nil
	case AllStr:
		return All, nil
	default:
		return None, errors.New("found unexpected preset string")
	}
}

func PresetStrToType(pr string) (Preset_t, error) {
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
		return None, errors.New("found unexpected preset string")
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

func (prStr *PresetStr_t) String() string {
	return string(*prStr)
}

func (prStr *PresetStr_t) Set(v string) error {
	switch v {
	case "configure", "build", "test", "package", "workflow", "all":
		*prStr = PresetStr_t(v)
		return nil
	default:
		return errors.New(`must be one of {"configure", "build", "test", "package", "workflow", "all"}`)
	}
}

func (prStr *PresetStr_t) Type() string {
	return "PresetStr_t"
}
