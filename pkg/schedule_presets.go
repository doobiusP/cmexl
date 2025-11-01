package cmexl_utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os/exec"
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
		return "-UNKNOWN-"
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
		return
	}
}

var defaultTickerFreq = float32(5.0)
var tickerFreq = float32(10.0)
var defaultEventChScale = 10

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
		freqHz = defaultTickerFreq
	}
	period := time.Second / time.Duration(freqHz)

	ctx, cancel := context.WithCancel(parentCtx)
	ticker := time.NewTicker(period)
	start := time.Now()

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done(): // if parent cancels, so will this
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
	cmakeErrLog := fmt.Sprintf("{%s, %s} error", prKey.Name, prKey.Type.String())
	cmakeErrEventData := CmexlEventData{Log: cmakeErrLog, ElapsedTime: -1}
	cmakeErrEvent := CmexlEvent{Key: prKey, Type: ExecKilled, Data: cmakeErrEventData, Result: err}
	TrySend(eventsCh, cmakeErrEvent)
}

func startCMakeCommand(parentCtx context.Context, eventsCh chan<- CmexlEvent, prKey PresetInfoKey, wg *sync.WaitGroup) (PresetInfoKey, error) {
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

	stopTicker := startCmakeTicker(parentCtx, eventsCh, prKey, tickerFreq)

	// Handles shutdown mechanics of the CMake Cmd
	go func() {
		defer stopTicker()
		defer wg.Done()
		err := cmakeCmd.Wait()
		select {
		case <-parentCtx.Done():
			stdout.Close()
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
			select {
			case <-parentCtx.Done():
				return
			default:
				cmakeLogEventData := CmexlEventData{Log: line, ElapsedTime: -1}
				cmakeLogEvent := CmexlEvent{Key: prKey, Type: LogLineUpdate, Data: cmakeLogEventData, Result: nil}
				TrySend(eventsCh, cmakeLogEvent)
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

func updateState(ev CmexlEvent, errReport map[PresetInfoKey]error) {
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
		errReport[ev.Key] = ev.Result
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

func ScheduleCmakePresets(prType Preset_t, prKeys []PresetInfoKey) error {
	prMap, prErr := GetCmakePresets(prType)
	if prErr != nil {
		return prErr
	}

	numPrs := len(prKeys)
	if numPrs < 1 {
		return errors.New("no presets to execute")
	}

	eventChSize := numPrs * (int(math.Ceil(float64(tickerFreq))) + 1) * defaultEventChScale
	eventsCh := make(chan CmexlEvent, eventChSize)
	var wg sync.WaitGroup
	wg.Add(numPrs)

	var uiWg sync.WaitGroup
	uiWg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errReport := make(map[PresetInfoKey]error)

	for _, key := range prKeys {
		// TODO: For now force all presets to fail if any one fails. Later make it redundant
		if _, ok := prMap[key]; !ok {
			errReport[key] = errors.New("can't find preset for this preset-type")
		}
		prKey, err := startCMakeCommand(ctx, eventsCh, key, &wg)
		if err != nil {
			errReport[key] = fmt.Errorf("{%s, %s}: %w", prKey.Name, prKey.Type.String(), err)
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

	fmt.Println("Error Report")
	fmt.Println("==============")
	for _, k := range keys {
		if err := errReport[k]; err != nil {
			fmt.Printf("%s (%v): %v\n", k.Name, k.Type, err)
		}
	}
	return nil
}
