package bigif

import (
	"encoding/json"
	"fmt"
)

// StoryGraph is the final, processed output of the engine. It contains only reachable states.
type StoryGraph struct {
	Metadata map[string]string      `json:"metadata"`
	Graph    map[string]*StoryNode `json:"nodes"`
}

// StoryNode represents a single, unique, and reachable state in the narrative.
type StoryNode struct {
	KnotName string          `json:"knotName"`
	Scene    string          `json:"scene"`
	State    map[string]bool `json:"state"`
	Content  string          `json:"content"`
	Edges    []*StoryEdge    `json:"edges"`
	IsEnd    bool            `json:"isEnd"`
	Stitch   string          `json:"stitch,omitempty"`
}

// StoryEdge represents a choice leading from one StoryNode to another.
type StoryEdge struct {
	Text         string `json:"text"`
	TargetNodeID string `json:"targetNodeId"`
	Stitch       string `json:"stitch,omitempty"`
}

// Compile is the main public entry point for the BigIF engine.
// It takes a script as a string and returns the fully processed StoryGraph as a JSON byte slice.
func Compile(scriptContent string) ([]byte, error) {
	// 1. Parse the script into an AST
	ast, err := parse(scriptContent)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	// 2. Analyze the AST to build the graph of reachable states
	graph, err := buildGraph(ast)
	if err != nil {
		return nil, fmt.Errorf("graph analysis error: %w", err)
	}

	// 3. Serialize the final graph to JSON
	output := map[string]interface{}{
		"metadata": ast.Metadata,
		"graph":    graph.Graph, // Note: directly embedding the map
	}

	return json.MarshalIndent(output, "", "  ")
}

