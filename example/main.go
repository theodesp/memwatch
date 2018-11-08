package main

import (
	"github.com/theodesp/memwatch"
	"fmt"
	"time"
)

func main()  {
	watch := memwatch.New(&memwatch.WatchConfig{
		// Will cycle 10 times if memory is over the WarningLimit
		Cycle: 10,
		WarningLimit: 300 * memwatch.KiloByte,
		// Once we reach the CriticalLimit though we bail immediately
		CriticalLimit: 300 * memwatch.KiloByte,
		// Will exit after 2 seconds
		ExitTime: 2 * time.Second,
	})

	events := watch.Start()
	time.Sleep(2*time.Second)

	watch.Stop()
	time.Sleep(6*time.Second)
	events = watch.Start()

	// Watch for memory overload
	boom := <-events

	// boom you have 2 seconds to evacuate
	fmt.Print(boom)
	time.Sleep(10 * time.Second) // too late!

	fmt.Println("clean up") // sorry won't happen!
}
