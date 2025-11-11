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

type CmexlEvent_t int

const (
	TimerUpdate CmexlEvent_t = iota
	LogLineUpdate
	ExecFinished
	ExecKilled
)

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

func (e CmexlEvent) String() string {
	return fmt.Sprintf("[%s:%s](%s) Elapsed time: %v, Log: %s", e.Key.Name, e.Key.Type.String(), e.Type.String(), e.Data.ElapsedTime, e.Data.Log)
}

func TrySend(events chan<- CmexlEvent, event CmexlEvent) {
	select {
	case events <- event:
	default:
		fmt.Printf("overflow")
		return
	}
}

var (
	defaultTickerFreqHz   = float32(5.0)
	tickerFreqHz          = float32(12.0)
	eventChannelSizeScale = 20

	cmexlRegex                 = regexp.MustCompile(`\[CMEXL\]\s*(?P<log>.*)$`)
	vcpkgPkgDetailsRegex       = regexp.MustCompile(`(?P<package>[\w\-]+(?:\[[^\]]*\])?):(?P<triplet>[\w\-]+)(?:@(?P<version>[\w\.\-\+]+)(?:#(?P<patch>\d+))?)?`)
	vcpkgLockRegex             = regexp.MustCompile(`waiting to take filesystem lock`)
	vcpkgAlreadyInstalledRegex = regexp.MustCompile(`The following packages are already installed`)
	vcpkgNeedInstalledRegex    = regexp.MustCompile(`The following packages will be (?:built and installed|rebuilt|removed)`)
	vcpkgCompilerHashRegex     = regexp.MustCompile(`Detecting compiler hash`)
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
				tickerEventData := CmexlEventData{Log: "", ElapsedTime: elapsedTime}
				tickerEvent := CmexlEvent{Key: prKey, Type: TimerUpdate, Data: tickerEventData, Result: nil}
				TrySend(eventsCh, tickerEvent)
			}
		}
	}()

	return func() {
		cancel()
	}
}

func reportCMakeErr(err error, eventsCh chan<- CmexlEvent, prKey PresetInfoKey) {
	cmakeErrEventData := CmexlEventData{Log: "error", ElapsedTime: -1}
	cmakeErrEvent := CmexlEvent{Key: prKey, Type: ExecKilled, Data: cmakeErrEventData, Result: err}
	TrySend(eventsCh, cmakeErrEvent)
}

/* ---------- your existing handler ---------- */
func handleLine(line string, eventsCh chan<- CmexlEvent, prKey PresetInfoKey, prState map[PresetInfoKey]PresetState) {
	sendLogUpdate := func(logLine string) {
		cmakeLogEventData := CmexlEventData{Log: logLine, ElapsedTime: -1}
		cmakeLogEvent := CmexlEvent{Key: prKey, Type: LogLineUpdate, Data: cmakeLogEventData, Result: nil}
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
		out := fmt.Sprintf("Now processing %s @ %s with vcpkg patch %s",
			match[pkgIdx], match[versionIdx], patchStr)

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

func startCMakeCommand(parentCtx context.Context, eventsCh chan<- CmexlEvent, prKey PresetInfoKey, wg *sync.WaitGroup, prState map[PresetInfoKey]PresetState) (PresetInfoKey, error) {
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
			file.Close()
			stdout.Close()
		}()
		err := cmakeCmd.Wait()
		select {
		case <-parentCtx.Done():
			if cmakeCmd.Process != nil {
				_ = cmakeCmd.Process.Kill()
				reportCMakeErr(parentCtx.Err(), eventsCh, prKey)
			}
		default:
			if err != nil {
				reportCMakeErr(err, eventsCh, prKey)
			} else {
				cmakeFinEventData := CmexlEventData{Log: "finished", ElapsedTime: -1}
				cmakeFinEvent := CmexlEvent{Key: prKey, Type: ExecFinished, Data: cmakeFinEventData, Result: nil}
				TrySend(eventsCh, cmakeFinEvent)
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
					reportCMakeErr(err, eventsCh, prKey)
				}
			}
			bytesWritten, err := buildLogger.WriteString(line)
			if bytesWritten < len(line) {
				reportCMakeErr(err, eventsCh, prKey)
			}

			err = buildLogger.WriteByte('\n')
			if err != nil {
				reportCMakeErr(err, eventsCh, prKey)
			}

			select {
			case <-parentCtx.Done():
				return
			default:
				handleLine(line, eventsCh, prKey, prState)
			}
		}
		if err := sc.Err(); err != nil {
			reportCMakeErr(err, eventsCh, prKey)
		}
	}()

	return prKey, nil
}

type presetState struct {
	ElapsedTime float32
	Log         string
}

var logDoubleBuffer [2]map[PresetInfoKey]presetState
var activeIndex atomic.Uint32 // active means the read-allowed buffer

func init() {
	logDoubleBuffer[0] = make(map[PresetInfoKey]presetState)
	logDoubleBuffer[1] = make(map[PresetInfoKey]presetState)
}

func updateState(ev CmexlEvent, errReport map[PresetInfoKey][]error) {
	cur := activeIndex.Load()
	next := (cur + 1) % 2

	for key, val := range logDoubleBuffer[cur] {
		logDoubleBuffer[next][key] = val
	}

	switch ev.Type {
	case ExecKilled:
		s := logDoubleBuffer[cur][ev.Key]
		s.Log = ev.Data.Log
		logDoubleBuffer[next][ev.Key] = s
		errReport[ev.Key] = append(errReport[ev.Key], ev.Result)
	case ExecFinished, LogLineUpdate:
		s := logDoubleBuffer[cur][ev.Key]
		s.Log = ev.Data.Log
		logDoubleBuffer[next][ev.Key] = s
	case TimerUpdate:
		s := logDoubleBuffer[cur][ev.Key]
		s.ElapsedTime = float32(ev.Data.ElapsedTime)
		logDoubleBuffer[next][ev.Key] = s
	default:
	}

	activeIndex.Store(next)
}

func getActiveBuffer() map[PresetInfoKey]presetState {
	idx := activeIndex.Load()
	return logDoubleBuffer[idx]
}

func drawUI(uiWg *sync.WaitGroup, uiDone <-chan struct{}) {
	defer uiWg.Done()
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	render := func() {
		snapshot := getActiveBuffer()

		var b strings.Builder
		b.WriteString("Preset Status\n")
		b.WriteString("==============\n")

		keys := make([]PresetInfoKey, 0, len(snapshot))
		for k := range snapshot {
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool {
			ni, nj := keys[i].Name, keys[j].Name
			if ni == nj {
				return keys[i].Type < keys[j].Type
			}
			return ni < nj
		})

		for i, k := range keys {
			v := snapshot[k]
			fmt.Fprintf(&b, "%d. %s (%v, %.2fs) : %s\n", i+1, k.Name, k.Type, v.ElapsedTime, v.Log)
		}

		fmt.Print("\033[H\033[0J")
		fmt.Print(b.String())
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

	eventChSize := numPrs * (int(math.Ceil(float64(tickerFreqHz))) + 1) * eventChannelSizeScale
	eventsCh := make(chan CmexlEvent, eventChSize)
	var wg sync.WaitGroup
	wg.Add(numPrs)

	var uiWg sync.WaitGroup
	uiWg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errReport := make(map[PresetInfoKey][]error)
	presetExecStates := make(map[PresetInfoKey]PresetState)

	err := CreateCmexlStore()
	if err != nil {
		return err
	}
	for _, key := range prKeys {
		if _, ok := prMap[key]; !ok {
			errReport[key] = append(errReport[key], errors.New("can't find preset for this preset-type"))
		}
		/*
			url=https://github.com/microsoft/vcpkg-tool/blob/4819c236bb6a9e51be6e46a226683a64b11fc409/include/vcpkg/base/message-data.inc.h
			DECLARE_MESSAGE(PackagesToInstall, (), "", "The following packages will be built and installed:")
			DECLARE_MESSAGE(PackagesToModify, (), "", "Additional packages (*) will be modified to complete this operation.")
			DECLARE_MESSAGE(PackagesToRebuild, (), "", "The following packages will be rebuilt:")
			DECLARE_MESSAGE(PackagesToRemove, (), "", "The following packages will be removed:")
			// TODO: Sort in ascending order of work
			// `(?P<package>[\w\-]+(?:\[[^\]]*\])?):(?P<triplet>[\w\-]+)@(?P<version>[\w\.\-\+]+)(?:#(?P<patch>\d+))?
		*/
		prKey, err := startCMakeCommand(ctx, eventsCh, key, &wg, presetExecStates)
		if err != nil {
			errReport[key] = append(errReport[key], fmt.Errorf("{%s, %s}: %w", prKey.Name, prKey.Type.String(), err))
		}
	}

	go func() {
		for ev := range eventsCh {
			updateState(ev, errReport)
		}
	}()

	uiDone := make(chan struct{})
	go drawUI(&uiWg, uiDone)

	wg.Wait()
	for len(eventsCh) > 0 {
		ev := <-eventsCh
		updateState(ev, errReport)
	}
	close(eventsCh)
	close(uiDone)
	uiWg.Wait()

	keys := make([]PresetInfoKey, 0, len(errReport))
	for k := range errReport {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Name == keys[j].Name {
			return keys[i].Type < keys[j].Type
		}
		return keys[i].Name < keys[j].Name
	})

	fmt.Println("Packages")
	fmt.Println("==============")
	for key, val := range presetExecStates {
		fmt.Printf("{%s, %s}: Already installed: %d, Needed installation/removal: %d\n", key.Name, key.Type.String(), val.VcpkgAlreadyInstalledCount, val.VcpkgNeedInstalledCount)
	}

	fmt.Println("Error Report")
	fmt.Println("==============")
	if len(keys) <= 0 {
		fmt.Print("None")
	} else {
		for _, k := range keys {
			if err := errReport[k]; err != nil {
				for each_err := range err {
					fmt.Printf("%s (%v): %v\n", k.Name, k.Type, each_err)
				}
			}
		}
	}
	return nil
}
