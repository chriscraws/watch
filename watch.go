// package watch implements a type for watching files that import other
// files. It has no dependencies and works for a variety of coding systems.
package watch

import (
	"io/fs"
	"os"
)

// Node is an interface for a set of files that should be watched. A Node is
// attached to a Watcher via the Register method on Watcher. When Scan is
// called on the Watcher, Paths is called to determine which paths to check for
// a new modification time. If the modification time is different, or if the
// file previously did not exist, Updated is called on the Node.
type Node interface {
	// Paths returns all paths that should be scanned for updates. Paths
	// can return new values, but should be consistent between calls to Updated.
	Paths() []string

	// Updated is called by Watcher when a change is detected at one of the
	// paths last returned by Paths during a call to Scan.
	Updated() error
}

// Watcher is a struct that can notify Node objects when the paths they
// reference have been updated. Nodes are registered via Register and
// unregistered via Unregister. The zero-value of Watcher is ready to
// use. Scan is used to check for file updates and calls Updated
// synchronously on all registerd nodes with updates.
type Watcher struct {
	FS          fs.FS
	initialized bool
	nodes       map[Node]struct{}
	paths       map[string]*pathStat
}

type pathStat struct {
	info    fs.FileInfo
	visited bool
	updated bool
	nodes   map[Node]struct{}
}

func (w *Watcher) init() {
	w.initialized = true
	w.nodes = make(map[Node]struct{})
	w.paths = make(map[string]*pathStat)
}

// Empty returns true if the watcher is not observing any nodes.
func (w *Watcher) Empty() bool {
	return len(w.nodes) == 0
}

// Register registers a node to be observed on sucessive calls to Scan.
func (w *Watcher) Register(node Node) {
	if !w.initialized {
		w.init()
	}
	if _, ok := w.nodes[node]; ok {
		return
	}
	w.nodes[node] = struct{}{}
}

// Unregister unregisters a node from being observed on sucessive calls to Scan.
func (w *Watcher) Unregister(node Node) {
	if !w.initialized {
		w.init()
	}
	delete(w.nodes, node)
}

// UpdateAll calls Updated on all registered nodes. Does not modify the files,
// so Scan may still trigger changes.
func (w *Watcher) UpdateAll() []error {
	var errors []error
	for node := range w.nodes {
		if err := node.Updated(); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// Scan synchronously calls Updated on each registered Node that references a path
// where a file has been updated or created since the last call to Scan.
// The first time Scan is called, Updated will not be called for existing
// files.
func (w *Watcher) Scan() (bool, []error) {
	if !w.initialized {
		w.init()
	}

	// reset all paths
	for _, stat := range w.paths {
		stat.visited = false
		stat.updated = false
	}

	// scan all paths and determine which have changed
	for node := range w.nodes {
		for _, path := range node.Paths() {
			stat, pathExistedAlready := w.paths[path]
			if stat == nil {
				stat = new(pathStat)
				w.paths[path] = stat
			}
			if stat.visited {
				stat.nodes[node] = struct{}{}
				continue
			}
			stat.visited = true
			stat.nodes = map[Node]struct{}{node: {}}
			var info os.FileInfo
			if fsys, ok := w.FS.(fs.StatFS); ok {
				info, _ = fsys.Stat(path)
			} else {
				info, _ = os.Stat(path)
			}
			if info != nil {
				if stat.info != nil {
					if !stat.info.ModTime().Equal(info.ModTime()) {
						stat.updated = true
					}
				} else if pathExistedAlready {
					stat.updated = true
				}
				stat.info = info
			}
		}
	}

	// delete unused paths and collect updated nodes
	updatedNodes := map[Node]struct{}{}
	for path, stat := range w.paths {
		if !stat.visited {
			delete(w.paths, path)
		}
		if stat.updated {
			for node := range stat.nodes {
				updatedNodes[node] = struct{}{}
			}
		}
	}

	// notify nodes
	var errors []error
	for node := range updatedNodes {
		if err := node.Updated(); err != nil {
			errors = append(errors, err)
		}
	}

	return len(updatedNodes) > 0, errors
}
