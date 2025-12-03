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
	VcpkgWorkProgressRegex       = regexp.MustCompile(`(?P<action>Installing|Removing)\s+(?P<current>\d+)\/(?P<total>\d+)`)
	VcpkgNeedInstalledRegex      = regexp.MustCompile(`The following packages will be (?:built and installed|rebuilt|removed|installed)`)
	VcpkgNeedRemovedRegex        = regexp.MustCompile(`The following packages will be removed`)

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

type CmexlPresetData struct {
	Errors    []error
	ExecState *CmexlStateMachine
	EventsLog *os.File
	StdoutLog *os.File
	StderrLog *os.File
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
	VcpkgNeedRemovedCount      int16
	VcpkgNeedInstalledCount    int16
}

type CmexlStateFunc func(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc

func send(log string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) {
	cmakeLogEvent := NewLogLineEvent(smPtr.PrKey, log)
	TrySend(eventsCh, cmakeLogEvent)
}

func sendPackageDetails(action string, match []string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) {
	pkgIdx := VcpkgPkgDetailsRegex.SubexpIndex("package")
	versionIdx := VcpkgPkgDetailsRegex.SubexpIndex("version")
	patchIdx := VcpkgPkgDetailsRegex.SubexpIndex("patch")

	trimmedVersion := strings.TrimSuffix(match[versionIdx], "...")

	patchStr := match[patchIdx]

	var out string
	if patchStr == "" {
		if trimmedVersion == "" {
			out = fmt.Sprintf("%s %s", action, match[pkgIdx])
		} else {
			out = fmt.Sprintf("%s %s @ %s", action, match[pkgIdx], trimmedVersion)
		}
	} else {
		out = fmt.Sprintf("%s %s @ %s with vcpkg patch %s", action, match[pkgIdx], trimmedVersion, patchStr)
	}
	send(out, smPtr, eventsCh)
}

// NI stands for need to install
func CmexlVcpkgNIStateFn(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc {
	if match := VcpkgPkgDetailsRegex.FindStringSubmatch(line); match != nil {
		sendPackageDetails("Found needed to install", match, smPtr, eventsCh)
		smPtr.VcpkgNeedInstalledCount += 1
		return CmexlVcpkgNIStateFn
	}

	if match := VcpkgInstalledDelimeterRegex.FindString(line); len(match) != 0 {
		send(fmt.Sprintf("now building %d required packages and removing %d unecessary packages", smPtr.VcpkgNeedInstalledCount, smPtr.VcpkgNeedRemovedCount), smPtr, eventsCh)
		return CmexlVcpkgWorkingStateFn
	}

	return CmexlVcpkgNIStateFn
}

func CmexlVcpkgWorkingStateFn(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc {
	packageMatch := VcpkgPkgDetailsRegex.FindStringSubmatch(line)
	progressMatch := VcpkgWorkProgressRegex.FindStringSubmatch(line)

	if packageMatch != nil && progressMatch != nil {
		actionStr := strings.ToLower(progressMatch[VcpkgWorkProgressRegex.SubexpIndex("action")])
		currentStr := progressMatch[VcpkgWorkProgressRegex.SubexpIndex("current")]
		totalStr := progressMatch[VcpkgWorkProgressRegex.SubexpIndex("total")]

		finalAction := fmt.Sprintf("(%s/%s) Now %s", currentStr, totalStr, actionStr)
		sendPackageDetails(finalAction, packageMatch, smPtr, eventsCh)
		return CmexlVcpkgWorkingStateFn
	}

	if match := VcpkgFailedRegex.FindString(line); len(match) != 0 {
		send("vcpkg failed - check logs", smPtr, eventsCh)
		return CmexlDefaultStateFn
	}
	if match := VcpkgSuccessRegex.FindString(line); len(match) != 0 {
		send("vcpkg success", smPtr, eventsCh)
		return CmexlDefaultStateFn
	}

	return CmexlVcpkgWorkingStateFn
}

// AI stands for already installed
func CmexlVcpkgAIStateFn(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc {
	if match := VcpkgPkgDetailsRegex.FindStringSubmatch(line); match != nil {
		sendPackageDetails("Found installed", match, smPtr, eventsCh)
		smPtr.VcpkgAlreadyInstalledCount += 1
		return CmexlVcpkgAIStateFn
	}

	if match := VcpkgNeedRemovedRegex.FindString(line); len(match) != 0 {
		send("checking for packages that need to be removed", smPtr, eventsCh)
		return CmexlVcpkgNRStateFn
	}

	if match := VcpkgNeedInstalledRegex.FindString(line); len(match) != 0 {
		send("checking for packages that need to be installed", smPtr, eventsCh)
		return CmexlVcpkgNIStateFn
	}

	if match := VcpkgFailedRegex.FindString(line); len(match) != 0 {
		send("vcpkg failed - check logs", smPtr, eventsCh)
		return CmexlDefaultStateFn
	}
	if match := VcpkgSuccessRegex.FindString(line); len(match) != 0 {
		send("vcpkg success", smPtr, eventsCh)
		return CmexlDefaultStateFn
	}

	return CmexlVcpkgAIStateFn
}

// NR stands for need removed
func CmexlVcpkgNRStateFn(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc {
	if match := VcpkgPkgDetailsRegex.FindStringSubmatch(line); match != nil {
		sendPackageDetails("Found needed to remove", match, smPtr, eventsCh)
		smPtr.VcpkgNeedRemovedCount += 1
		return CmexlVcpkgNRStateFn
	}

	if match := VcpkgNeedInstalledRegex.FindString(line); len(match) != 0 {
		send("checking for packages that need to be installed", smPtr, eventsCh)
		return CmexlVcpkgNIStateFn
	}

	if match := VcpkgInstalledDelimeterRegex.FindString(line); len(match) != 0 {
		send(fmt.Sprintf("now building %d required packages and removing %d unecessary packages", smPtr.VcpkgNeedInstalledCount, smPtr.VcpkgNeedRemovedCount), smPtr, eventsCh)
		return CmexlVcpkgWorkingStateFn
	}
	return CmexlVcpkgNRStateFn
}

func CmexlVcpkgStateFn(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc {
	if match := VcpkgLockRegex.FindString(line); len(match) != 0 {
		send("waiting for vcpkg lock", smPtr, eventsCh)
		return CmexlVcpkgStateFn
	}
	if match := VcpkgCompilerHashRegex.FindString(line); len(match) != 0 {
		send("checking for change in build environment", smPtr, eventsCh)
		return CmexlVcpkgStateFn
	}

	if match := VcpkgFailedRegex.FindString(line); len(match) != 0 {
		send("vcpkg failed - check logs", smPtr, eventsCh)
		return CmexlDefaultStateFn
	}
	if match := VcpkgSuccessRegex.FindString(line); len(match) != 0 {
		send("vcpkg success", smPtr, eventsCh)
		return CmexlDefaultStateFn
	}

	if match := VcpkgAlreadyInstalledRegex.FindString(line); len(match) != 0 {
		send("checking for already installed packages", smPtr, eventsCh)
		return CmexlVcpkgAIStateFn
	}

	if match := VcpkgNeedRemovedRegex.FindString(line); len(match) != 0 {
		send("checking for packages that need to be removed", smPtr, eventsCh)
		return CmexlVcpkgNRStateFn
	}

	if match := VcpkgNeedInstalledRegex.FindString(line); len(match) != 0 {
		send("checking for packages that need to be installed", smPtr, eventsCh)
		return CmexlVcpkgNIStateFn
	}

	return CmexlVcpkgStateFn
}

func CmexlDefaultStateFn(line string, smPtr *CmexlStateMachine, eventsCh chan<- CmexlEvent) CmexlStateFunc {
	if match := VcpkgStartRegex.FindString(line); len(match) != 0 {
		send("starting vcpkg", smPtr, eventsCh)
		return CmexlVcpkgStateFn
	}
	if match := CmexlRegex.FindStringSubmatch(line); match != nil {
		logIdx := CmexlRegex.SubexpIndex("log")
		out := match[logIdx]
		send(out, smPtr, eventsCh)
	}
	return CmexlDefaultStateFn
}

func CreateCmexlStore(flags ScheduleFlags) error {
	err := os.MkdirAll(".cmexl", 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll(".cmexl/stderr", 0755)
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
