package tooling

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

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

func (p *exprParser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}
