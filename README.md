memwatch
---
[![All Contributors](https://img.shields.io/badge/all_contributors-0-orange.svg?style=flat-square)](#contributors)

<a href="https://godoc.org/github.com/theodesp/memwatch">
<img src="https://godoc.org/github.com/theodesp/memwatch?status.svg" alt="GoDoc">
</a>

<a href="https://opensource.org/licenses/MIT" rel="nofollow">
<img src="https://img.shields.io/github/license/mashape/apistatus.svg" alt="License"/>
</a>

<a href="https://travis-ci.org/theodesp/memwatch" rel="nofollow">
<img src="https://travis-ci.org/theodesp/memwatch.svg?branch=master" />
</a>

<a href="https://codecov.io/gh/theodesp/memwatch">
  <img src="https://codecov.io/gh/theodesp/memwatch/branch/master/graph/badge.svg" />
</a>

Trips with an event when the runtime memory usage of the process is over the limit and exits after a while.

## Installation

```
go get -u github/theodesp/memwatch
```

## How To Use

```go
package main

import (
	"fmt"
	"time"
	
	"github.com/theodesp/memwatch"
)

func main()  {
	watch := memwatch.New(&memwatch.WatchConfig{
		// Will cycle 10 times if memory is over the WarningLimit
		Cycle: 10,
		WarningLimit: 300 * memwatch.KiloByte,
		// Once we reach the CriticalLimit though we bail immediately
		CriticalLimit: 400 * memwatch.KiloByte,
		// Will exit after 2 seconds
		ExitTime: 2 * time.Second,
	})

	events := watch.Start()

	// Watch for memory overload
	boom := <-events

	// boom you have 2 seconds to evacuate
	fmt.Print(boom)
	time.Sleep(10 * time.Second) // too late!

	fmt.Println("clean up") // sorry won't happen!
}

```

## API
* **Start()**: Starts the monitor
* **Stop()**: Stops the monitor

The code supports building under Go >= 1.8.

## Contributors
Thanks goes to these wonderful people ([emoji key](https://github.com/kentcdodds/all-contributors#emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/kentcdodds/all-contributors) specification. Contributions of any kind welcome!

### LICENCE

Copyright © 2017 Theo Despoudis BSD license
