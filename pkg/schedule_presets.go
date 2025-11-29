package cmexl_utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	defaultTickerFreqHz   = float32(5.0)
	tickerFreqHz          = float32(12.0)
	eventChannelSizeScale = 20

	cursorHide      = "\033[?25l"
	cursorShow      = "\033[?25h"
	cursorHome      = "\033[H"
	clearFromCursor = "\033[0J"

	cmexlRegex = regexp.MustCompile(`\[CMEXL\]\s*(?P<log>.*)$`)

	vcpkgPkgDetailsRegex       = regexp.MustCompile(`(?P<package>[\w\-]+(?:\[[^\]]*\])?):(?P<triplet>[\w\-]+)(?:@(?P<version>[\w\.\-\+]+)(?:#(?P<patch>\d+))?)?`)
	vcpkgLockRegex             = regexp.MustCompile(`waiting to take filesystem lock`)
	vcpkgAlreadyInstalledRegex = regexp.MustCompile(`The following packages are already installed`)
	vcpkgNeedInstalledRegex    = regexp.MustCompile(`The following packages will be (?:built and installed|rebuilt|removed)`)
	vcpkgCompilerHashRegex     = regexp.MustCompile(`Detecting compiler hash`)
	vcpkgFailedRegex           = regexp.MustCompile(`Running vcpkg install - failed`)
	vcpkgManifestLogRegex      = regexp.MustCompile(`(?P<manifest_log>\S*vcpkg-manifest-install\.log)`)
)

func getCmakeCommand(ctx context.Context, prKey PresetInfoKey) (*exec.Cmd, error) {
	var cmakeArgs []string
	var cmakeCmd string

	cmakeCmd = "cmake"
	switch prKey.Type {
	case Configure:
	case Build:
		cmakeArgs = append(cmakeArgs, "--build")
	case Workflow:
		cmakeArgs = append(cmakeArgs, "--workflow")
	case Test:
		cmakeCmd = "ctest"
	case Package:
		cmakeCmd = "cpack"
	default:
		return nil, errors.New("got unexpected Preset_t type")
	}
	cmakeArgs = append(cmakeArgs, "--preset")
	cmakeArgs = append(cmakeArgs, prKey.Name)

	return exec.CommandContext(ctx, cmakeCmd, cmakeArgs...), nil
}

func startCmakeTicker(parentCtx context.Context, eventsCh chan<- CmexlEvent, prKey PresetInfoKey, freqHz float32) func() {
	if freqHz <= 0 {
		freqHz = defaultTickerFreqHz
	}
	period := time.Second / time.Duration(freqHz)

	ctx, cancel := context.WithCancel(parentCtx)
	ticker := time.NewTicker(period)
	start := time.Now()

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				elapsedTime := t.Sub(start).Seconds()
				tickerEvent := NewTimerUpdateEvent(prKey, elapsedTime)
				TrySend(eventsCh, tickerEvent)
			}
		}
	}()

	return func() {
		cancel()
	}
}

func handleLine(line string, eventsCh chan<- CmexlEvent, prKey PresetInfoKey, prState map[PresetInfoKey]PresetEventState) {
	sendLogUpdate := func(logLine string) {
		cmakeLogEvent := NewLogLineEvent(prKey, logLine)
		TrySend(eventsCh, cmakeLogEvent)
	}

	if match := cmexlRegex.FindStringSubmatch(line); match != nil {
		logIdx := cmexlRegex.SubexpIndex("log")
		out := match[logIdx]

		switch out {
		case "starting vcpkg":
			vr := prState[prKey]
			vr.VcpkgRunning = true
			prState[prKey] = vr
		case "vcpkg finished":
			vr := prState[prKey]
			vr.VcpkgRunning = false
			prState[prKey] = vr
		default:
		}

		sendLogUpdate(out)
	}

	if match := vcpkgLockRegex.FindString(line); len(match) != 0 && prState[prKey].VcpkgRunning {
		sendLogUpdate("waiting for vcpkg lock")
	}

	if match := vcpkgCompilerHashRegex.FindString(line); len(match) != 0 && prState[prKey].VcpkgRunning {
		sendLogUpdate("checking for change in build environment")
	}

	if match := vcpkgFailedRegex.FindString(line); len(match) != 0 && prState[prKey].VcpkgRunning {
		sendLogUpdate("vcpkg failed")
	}

	if match := vcpkgAlreadyInstalledRegex.FindString(line); len(match) != 0 && prState[prKey].VcpkgRunning {
		vr := prState[prKey]
		vr.VcpkgReadingInstalled = true
		vr.VcpkgReadingNeeded = false
		prState[prKey] = vr
	}

	if match := vcpkgNeedInstalledRegex.FindString(line); len(match) != 0 && prState[prKey].VcpkgRunning {
		vr := prState[prKey]
		vr.VcpkgReadingInstalled = false
		vr.VcpkgReadingNeeded = true
		prState[prKey] = vr
	}

	if match := vcpkgPkgDetailsRegex.FindStringSubmatch(line); match != nil && prState[prKey].VcpkgRunning {
		pkgIdx := vcpkgPkgDetailsRegex.SubexpIndex("package")
		versionIdx := vcpkgPkgDetailsRegex.SubexpIndex("version")
		patchIdx := vcpkgPkgDetailsRegex.SubexpIndex("patch")

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

		if prState[prKey].VcpkgReadingInstalled {
			vr := prState[prKey]
			vr.VcpkgAlreadyInstalledCount += 1
			prState[prKey] = vr
		} else if prState[prKey].VcpkgReadingNeeded {
			vr := prState[prKey]
			vr.VcpkgNeedInstalledCount += 1
			prState[prKey] = vr
		}
		sendLogUpdate(out)
	}
}

func startCMakeCommand(parentCtx context.Context, eventsCh chan<- CmexlEvent, prKey PresetInfoKey, wg *sync.WaitGroup, prState map[PresetInfoKey]PresetEventState) (PresetInfoKey, error) {
	cmakeCmd, err := getCmakeCommand(parentCtx, prKey)
	if err != nil {
		return prKey, err
	}

	stdout, err := cmakeCmd.StdoutPipe()
	if err != nil {
		return prKey, fmt.Errorf("stdoutpipe: %w", err)
	}

	if err := cmakeCmd.Start(); err != nil {
		return prKey, fmt.Errorf("failed to start cmake: %w", err)
	}

	filename := fmt.Sprintf(".cmexl/%s.log", prKey.Name)
	file, err := os.Create(filename)
	if err != nil {
		return prKey, err
	}
	buildLogger := bufio.NewWriterSize(file, 0.5*1024)

	stopTicker := startCmakeTicker(parentCtx, eventsCh, prKey, tickerFreqHz)

	// Handles shutdown mechanics of the CMake Cmd
	go func() {
		defer stopTicker()
		defer wg.Done()
		defer func() {
			buildLogger.Flush()
			stdout.Close()
			file.Close()
		}()
		err := cmakeCmd.Wait()
		select {
		case <-parentCtx.Done():
			_ = cmakeCmd.Process.Kill()
			TrySend(eventsCh, NewExecExitEvent(prKey, parentCtx.Err(), err))
		default:
			if err != nil {
				TrySend(eventsCh, NewExecExitEvent(prKey, parentCtx.Err(), err))
			} else {
				TrySend(eventsCh, NewExecExitEvent(prKey, parentCtx.Err(), err))
			}
		}
	}()

	go func() {
		sc := bufio.NewScanner(stdout)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1<<20) // ~1 MB

		for sc.Scan() {
			line := sc.Text()
			if buildLogger.Available() < len(line) {
				err := buildLogger.Flush()
				if err != nil {
					TrySend(eventsCh, NewExecErrEvent(prKey, err))
				}
			}
			bytesWritten, err := buildLogger.WriteString(line)
			if bytesWritten < len(line) {
				TrySend(eventsCh, NewExecErrEvent(prKey, err))
			}

			err = buildLogger.WriteByte('\n')
			if err != nil {
				TrySend(eventsCh, NewExecErrEvent(prKey, err))
			}

			select {
			case <-parentCtx.Done():
				return
			default:
				handleLine(line, eventsCh, prKey, prState)
			}
		}
		if err := sc.Err(); err != nil {
			TrySend(eventsCh, NewExecErrEvent(prKey, err))
		}
	}()

	return prKey, nil
}

var logDoubleBuffer [2]map[PresetInfoKey]DisplayState
var activeIndex atomic.Uint32 // active means the read-allowed buffer

func init() {
	logDoubleBuffer[0] = make(map[PresetInfoKey]DisplayState)
	logDoubleBuffer[1] = make(map[PresetInfoKey]DisplayState)
}

func updateState(ev CmexlEvent, errReport map[PresetInfoKey][]error) {
	cur := activeIndex.Load()
	next := (cur + 1) % 2

	for key, val := range logDoubleBuffer[cur] {
		logDoubleBuffer[next][key] = val
	}

	switch ev.Type {
	case TimerUpdate:
		s := logDoubleBuffer[cur][ev.Key]
		s.ElapsedTime = ev.Payload.(TimerUpdatePayload).ElapsedTime
		logDoubleBuffer[next][ev.Key] = s
	case LogLineUpdate:
		s := logDoubleBuffer[cur][ev.Key]
		s.Log = ev.Payload.(LogLinePayload).Log
		logDoubleBuffer[next][ev.Key] = s
	case ExecErr:
		err := ev.Payload.(ExecErrPayload).Err
		s := logDoubleBuffer[cur][ev.Key]
		s.Log = fmt.Sprintf("error during execution: %w", err)
		logDoubleBuffer[next][ev.Key] = s
		errReport[ev.Key] = append(errReport[ev.Key], err)
	case ExecExit:
		exitStatus := ev.Payload.(ExecExitPayload)
		s := logDoubleBuffer[cur][ev.Key]
		if exitStatus.Err != nil || exitStatus.ExitCode != nil {
			err := fmt.Errorf("error after execution: %w, %w", exitStatus.Err, exitStatus.ExitCode)
			s.Log = err.Error()
			errReport[ev.Key] = append(errReport[ev.Key], err)
		} else {
			s.Log = "no errors occurred after execution"
		}
		logDoubleBuffer[next][ev.Key] = s
	default:
	}

	activeIndex.Store(next)
}

func getActiveBuffer() map[PresetInfoKey]DisplayState {
	idx := activeIndex.Load()
	return logDoubleBuffer[idx]
}

func drawUI(uiWg *sync.WaitGroup, uiDone <-chan struct{}, keys []PresetInfoKey) {
	fmt.Print(cursorHide)
	defer fmt.Print(cursorShow)
	defer uiWg.Done()

	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	render := func() {
		snapshot := getActiveBuffer()

		var frameData strings.Builder
		frameData.WriteString("Preset Status\n")
		frameData.WriteString("==============\n")

		for i, k := range keys {
			v := snapshot[k]
			fmt.Fprintf(&frameData, "%d. %s (%v, %.2fs) : %s\n", i+1, k.Name, k.Type, v.ElapsedTime, v.Log)
		}

		fmt.Print(cursorHome, clearFromCursor)
		fmt.Print(frameData.String())
	}

	for {
		select {
		case <-uiDone:
			render()
			return
		case <-ticker.C:
			render()
		}
	}
}

func ScheduleCmakePresets(prType Preset_t, prKeys []PresetInfoKey, prMap PresetMap_t) error {
	numPrs := len(prKeys)
	if numPrs < 1 {
		return errors.New("no presets to execute")
	}
	sort.Slice(prKeys, func(i, j int) bool {
		ni, nj := prKeys[i].Name, prKeys[j].Name
		if ni == nj {
			return prKeys[i].Type < prKeys[j].Type
		}
		return ni < nj
	})

	eventChSize := numPrs * (int(math.Ceil(float64(tickerFreqHz))) + 1) * eventChannelSizeScale
	eventsCh := make(chan CmexlEvent, eventChSize)

	var cmakeWg sync.WaitGroup
	cmakeWg.Add(numPrs)

	var uiWg sync.WaitGroup
	uiWg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errReport := make(map[PresetInfoKey][]error)
	presetExecStates := make(map[PresetInfoKey]PresetEventState)

	err := CreateCmexlStore()
	if err != nil {
		return err
	}

	// Beyond this point, we should not abruptly halt this parent process since we want any working preset to at least finish
	for _, key := range prKeys {
		prKey, initErr := startCMakeCommand(ctx, eventsCh, key, &cmakeWg, presetExecStates)
		if initErr != nil {
			errReport[key] = append(errReport[key], fmt.Errorf("{%s, %s}: %w", prKey.Name, prKey.Type.String(), initErr))
		}
	}

	go func() {
		for ev := range eventsCh {
			updateState(ev, errReport)
		}
	}()

	uiDone := make(chan struct{})
	go drawUI(&uiWg, uiDone, prKeys)

	// event draining
	cmakeWg.Wait()
	for len(eventsCh) > 0 {
		ev := <-eventsCh
		updateState(ev, errReport)
	}
	close(eventsCh)
	close(uiDone)
	uiWg.Wait()

	// TODO: Remove this for projects that dont need vcpkg. Will have to read from viper for this
	fmt.Println("Packages")
	fmt.Println("==============")
	for key, val := range presetExecStates {
		fmt.Printf("{%s, %s}: Already installed: %d, Needed installation/removal: %d\n", key.Name, key.Type.String(), val.VcpkgAlreadyInstalledCount, val.VcpkgNeedInstalledCount)
	}

	fmt.Println("Error Report")
	fmt.Println("==============")
	if len(errReport) <= 0 {
		fmt.Print("none")
	} else {
		for prKey, errList := range errReport {
			for _, err := range errList {
				fmt.Printf("{%s, %v}: %s\n", prKey.Name, prKey.Type, err)
			}
		}
	}
	return nil
}
