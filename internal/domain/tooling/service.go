package tooling

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/andygeiss/go-agent/pkg/agent"
)

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

// GetCurrentTime returns the current date and time in RFC3339 format.
func GetCurrentTime(_ context.Context, _ string) (string, error) {
	return time.Now().Format(time.RFC3339), nil
}
