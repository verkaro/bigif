package bigif

import (
	"fmt"
	"sort"
	"strings"
)

// buildGraph performs the reachable state analysis to create the final graph.
func buildGraph(ast *Script) (*StoryGraph, error) {
	if _, ok := ast.Knots["index"]; !ok {
		return nil, fmt.Errorf("script must contain a starting knot named 'index'")
	}

	graph := &StoryGraph{
		Graph: make(map[string]*StoryNode),
	}
	queue := []*StoryNode{}
	visited := make(map[string]bool)

	// Create the initial state
	initialState := make(map[string]bool)
	for state := range ast.GlobalStates {
		initialState[state] = false
	}
	for state := range ast.LocalStates {
		initialState[state] = false
	}

	// Create the root node of the graph (index knot, all states false)
	rootNode, err := createNode("index", ast.Knots["index"], initialState)
	if err != nil {
		return nil, err
	}
	nodeID := generateNodeID(rootNode.KnotName, rootNode.State)

	graph.Graph[nodeID] = rootNode
	queue = append(queue, rootNode)
	visited[nodeID] = true

	// Breadth-First Search (BFS) to discover all reachable nodes
	for len(queue) > 0 {
		currentNode := queue[0]
		queue = queue[1:]

		currentKnot := ast.Knots[currentNode.KnotName]

		// Process each available choice from the current node
		for _, choice := range currentKnot.Choices {
			// Check if the choice's condition is met by the current state
			if choice.Condition != "" && !evaluateCondition(choice.Condition, currentNode.State) {
				continue
			}

			// This choice is available. Calculate the next state.
			nextState := applyStateChanges(currentNode.State, choice, ast)

			// Determine the target knot and create the next node
			targetKnotName := choice.TargetKnot
			if targetKnotName == "" && len(choice.StateChanges) > 0 {
				targetKnotName = currentNode.KnotName // Self-link on state change
			} else if targetKnotName == "" {
				continue // Skip choices without target and without state change
			}
			
			targetKnot, exists := ast.Knots[targetKnotName]
			if !exists {
				return nil, fmt.Errorf("choice leads to non-existent knot: '%s'", targetKnotName)
			}
			
			// If moving to a new scene, purge local states
			if currentKnot.Scene != targetKnot.Scene {
				for state := range ast.LocalStates {
					nextState[state] = false
				}
			}

			// Create the target node and its ID
			nextNode, err := createNode(targetKnotName, targetKnot, nextState)
			if err != nil {
				return nil, err
			}
			nextNodeID := generateNodeID(nextNode.KnotName, nextNode.State)
			
			// Add an edge from the current node to the target node
			edge := &StoryEdge{Text: choice.Text, TargetNodeID: nextNodeID, Stitch: choice.Stitch}
			currentNode.Edges = append(currentNode.Edges, edge)
			
			// If we haven't visited this node before, add it to the graph and the queue
			if !visited[nextNodeID] {
				visited[nextNodeID] = true
				graph.Graph[nextNodeID] = nextNode
				queue = append(queue, nextNode)
			}
		}
	}
	return graph, nil
}

// createNode generates a StoryNode for a given knot and state.
func createNode(knotName string, knot *Knot, state map[string]bool) (*StoryNode, error) {
	node := &StoryNode{
		KnotName: knotName,
		Scene:    knot.Scene,
		State:    state,
		IsEnd:    knot.IsEnd,
		Edges:    []*StoryEdge{},
	}
	// Determine the body content based on conditions
	for _, block := range knot.Body {
		if block.Condition == "" || evaluateCondition(block.Condition, state) {
			node.Content = block.Content
			break // First match wins
		}
	}
	return node, nil
}

// generateNodeID creates a unique, deterministic ID for a node.
func generateNodeID(knotName string, state map[string]bool) string {
	keys := make([]string, 0, len(state))
	for k := range state {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var stateParts []string
	for _, k := range keys {
		stateParts = append(stateParts, fmt.Sprintf("%s=%t", k, state[k]))
	}
	
	return fmt.Sprintf("%s|%s", knotName, strings.Join(stateParts, ","))
}

// evaluateCondition checks if a condition string is true for a given state.
func evaluateCondition(condition string, state map[string]bool) bool {
	parts := strings.Split(condition, "&&")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		
		var op, stateName, valueStr string
		if strings.Contains(part, "!=") {
			op = "!="
			vals := strings.Split(part, "!=")
			stateName, valueStr = strings.TrimSpace(vals[0]), strings.TrimSpace(vals[1])
		} else if strings.Contains(part, "==") {
			op = "=="
			vals := strings.Split(part, "==")
			stateName, valueStr = strings.TrimSpace(vals[0]), strings.TrimSpace(vals[1])
		} else {
			return false // Invalid condition format
		}

		expectedValue := valueStr == "true"
		actualValue := state[stateName]

		var result bool
		if op == "==" {
			result = actualValue == expectedValue
		} else {
			result = actualValue != expectedValue
		}
		if !result {
			return false // Early exit if any part of an AND condition is false
		}
	}
	return true
}

// applyStateChanges calculates the next state based on a choice.
func applyStateChanges(currentState map[string]bool, choice Choice, ast *Script) map[string]bool {
	nextState := make(map[string]bool)
	for k, v := range currentState {
		nextState[k] = v
	}

	for _, change := range choice.StateChanges {
		parts := strings.Split(change, "=")
		stateName := strings.TrimSpace(parts[0])
		newValue := strings.TrimSpace(parts[1]) == "true"

		// Enforce FLAG-STATE rule: can't be set to false
		if isFlag, ok := ast.GlobalStates[stateName]; ok && isFlag && !newValue {
			continue // Silently ignore attempt to set flag to false
		}

		nextState[stateName] = newValue
	}
	return nextState
}

