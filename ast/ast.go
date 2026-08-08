package ast

import (
	"bytes"
	"fmt"
	"strings"

	tk "github.com/theawakener0/Zod/token"
)

type Node interface {
	TokenLiteral() 	string
	String()		string
}

type Statement interface {
	Node 
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type LetStatement struct {
	Token 	tk.Token
	Name	*Identifier
	Value 	Expression
}

func (ls *LetStatement) statementNode() {}
func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")

	return out.String()
}

type AssignStatement struct {
	Token 	tk.Token
	Left	Expression
	Value 	Expression
}

func (as *AssignStatement) statementNode() {}
func (as *AssignStatement) TokenLiteral() string {
	return as.Token.Literal
}
func (as *AssignStatement) String() string {
	var out bytes.Buffer

	out.WriteString(as.Left.String())
	out.WriteString(" " + as.TokenLiteral() + " ")

	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	out.WriteString(";")

	return out.String()
}

type ReturnStatement struct {
	Token 			tk.Token
	ReturnValue 	Expression
}

func (rs *ReturnStatement) statementNode() {}
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")
	
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")

	return out.String()
}

type ExpressionStatement struct {
	Token 		tk.Token
	Expression 	Expression
}

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BreakStatement struct {
	Token tk.Token
}

func (bs *BreakStatement) statementNode() {}
func (bs *BreakStatement) TokenLiteral() string {
	return bs.Token.Literal
}
func (bs *BreakStatement) String() string {
	return "break;"
}

type ContinueStatement struct {
	Token tk.Token
}

func (cs *ContinueStatement) statementNode() {}
func (cs *ContinueStatement) TokenLiteral() string {
	return cs.Token.Literal
}
func (cs *ContinueStatement) String() string {
	return "continue;"
}

type IntegerLiteral struct {
	Token tk.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
}
func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

type FloatLiteral struct {
	Token tk.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode() {}
func (fl *FloatLiteral) TokenLiteral() string {
	return fl.Token.Literal
}
func (fl *FloatLiteral) String() string {
	return fl.Token.Literal
}

type PrefixExpression struct {
	Token 	tk.Token
	Opt 	string
	Right 	Expression
}

func (pe *PrefixExpression) expressionNode() {}
func (pe *PrefixExpression) TokenLiteral() string {
	return pe.Token.Literal
}
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Opt)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token 	tk.Token
	Left 	Expression
	Opt 	string
	Right 	Expression
}

func (ie *InfixExpression) expressionNode() {}
func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Opt + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type Boolean struct {
	Token 	tk.Token
	Value 	bool
}

func (b *Boolean) expressionNode() {}
func (b *Boolean) TokenLiteral() string {
	return b.Token.Literal
}
func (b *Boolean) String() string {
	return b.TokenLiteral()
}

type NullLiteral struct {
	Token tk.Token
}

func (nl *NullLiteral) expressionNode() {}
func (nl *NullLiteral) TokenLiteral() string {
	return nl.Token.Literal
}
func (nl *NullLiteral) String() string {
	return nl.TokenLiteral()
}

type BlockStatement struct {
	Token 	tk.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}
func (bs *BlockStatement) TokenLiteral() string {
	return bs.Token.Literal
}
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type IfExpression struct {
	Token 			tk.Token
	Condition 		Expression
	Consequence 	*BlockStatement
	ElseIf 			*IfExpression
	Alternative 	*BlockStatement
}

func (ie *IfExpression) expressionNode() {}
func (ie *IfExpression) TokenLiteral() string {
	return ie.Token.Literal
}
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.ElseIf != nil {
		out.WriteString("else ")
		out.WriteString(ie.ElseIf.String())
	}

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

type FunctionLiteral struct {
	Token		tk.Token
	Parameters	[]*Identifier
	Body		*BlockStatement
}

func (fl *FunctionLiteral) expressionNode() {}
func (fl *FunctionLiteral) TokenLiteral() string {
	return fl.Token.Literal
}
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := make([]string, 0, len(fl.Parameters))
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

type CallExpression struct {
	Token		tk.Token
	Function	Expression
	Arguments	[]Expression
}

func (ce *CallExpression) expressionNode() {}

func (ce *CallExpression) TokenLiteral() string {
	return ce.Token.Literal
}
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := make([]string, 0, len(ce.Arguments))
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token	tk.Token
	Value	string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) TokenLiteral() string {
	return sl.Token.Literal
}
func (sl *StringLiteral) String() string {
	return sl.Token.Literal
}

type ForExpression struct {
	Token     tk.Token
	Init      Statement
	Condition Expression
	Update    Statement
	Body      *BlockStatement
}

func (fe *ForExpression) expressionNode() {}
func (fe *ForExpression) TokenLiteral() string {
	return fe.Token.Literal
}
func (fe *ForExpression) String() string {
	var out bytes.Buffer

	out.WriteString("for")
	out.WriteString("(")
	if fe.Init != nil {
		out.WriteString(strings.TrimSuffix(fe.Init.String(), ";"))
	}
	out.WriteString(";")
	if fe.Condition != nil {
		out.WriteString(fe.Condition.String())
	}
	out.WriteString(";")
	if fe.Update != nil {
		out.WriteString(strings.TrimSuffix(fe.Update.String(), ";"))
	}
	out.WriteString(")")
	out.WriteString(fe.Body.String())

	return out.String()
}

type PostfixExpression struct {
	Token tk.Token
	Opt   string
	Left  Expression
}

func (pe *PostfixExpression) expressionNode() {}
func (pe *PostfixExpression) TokenLiteral() string {
	return pe.Token.Literal
}
func (pe *PostfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Left.String())
	out.WriteString(pe.Opt)
	out.WriteString(")")

	return out.String()
}

type LoopExpression struct {
	Token tk.Token
	Body  *BlockStatement
}

func (le *LoopExpression) expressionNode() {}
func (le *LoopExpression) TokenLiteral() string {
	return le.Token.Literal
}
func (le *LoopExpression) String() string {
	var out bytes.Buffer

	out.WriteString("loop")
	out.WriteString(le.Body.String())

	return out.String()
}

type ArrayLiteral struct {
	Token tk.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode() {}
func (al *ArrayLiteral) TokenLiteral() string {
	return al.Token.Literal
}
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := make([]string, 0, len(al.Elements))
	for _, e := range al.Elements {
		elements = append(elements, e.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type IndexExpression struct {
	Token 	tk.Token
	Left	Expression
	Index 	Expression
}

func (ie *IndexExpression) expressionNode() {}
func (ie *IndexExpression) TokenLiteral() string {
	return ie.Token.Literal
}
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

type HashLiteralPair struct {
	Key		Expression
	Value	Expression
}

type HashLiteral struct {
	Token	tk.Token
	Pairs	[]HashLiteralPair
}

func (hl *HashLiteral) expressionNode() {}
func (hl *HashLiteral) TokenLiteral() string {
	return hl.Token.Literal
}
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := make([]string, 0, len(hl.Pairs))
	for _, p := range hl.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s : %s", p.Key, p.Value))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

type Identifier struct {
	Token tk.Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}
func (i *Identifier) String() string {
	return i.Value
}


