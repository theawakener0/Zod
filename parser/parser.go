package parser

import (
	"fmt"
	"strconv"

	"github.com/theawakener0/zod/ast"
	lx "github.com/theawakener0/zod/lexer"
	tk "github.com/theawakener0/zod/token"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l 				*lx.Lexer
	
	errors			[]string
	
	curToken 		tk.Token
	peekToken 		tk.Token

	prefixParseFn 	map[tk.TokenType]prefixParseFn
	infixParseFn  	map[tk.TokenType]infixParseFn
}

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[tk.TokenType]int {
	tk.EQ: 			EQUALS,
	tk.NOTEQ:  		EQUALS,
	tk.LT: 			LESSGREATER,
	tk.GT: 			LESSGREATER,
	tk.PLUS: 		SUM,
	tk.MINUS: 		SUM,
	tk.SLASH: 		PRODUCT,
	tk.ASTERISK: 	PRODUCT,
}

func New(l *lx.Lexer) *Parser {
	p := &Parser{
		l: l,
		errors: []string{},
	}

	p.prefixParseFn = make(map[tk.TokenType]prefixParseFn)
	p.registerPrefix(tk.IDENT, p.parseIdentifier)
	p.registerPrefix(tk.INT, p.parseIntegerLiteral)
	p.registerPrefix(tk.BANG, p.parsePrefixExpression)
	p.registerPrefix(tk.MINUS, p.parsePrefixExpression)

	p.infixParseFn = make(map[tk.TokenType]infixParseFn)
	p.registerInfix(tk.PLUS, p.parseInfixExpression)
	p.registerInfix(tk.MINUS, p.parseInfixExpression)
	p.registerInfix(tk.SLASH, p.parseInfixExpression)
	p.registerInfix(tk.ASTERISK, p.parseInfixExpression)
	p.registerInfix(tk.EQ, p.parseInfixExpression)
	p.registerInfix(tk.NOTEQ, p.parseInfixExpression)
	p.registerInfix(tk.LT, p.parseInfixExpression)
	p.registerInfix(tk.GT, p.parseInfixExpression)
	
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(tok tk.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", tok, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != tk.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) registerPrefix(tokenType tk.TokenType, fn prefixParseFn) {
	p.prefixParseFn[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType tk.TokenType, fn infixParseFn) {
	p.infixParseFn[tokenType] = fn
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case tk.LET:
		return p.parseLetStatement()
	case tk.IDENT:
		if p.peekTokenIs(tk.ASSIGNCHAR) { 
			return p.parseAssignCharStatement()
		} else {
			return p.parseExpressionStatement()
		}
	case tk.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(tk.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(tk.ASSIGN) {
		return nil
	}

	for !p.curTokenIs(tk.SEMICOLON) && !p.curTokenIs(tk.EOF){
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseAssignCharStatement() *ast.AssignStatement {
	nameTok := p.curToken

	if !p.expectPeek(tk.ASSIGNCHAR) {
		return nil
	}

	stmt := &ast.AssignStatement{Token: p.curToken}
	stmt.Name = &ast.Identifier{Token: nameTok, Value: nameTok.Literal}

	p.nextToken()
	
	for !p.curTokenIs(tk.SEMICOLON) && !p.curTokenIs(tk.EOF) {
		p.nextToken()
	}
	
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	for !p.curTokenIs(tk.SEMICOLON) && !p.curTokenIs(tk.EOF) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(tk.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFn[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExpr := prefix()

	for !p.peekTokenIs(tk.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFn[p.peekToken.Type]
		if infix == nil {
			return leftExpr
		}

		p.nextToken()

		leftExpr = infix(leftExpr)
	}

	return leftExpr
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)

		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token: p.curToken,
		Opt: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token: p.curToken,
		Opt: p.curToken.Literal,
		Left: left,
	}
	
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) noPrefixParseFnError(t tk.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) curTokenIs(tok tk.TokenType) bool {
	return p.curToken.Type == tok
}

func (p *Parser) peekTokenIs(tok tk.TokenType) bool {
	return p.peekToken.Type == tok
}

func (p *Parser) expectPeek(tok tk.TokenType) bool {
	if p.peekTokenIs(tok) {
		p.nextToken()
		return true
	} else {
		p.peekError(tok)
		return false
	}
}





