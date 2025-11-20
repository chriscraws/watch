package watch_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/chriscraws/watch"
)

type testNode struct {
	path    string
	deps    []string
	updated int
}

var _ = (*testNode)(nil)

func (tn *testNode) Paths() []string {
	return append(tn.deps, tn.path)
}

func (tn *testNode) Updated() error {
	tn.updated++
	return nil
}

func TestWatcher(t *testing.T) {
	wd := os.TempDir()
	defer os.RemoveAll(wd)

	t.Run("doesn't notify existing file", func(t *testing.T) {
		w := new(watch.Watcher)
		p := path.Join(wd, "single_file.txt")
		defer os.Remove(p)
		n := testNode{path: p}
		w.Register(&n)
		os.Create(p)
		w.Scan()
		if n.updated != 0 {
			t.Errorf("updated should be 0")
		}
	})

	t.Run("watches single file update", func(t *testing.T) {
		w := new(watch.Watcher)
		p := path.Join(wd, "single_file.txt")
		defer os.Remove(p)
		n := testNode{path: p}
		w.Register(&n)
		w.Scan()
		if n.updated != 0 {
			t.Errorf("updated should be 0")
		}

		os.Create(p)
		w.Scan()
		if n.updated != 1 {
			t.Errorf("updated should be 1")
		}

		os.Chtimes(p, time.Now(), time.Now())
		w.Scan()
		if n.updated != 2 {
			t.Errorf("updated should be 2")
		}
	})

	t.Run("register and unregister node", func(t *testing.T) {
		w := new(watch.Watcher)
		p := path.Join(wd, "single_file.txt")
		defer os.Remove(p)
		n := testNode{path: p}
		w.Register(&n)
		w.Scan()
		os.Create(p)
		w.Scan()
		w.Unregister(&n)
		os.Chtimes(p, time.Now(), time.Now())
		w.Scan()
		if n.updated != 1 {
			t.Errorf("updated should be 1")
		}
	})

	t.Run("watches single file with dependency update", func(t *testing.T) {
		w := new(watch.Watcher)
		mainPath := path.Join(wd, "main_file.txt")
		depPath := path.Join(wd, "dep_file.txt")
		depPath2 := path.Join(wd, "dep_file_2.txt")
		defer os.Remove(mainPath)
		defer os.Remove(depPath)
		defer os.Remove(depPath2)
		n := testNode{path: mainPath, deps: []string{depPath, depPath2}}

		os.Create(mainPath)
		w.Register(&n)
		w.Scan()
		if n.updated != 0 {
			t.Errorf("updated should be 0")
		}

		os.Create(depPath)
		w.Scan()
		if n.updated != 1 {
			t.Errorf("updated should be 1")
		}

		os.Create(depPath2)
		w.Scan()
		if n.updated != 2 {
			t.Errorf("updated should be 2")
		}

		os.Chtimes(depPath, time.Now(), time.Now())
		w.Scan()
		if n.updated != 3 {
			t.Errorf("updated should be 3")
		}

		os.Chtimes(depPath2, time.Now(), time.Now())
		w.Scan()
		if n.updated != 4 {
			t.Errorf("updated should be 4")
		}
	})
}
