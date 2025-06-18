# BigIF Engine

BigIF is a headless logic pre-processor designed to parse a specialized scripting language (`.biff`) and compile it into a structured, navigable graph of story states. Its sole purpose is to handle the complex logic of a branching, stateful narrative, providing a clean API for other applications to consume, render, and use. It is presentation and application agnostic.

This engine is the core component that can power applications like static site generators, interactive command-line games, training simulators, and more.

## Features

* **Headless by Design:** The engine's only output is a JSON data structure (the State Graph). It does not generate files, handle user interfaces, or run servers.
* **Simple Scripting Language:** Define complex narratives with an intuitive, author-friendly `.biff` script format.
* **Intelligent State Management:** Automatically prunes unreachable states and impossible state combinations, solving the "state explosion" problem common in branching narratives.
* **Advanced State Control:** Supports global states, one-way "flag" states for major events, and scene-scoped "local" states for managing complexity.
* **Testable & Deterministic:** Designed as a pure function (script in, graph out) for robust testing and predictable output.

## Getting Started

### Prerequisites

* Go 1.18 or higher installed.

### Installation

To add the BigIF engine to your project, use `go get`:

```

go get [github.com/verkaro/bigif](https://www.google.com/url?sa=E&source=gmail&q=https://github.com/verkaro/bigif)

````

*(Note: Replace `verkaro` with the actual repository path once available.)*

### Usage Example

The engine exposes a single public function, `Compile`. Here is a minimal example of how an application would use it:

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"[github.com/verkaro/bigif](https://github.com/verkaro/bigif)" // Import the engine package
)

func main() {
	// 1. Read a script file from disk.
	scriptBytes, err := ioutil.ReadFile("my_story.biff")
	if err != nil {
		log.Fatalf("Failed to read script: %v", err)
	}

	// 2. Call the engine's Compile function.
	storyGraphJSON, err := bigif.Compile(string(scriptBytes))
	if err != nil {
		log.Fatalf("Engine failed to compile script: %v", err)
	}

	// 3. The output is a JSON byte slice, ready to be used.
	fmt.Println(string(storyGraphJSON))

    // An application could now parse this JSON to render a web page,
    // power a game, etc.
}
````

## Architectural Overview

The engine follows a classic compiler design pattern for clarity and testability.

1.  **Parsing (`parser.go`):** The engine first reads the raw `.biff` script string and parses it into an Abstract Syntax Tree (AST) defined in `ast.go`. This provides a structured, in-memory representation of all knots, choices, and conditions.
2.  **Graph Analysis (`graph.go`):** This is the core of the engine. It performs a **Reachable State Analysis** using a breadth-first search algorithm.
      * It starts at the `index` knot with all states `false`.
      * It explores the story choice-by-choice, tracking the state at each step.
      * It builds a directed graph of `StoryNode` objects, creating nodes **only** for knot/state combinations that are actually reachable. This prevents the "state explosion" problem and automatically prunes impossible branches.
      * State management rules (`FLAG-STATES`, `LOCAL-STATES`) are applied during this traversal.
3.  **API Generation (`engine.go`):** The public `Compile` function orchestrates the process. Once the in-memory graph is complete, it serializes the final structure into a JSON object according to the API specification.

## Development

### Fetching Test Dependencies

The test harness uses `stretchr/testify`. You can fetch it with:

```
go get [github.com/stretchr/testify](https://github.com/stretchr/testify)
```

### Testing

To run the comprehensive test harness and verify that all engine logic conforms to the specification, run the following command from the root of the project:

```
go test ./...
```

A successful run will show `PASS` for all tests, giving you verifiable proof that the engine is working as designed.

## Contributing

Contributions are welcome\! Please feel free to submit a pull request or open an issue for any bugs, feature requests, or suggestions.

## Credits

This project's initial specification, architecture, and core implementation were collaboratively developed with Google's **Gemini**.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

