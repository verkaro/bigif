package bigif

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleCompilation(t *testing.T) {
	script := `
// title: My Story
// STATES: has_key

=== index ===
The door is locked.
* {has_key == false} Look for a key. ~ has_key = true
* {has_key == true} Open the door. -> victory

=== victory ===
You opened the door!
END
`
	outputJSON, err := Compile(script)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(outputJSON, &result)
	require.NoError(t, err)

	// Check metadata
	metadata := result["metadata"].(map[string]interface{})
	assert.Equal(t, "My Story", metadata["title"])

	// Check graph nodes
	graph := result["graph"].(map[string]interface{})

	// Should generate 3 reachable nodes
	// 1. index, key=false
	// 2. index, key=true
	// 3. victory, key=true
	assert.Len(t, graph, 3, "Should have exactly 3 reachable nodes")

	// Verify one of the nodes
	node1 := graph["index|has_key=false"].(map[string]interface{})
	assert.Equal(t, "index", node1["knotName"])
	assert.Equal(t, "The door is locked.", node1["content"])

	edges := node1["edges"].([]interface{})
	assert.Len(t, edges, 1) // Only one choice is available
	edge1 := edges[0].(map[string]interface{})
	assert.Equal(t, "Look for a key.", edge1["text"])
	assert.Equal(t, "index|has_key=true", edge1["targetNodeId"])
}

func TestFlagState(t *testing.T) {
	script := `
// FLAG-STATES: major_event

=== index ===
* Do the thing. ~ major_event = true -> next

=== next ===
You did the thing.
* Try to undo it. ~ major_event = false -> index
`
	outputJSON, err := Compile(script)
	require.NoError(t, err)

	var result map[string]interface{}
	json.Unmarshal(outputJSON, &result)
	graph := result["graph"].(map[string]interface{})
	
	// The edge from 'next' should lead to a state where major_event is STILL true
	// because flags cannot be set to false.
	nextNode := graph["next|major_event=true"].(map[string]interface{})
	edges := nextNode["edges"].([]interface{})
	edge := edges[0].(map[string]interface{})
	
	assert.Equal(t, "index|major_event=true", edge["targetNodeId"])
}

func TestLocalState(t *testing.T) {
	script := `
// LOCAL-STATES: has_room_key
// STATES: global_quest_active

=== room1 ===
// scene: bedroom
* Pick up key. ~ has_room_key = true
* Leave room. -> hallway

=== hallway ===
// scene: corridor
* Go back. -> room1
`
	outputJSON, err := Compile(script)
	require.NoError(t, err)

	var result map[string]interface{}
	json.Unmarshal(outputJSON, &result)
	graph := result["graph"].(map[string]interface{})

	// From room1 with key, going to hallway
	node1 := graph["room1|global_quest_active=false,has_room_key=true"].(map[string]interface{})
	edgeToHallway := node1["edges"].([]interface{})[1].(map[string]interface{})
	
	// The target node in the hallway should not have has_room_key=true, it should be purged
	expectedTargetID := "hallway|global_quest_active=false,has_room_key=false"
	assert.Equal(t, expectedTargetID, edgeToHallway["targetNodeId"])
}

func TestConditionalText(t *testing.T) {
	script := `
// STATES: power_on

=== index ===
- {power_on == false} The room is dark.
  It is very spooky.
- {power_on == true} The lights are on.
* Flip switch. ~ power_on = true
`
	outputJSON, err := Compile(script)
	require.NoError(t, err)
	var result map[string]interface{}
	json.Unmarshal(outputJSON, &result)
	graph := result["graph"].(map[string]interface{})

	darkNode := graph["index|power_on=false"].(map[string]interface{})
	assert.Equal(t, "The room is dark.\nIt is very spooky.", darkNode["content"])

	lightNode := graph["index|power_on=true"].(map[string]interface{})
	assert.Equal(t, "The lights are on.", lightNode["content"])
}

func TestUnreachableStatePruning(t *testing.T) {
	script := `
// STATES: has_key

=== index ===
* Get the key. ~ has_key = true -> door

=== door ===
This door requires a key.
* {has_key == true} Open it. -> victory

=== victory ===
You win.
END
`
	outputJSON, err := Compile(script)
	require.NoError(t, err)
	var result map[string]interface{}
	json.Unmarshal(outputJSON, &result)
	graph := result["graph"].(map[string]interface{})

	// The state `door|has_key=false` should NOT exist, because it's impossible to reach
	// the door knot without having the key.
	_, exists := graph["door|has_key=false"]
	assert.False(t, exists, "An unreachable node was generated")
	assert.Len(t, graph, 3, "Should only have 3 reachable nodes")
}

