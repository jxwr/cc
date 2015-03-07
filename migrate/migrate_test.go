package migrate

import (
	"fmt"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	m := NewMigrateManager()
	m.Create("abc", "def", []Range{Range{0, 20}})

	go m.RunTask("abc")

	fmt.Println("=======")
	time.Sleep(3 * time.Second)
	fmt.Println("pause:", m.Pause("abc"))
	time.Sleep(3 * time.Second)
	fmt.Println("resume:", m.Resume("abc"))
	fmt.Println("cancel:", m.Cancel("abc"))
	fmt.Println("resume:", m.Resume("abc"))
}
