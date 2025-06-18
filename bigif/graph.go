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

	rootNode, err := createNode("index", ast.Knots["index"], initialState)
	if err != nil {
		return nil, err
	}
	nodeID := generateNodeID(rootNode.KnotName, rootNode.State)

	graph.Graph[nodeID] = rootNode
	queue = append(queue, rootNode)
	visited[nodeID] = true

	for len(queue) > 0 {
		currentNode := queue[0]
		queue = queue[1:]

		currentKnot := ast.Knots[currentNode.KnotName]

		for _, choice := range currentKnot.Choices {
			if choice.Condition != "" && !evaluateCondition(choice.Condition, currentNode.State) {
				continue
			}

			nextState := applyStateChanges(currentNode.State, choice, ast)

			var targetKnotName string
			if choice.Stitch != "" {
				// Stitches are local jumps, so the "knot" doesn't change, but we need a new node for the stitch content.
				// This is a simplification for the POC; a full implementation might handle this differently.
				// For now, we treat a stitch as a choice leading to a new "knot" with the stitch name.
				targetKnotName = strings.TrimPrefix(choice.Stitch, ".")
			} else {
				targetKnotName = choice.TargetKnot
			}

			if targetKnotName == "" {
				if len(choice.StateChanges) > 0 {
					targetKnotName = currentNode.KnotName
				} else {
					continue
				}
			}
			
			targetKnot, exists := ast.Knots[targetKnotName]
			if !exists {
				return nil, fmt.Errorf("choice leads to non-existent knot: '%s'", targetKnotName)
			}
			
			if currentKnot.Scene != targetKnot.Scene {
				for state := range ast.LocalStates {
					nextState[state] = false
				}
			}

			nextNode, err := createNode(targetKnotName, targetKnot, nextState)
			if err != nil {
				return nil, err
			}
			nextNodeID := generateNodeID(nextNode.KnotName, nextNode.State)
			
			edge := &StoryEdge{Text: choice.Text, TargetNodeID: nextNodeID, Stitch: choice.Stitch}
			currentNode.Edges = append(currentNode.Edges, edge)
			
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
	for _, block := range knot.Body {
		if block.Condition == "" || evaluateCondition(block.Condition, state) {
			node.Content = block.Content
			break
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
			return false
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
			return false
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

		if isFlag, ok := ast.GlobalStates[stateName]; ok && isFlag && !newValue {
			continue
		}

		nextState[stateName] = newValue
	}
	return nextState
}

