# BigIF Engine Specification v0.1

## 1. Core Philosophy

BigIF is a headless logic pre-processor designed to parse a specialized scripting language and compile it into a structured, navigable graph of story states. Its sole purpose is to handle the complex logic of a branching, stateful narrative, providing a clean API for other applications to consume, render, and use. It is presentation and application agnostic.

**Core Principles:**

* **Engine, Not Application:** BigIF's only output is a data structure (the Story Graph). It does not generate files, handle user interfaces, or run servers.
* **Serviceable Simplicity:** The engine's architecture and API are designed to be as simple as possible while robustly handling complex state interactions.
* **Deterministic Output:** For a given script and input, the engine will always produce the exact same Story Graph.
* **Testability:** The engine is designed as a pure function: script in, graph out. This makes it highly testable and verifiable.

## 2. Input: BigIF Script Syntax

The engine ingests text-based scripts conforming to the following syntax.

### 2.1. Fundamental Structure

* **`// KEY: VALUE`:** Top-level comments define metadata (e.g., `title`, `author`) or declare state variables.
* **`=== knot_name ===`:** Defines a content block, or "knot." Every script **must** have a starting knot named `index`.
* **`END`:** Explicitly marks the termination of a narrative path.

### 2.2. State Management

* **Global States (`// STATES: ...`):** A comma-separated list of globally tracked boolean state variables. All states default to `false`.
* **Flag States (`// FLAG-STATES: ...`):** Global boolean states that can only transition from `false` to `true`. Attempts to set a flag to `false` will be ignored.
* **Local States (`// LOCAL-STATES: ...`):** Boolean states scoped to a `scene`. They are reset to `false` when a choice leads to a knot in a different scene.
* **State Manipulation (`~`):** `~ state_name = true/false` modifies a state. Multiple modifications are separated by `~` and evaluated left-to-right. State assignment must use a single equals sign (`=`).

### 2.3. Knot Content & Logic

The body of a knot consists of optional descriptive text followed by a list of choices.

* **Conditional Text (`- {condition} text...`):** The engine evaluates these blocks top-to-bottom and selects the first one whose condition is met. The text block can span multiple indented lines. A block with no condition (`- text...`) is a fallback.
* **Choices (`* text...`):** A list of options available to the user.
    * **Conditional Choices (`* {condition} text...`):** A choice is only available if its condition is met. Conditions support `==`, `!=`, and the `&&` operator for multiple checks.
    * **Diversion (`-> knot_name`):** Navigates to another knot.
    * **Stitches (`-> .stitch_name`):** A local anchor jump. The engine will note this, but the consuming application is responsible for rendering it as an HTML anchor.

## 3. Core Engine Architecture

The engine's primary responsibility is to avoid the "state explosion" problem by intelligently analyzing the script.

* **Reachable State Analysis:** The engine **must not** generate permutations naively. It will build a **directed graph** starting from the `index` knot (with all states `false`) and explore the story choice by choice. Only knot/state combinations that are actually reachable will be instantiated as nodes in the graph.
* **State Pruning:** The engine will correctly apply `FLAG-STATES` and `LOCAL-STATES` rules during its graph traversal to further manage and prune the state space.

## 4. Output: The Story Graph API

The engine's sole output is a data structure representing the fully analyzed and pruned Story Graph. This structure can be serialized (e.g., to JSON) for consumption by other tools.

### 4.1. Graph Structure

```json
{
  "metadata": {
    "title": "The Library of Secrets",
    "author": "AI"
  },
  "graph": {
    "nodes": {
      "index|has_torch=false,has_read_tome=false": {
        "knotName": "index",
        "scene": "library/index",
        "state": { "has_torch": false, "has_read_tome": false, ... },
        "content": "The great library is hushed and still...",
        "edges": [
          { "text": "Read the tome.", "targetNodeId": "index|has_torch=false,has_read_tome=true" },
          { "text": "Search for a torch.", "targetNodeId": "index|has_torch=true,has_read_tome=false" }
        ],
        "isEnd": false
      },
      "index|has_torch=true,has_read_tome=false": { ... }
    }
  }
}
