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

		// A blank line can be part of a multi-paragraph block, so only reset
		// currentTextBlock when a new structural element is found.
		if trimmedLine == "" {
			// If we are in a text block, treat the blank line as a paragraph break.
			if currentTextBlock != nil {
				currentTextBlock.Content += "\n"
			}
			continue
		}

		// --- Header Parsing ---
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

		// --- Knot Declaration ---
		if strings.HasPrefix(trimmedLine, "===") && strings.HasSuffix(trimmedLine, "===") {
			knotName := strings.TrimSpace(trimmedLine[3 : len(trimmedLine)-3])
			if knotName == "" {
				return nil, fmt.Errorf("found knot with empty name")
			}
			currentKnot = &Knot{Name: knotName}
			script.Knots[knotName] = currentKnot
			currentTextBlock = nil // Reset for the new knot
			continue
		}

		if currentKnot == nil {
			continue // Skip lines before the first knot that aren't headers
		}

		// --- Knot Content Parsing ---

		// Reset currentTextBlock pointer if a new structural element is found.
		if strings.HasPrefix(trimmedLine, "*") || strings.HasPrefix(trimmedLine, "//") || trimmedLine == "END" {
			currentTextBlock = nil
		}
		
		switch {
		// Comment or scene directive
		case strings.HasPrefix(trimmedLine, "//"):
			lineContent := strings.TrimSpace(trimmedLine[2:])
			if parts := strings.SplitN(lineContent, ":", 2); len(parts) == 2 && strings.TrimSpace(parts[0]) == "scene" {
				currentKnot.Scene = strings.TrimSpace(parts[1])
			}
		// End of path
		case trimmedLine == "END":
			currentKnot.IsEnd = true
		// Choice line
		case strings.HasPrefix(trimmedLine, "*"):
			choice, err := parseChoice(trimmedLine)
			if err != nil {
				return nil, fmt.Errorf("failed to parse choice '%s': %w", trimmedLine, err)
			}
			currentKnot.Choices = append(currentKnot.Choices, *choice)
		// Conditional text block
		case strings.HasPrefix(trimmedLine, "-"):
			block, err := parseTextBlock(trimmedLine)
			if err != nil {
				return nil, err
			}
			currentKnot.Body = append(currentKnot.Body, *block)
			// Point to the actual copy in the slice
			currentTextBlock = &currentKnot.Body[len(currentKnot.Body)-1]
		// Default case for text content
		default:
			if currentTextBlock != nil {
				// This is a subsequent line of a multi-paragraph block.
				currentTextBlock.Content += "\n" + trimmedLine
			} else { 
				// This is the first line of an unconditional body text.
				block := TextBlock{Content: trimmedLine}
				currentKnot.Body = append(currentKnot.Body, block)
				// Point to the actual copy in the slice
				currentTextBlock = &currentKnot.Body[len(currentKnot.Body)-1]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	// Final cleanup of content whitespace
	for _, knot := range script.Knots {
		for i := range knot.Body {
			knot.Body[i].Content = strings.TrimSpace(knot.Body[i].Content)
		}
	}

	return script, nil
}

// parseChoice correctly parses a choice line by handling target, state changes, and conditions.
func parseChoice(line string) (*Choice, error) {
	c := &Choice{}
	remainder := strings.TrimSpace(line[1:])

	// --- Logic Correction: Parse from right to left to avoid clobbering ---

	// 1. Extract Target (->) first
	if parts := strings.SplitN(remainder, "->", 2); len(parts) > 1 {
		remainder = strings.TrimSpace(parts[0])
		target := strings.TrimSpace(parts[1])
		if strings.HasPrefix(target, ".") {
			c.Stitch = target
		} else {
			c.TargetKnot = target
		}
	}

	// 2. Extract State Changes (~) from the remaining string
	if parts := strings.Split(remainder, "~"); len(parts) > 1 {
		remainder = strings.TrimSpace(parts[0])
		for _, change := range parts[1:] {
			trimmedChange := strings.TrimSpace(change)
			if trimmedChange != "" {
				c.StateChanges = append(c.StateChanges, trimmedChange)
			}
		}
	}

	// 3. Extract Condition ({}) from the remaining string
	if start := strings.Index(remainder, "{"); start != -1 {
		end := strings.Index(remainder, "}")
		if end == -1 || end < start {
			return nil, fmt.Errorf("mismatched braces in condition")
		}
		c.Condition = strings.TrimSpace(remainder[start+1 : end])
		remainder = remainder[:start] + remainder[end+1:]
	}

	// 4. Whatever is left is the choice text
	c.Text = strings.TrimSpace(remainder)

	if c.Text == "" && c.TargetKnot == "" && len(c.StateChanges) == 0 {
		return nil, fmt.Errorf("choice appears to be empty")
	}

	return c, nil
}

// parseTextBlock parses a conditional text block line.
func parseTextBlock(line string) (*TextBlock, error) {
	b := &TextBlock{}
	remainder := strings.TrimSpace(line[1:])
	
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

