package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/theawakener0/Zod/ast"
	lx "github.com/theawakener0/Zod/lexer"
	tk "github.com/theawakener0/Zod/token"
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
	OR
	AND
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
	INDEX
)

var precedences = map[tk.TokenType]int {
	tk.EQ: 			EQUALS,
	tk.NOTEQ:  		EQUALS,
	tk.LAND: 		AND,
	tk.LOR: 		OR,
	tk.LT: 			LESSGREATER,
	tk.GT: 			LESSGREATER,
	tk.GTEQ: 		LESSGREATER,
	tk.LTEQ: 		LESSGREATER,
	tk.PLUS: 		SUM,
	tk.MINUS: 		SUM,
	tk.SLASH: 		PRODUCT,
	tk.ASTERISK: 	PRODUCT,
	tk.LPAREN:		CALL,
	tk.INC:			CALL,
	tk.DEC:			CALL,
	tk.LBRACKET:	INDEX,
}

func New(l *lx.Lexer) *Parser {
	p := &Parser{
		l: l,
		errors: []string{},
	}

	p.prefixParseFn = make(map[tk.TokenType]prefixParseFn)
	p.registerPrefix(tk.IDENT, p.parseIdentifier)
	p.registerPrefix(tk.INT, p.parseIntegerLiteral)
	p.registerPrefix(tk.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(tk.BANG, p.parsePrefixExpression)
	p.registerPrefix(tk.MINUS, p.parsePrefixExpression)

	p.infixParseFn = make(map[tk.TokenType]infixParseFn)
	p.registerInfix(tk.PLUS, p.parseInfixExpression)
	p.registerInfix(tk.MINUS, p.parseInfixExpression)
	p.registerInfix(tk.SLASH, p.parseInfixExpression)
	p.registerInfix(tk.ASTERISK, p.parseInfixExpression)
	p.registerInfix(tk.EQ, p.parseInfixExpression)
	p.registerInfix(tk.NOTEQ, p.parseInfixExpression)
	p.registerInfix(tk.LAND, p.parseInfixExpression)
	p.registerInfix(tk.LOR, p.parseInfixExpression)
	p.registerInfix(tk.LT, p.parseInfixExpression)
	p.registerInfix(tk.GT, p.parseInfixExpression)
	p.registerInfix(tk.GTEQ, p.parseInfixExpression)
	p.registerInfix(tk.LTEQ, p.parseInfixExpression)
	p.registerInfix(tk.LPAREN, p.parseCallExpression)
	p.registerInfix(tk.INC, p.parsePostfixExpression)
	p.registerInfix(tk.DEC, p.parsePostfixExpression)
	p.registerInfix(tk.LBRACKET, p.parseIndexExpression)

	p.registerPrefix(tk.TRUE, p.parseBoolean)
	p.registerPrefix(tk.FALSE, p.parseBoolean)
	p.registerPrefix(tk.NULL, p.parseNullLiteral)

	p.registerPrefix(tk.LPAREN, p.parseGroupedExpression)
	
	p.registerPrefix(tk.IF, p.parseIfExpression)
	
	p.registerPrefix(tk.FUNCTION, p.parseFunctionLiteral)

	p.registerPrefix(tk.STRING, p.parseStringLiteral)

	p.registerPrefix(tk.FOR, p.parseForExpression)
	p.registerPrefix(tk.LOOP, p.parseLoopExpression)

	p.registerPrefix(tk.INC, p.parsePrefixExpression)
	p.registerPrefix(tk.DEC, p.parsePrefixExpression)

	p.registerPrefix(tk.LBRACKET, p.parseArrayLiteral)

	p.registerPrefix(tk.LBRACE, p.parseHashLiteral)
	
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
	if p.curTokenIs(tk.SEMICOLON) {
		return nil
	}

	switch p.curToken.Type {
	case tk.LET:
		return p.parseLetStatement()
	case tk.IDENT:
		if p.peekTokenIs(tk.ASSIGNCHAR) { 
			return p.parseAssignCharStatement()
		} else if p.peekTokenIs(tk.INCASSIGN) || p.peekTokenIs(tk.DECDASSIGN) ||
			p.peekTokenIs(tk.MLTASSIGN) || p.peekTokenIs(tk.DIVASSIGN) {
			return p.parseCompoundAssignStatement()
		} else if p.peekTokenIs(tk.ASSIGN) {
			return p.parseAssignStatement()
		} else {
			return p.parseExpressionStatement()
		}
	case tk.RETURN:
		return p.parseReturnStatement()
	case tk.BREAK:
		return p.parseBreakStatement()
	case tk.CONTINUE:
		return p.parseContinueStatement()
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

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(tk.SEMICOLON) {
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
	stmt.Left = &ast.Identifier{Token: nameTok, Value: nameTok.Literal}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	
	if p.peekTokenIs(tk.SEMICOLON) {
		p.nextToken()
	}
	
	return stmt
}

func (p *Parser) parseAssignStatement() *ast.AssignStatement {
	nameTok := p.curToken

	if !p.expectPeek(tk.ASSIGN) {
		return nil
	}

	stmt := &ast.AssignStatement{Token: p.curToken}
	stmt.Left = &ast.Identifier{Token: nameTok, Value: nameTok.Literal}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	
	if p.peekTokenIs(tk.SEMICOLON) {
		p.nextToken()
	}
	
	return stmt
}

func (p *Parser) parseCompoundAssignStatement() *ast.AssignStatement {
	nameTok := p.curToken

	p.nextToken()

	stmt := &ast.AssignStatement{Token: p.curToken}
	stmt.Left = &ast.Identifier{Token: nameTok, Value: nameTok.Literal}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(tk.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(tk.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseBreakStatement() ast.Statement {
	stmt := &ast.BreakStatement{Token: p.curToken}

	if p.peekTokenIs(tk.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseContinueStatement() ast.Statement {
	stmt := &ast.ContinueStatement{Token: p.curToken}

	if p.peekTokenIs(tk.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if isAssignOp(p.peekToken.Type) {
		if idx, ok := stmt.Expression.(*ast.IndexExpression); ok {
			return p.parseIndexAssignStatement(idx)
		}
	}

	if p.peekTokenIs(tk.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseIndexAssignStatement(idx *ast.IndexExpression) *ast.AssignStatement {
	p.nextToken()

	stmt := &ast.AssignStatement{Token: p.curToken, Left: idx}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(tk.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func isAssignOp(tok tk.TokenType) bool {
	switch tok {
	case tk.ASSIGN, tk.INCASSIGN, tk.DECDASSIGN, tk.MLTASSIGN, tk.DIVASSIGN:
		return true
	default:
		return false
	}
}

func (p *Parser) skipPeekSemicolons() {
	for p.peekTokenIs(tk.SEMICOLON) && p.peekToken.Literal == "\n" {
		p.nextToken()
	}
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

	literal := p.curToken.Literal
	var value int64
	var err error

	if len(literal) >= 2 && literal[0] == '0' && strings.Contains("xXbBoO", string(literal[1])) {
		value, err = strconv.ParseInt(literal, 0, 64)
	} else {
		value, err = strconv.ParseInt(literal, 10, 64)
	}
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)

		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
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

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.PostfixExpression{
		Token: p.curToken,
		Opt: p.curToken.Literal,
		Left: left,
	}

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

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(tk.TRUE)}
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(tk.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(tk.LPAREN) {
		return nil
	}

	p.nextToken()
	exp.Condition = p.parseExpression(LOWEST)

	p.skipPeekSemicolons()
	if !p.expectPeek(tk.RPAREN) {
		return nil
	}

	p.skipPeekSemicolons()
	if !p.expectPeek(tk.LBRACE) {
		return nil
	}

	exp.Consequence = p.parseBlockStatement()

	p.skipPeekSemicolons()
	if p.peekTokenIs(tk.ELSEIF) {
		p.nextToken()

		if nested, ok := p.parseIfExpression().(*ast.IfExpression); ok {
			exp.ElseIf = nested
		}
	} else if p.peekTokenIs(tk.ELSE) {
		p.nextToken()

		if p.peekTokenIs(tk.IF) {
			p.nextToken()

			if nested, ok := p.parseIfExpression().(*ast.IfExpression); ok {
				exp.ElseIf = nested
			}
		} else {
			p.skipPeekSemicolons()
			if !p.expectPeek(tk.LBRACE) {
				return nil
			}

			exp.Alternative = p.parseBlockStatement()
		}
	}

	return exp
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(tk.RBRACE) && !p.curTokenIs(tk.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseForExpression() ast.Expression {
	exp := &ast.ForExpression{Token: p.curToken}

	if !p.expectPeek(tk.LPAREN) {
		return nil
	}

	p.nextToken()

	if p.curTokenIs(tk.SEMICOLON) {
		p.nextToken()
	} else {
		exp.Init = p.parseStatement()

		if p.curTokenIs(tk.SEMICOLON) {
			p.nextToken()
		} else if p.peekTokenIs(tk.RPAREN) {
			if es, ok := exp.Init.(*ast.ExpressionStatement); ok {
				exp.Condition = es.Expression
				exp.Init = nil
				p.nextToken()
				exp.Body = p.parseBraceBlock()
				if exp.Body == nil {
					return nil
				}
				return exp
			}
			p.errors = append(p.errors, "expected ; or ) in for clause")
			return nil
		} else {
			p.errors = append(p.errors, "expected ; or ) in for clause")
			return nil
		}
	}

	if p.curTokenIs(tk.SEMICOLON) {
		p.nextToken()
	} else {
		exp.Condition = p.parseExpression(LOWEST)

		if p.peekTokenIs(tk.SEMICOLON) {
			p.nextToken()
			p.nextToken()
			if p.curTokenIs(tk.SEMICOLON) {
				p.nextToken()
			}
		} else if p.peekTokenIs(tk.RPAREN) {
			p.nextToken()
			exp.Body = p.parseBraceBlock()
			if exp.Body == nil {
				return nil
			}
			return exp
		} else {
			p.errors = append(p.errors, "expected ; or ) in for condition")
			return nil
		}
	}

	if p.curTokenIs(tk.RPAREN) {
		// empty update
	} else {
		exp.Update = p.parseStatement()
		if !p.expectPeek(tk.RPAREN) {
			return nil
		}
	}

	exp.Body = p.parseBraceBlock()
	if exp.Body == nil {
		return nil
	}
	return exp
}

func (p *Parser) parseLoopExpression() ast.Expression {
	exp := &ast.LoopExpression{Token: p.curToken}
	exp.Body = p.parseBraceBlock()
	if exp.Body == nil {
		return nil
	}
	return exp
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fl := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(tk.LPAREN) {
		return nil
	}

	fl.Parameters = p.parseFunctionParameters()

	p.skipPeekSemicolons()
	if !p.expectPeek(tk.LBRACE) {
		return nil
	}
	
	fl.Body = p.parseBlockStatement()

	return fl
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(tk.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	if p.curToken.Type != tk.IDENT {
		p.errors = append(p.errors, fmt.Sprintf("expected identifier in function parameters, got %s", p.curToken.Type))
	}
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(tk.COMMA) {
		p.nextToken()
		p.nextToken()

		if p.curToken.Type != tk.IDENT {
			p.errors = append(p.errors, fmt.Sprintf("expected identifier in function parameters, got %s", p.curToken.Type))
		}
		nextIdent := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, nextIdent)
	}

	p.skipPeekSemicolons()
	if !p.expectPeek(tk.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(tk.RPAREN)

	return exp
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	array.Elements = p.parseExpressionList(tk.RBRACKET)

	return array
}

func (p *Parser) parseExpressionList(endTok tk.TokenType) []ast.Expression {
	elements := []ast.Expression{}

	if p.peekTokenIs(endTok) {
		p.nextToken()
		return elements
	}
	
	p.nextToken()
	elements = append(elements, p.parseExpression(LOWEST))

	for p.peekTokenIs(tk.COMMA) {
		p.nextToken()
		p.nextToken()
		nextArg := p.parseExpression(LOWEST)
		elements = append(elements, nextArg)
	}

	p.skipPeekSemicolons()
	if !p.expectPeek(endTok) {
		return nil
	}

	return elements
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	p.skipPeekSemicolons()
	if !p.expectPeek(tk.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = []ast.HashLiteralPair{}

	p.skipPeekSemicolons()
	for !p.peekTokenIs(tk.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(tk.COLOMN) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs = append(hash.Pairs, ast.HashLiteralPair{Key: key, Value: value})

		p.skipPeekSemicolons()
		if !p.peekTokenIs(tk.RBRACE) && !p.expectPeek(tk.COMMA) {
			return nil
		}

	}

	if !p.expectPeek(tk.RBRACE) {
		return nil
	}
	
	return hash
}

func (p *Parser) parseBraceBlock() *ast.BlockStatement {
	p.skipPeekSemicolons()
	if !p.expectPeek(tk.LBRACE) {
		return nil
	}
	return p.parseBlockStatement()
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





