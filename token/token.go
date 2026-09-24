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
	FLOAT = "FLOAT"
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
	DOT = "."

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

	IMPORT = "IMPORT"
	FROM = "FROM"
	PUB = "PUB"
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
	"import": IMPORT,
	"from": FROM,
	"pub": PUB,
}

func LookupIdent(ident string) TokenType {
	// Fast-path switch for common keywords avoids a map hash;
	// map fallback preserves behavior for the full keyword set.
	switch ident {
	case "fn":
		return FUNCTION
	case "let":
		return LET
	case "true":
		return TRUE
	case "false":
		return FALSE
	case "if":
		return IF
	case "else":
		return ELSE
	case "return":
		return RETURN
	case "for":
		return FOR
	}
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

