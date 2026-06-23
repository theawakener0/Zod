package parser

import (
	"github.com/theawakener0/zod/ast"
	lx "github.com/theawakener0/zod/lexer"
	tk "github.com/theawakener0/zod/token"
)

type Parser struct {
	l 			*lx.Lexer
	
	curToken 	tk.Token
	peekToken 	tk.Token
}

func New(l *lx.Lexer) *Parser {
	p := &Parser{l: l}
	
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
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

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case tk.LET:
		return p.parseLetStatement()
	case tk.IDENT:
		if p.peekTokenIs(tk.ASSIGNCHAR) { 
			return p.parseAssignCharStatement()
		} else {
			// it returns nil for now.
			return nil
		}
	default:
		return nil
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
		return false
	}
}





