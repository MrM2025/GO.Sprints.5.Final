package application

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

type ASTNode struct {
	IsLeaf        bool
	Value         float64
	Operator      string
	Left, Right   *ASTNode
	TaskScheduled bool
}

type Token struct {
	Type  TokenType
	Value string
}

type TokenType int

const (
	Number TokenType = iota
	Operator
	LParen
	RParen
	EOF
)

type Parser struct {
	tokens []Token
	pos    int
}

func NewNumberNode(value float64) *ASTNode {
	return &ASTNode{
		IsLeaf: true,
		Value:  value,
	}
}

func NewOperatorNode(operator string, left, right *ASTNode) *ASTNode {
	return &ASTNode{
		IsLeaf:   false,
		Operator: operator,
		Left:     left,
		Right:    right,
	}
}

func tokenize(expr string) ([]Token, error) {
	var tokens []Token
	runes := []rune(expr)
	n := len(runes)
	i := 0

	for i < n {
		r := runes[i]
		switch {
		case unicode.IsSpace(r):
			i++

		case r == '(':
			tokens = append(tokens, Token{Type: LParen, Value: "("})
			i++

		case r == ')':
			tokens = append(tokens, Token{Type: RParen, Value: ")"})
			i++

		case r == '+' || r == '-' || r == '*' || r == '/':
			tokens = append(tokens, Token{Type: Operator, Value: string(r)})
			i++

		case unicode.IsDigit(r) || r == '.':
			start := i
			for i < n && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
				i++
			}
			tokens = append(tokens, Token{Type: Number, Value: string(runes[start:i])})

		default:
			return nil, fmt.Errorf("неизвестный символ: %c", r)
		}
	}
	tokens = append(tokens, Token{Type: EOF})
	return tokens, nil
}

func (p *Parser) parseExpression() (*ASTNode, error) {
	node, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.peek()
		if tok.Type != Operator || (tok.Value != "+" && tok.Value != "-") {
			break
		}
		op := tok.Value
		p.consume()

		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		node = NewOperatorNode(op, node, right)
	}
	return node, nil
}

func (p *Parser) parseTerm() (*ASTNode, error) {
	node, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.peek()
		if tok.Type != Operator || (tok.Value != "*" && tok.Value != "/") {
			break
		}
		op := tok.Value
		p.consume()

		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		node = NewOperatorNode(op, node, right)
	}
	return node, nil
}

func (p *Parser) parseFactor() (*ASTNode, error) {
	var ops []string
	for {
		tok := p.peek()
		if tok.Type == Operator && (tok.Value == "+" || tok.Value == "-") {
			ops = append(ops, tok.Value)
			p.consume()
		} else {
			break
		}
	}

	node, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == LParen {
		p.consume()
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if next := p.consume(); next.Type != RParen {
			return nil, errors.New("ожидалась закрывающая скобка")
		}
		node = NewOperatorNode("*", node, right)
	}

	for i := len(ops) - 1; i >= 0; i-- {
		op := ops[i]
		node = NewOperatorNode(op, NewNumberNode(0), node)
	}

	return node, nil
}

func (p *Parser) parsePrimary() (*ASTNode, error) {
	tok := p.consume()
	switch tok.Type {
	case Number:
		val, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("неверное число: %s", tok.Value)
		}
		return NewNumberNode(val), nil

	case LParen:
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if next := p.consume(); next.Type != RParen {
			return nil, errors.New("ожидалась закрывающая скобка")
		}
		return expr, nil

	default:
		return nil, errors.New("неожиданный токен в выражении")
	}
}

func (p *Parser) peek() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: EOF}
}

func (p *Parser) consume() Token {
	if p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		p.pos++
		return tok
	}
	return Token{Type: EOF}
}

func ParseAST(expr string) (*ASTNode, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return nil, err
	}

	parser := &Parser{tokens: tokens}
	node, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	if parser.peek().Type != EOF {
		return nil, errors.New("неполный разбор выражения")
	}
	return node, nil
}

func Evaluate(node *ASTNode) (float64, error) {
	if node == nil {
		return 0, errors.New("пустой узел")
	}

	if node.IsLeaf {
		return node.Value, nil
	}

	leftVal, err := Evaluate(node.Left)
	if err != nil {
		return 0, err
	}
	rightVal, err := Evaluate(node.Right)
	if err != nil {
		return 0, err
	}

	switch node.Operator {
	case "+":
		return leftVal + rightVal, nil
	case "-":
		return leftVal - rightVal, nil
	case "*":
		return leftVal * rightVal, nil
	case "/":
		if rightVal == 0 {
			return 0, errors.New("деление на ноль")
		}
		return leftVal / rightVal, nil
	default:
		return 0, fmt.Errorf("неизвестный оператор: %s", node.Operator)
	}
}
