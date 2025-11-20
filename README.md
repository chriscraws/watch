# watch

A minimal, dependency-free Go package for watching files and notifying registered objects when changes occur. Designed for flexibility, it supports custom file systems and tracks multiple files and their dependencies.

## Overview

The `watch` package provides a simple interface for monitoring files and triggering callbacks when those files are updated or created. It is suitable for build systems, live-reload tools, or any application that needs to react to file changes.

## Features

- **No dependencies:** Pure Go, no external libraries required.
- **Custom file system support:** Works with any `fs.FS` implementation.
- **Multiple file tracking:** Watch many files and their dependencies.
- **Flexible notification:** Register any object implementing the `Node` interface.
- **Synchronous updates:** All notifications are handled synchronously.

## Usage

### Basic Example

```go
package main

import (
	"fmt"
	"github.com/chriscraws/sol/go/watch"
)

type myNode struct {
	files []string
}

func (n *myNode) Paths() []string {
	return n.files
}

func (n *myNode) Updated() error {
	fmt.Println("Files updated!")
	return nil
}

func main() {
	w := new(watch.Watcher)
	node := &myNode{files: []string{"file1.txt", "file2.txt"}}
	w.Register(node)
	// Call w.Scan() periodically to check for updates
	changed, errs := w.Scan()
	if changed {
		fmt.Println("Some files changed!")
	}
	if len(errs) > 0 {
		fmt.Println("Errors:", errs)
	}
}
```

### Registering and Unregistering Nodes

- `Register(node Node)`: Start watching a node.
- `Unregister(node Node)`: Stop watching a node.

### Scanning for Changes

- `Scan() (bool, []error)`: Checks all registered nodes for file changes and calls their `Updated()` method if needed.

## API Summary

- **Node interface**
  - `Paths() []string`: Returns the list of file paths to watch.
  - `Updated() error`: Called when any watched file changes.

- **Watcher struct**
  - `Register(node Node)`: Register a node for updates.
  - `Unregister(node Node)`: Unregister a node.
  - `Scan() (bool, []error)`: Scan for file changes and notify nodes.
  - `UpdateAll() []error`: Call `Updated()` on all nodes.
  - `Empty() bool`: Returns true if no nodes are registered.

## Testing

Unit tests are provided in [`watch_test.go`](./watch_test.go), covering:

- Registering/unregistering nodes
- Watching single and multiple files
- Dependency tracking
- Correct notification behavior

Run tests with:

```sh
go test
```

## Installation

```sh
go get github.com/chriscraws/sol/go/watch
```

## License

[MIT](./LICENSE)
