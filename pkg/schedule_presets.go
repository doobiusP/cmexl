package cmexl_utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
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
)

type ScheduleFlags struct {
	SaveEvents *bool
	Refresh    *bool
	Serial     *bool
}

func getCmakeCommand(ctx context.Context, prKey PresetInfoKey, flags ScheduleFlags) (*exec.Cmd, error) {
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
	if *flags.Refresh {
		cmakeArgs = append(cmakeArgs, "--fresh")
	}

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

	// We provide the ability for the state of the cmake command to dictate when to stop timing
	return func() {
		cancel()
	}
}

func startCMakeCommand(parentCtx context.Context, eventsCh chan<- CmexlEvent, prKey PresetInfoKey, wg *sync.WaitGroup, cmexlDataMap map[PresetInfoKey]CmexlPresetData, flags ScheduleFlags) error {
	cmakeCmd, err := getCmakeCommand(parentCtx, prKey, flags)
	if err != nil {
		return err
	}

	cmdStateMachinePtr := new(CmexlStateMachine)
	cmdStateMachinePtr.PrKey = prKey

	v := cmexlDataMap[prKey]
	v.ExecState = cmdStateMachinePtr
	cmexlDataMap[prKey] = v

	file := cmexlDataMap[prKey].StdoutLog
	stdout, err := cmakeCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdoutpipe: %w", err)
	}

	cmakeCmd.Stderr = cmexlDataMap[prKey].StderrLog

	stopTicker := startCmakeTicker(parentCtx, eventsCh, prKey, tickerFreqHz)
	if err := cmakeCmd.Start(); err != nil {
		return fmt.Errorf("failed to start cmake: %w", err)
	}
	buildLogger := bufio.NewWriterSize(file, 0.5*1024)

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
			cmakeCmd.Process.Signal(os.Interrupt)
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
		sc.Buffer(buf, 2<<20) // ~2 MB
		cmexlStateFn := CmexlDefaultStateFn

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
				cmexlStateFn = cmexlStateFn(line, cmdStateMachinePtr, eventsCh)
			}
		}
		if err := sc.Err(); err != nil {
			TrySend(eventsCh, NewExecErrEvent(prKey, err))
		}
	}()

	return nil
}

var logDoubleBuffer [2]map[PresetInfoKey]DisplayState
var activeIndex atomic.Uint32 // active means the read-allowed buffer

func init() {
	logDoubleBuffer[0] = make(map[PresetInfoKey]DisplayState)
	logDoubleBuffer[1] = make(map[PresetInfoKey]DisplayState)
}

func updateState(ev CmexlEvent, cmexlDataMap map[PresetInfoKey]CmexlPresetData, flags ScheduleFlags) {
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
		s.Log = fmt.Sprintf("error during execution: %s", err)
		logDoubleBuffer[next][ev.Key] = s
		v := cmexlDataMap[ev.Key]
		v.Errors = append(v.Errors, err)
		cmexlDataMap[ev.Key] = v
	case ExecExit:
		exitStatus := ev.Payload.(ExecExitPayload)
		s := logDoubleBuffer[cur][ev.Key]
		if exitStatus.Err != nil || exitStatus.ExitCode != nil {
			err := fmt.Errorf("error after execution: check error report later, %s", exitStatus.ExitCode)
			s.Log = err.Error()
			v := cmexlDataMap[ev.Key]
			v.Errors = append(v.Errors, err)
			cmexlDataMap[ev.Key] = v
		} else {
			s.Log = "no errors occurred after execution"
		}
		logDoubleBuffer[next][ev.Key] = s
	default:
	}

	if *flags.SaveEvents {
		switch ev.Type {
		case LogLineUpdate, ExecErr, ExecExit:
			file := cmexlDataMap[ev.Key].EventsLog
			evLog := fmt.Sprintf("%fs : %s\n", logDoubleBuffer[next][ev.Key].ElapsedTime, ev)
			if _, err := file.WriteString(evLog); err != nil {
				panicMsg := fmt.Sprintf("failed to write event log (%s) for {%s,%s}", evLog, ev.Key.Name, ev.Key.Type)
				panic(panicMsg)
			}
		}
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
		frameData.WriteString("Preset Status\r\n")
		frameData.WriteString("==============\r\n")

		for i, k := range keys {
			v := snapshot[k]
			fmt.Fprintf(&frameData, "%d. %s (%v, %.2fs) : %s\r\n", i+1, k.Name, k.Type, v.ElapsedTime, strings.TrimRight(v.Log, "\r\n"))
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

func printKnownErrors(prKey PresetInfoKey) {
	errLog := fmt.Sprintf(".cmexl/stderr/%s.log", prKey.Name)
	errFile, err := os.Open(errLog)
	if err != nil {
		panic("can't open error log")
	}
	defer errFile.Close()

	scanner := bufio.NewScanner(errFile)
	var vcpkgManifestLog string

	fmt.Printf("(%s,%s):\n", prKey.Name, prKey.Type)
	fmt.Printf("Checking %s\n", errLog)
	for scanner.Scan() {
		line := scanner.Text()
		if match := VcpkgManifestLogRegex.FindStringSubmatch(line); match != nil {
			vcpkgManifestLog = match[VcpkgManifestLogRegex.SubexpIndex("manifest_log")]
		}
		if StderrFileRegex.MatchString(line) {
			fmt.Println(line)
		}
	}

	if vcpkgManifestLog != "" {
		fmt.Printf("Checking vcpkg-manifest-log: %s\n", vcpkgManifestLog)
		manifestFile, err := os.Open(vcpkgManifestLog)
		if err != nil {
			panic("can't open vcpkg manifest log")
		}
		scanner := bufio.NewScanner(manifestFile)
		for scanner.Scan() {
			line := scanner.Text()
			if StderrFileRegex.MatchString(line) {
				fmt.Println(line)
			}
		}
	}
}

// func ScheduleSerialPresets(prType Preset_t, prKeys []PresetInfoKey, prMap PresetMap_t, flags ScheduleFlags) error {
// 	numPrs := len(prKeys)
// 	if numPrs < 1 {
// 		return errors.New("no presets to execute")
// 	}
// 	sort.Slice(prKeys, func(i, j int) bool {
// 		ni, nj := prKeys[i].Name, prKeys[j].Name
// 		if ni == nj {
// 			return prKeys[i].Type < prKeys[j].Type
// 		}
// 		return ni < nj
// 	})

// 	err := CreateCmexlStore(flags)
// 	if err != nil {
// 		return err
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	cmexlDataMap := make(map[PresetInfoKey]CmexlPresetData, numPrs)
// 	var cmexlPresetLock sync.Mutex

// 	for _, key := range prKeys {
// 		v := cmexlDataMap[key]
// 		errFilename := fmt.Sprintf(".cmexl/stderr/%s.log", key.Name)
// 		errFile, err := os.Create(errFilename)
// 		if err != nil {
// 			v.Errors = append(v.Errors,
// 				fmt.Errorf("{%s, %s}: %w", key.Name, key.Type.String(), err))
// 		}
// 		v.StderrLog = errFile
// 		defer errFile.Close()

// 		outFilename := fmt.Sprintf(".cmexl/%s.log", key.Name)
// 		outFile, err := os.Create(outFilename)
// 		if err != nil {
// 			v.Errors = append(v.Errors,
// 				fmt.Errorf("{%s, %s}: %w", key.Name, key.Type.String(), err))
// 		}
// 		v.StdoutLog = outFile
// 		defer outFile.Close()

// 		if *flags.SaveEvents {
// 			eventFilename := fmt.Sprintf(".cmexl/events/%s.log", key.Name)
// 			eventFile, err := os.Create(eventFilename)
// 			if err != nil {
// 				v.Errors = append(v.Errors,
// 					fmt.Errorf("{%s, %s}: %w", key.Name, key.Type.String(), err))
// 			}
// 			v.EventsLog = eventFile
// 			defer eventFile.Close()
// 		}
// 		cmexlDataMap[key] = v
// 		err := startSerialCMakeCommand(ctx, eventsCh, key, &cmexlPresetLock, cmexlDataMap, flags)
// 		if err != nil {
// 			v.Errors = append(v.Errors,
// 				fmt.Errorf("{%s, %s}: %w", key.Name, key.Type.String(), err))
// 		}
// 	}
// }

func ScheduleCmakePresets(prType Preset_t, prKeys []PresetInfoKey, prMap PresetMap_t, flags ScheduleFlags) error {
	numPrs := len(prKeys)
	if numPrs < 1 {
		return errors.New("no presets to execute")
	}
	// TODO: This is the point where we can execute an early vcpkg install as an optimisation
	sort.Slice(prKeys, func(i, j int) bool {
		ni, nj := prKeys[i].Name, prKeys[j].Name
		if ni == nj {
			return prKeys[i].Type < prKeys[j].Type
		}
		return ni < nj
	})

	err := CreateCmexlStore(flags)
	if err != nil {
		return err
	}
	eventChSize := numPrs * (int(math.Ceil(float64(tickerFreqHz))) + 1) * eventChannelSizeScale
	eventsCh := make(chan CmexlEvent, eventChSize)

	var cmakeWg sync.WaitGroup

	if *flags.Serial {
		cmakeWg.Add(1)
	} else {
		cmakeWg.Add(numPrs)
	}

	var uiWg sync.WaitGroup
	uiWg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmexlDataMap := make(map[PresetInfoKey]CmexlPresetData, numPrs)

	go func() {
		for ev := range eventsCh {
			updateState(ev, cmexlDataMap, flags)
		}
	}()

	uiDone := make(chan struct{})
	go drawUI(&uiWg, uiDone, prKeys)

	for _, key := range prKeys {
		v := cmexlDataMap[key]
		errFilename := fmt.Sprintf(".cmexl/stderr/%s.log", key.Name)
		errFile, err := os.Create(errFilename)
		if err != nil {
			v.Errors = append(v.Errors,
				fmt.Errorf("{%s, %s}: %w", key.Name, key.Type.String(), err))
		}
		v.StderrLog = errFile
		defer errFile.Close()

		outFilename := fmt.Sprintf(".cmexl/%s.log", key.Name)
		outFile, err := os.Create(outFilename)
		if err != nil {
			v.Errors = append(v.Errors,
				fmt.Errorf("{%s, %s}: %w", key.Name, key.Type.String(), err))
		}
		v.StdoutLog = outFile
		defer outFile.Close()

		if *flags.SaveEvents {
			eventFilename := fmt.Sprintf(".cmexl/events/%s.log", key.Name)
			eventFile, err := os.Create(eventFilename)
			if err != nil {
				v.Errors = append(v.Errors,
					fmt.Errorf("{%s, %s}: %w", key.Name, key.Type.String(), err))
			}
			v.EventsLog = eventFile
			defer eventFile.Close()
		}
		cmexlDataMap[key] = v

		if *flags.Serial {
			TrySend(eventsCh, NewLogLineEvent(key, "waiting for cmexl lock"))
		}
	}

	// Beyond this point, we should not abruptly halt this parent process since
	// we want any working preset to at least finish in case of parallel build
	for _, key := range prKeys {
		initErr := startCMakeCommand(ctx, eventsCh, key, &cmakeWg, cmexlDataMap, flags)
		if *flags.Serial {
			cmakeWg.Wait()
			cmakeWg.Add(1)
		}
		if initErr != nil {
			v := cmexlDataMap[key]
			v.Errors = append(v.Errors,
				fmt.Errorf("{%s, %s}: %w", key.Name, key.Type.String(), err))
			cmexlDataMap[key] = v
		}
	}

	// event draining
	if !(*flags.Serial) {
		cmakeWg.Wait()
	}

	for len(eventsCh) > 0 {
		ev := <-eventsCh
		updateState(ev, cmexlDataMap, flags)
	}
	close(eventsCh)
	close(uiDone)
	uiWg.Wait()

	// TODO: Remove this for projects that dont need vcpkg. Will have to read from viper for this
	fmt.Println("Packages")
	fmt.Println("==============")
	for key, val := range cmexlDataMap {
		fmt.Printf("(%s, %s): Already installed: %d, Needed removal: %d, Needed installation: %d\n",
			key.Name, key.Type.String(),
			val.ExecState.VcpkgAlreadyInstalledCount,
			val.ExecState.VcpkgNeedRemovedCount,
			val.ExecState.VcpkgNeedInstalledCount)
	}

	fmt.Println("Error Report")
	fmt.Println("==============")

	for prKey, data := range cmexlDataMap {
		if len(data.Errors) > 0 {
			printKnownErrors(prKey)
		}
	}
	return nil
}
