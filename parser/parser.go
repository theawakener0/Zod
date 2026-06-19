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
	return nil
}





