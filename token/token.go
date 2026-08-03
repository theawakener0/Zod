package token


type TokenType string

type Token struct {
	Type	TokenType
	Literal	string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF = "EOF"

	IDENT = "IDENT"
	INT = "INT"
	STRING = "STRING"

	ASSIGN = "="
	PLUS = "+"
	MINUS = "-"
	BANG = "!"
	ASTERISK = "*"
	SLASH = "/"

	LT = "<"
	GT = ">"

	EQ = "=="
	NOTEQ = "!="
	LTEQ = "<="
	GTEQ = ">="
	INCASSIGN = "+="
	DECDASSIGN = "-="
	MLTASSIGN = "*="
	DIVASSIGN = "/="
	LAND = "&&"
	LOR = "||"
	INC = "++"
	DEC = "--"
	ASSIGNCHAR = ":="

	COMMA = ","
	COLOMN = ":"
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"
	LBRACKET = "["
	RBRACKET = "]"

	FUNCTION = "FUNCTION"
	LET = "LET"
	TRUE = "TRUE"
	FALSE = "FALSE"
	IF = "IF"
	ELSEIF = "ELSEIF"
	ELSE = "ELSE"
	RETURN = "RETURN"
	FOR = "FOR"
	LOOP = "LOOP"
	BREAK = "BREAK"
	CONTINUE = "CONTINUE"
	NULL = "NULL"
)

var keywords = map[string]TokenType {
	"fn" : FUNCTION,
	"let": LET,
	"true": TRUE,
	"false": FALSE,
	"if": IF,
	"elseif": ELSEIF,
	"else": ELSE,
	"return": RETURN,
	"for": FOR,
	"loop": LOOP,
	"break": BREAK,
	"continue": CONTINUE,
	"null": NULL,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

