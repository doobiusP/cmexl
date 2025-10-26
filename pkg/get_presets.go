package cmexl_utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
)

type Preset_t int

const (
	Configure Preset_t = iota
	Build
	Test
	Package
	Workflow
	All
)

var AllowedPresetTypes = []string{"configure", "build", "test", "package", "workflow", "all"}

func PresetIsAllowed(pr string) bool {
	return slices.Contains(AllowedPresetTypes, pr)
}

func MapPresetToType(pr string) (Preset_t, error) {
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
		return All, errors.New("found unexpected preset string")
	}
}

type PresetInfoKey struct {
	Name string
	Type Preset_t
}

type PresetInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	File        string
	Type        Preset_t
}

func (pt Preset_t) String() (string, error) {
	switch pt {
	case Configure:
		return "configure", nil
	case Build:
		return "build", nil
	case Test:
		return "test", nil
	case Package:
		return "package", nil
	case Workflow:
		return "workflow", nil
	case All:
		return "all", nil
	default:
		return "", errors.New("unknown cmake preset type")
	}
}
func (pIt PresetInfo) String() string {
	var msg string
	msg += fmt.Sprintf("Name: %s, ", pIt.Name)
	msg += fmt.Sprintf("DisplayName: %s, ", pIt.DisplayName)
	msg += fmt.Sprintf("File: %s, ", pIt.File)
	prType, _ := pIt.Type.String()
	msg += fmt.Sprintf("Type: %s", prType)
	return msg
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Unknown error while trying to read %s: %v", path, err)
		}
		return false
	}
	return true
}

func extractPresets(presetType Preset_t, path string, presetsMap map[PresetInfoKey]PresetInfo, obj map[string]json.RawMessage) error {
	presetStr, pErr := presetType.String()
	if pErr != nil {
		return pErr
	}
	presetField := fmt.Sprintf("%sPresets", presetStr)
	var currPresetList []PresetInfo
	if data, ok := obj[presetField]; ok {
		if err := json.Unmarshal(data, &currPresetList); err != nil {
			return fmt.Errorf("unmarshall error: %w", err)
		}
	}
	for _, preset := range currPresetList {
		if len(preset.DisplayName) == 0 {
			preset.DisplayName = "-UNKNOWN-"
		}
		preset.File = path
		preset.Type = presetType
		presetsMap[PresetInfoKey{preset.Name, preset.Type}] = preset
	}
	return nil
}

func getPresetsRecur(path string, presetType Preset_t, presetsMap map[PresetInfoKey]PresetInfo) error {
	f, fileErr := os.Open(path)
	if fileErr != nil {
		return fileErr
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, 10<<20))

	var obj map[string]json.RawMessage
	if err := dec.Decode(&obj); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}

	if presetType == All {
		allPresets := []Preset_t{Configure, Build, Test, Package, Workflow}
		for _, pr := range allPresets {
			err := extractPresets(pr, path, presetsMap, obj)
			if err != nil {
				return err
			}
		}
	} else {
		err := extractPresets(presetType, path, presetsMap, obj)
		if err != nil {
			return err
		}
	}

	var nextIncludes []string
	if data, ok := obj["include"]; ok {
		if err := json.Unmarshal(data, &nextIncludes); err != nil {
			return fmt.Errorf("unmarshall error: %w", err)
		}
	}
	for _, nextIncludePath := range nextIncludes {
		if !fileExists(nextIncludePath) {
			errMsg := fmt.Sprintf("can't find the next include presets file at %s", nextIncludePath)
			return errors.New(errMsg)
		}
		err := getPresetsRecur(nextIncludePath, presetType, presetsMap)

		if err != nil {
			return err
		}
	}
	return nil
}

func GetCmakePresets(presetType Preset_t) (map[PresetInfoKey]PresetInfo, error) {
	presetsMap := make(map[PresetInfoKey]PresetInfo)
	if fileExists("CMakeUserPresets.json") {
		err := getPresetsRecur("CMakeUserPresets.json", presetType, presetsMap)
		if err != nil {
			return nil, err
		}
	} else if fileExists("CMakePresets.json") {
		err := getPresetsRecur("CMakePresets.json", presetType, presetsMap)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("can't find either CMakeUserPresets or CMakePresets")
	}
	return presetsMap, nil
}
