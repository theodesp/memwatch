package memwatch

import (
	"os"
	"runtime"
	"sync"
	"time"
)

// A MemoryUnit represents a memory size in bytes
type MemoryUnit int64

const (
	Byte     MemoryUnit = 1
	KiloByte            = 1024 * Byte
	MegaByte            = 1024 * KiloByte
	GigaByte            = 1024 * MegaByte
	TeraByte            = 1024 * GigaByte
)

// EventType represents a Memory Watch Event Type
type EventType struct{}

// Throws on Memory overload
var Boom EventType

var defaultWatchConfig = WatchConfig{
	WarningLimit:  512 * MegaByte,
	CriticalLimit: 768 * MegaByte,
	Cycle:         10,
	Interval:      5 * time.Second,
	ExitCode:      101,
	ExitTime:      10 * time.Second,
}

// WatchConfig is used to configure the memory watcher
type WatchConfig struct {
	// The amount of memory units that are required to trigger a Warning
	WarningLimit MemoryUnit

	// The amount of memory units that are required to trigger a Critical
	CriticalLimit MemoryUnit

	// Consecutive warnings that need to continuously meet, or a Critical is Triggered
	Cycle int

	// Memory check interval
	Interval time.Duration

	// Exit after ExitTime
	ExitTime time.Duration

	// Exit with ExitCode
	ExitCode int
}

func mergeWithDefaults(base *WatchConfig) *WatchConfig {
	if base.WarningLimit == MemoryUnit(0) {
		base.WarningLimit = defaultWatchConfig.WarningLimit
	}
	if base.CriticalLimit == MemoryUnit(0) {
		base.CriticalLimit = defaultWatchConfig.CriticalLimit
	}
	if base.Cycle == 0 {
		base.Cycle = defaultWatchConfig.Cycle
	}
	if base.Interval == time.Duration(0) {
		base.Interval = defaultWatchConfig.Interval
	}
	if base.ExitTime == time.Duration(0) {
		base.ExitTime = defaultWatchConfig.ExitTime
	}
	if base.ExitCode == 0 {
		base.ExitCode = defaultWatchConfig.ExitCode
	}
	return base
}

// MemoryWatcher watches the memory consumption for you
type MemoryWatcher struct {
	cfg    WatchConfig
	count  int
	ticker *time.Ticker
	events chan EventType
	once   sync.Once
	stopped chan struct{}
}

// Creates a new MemoryWatcher
func New(opt *WatchConfig) *MemoryWatcher {
	mw := MemoryWatcher{
		count:  0,
		events: make(chan EventType, 1),
		stopped: make(chan struct{}),
	}
	if opt == nil {
		mw.cfg = defaultWatchConfig
	} else {
		mw.cfg = *mergeWithDefaults(opt)
	}

	return &mw
}

// GetConfig return current configuration
func (mw *MemoryWatcher) GetConfig() WatchConfig {
	return mw.cfg
}

// CalculateTotalMemory returns the total memory usage
func CalculateTotalMemory(stats runtime.MemStats) MemoryUnit {
	// Sys related stats may be released to the OS so the runtime
	// memory usage would not be close with the one observed via
	// ps or activity monitor
	return MemoryUnit(stats.HeapInuse +
		stats.StackInuse +
		stats.MSpanInuse +
		stats.MCacheInuse +
		stats.BuckHashSys)
}

// ReachCritical returns whether total memory reached critical threshold
func (mw *MemoryWatcher) ReachCritical(total MemoryUnit) bool {
	return total > mw.cfg.CriticalLimit
}

// ReachWarning returns whether total memory reached warning threshold
func (mw *MemoryWatcher) ReachWarning(total MemoryUnit) bool {
	return total > mw.cfg.WarningLimit
}

// Starts the monitoring
func (mw *MemoryWatcher) Start() <-chan EventType {
	go func() {
		mw.ticker = time.NewTicker(mw.cfg.Interval)
		for {
			select {
			case <-mw.ticker.C:
				mw.tick()
			case <-mw.stopped:
				return
			}
		}
	}()
	return mw.events
}

// Stops the monitoring without recycling the watcher.
func (mw *MemoryWatcher) Stop() {
	mw.count = 0
	mw.ticker.Stop()
}

// Call in every circle, check the memory usage
func (mw *MemoryWatcher) tick() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	total := CalculateTotalMemory(rtm)
	if mw.ReachCritical(total) {
		mw.once.Do(mw.trigger)
	} else if mw.ReachWarning(total) {
		mw.count = 0
	}
	mw.count++
	if mw.count >= mw.cfg.Cycle {
		mw.once.Do(mw.trigger)
	}
}

// Trigger boom and exit after the ExitTime duration
func (mw *MemoryWatcher) trigger() {
	mw.boom()
	mw.count = 0
	mw.Stop()
	mw.stopped <- struct{}{}

	<-time.After(mw.cfg.ExitTime)
	os.Exit(mw.cfg.ExitCode)
}

func (mw *MemoryWatcher) boom() {
	mw.events <- Boom
	close(mw.events)
}
