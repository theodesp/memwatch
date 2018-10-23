package memwatch

import (
	"time"
	"runtime"
	"os"
	"sync"
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
}

// Creates a new MemoryWatcher
func New(opt *WatchConfig) *MemoryWatcher {
	mw := MemoryWatcher{
		count:  0,
		events: make(chan EventType, 1),
	}
	if opt == nil {
		mw.cfg = defaultWatchConfig
	} else {
		mw.cfg = *mergeWithDefaults(opt)
	}

	return &mw
}

// Starts the monitoring
func (mw *MemoryWatcher) Start() (<-chan EventType) {
	go func() {
		mw.ticker = time.NewTicker(mw.cfg.Interval)
		for {
			select {
			case <-mw.ticker.C:
				mw.tick()
			}
		}
	}()
	return mw.events
}

func (mw *MemoryWatcher) stop() {
	mw.count = 0
	mw.ticker.Stop()
}

// Call in every circle, check the memory usage
func (mw *MemoryWatcher) tick() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	total := calculateTotalMemory(rtm)
	if MemoryUnit(total) > mw.cfg.CriticalLimit {
		mw.once.Do(mw.trigger)
	} else if MemoryUnit(total) < mw.cfg.WarningLimit {
		mw.count = 0
	}
	mw.count += 1
	if mw.count >= mw.cfg.Cycle {
		mw.once.Do(mw.trigger)
	}
}

func calculateTotalMemory(stats runtime.MemStats) uint64  {
	// Sys related stats may be released to the OS so the runtime
	// memory usage would not be close with the one observed via
	// ps or activity monitor
	return stats.HeapInuse +
		stats.StackInuse +
		stats.MSpanInuse +
		stats.MCacheInuse +
		stats.BuckHashSys
}

// Trigger boom and exit after the ExitTime duration
func (mw *MemoryWatcher) trigger() {
	mw.boom()
	mw.count = 0
	mw.stop()

	<-time.After(mw.cfg.ExitTime)
	os.Exit(mw.cfg.ExitCode)
}

func (mw *MemoryWatcher) boom() {
	mw.events <- Boom
	close(mw.events)
}
