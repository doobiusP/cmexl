package cmexl_utils

import (
	"errors"
	"fmt"
	"os"
)

type PresetEventState struct {
	VcpkgRunning               bool
	VcpkgReadingInstalled      bool
	VcpkgReadingNeeded         bool
	VcpkgAlreadyInstalledCount int16
	VcpkgNeedInstalledCount    int16
}

type CmexlEventData struct {
	Log         string
	ElapsedTime float64
}

type CmexlEvent struct {
	Key    PresetInfoKey
	Type   CmexlEvent_t
	Data   CmexlEventData
	Result error
}

type presetState struct {
	ElapsedTime float32
	Log         string
}

type CmexlEvent_t int

const (
	TimerUpdate CmexlEvent_t = iota
	LogLineUpdate
	ExecFinished
	ExecKilled
)

func CreateCmexlStore() error {
	err := os.MkdirAll(".cmexl", 0755)
	return err
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

func (et CmexlEvent_t) String() string {
	switch et {
	case TimerUpdate:
		return "TimerUpdate"
	case LogLineUpdate:
		return "LogLineUpdate"
	case ExecFinished:
		return "ExecFinished"
	case ExecKilled:
		return "ExecKilled"
	default:
		return "UNKNOWN"
	}
}

func (e CmexlEvent) String() string {
	return fmt.Sprintf("[%s:%s](%s) Elapsed time: %v, Log: %s", e.Key.Name, e.Key.Type.String(), e.Type.String(), e.Data.ElapsedTime, e.Data.Log)
}

func TrySend(events chan<- CmexlEvent, event CmexlEvent) {
	select {
	case events <- event:
	default:
		return
	}
}

func reportCMakeErr(err error, eventsCh chan<- CmexlEvent, prKey PresetInfoKey) {
	cmakeErrEventData := CmexlEventData{Log: "error", ElapsedTime: -1}
	cmakeErrEvent := CmexlEvent{Key: prKey, Type: ExecKilled, Data: cmakeErrEventData, Result: err}
	TrySend(eventsCh, cmakeErrEvent)
}
