package bigif

// Script represents the entire parsed script as an Abstract Syntax Tree (AST).
type Script struct {
	Metadata     map[string]string
	GlobalStates map[string]bool // True if a state is a FLAG-STATE
	LocalStates  map[string]bool // True if a state is a LOCAL-STATE
	Knots        map[string]*Knot
}

// Knot represents a single content block, e.g., === knot_name ===
type Knot struct {
	Name    string
	Scene   string
	Body    []TextBlock
	Choices []Choice
	IsEnd   bool
}

// TextBlock represents a conditional block of text in a Knot's body.
type TextBlock struct {
	Condition string // Raw condition text, e.g., "has_key == true"
	Content   string // The multi-line body text
}

// Choice represents a single choice line, e.g., * Text {condition} ~ state_change -> target
type Choice struct {
	Text         string
	Condition    string   // Raw condition text, e.g., "has_key == true && has_torch == true"
	StateChanges []string // e.g., ["has_key = false", "torch_lit = true"]
	TargetKnot   string
	Stitch       string // e.g., ".stitch_name"
}

