package tooling

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"unicode"

	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// calculateArgs represents the arguments for the calculate tool.
type calculateArgs struct {
	Expression string `json:"expression"`
}

// NewCalculateTool creates a new calculate tool that evaluates arithmetic expressions.
// Supports +, -, *, / operators with proper precedence and parentheses.
func NewCalculateTool() agent.Tool {
	return agent.Tool{
		ID: "calculate",
		Definition: agent.NewToolDefinition("calculate", "Perform a simple arithmetic calculation").
			WithParameter("expression", "The arithmetic expression to evaluate (e.g., '2 + 2')"),
		Func: Calculate,
	}
}

// Calculate performs a simple arithmetic calculation using a safe evaluator.
// Supports +, -, *, / operators with proper precedence and parentheses.
func Calculate(_ context.Context, arguments string) (string, error) {
	var args calculateArgs
	if err := agent.DecodeArgs(arguments, &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := evaluateExpression(args.Expression)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate expression: %w", err)
	}

	// Format result nicely (remove trailing zeros for whole numbers)
	if result == float64(int64(result)) {
		return strconv.FormatInt(int64(result), 10), nil
	}
	return fmt.Sprintf("%g", result), nil
}

// evaluateExpression safely evaluates a simple arithmetic expression.
// Supports +, -, *, / operators with proper precedence and parentheses.
// This is a minimal recursive descent parser for safety (no eval/reflection).
func evaluateExpression(expr string) (float64, error) {
	p := &exprParser{input: expr, pos: 0}
	result, err := p.parseExpression()
	if err != nil {
		return 0, err
	}
	p.skipWhitespace()
	if p.pos < len(p.input) {
		return 0, fmt.Errorf("unexpected character at position %d: %c", p.pos, p.input[p.pos])
	}
	return result, nil
}

// exprParser is a simple recursive descent parser for arithmetic expressions.
type exprParser struct {
	input string
	pos   int
}

// parseExpression handles addition and subtraction (lowest precedence).
func (p *exprParser) parseExpression() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}

	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return left, nil
		}

		op := p.input[p.pos]
		if op != '+' && op != '-' {
			return left, nil
		}
		p.pos++

		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}

		if op == '+' {
			left += right
		} else {
			left -= right
		}
	}
}

// parseTerm handles multiplication and division (higher precedence).
func (p *exprParser) parseTerm() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}

	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return left, nil
		}

		op := p.input[p.pos]
		if op != '*' && op != '/' {
			return left, nil
		}
		p.pos++

		right, err := p.parseFactor()
		if err != nil {
			return 0, err
		}

		if op == '*' {
			left *= right
		} else {
			if right == 0 {
				return 0, errors.New("division by zero")
			}
			left /= right
		}
	}
}

// parseFactor handles numbers, parentheses, and unary minus (highest precedence).
func (p *exprParser) parseFactor() (float64, error) {
	p.skipWhitespace()

	if p.pos >= len(p.input) {
		return 0, errors.New("unexpected end of expression")
	}

	// Handle unary minus
	if p.input[p.pos] == '-' {
		p.pos++
		factor, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		return -factor, nil
	}

	// Handle unary plus
	if p.input[p.pos] == '+' {
		p.pos++
		return p.parseFactor()
	}

	// Handle parentheses
	if p.input[p.pos] == '(' {
		p.pos++
		result, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] != ')' {
			return 0, errors.New("missing closing parenthesis")
		}
		p.pos++
		return result, nil
	}

	// Parse number
	return p.parseNumber()
}

// parseNumber extracts and parses a numeric value.
func (p *exprParser) parseNumber() (float64, error) {
	p.skipWhitespace()
	start := p.pos

	// Handle leading digits
	for p.pos < len(p.input) && (unicode.IsDigit(rune(p.input[p.pos])) || p.input[p.pos] == '.') {
		p.pos++
	}

	if start == p.pos {
		if p.pos < len(p.input) {
			return 0, fmt.Errorf("expected number at position %d, got '%c'", p.pos, p.input[p.pos])
		}
		return 0, fmt.Errorf("expected number at position %d", p.pos)
	}

	numStr := p.input[start:p.pos]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", numStr)
	}
	return num, nil
}

func (p *exprParser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}
