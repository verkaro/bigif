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

		if trimmedLine == "" {
			if currentTextBlock != nil {
				currentTextBlock.Content += "\n"
			}
			continue
		}

		// --- Header Parsing ---
		if currentKnot == nil && strings.HasPrefix(trimmedLine, "//") {
			parseHeaderLine(trimmedLine, script)
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
			currentTextBlock = nil
			continue
		}

		if currentKnot == nil {
			continue
		}

		if strings.HasPrefix(trimmedLine, "*") || strings.HasPrefix(trimmedLine, "//") || trimmedLine == "END" {
			currentTextBlock = nil
		}
		
		switch {
		case strings.HasPrefix(trimmedLine, "//"):
			lineContent := strings.TrimSpace(trimmedLine[2:])
			if parts := strings.SplitN(lineContent, ":", 2); len(parts) == 2 && strings.TrimSpace(parts[0]) == "scene" {
				currentKnot.Scene = strings.TrimSpace(parts[1])
			}
		case trimmedLine == "END":
			currentKnot.IsEnd = true
		case strings.HasPrefix(trimmedLine, "*"):
			choice, err := parseChoice(trimmedLine)
			if err != nil {
				return nil, fmt.Errorf("failed to parse choice '%s': %w", trimmedLine, err)
			}
			currentKnot.Choices = append(currentKnot.Choices, *choice)
		case strings.HasPrefix(trimmedLine, "-"):
			block, err := parseTextBlock(trimmedLine)
			if err != nil {
				return nil, err
			}
			currentKnot.Body = append(currentKnot.Body, *block)
			currentTextBlock = &currentKnot.Body[len(currentKnot.Body)-1]
		default:
			if currentTextBlock != nil {
				currentTextBlock.Content += "\n" + trimmedLine
			} else {
				block := TextBlock{Content: trimmedLine}
				currentKnot.Body = append(currentKnot.Body, block)
				currentTextBlock = &currentKnot.Body[len(currentKnot.Body)-1]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	for _, knot := range script.Knots {
		for i := range knot.Body {
			knot.Body[i].Content = strings.TrimSpace(knot.Body[i].Content)
		}
	}

	return script, nil
}

// parseHeaderLine processes a single line from the script header.
func parseHeaderLine(line string, script *Script) {
	headerLine := strings.TrimSpace(line[2:])
	parts := strings.SplitN(headerLine, ":", 2)
	if len(parts) != 2 {
		return // It's a simple comment, not a key-value directive.
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
		// This correctly captures any other metadata like 'title', 'author', or 'description'.
		script.Metadata[key] = value
	}
}

func parseChoice(line string) (*Choice, error) {
	c := &Choice{}
	remainder := strings.TrimSpace(line[1:])

	if parts := strings.SplitN(remainder, "->", 2); len(parts) > 1 {
		remainder = strings.TrimSpace(parts[0])
		target := strings.TrimSpace(parts[1])
		if strings.HasPrefix(target, ".") {
			c.Stitch = target
			c.TargetKnot = ""
		} else {
			c.TargetKnot = target
		}
	}

	if parts := strings.Split(remainder, "~"); len(parts) > 1 {
		remainder = strings.TrimSpace(parts[0])
		for _, change := range parts[1:] {
			trimmedChange := strings.TrimSpace(change)
			if trimmedChange != "" {
				c.StateChanges = append(c.StateChanges, trimmedChange)
			}
		}
	}

	if start := strings.Index(remainder, "{"); start != -1 {
		end := strings.Index(remainder, "}")
		if end == -1 || end < start {
			return nil, fmt.Errorf("mismatched braces in condition")
		}
		c.Condition = strings.TrimSpace(remainder[start+1 : end])
		remainder = remainder[:start] + remainder[end+1:]
	}

	c.Text = strings.TrimSpace(remainder)

	if c.Text == "" && c.TargetKnot == "" && len(c.StateChanges) == 0 && c.Stitch == "" {
		return nil, fmt.Errorf("choice appears to be empty")
	}

	return c, nil
}

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

