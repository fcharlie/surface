package surface_test

import (
	"os"

	"gitee.com/oscstudio/surface"
)

func slot_test() {
	var slot surface.Slot
	slot.INFO("%s", os.Args[0])
}
