package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/theawakener0/Zod/ast"
)

type ObjectType string

const (
	INTEGER_OBJ = "INTEGER"
	FLOAT_OBJ = "FLOAT"
	BOOLEAN_OBJ = "BOOLEAN"
	NULL_OBJ = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ = "ERROR"
	BREAK_OBJ = "BREAK"
	CONTINUE_OBJ = "CONTINUE"
	FUNCTION_OBJ = "FUNCTION"
	STRING_OBJ = "STRING"
	BUILTIN_OBJ = "BUILTIN"
	ARRAY_OBJ = "ARRAY"
	HASH_OBJ = "HASH"
	MATRIX_OBJ = "MATRIX"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type HashKey struct {
	Type ObjectType
	Value uint64
}

type Hashable interface {
	HashKey() HashKey
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}
func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType {
	return FLOAT_OBJ
}
func (f *Float) Inspect() string {
	s := strconv.FormatFloat(f.Value, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}
func (f *Float) HashKey() HashKey {
	if f.Value == math.Trunc(f.Value) {
		iv := int64(f.Value)
		if float64(iv) == f.Value {
			return HashKey{Type: INTEGER_OBJ, Value: uint64(iv)}
		}
	}
	return HashKey{Type: f.Type(), Value: math.Float64bits(f.Value)}
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType {
	return BOOLEAN_OBJ
}
func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
}
func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

type Null struct {}

func (n *Null) Type() ObjectType {
	return NULL_OBJ
}
func (n *Null) Inspect() string {
	return "null"
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType {
	return RETURN_VALUE_OBJ
}
func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType {
	return ERROR_OBJ
}

func (e *Error) Inspect() string {
	return fmt.Sprintf("Error: %s", e.Message)
}

type Break struct{}

func (b *Break) Type() ObjectType {
	return BREAK_OBJ
}
func (b *Break) Inspect() string {
	return "break"
}

type Continue struct{}

func (c *Continue) Type() ObjectType {
	return CONTINUE_OBJ
}
func (c *Continue) Inspect() string {
	return "continue"
}

type Function struct {
	Parameters 	[]*ast.Identifier
	Body 		*ast.BlockStatement
	Env 		*Enviroment
}

func (f *Function) Type() ObjectType {
	return FUNCTION_OBJ
}

func (f *Function) Inspect() string {
	var out bytes.Buffer
	
	params := make([]string, 0, len(f.Parameters))
	for _, param := range f.Parameters {
		params = append(params, param.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType {
	return STRING_OBJ
}
func (s *String) Inspect() string {
	return s.Value
}
func (s *String) HashKey() HashKey {
	hash := fnv.New64a()
	hash.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: hash.Sum64()}
}

type BuiltinFn func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFn
}

func (b *Builtin) Type() ObjectType {
	return BUILTIN_OBJ
}
func (b *Builtin) Inspect() string {
	return "builtin function"
}

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType {
	return ARRAY_OBJ
}
func (a *Array) Inspect() string {
	var out bytes.Buffer

	elements := make([]string, 0, len(a.Elements))
	for _, element := range a.Elements {
		elements = append(elements, element.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type Matrix struct {
	Rows int
	Cols int
	Data [][]Object
}

func (m *Matrix) Type() ObjectType {
	return MATRIX_OBJ
}
func (m *Matrix) Inspect() string {
	var out bytes.Buffer

	rows := make([]string, 0, m.Rows)
	for _, row := range m.Data {
		elements := make([]string, 0, m.Cols)
		for _, element := range row {
			elements = append(elements, element.Inspect())
		}
		rows = append(rows, "["+strings.Join(elements, ", ")+"]")
	}

	out.WriteString("[")
	out.WriteString(strings.Join(rows, ", "))
	out.WriteString("]")

	return out.String()
}

type HashPair struct {
	Key		Object
	Value	Object
}

type Hash struct {
	Pairs	map[HashKey]HashPair
	Order	[]HashKey
}

func (h *Hash) Type() ObjectType {
	return HASH_OBJ
}
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := make([]string, 0, len(h.Pairs))
	if len(h.Order) == 0 && len(h.Pairs) > 0 {
		// Fallback for hashes built without Order 
		// Sort by rendered key for deterministic output.
		rendered := make([]string, 0, len(h.Pairs))
		byRendered := make(map[string]HashPair, len(h.Pairs))
		for _, pair := range h.Pairs {
			r := pair.Key.Inspect()
			if _, exists := byRendered[r]; !exists {
				rendered = append(rendered, r)
				byRendered[r] = pair
			}
		}
		sort.Strings(rendered)
		for _, r := range rendered {
			pair := byRendered[r]
			pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
		}
	} else {
		for _, key := range h.Order {
			pair := h.Pairs[key]
			pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
		}
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

