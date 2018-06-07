# Go basic libary, Aka Surface

TODO


```go
package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"gitee.com/oscstudio/surface"
)

var slot surface.Slot

func main() {
	f, err := os.Create(fmt.Sprintf("cpu_%d_%s.prof", os.Getpid(), time.Now().Format("2006_01_02_03_04_05")))
	if err != nil {
		return
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	slot.Initialize("/tmp/slot_access.log", "/tmp/slot_error.log")
	slot.RolateSize(4096)
	for i := 0; i < 200000; i++ {
		slot.INFO("test %s %d", os.Args[0], i)
		slot.Access("POST /api/v3/internal/test %d", i)
	}
}

```