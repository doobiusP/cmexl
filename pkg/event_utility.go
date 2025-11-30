package cmexl_utils

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	CmexlRegex = regexp.MustCompile(`\[CMEXL\]\s*(?P<log>.*)$`)

	VcpkgPkgDetailsRegex         = regexp.MustCompile(`(?P<package>[\w\-]+(?:\[[^\]]*\])?):(?P<triplet>[\w\-]+)(?:@(?P<version>[\w\.\-\+]+)(?:#(?P<patch>\d+))?)?`)
	VcpkgAlreadyInstalledRegex   = regexp.MustCompile(`The following packages are already installed`)
	VcpkgInstalledDelimeterRegex = regexp.MustCompile(`Additional packages \(\*\) will be modified to complete this operation`)
	VcpkgNeedInstalledRegex      = regexp.MustCompile(`The following packages will be (?:built and installed|rebuilt|removed|installed)`)

	VcpkgLockRegex         = regexp.MustCompile(`waiting to take filesystem lock`)
	VcpkgCompilerHashRegex = regexp.MustCompile(`Detecting compiler hash`)

	VcpkgStartRegex   = regexp.MustCompile(`Running vcpkg install`)
	VcpkgFailedRegex  = regexp.MustCompile(`Running vcpkg install - failed`)
	VcpkgSuccessRegex = regexp.MustCompile(`Running vcpkg install - done`)

	VcpkgManifestLogRegex = regexp.MustCompile(`(?P<manifest_log>\S*vcpkg-manifest-install\.log)`)
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
		eventInfo = fmt.Sprintf("Exit code: %s, Error: %s", s.ExitCode, s.Err)
	case ExecErr:
		s := e.Payload.(ExecErrPayload)
		eventInfo = fmt.Sprintf("Error: %s", s.Err)
	default:
		eventInfo = "UNABLE TO DECODE PAYLOAD"
	}
	return fmt.Sprintf("[%s:%s](%s) %s", e.Key.Name, e.Key.Type.String(), e.Type.String(), eventInfo)
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

type CmexlStateMachine struct {
	PrKey                      PresetInfoKey
	VcpkgAlreadyInstalledCount int16
	VcpkgNeedInstalledCount    int16
}

type CmexlStateFunc func(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc

// NI stands for need to install
func CmexlVcpkgNIStateFn(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc {
	send := func(log string) {
		cmakeLogEvent := NewLogLineEvent(smPtr.PrKey, log)
		TrySend(eventsCh, cmakeLogEvent)
	}

	if match := VcpkgPkgDetailsRegex.FindStringSubmatch(line); match != nil {
		pkgIdx := VcpkgPkgDetailsRegex.SubexpIndex("package")
		versionIdx := VcpkgPkgDetailsRegex.SubexpIndex("version")
		patchIdx := VcpkgPkgDetailsRegex.SubexpIndex("patch")

		patchStr := match[patchIdx]
		if patchStr == "" {
			patchStr = "/NA"
		}

		var action string
		if strings.ContainsAny(line, "install") {
			action = "installing"
		} else if strings.Contains(line, "rebuilt") {
			action = "rebuilding"
		} else if strings.Contains(line, "remove") {
			action = "removing"
		}

		out := fmt.Sprintf("Now %s %s @ %s with vcpkg patch %s", action, match[pkgIdx], match[versionIdx], patchStr)
		smPtr.VcpkgNeedInstalledCount += 1
		send(out)
		return CmexlVcpkgNIStateFn
	}
	if match := VcpkgFailedRegex.FindString(line); len(match) != 0 {
		send("vcpkg failed - check logs")
		return CmexlDefaultStateFn
	}
	if match := VcpkgSuccessRegex.FindString(line); len(match) != 0 {
		send("vcpkg success")
		return CmexlDefaultStateFn
	}

	return CmexlVcpkgNIStateFn
}

// AI stands for already installed
func CmexlVcpkgAIStateFn(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc {
	send := func(log string) {
		cmakeLogEvent := NewLogLineEvent(smPtr.PrKey, log)
		TrySend(eventsCh, cmakeLogEvent)
	}

	if match := VcpkgPkgDetailsRegex.FindStringSubmatch(line); match != nil {
		pkgIdx := VcpkgPkgDetailsRegex.SubexpIndex("package")
		versionIdx := VcpkgPkgDetailsRegex.SubexpIndex("version")
		patchIdx := VcpkgPkgDetailsRegex.SubexpIndex("patch")

		patchStr := match[patchIdx]
		if patchStr == "" {
			patchStr = "/NA"
		}

		var action string
		if strings.ContainsAny(line, "install") {
			action = "installing"
		} else if strings.Contains(line, "rebuilt") {
			action = "rebuilding"
		} else if strings.Contains(line, "remove") {
			action = "removing"
		}

		out := fmt.Sprintf("Now %s %s @ %s with vcpkg patch %s", action, match[pkgIdx], match[versionIdx], patchStr)
		smPtr.VcpkgAlreadyInstalledCount += 1
		send(out)
		return CmexlVcpkgAIStateFn
	}

	if match := VcpkgNeedInstalledRegex.FindString(line); len(match) != 0 {
		send("checking for packages that need to be installed")
		return CmexlVcpkgNIStateFn
	}

	if match := VcpkgFailedRegex.FindString(line); len(match) != 0 {
		send("vcpkg failed - check logs")
		return CmexlDefaultStateFn
	}
	if match := VcpkgSuccessRegex.FindString(line); len(match) != 0 {
		send("vcpkg success")
		return CmexlDefaultStateFn
	}
	// TODO: Put VcpkgInstalledDelimeterRegex here as a check and move into a new state where you stop counting

	return CmexlVcpkgAIStateFn
}

func CmexlVcpkgStateFn(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc {
	send := func(log string) {
		cmakeLogEvent := NewLogLineEvent(smPtr.PrKey, log)
		TrySend(eventsCh, cmakeLogEvent)
	}
	if match := VcpkgLockRegex.FindString(line); len(match) != 0 {
		send("waiting for vcpkg lock")
		return CmexlVcpkgStateFn
	}
	if match := VcpkgCompilerHashRegex.FindString(line); len(match) != 0 {
		send("checking for change in build environment")
		return CmexlVcpkgStateFn
	}

	if match := VcpkgFailedRegex.FindString(line); len(match) != 0 {
		send("vcpkg failed - check logs")
		return CmexlDefaultStateFn
	}
	if match := VcpkgSuccessRegex.FindString(line); len(match) != 0 {
		send("vcpkg success")
		return CmexlDefaultStateFn
	}

	if match := VcpkgNeedInstalledRegex.FindString(line); len(match) != 0 {
		send("checking for packages that need to be installed")
		return CmexlVcpkgNIStateFn
	}

	if match := VcpkgAlreadyInstalledRegex.FindString(line); len(match) != 0 {
		send("checking for already installed packages")
		return CmexlVcpkgAIStateFn
	}

	return CmexlVcpkgStateFn
}

func CmexlDefaultStateFn(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc {
	send := func(log string) {
		cmakeLogEvent := NewLogLineEvent(smPtr.PrKey, log)
		TrySend(eventsCh, cmakeLogEvent)
	}

	if match := VcpkgStartRegex.FindString(line); len(match) != 0 {
		send("starting vcpkg")
		return CmexlVcpkgStateFn
	}
	if match := CmexlRegex.FindStringSubmatch(line); match != nil {
		logIdx := CmexlRegex.SubexpIndex("log")
		out := match[logIdx]
		send(out)
	}
	return CmexlDefaultStateFn
}

func CreateCmexlStore(flags ScheduleFlags) error {
	err := os.MkdirAll(".cmexl", 0755)
	if err != nil {
		return err
	}
	if *flags.SaveEvents {
		err = os.MkdirAll(".cmexl/events", 0755)
	}
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
