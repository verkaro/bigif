package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/verkaro/bigif/bigif" // Import the engine package
)

func main() {
	// 1. Read a script file from disk.
	scriptBytes, err := ioutil.ReadFile("story.biff")
	if err != nil {
		log.Fatalf("Failed to read script file: %v", err)
	}

	// 2. Call the engine's public Compile function.
	// This is the primary API for the BigIF engine.
	storyGraphJSON, err := bigif.Compile(string(scriptBytes))
	if err != nil {
		log.Fatalf("Engine failed to compile script: %v", err)
	}

	// 3. The output is a JSON byte slice, ready to be used.
	// We'll just print it to the console for this demonstration.
	fmt.Println("--- BigIF Engine Output ---")
	fmt.Println(string(storyGraphJSON))
}

