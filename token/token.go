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

	FUNCTION = "FUNCTION"
	LET = "LET"
	TRUE = "TRUE"
	FALSE = "FALSE"
	IF = "IF"
	ELSE = "ELSE"
	RETURN = "RETURN"
	FOR = "FOR"
	WHILE = "WHILE"
	LOOP = "LOOP"
)

var keywords = map[string]TokenType {
	"fn" : FUNCTION,
	"let": LET,
	"true": TRUE,
	"false": FALSE,
	"if": IF,
	"else": ELSE,
	"return": RETURN,
	"for": FOR,
	"while": WHILE,
	"loop": LOOP,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

