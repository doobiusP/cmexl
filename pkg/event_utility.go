package cmexl_utils

import (
	"errors"
	"fmt"
	"os"
)

type CmexlEvent_t int

const (
	TimerUpdate CmexlEvent_t = iota
	LogLineUpdate
	ExecExit
	ExecErr
)

func (et CmexlEvent_t) String() string {
	switch et {
	case TimerUpdate:
		return "TimerUpdate"
	case LogLineUpdate:
		return "LogLineUpdate"
	case ExecExit:
		return "ExecExit"
	case ExecErr:
		return "ExecErr"
	default:
		return "UNKNOWN"
	}
}

type TimerUpdatePayload struct {
	ElapsedTime float64
}

type LogLinePayload struct {
	Log string
}

type ExecExitPayload struct {
	Err      error
	ExitCode error
}

type ExecErrPayload struct {
	Err error
}

type CmexlEvent struct {
	Key     PresetInfoKey
	Type    CmexlEvent_t
	Payload any
}

// Minimum necessary information that UI needs to update
type DisplayState struct {
	Log         string
	ElapsedTime float64
}

func (e CmexlEvent) String() string {
	var eventInfo string
	switch e.Type {
	case TimerUpdate:
		s := e.Payload.(TimerUpdatePayload)
		eventInfo = fmt.Sprintf("Elapsed time: %f", s.ElapsedTime)
	case LogLineUpdate:
		s := e.Payload.(LogLinePayload)
		eventInfo = fmt.Sprintf("Log: %s", s.Log)
	case ExecExit:
		s := e.Payload.(ExecExitPayload)
		eventInfo = fmt.Sprintf("Exit code: %w, Error: %w", s.ExitCode, s.Err)
	case ExecErr:
		s := e.Payload.(ExecErrPayload)
		eventInfo = fmt.Sprintf("Error: %w", s.Err)
	default:
		eventInfo = "UNABLE TO DECODE PAYLOAD"
	}
	return fmt.Sprintf("[%s:%s](%s) Event Info: %s", e.Key.Name, e.Key.Type.String(), e.Type.String(), eventInfo)
}

func NewTimerUpdateEvent(key PresetInfoKey, elapsed float64) CmexlEvent {
	return CmexlEvent{
		Key:  key,
		Type: TimerUpdate,
		Payload: TimerUpdatePayload{
			ElapsedTime: elapsed,
		},
	}
}

func NewLogLineEvent(key PresetInfoKey, line string) CmexlEvent {
	return CmexlEvent{
		Key:  key,
		Type: LogLineUpdate,
		Payload: LogLinePayload{
			Log: line,
		},
	}
}

func NewExecExitEvent(key PresetInfoKey, err error, exitCode error) CmexlEvent {
	return CmexlEvent{
		Key:  key,
		Type: ExecExit,
		Payload: ExecExitPayload{
			Err:      err,
			ExitCode: exitCode,
		},
	}
}

func NewExecErrEvent(key PresetInfoKey, err error) CmexlEvent {
	return CmexlEvent{
		Key:  key,
		Type: ExecExit,
		Payload: ExecErrPayload{
			Err: err,
		},
	}
}

// type PresetEventState struct {
// 	VcpkgRunning               bool
// 	VcpkgReadingInstalled      bool
// 	VcpkgReadingNeeded         bool
// 	VcpkgAlreadyInstalledCount int16
// 	VcpkgNeedInstalledCount    int16
// }

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

func TrySend(events chan<- CmexlEvent, event CmexlEvent) {
	select {
	case events <- event:
	default:
		return
	}
}
