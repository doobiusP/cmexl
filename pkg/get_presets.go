package cmexl_utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

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

func extractPresets(prType Preset_t, path string, prMap map[PresetInfoKey]PresetInfo, prJson map[string]json.RawMessage) error {
	presetField := fmt.Sprintf("%sPresets", prType.String())
	var currPresetList []PresetInfo
	if data, ok := prJson[presetField]; ok {
		if err := json.Unmarshal(data, &currPresetList); err != nil {
			return fmt.Errorf("unmarshall error: %w", err)
		}
	}
	for _, preset := range currPresetList {
		if len(preset.DisplayName) == 0 {
			preset.DisplayName = "-UNKNOWN-"
		}
		preset.File = path
		preset.Type = prType
		prMap[PresetInfoKey{preset.Name, preset.Type}] = preset
	}
	return nil
}

func getPresetsRecur(path string, prType Preset_t, prMap map[PresetInfoKey]PresetInfo) error {
	f, fileErr := os.Open(path)
	if fileErr != nil {
		return fileErr
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, 10<<20))

	var prJson map[string]json.RawMessage
	if err := dec.Decode(&prJson); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}

	if prType == All {
		allPresets := []Preset_t{Configure, Build, Test, Package, Workflow}
		for _, pr := range allPresets {
			err := extractPresets(pr, path, prMap, prJson)
			if err != nil {
				return err
			}
		}
	} else {
		err := extractPresets(prType, path, prMap, prJson)
		if err != nil {
			return err
		}
	}

	var nextIncludes []string
	if data, ok := prJson["include"]; ok {
		if err := json.Unmarshal(data, &nextIncludes); err != nil {
			return fmt.Errorf("unmarshall error: %w", err)
		}
	}
	for _, nextIncludePath := range nextIncludes {
		if !fileExists(nextIncludePath) {
			errMsg := fmt.Sprintf("can't find the next include presets file at %s", nextIncludePath)
			return errors.New(errMsg)
		}
		err := getPresetsRecur(nextIncludePath, prType, prMap)

		if err != nil {
			return err
		}
	}
	return nil
}

func GetCmakePresets(prType Preset_t) (map[PresetInfoKey]PresetInfo, error) {
	prMap := make(map[PresetInfoKey]PresetInfo)
	if fileExists("CMakeUserPresets.json") {
		err := getPresetsRecur("CMakeUserPresets.json", prType, prMap)
		if err != nil {
			return nil, err
		}
	} else if fileExists("CMakePresets.json") {
		err := getPresetsRecur("CMakePresets.json", prType, prMap)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("can't find either CMakeUserPresets or CMakePresets")
	}
	return prMap, nil
}
