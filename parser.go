package bigif

import (
	"bufio"
	"fmt"
	"strings"
)

// parse takes the raw script string and converts it into an AST.
func parse(scriptContent string) (*Script, error) {
	script := &Script{
		Metadata:     make(map[string]string),
		GlobalStates: make(map[string]bool),
		LocalStates:  make(map[string]bool),
		Knots:        make(map[string]*Knot),
	}
	var currentKnot *Knot
	var currentTextBlock *TextBlock

	scanner := bufio.NewScanner(strings.NewReader(scriptContent))
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// End of a multi-paragraph block
		if trimmedLine == "" {
			currentTextBlock = nil
			continue
		}

		// Parse Header (Metadata and State Declarations)
		if currentKnot == nil && strings.HasPrefix(trimmedLine, "//") {
			headerLine := strings.TrimSpace(trimmedLine[2:])
			parts := strings.SplitN(headerLine, ":", 2)
			if len(parts) != 2 {
				continue // Simple comment, not a key-value pair
			}
			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			switch strings.ToUpper(key) {
			case "STATES":
				for _, state := range strings.Split(value, ",") {
					script.GlobalStates[strings.TrimSpace(state)] = false
				}
			case "FLAG-STATES":
				for _, state := range strings.Split(value, ",") {
					script.GlobalStates[strings.TrimSpace(state)] = true
				}
			case "LOCAL-STATES":
				for _, state := range strings.Split(value, ",") {
					script.LocalStates[strings.TrimSpace(state)] = true
				}
			default:
				script.Metadata[key] = value
			}
			continue
		}

		// Parse Knot Declaration
		if strings.HasPrefix(trimmedLine, "===") && strings.HasSuffix(trimmedLine, "===") {
			knotName := strings.TrimSpace(trimmedLine[3 : len(trimmedLine)-3])
			if knotName == "" {
				return nil, fmt.Errorf("found knot with empty name")
			}
			currentKnot = &Knot{Name: knotName}
			script.Knots[knotName] = currentKnot
			currentTextBlock = nil
			continue
		}

		if currentKnot == nil {
			continue // Still in header area, but not a recognized header line
		}

		// Parse lines within a Knot
		switch {
		// Knot-level comment or scene directive
		case strings.HasPrefix(trimmedLine, "//"):
			lineContent := strings.TrimSpace(trimmedLine[2:])
			if parts := strings.SplitN(lineContent, ":", 2); len(parts) == 2 && strings.TrimSpace(parts[0]) == "scene" {
				currentKnot.Scene = strings.TrimSpace(parts[1])
			}
			continue

		// END keyword
		case trimmedLine == "END":
			currentKnot.IsEnd = true
			currentTextBlock = nil
			continue

		// Choice
		case strings.HasPrefix(trimmedLine, "*"):
			currentTextBlock = nil
			choice, err := parseChoice(trimmedLine)
			if err != nil {
				return nil, fmt.Errorf("failed to parse choice '%s': %w", trimmedLine, err)
			}
			currentKnot.Choices = append(currentKnot.Choices, *choice)
			continue

		// Conditional Text Block
		case strings.HasPrefix(trimmedLine, "-"):
			block, err := parseTextBlock(trimmedLine)
			if err != nil {
				return nil, err
			}
			currentKnot.Body = append(currentKnot.Body, *block)
			currentTextBlock = block
			continue

		// Indented line (part of a multi-paragraph text block)
		case (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) && currentTextBlock != nil:
			currentTextBlock.Content += "\n" + trimmedLine
			continue

		// Default: Unconditional, non-indented text
		default:
			if currentKnot.Body == nil { // First line of body text
				block := &TextBlock{Content: trimmedLine}
				currentKnot.Body = append(currentKnot.Body, *block)
				currentTextBlock = block
			} else if currentTextBlock != nil { // Multi-line for a block that didn't use hyphens
				currentTextBlock.Content += "\n" + trimmedLine
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	// Clean up content whitespace
	for _, knot := range script.Knots {
		for i := range knot.Body {
			knot.Body[i].Content = strings.TrimSpace(knot.Body[i].Content)
		}
	}

	return script, nil
}

func parseChoice(line string) (*Choice, error) {
	c := &Choice{}
	remainder := strings.TrimSpace(line[1:])

	// Extract condition
	if start := strings.Index(remainder, "{"); start != -1 {
		end := strings.Index(remainder, "}")
		if end == -1 || end < start {
			return nil, fmt.Errorf("mismatched braces in condition")
		}
		c.Condition = strings.TrimSpace(remainder[start+1 : end])
		remainder = remainder[:start] + remainder[end+1:]
	}

	// Extract state changes
	if parts := strings.Split(remainder, "~"); len(parts) > 1 {
		remainder = strings.TrimSpace(parts[0])
		for _, change := range parts[1:] {
			c.StateChanges = append(c.StateChanges, strings.TrimSpace(change))
		}
	}
	
	// Extract target
	if parts := strings.Split(remainder, "->"); len(parts) > 1 {
		c.Text = strings.TrimSpace(parts[0])
		target := strings.TrimSpace(parts[1])
		if strings.HasPrefix(target, ".") {
			c.Stitch = target
		} else {
			c.TargetKnot = target
		}
	} else {
		c.Text = strings.TrimSpace(remainder)
	}

	if c.Text == "" && c.TargetKnot == "" && len(c.StateChanges) == 0 {
		return nil, fmt.Errorf("choice appears to be empty")
	}

	return c, nil
}

func parseTextBlock(line string) (*TextBlock, error) {
	b := &TextBlock{}
	remainder := strings.TrimSpace(line[1:])
	
	// Extract condition
	if start := strings.Index(remainder, "{"); start != -1 {
		end := strings.Index(remainder, "}")
		if end == -1 || end < start {
			return nil, fmt.Errorf("mismatched braces in condition")
		}
		b.Condition = strings.TrimSpace(remainder[start+1 : end])
		remainder = remainder[:start] + remainder[end+1:]
	}
	
	b.Content = strings.TrimSpace(remainder)
	return b, nil
}

