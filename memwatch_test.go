package memwatch_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/theodesp/memwatch"
)

func TestCalculateTotalMemory(t *testing.T) {
	var mstats runtime.MemStats
	runtime.ReadMemStats(&mstats)

	expectedTotal := memwatch.MemoryUnit(mstats.HeapInuse + mstats.StackInuse +
		mstats.MSpanInuse + mstats.MCacheInuse + mstats.BuckHashSys)

	totalmem := memwatch.CalculateTotalMemory(mstats)

	if totalmem != expectedTotal {
		t.Fatal("Expected total memory to match, but it didn't")
	}
}

func TestMemoryWatcher(t *testing.T) {
	t.Run("new", func(t *testing.T) {

		wcfgDefault := memwatch.WatchConfig{
			WarningLimit:  512 * memwatch.MegaByte,
			CriticalLimit: 768 * memwatch.MegaByte,
			Cycle:         10,
			Interval:      5 * time.Second,
			ExitCode:      101,
			ExitTime:      10 * time.Second,
		}

		t.Log("When no configuration is provided")
		{
			mwatch := memwatch.New(nil)
			if mwatch.GetConfig() != wcfgDefault {
				t.Fatal("Expected default configuration")
			}
		}

		t.Log("When complete new configuration is provided")
		{
			wcfgCustom := memwatch.WatchConfig{
				WarningLimit:  256 * memwatch.MegaByte,
				CriticalLimit: 512 * memwatch.MegaByte,
				Cycle:         5,
				Interval:      15 * time.Second,
				ExitCode:      102,
				ExitTime:      20 * time.Second,
			}

			mwatch := memwatch.New(&wcfgCustom)
			if mwatch.GetConfig() != wcfgCustom {
				t.Fatal("Expected custom configuration")
			}
		}

		t.Log("When partial configuration is provided")
		{
			wcfgExpected1 := memwatch.WatchConfig{
				WarningLimit:  256 * memwatch.MegaByte,
				CriticalLimit: 512 * memwatch.MegaByte,
				Cycle:         wcfgDefault.Cycle,
				Interval:      wcfgDefault.Interval,
				ExitCode:      wcfgDefault.ExitCode,
				ExitTime:      wcfgDefault.ExitTime,
			}

			wcfgExpected2 := memwatch.WatchConfig{
				WarningLimit:  wcfgDefault.WarningLimit,
				CriticalLimit: wcfgDefault.CriticalLimit,
				Cycle:         15,
				Interval:      15 * time.Second,
				ExitCode:      102,
				ExitTime:      20 * time.Second,
			}

			var mwatch *memwatch.MemoryWatcher
			mwatch = memwatch.New(&memwatch.WatchConfig{
				WarningLimit:  256 * memwatch.MegaByte,
				CriticalLimit: 512 * memwatch.MegaByte,
			})
			if mwatch.GetConfig() != wcfgExpected1 {
				t.Fatal("Expected first merged configuration to match, but it didn't")
			}

			mwatch = memwatch.New(&memwatch.WatchConfig{
				Cycle:    15,
				Interval: 15 * time.Second,
				ExitCode: 102,
				ExitTime: 20 * time.Second,
			})
			if mwatch.GetConfig() != wcfgExpected2 {
				t.Fatal("Expected second merged configuration to match, but it didn't")
			}
		}
	})

	t.Run("reach_critical", func(t *testing.T) {
		mwatch := memwatch.New(&memwatch.WatchConfig{
			CriticalLimit: 512 * memwatch.MegaByte,
		})

		notReachedTotal := 512 * memwatch.MegaByte
		if mwatch.ReachCritical(notReachedTotal) {
			t.Fatal("Expected total NOT to reach critical, but it did")
		}

		reachedTotal := 513 * memwatch.MegaByte
		if !mwatch.ReachCritical(reachedTotal) {
			t.Fatal("Expected total to reach critical, but it didn't")
		}
	})

	t.Run("reach_warning", func(t *testing.T) {
		mwatch := memwatch.New(&memwatch.WatchConfig{
			WarningLimit: 256 * memwatch.MegaByte,
		})

		notReachedTotal := 256 * memwatch.MegaByte
		if mwatch.ReachWarning(notReachedTotal) {
			t.Fatal("Expected total NOT to reach warning, but it did")
		}

		reachedTotal := 257 * memwatch.MegaByte
		if !mwatch.ReachWarning(reachedTotal) {
			t.Fatal("Expected total to reach warning, but it didn't")
		}
	})

}
