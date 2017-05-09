package frontend

import (
	"fmt"
	"strings"
)

var (
	nodeIds = map[string]bool{
		"grammar":          true,
		"identifier":       true,
		"number":           true,
		"string":           true,
		"newline":          true,
		"literal":          true,
		"pattern":          true,
		"patlist":          true,
		"patatom":          true,
		"consop":           true, // `:' in patterns
		"listop":           true, // `,' in patterns
		"listset":          true,
		"listsetmin":       true,
		"listsetmax":       true,
		"emptylist":        true,
		"expression":       true,
		"exprlist":         true,
		"expr":             true,
		"atom":             true,
		"rule":             true,
		"variants":         true,
		"variant":          true,
		"power":            true,
		"simplevariant":    true,
		"multilinevariant": true,
		"sentence":         true,
		"deflist":          true,
		"definition":       true,
		"condlist":         true,
	}
)

type SyntaxError struct {
	message string
}

func (e *SyntaxError) Error() string {
	return e.message
}

func NewSyntaxError(message string) error {
	return &SyntaxError{message}
}

// Syntax tree node
type Node struct {
	id        string
	token     *Token
	childrens []Node
	match     string
}

func NewNode(id string, token *Token) *Node {
	if !nodeIds[id] {
		fmt.Printf("%s is not valid type of nodes", id)
		return nil
	}
	return &Node{id: id, token: token, childrens: []Node{}, match: token.Value}
}

// Adding child to Node
func (n *Node) AddChild(child Node) {
	n.childrens = append(n.childrens, child)
	// Some optimizations
	if len(n.childrens) == 1 {
		n.match = child.match
	} else {
		n.match += " " + child.match
	}
}

// Checking node's id
func (n *Node) CheckId(id string) bool {
	// TODO: Add some kind of assertion id in Node.nodeIds
	return n.id == id
}

// Get Child with index
func (n *Node) GetChild(i int) Node {
	return n.childrens[i]
}

func (n *Node) Print(sandc ...string) {
	s := ""
	c := ""
	if len(sandc) > 0 {
		s = sandc[0]
		if len(sandc) > 1 {
			c = sandc[1]
		}
	}

	fmt.Printf("%s( %s | %s )\n", s, n.match, n.id)
	cnum := len(n.childrens) + 1
	if cnum > 1 {
		for _, child := range n.childrens[:cnum-2] {
			child.Print(c+"|--", c+"|  ")
		}
		n.childrens[cnum-2].Print(c+"`--", c+"   ")
	}
	//fmt.Printf("\n")
}

// SyntaxTree
type Tree struct {
	filename string
	tokens   []*Token
	lines    []string
	Rules    map[string]*Node
	ncur     int
	cur      *Token
	root     Node
}

func NewTree(reader Reader) *Tree {
	tokens := reader.Tokens
	rules := map[string]*Node{}
	return &Tree{filename: reader.Filename, tokens: tokens,
		lines: reader.Lines, Rules: rules, ncur: 0, cur: tokens[0]}
}

// go to the next token inside tree
// TODO: add error message about ncur < len(tokens)
func (t *Tree) NextToken() { //error{
	t.ncur += 1
	t.cur = t.tokens[t.ncur]
}

func (t *Tree) ReportError(message string) {
	cur := t.cur
	panic(
		fmt.Sprintf("%s:%d(%d): on token `%s' error: %s\n%s%s\n%s^",
			t.filename, cur.NLine, cur.NChar, cur.Value, message,
			strings.Repeat(" ", 8), strings.TrimSpace(t.lines[cur.NLine-1]),
			strings.Repeat(" ", (8*(cur.NChar-1)))))
}
func (t *Tree) NewNode(Id string) *Node {
	return NewNode(Id, t.cur)
}

func (t *Tree) CheckTokenId(Id string) bool {
	return t.cur.CheckId(Id)
}

func (t *Tree) CheckTokenIdsOr(Ids ...string) bool {
	for _, id := range Ids {
		if t.CheckTokenId(id) {
			return true
		}
	}
	return false
}

func (t *Tree) CheckTokenValue(Value string) bool {
	return t.cur.CheckValue(Value)
}

func (t *Tree) CheckTokenIDValue(Id string, Value string) bool {
	return (t.cur.CheckId(Id) && t.cur.CheckValue(Value))
}

func (t *Tree) PrintGrammar() {
	t.root.Print()
}

// Parse functions block
func (t *Tree) ParseGrammar() {
	t.root = *t.NewNode("grammar")
	t.SkipNewlines()
	for !t.CheckTokenId("endoffile") {
		t.root.AddChild(t.ParseRule())
	}
}

func (t *Tree) ParseIdentifier() Node {
	if !t.CheckTokenId("identifier") {
		t.ReportError("identifier expected")
	}
	node := *t.NewNode("identifier")
	t.NextToken()
	return node
}

func (t *Tree) ParseNumber() Node {
	if !t.CheckTokenId("number") {
		t.ReportError("number expected")
	}
	node := *t.NewNode("number")
	t.NextToken()
	return node
}

func (t *Tree) ParseString() Node {
	if !t.CheckTokenId("string") {
		t.ReportError("string expected")
	}
	node := *t.NewNode("string")
	t.NextToken()
	return node
}

func (t *Tree) SkipNewlines() {
	for t.CheckTokenId("newline") {
		t.NextToken()
	}
}

func (t *Tree) ParseLiteral() Node {
	node := *t.NewNode("literal")
	if t.CheckTokenId("string") {
		node.AddChild(t.ParseString())
	} else if t.CheckTokenId("number") {
		node.AddChild(t.ParseNumber())
	} else {
		t.ReportError("string or number expected")
	}
	return node
}

func (t *Tree) ParsePaterns() Node {
	node := *t.NewNode("pattern")
	for t.CheckTokenIdsOr("identifier", "string", "number", "(") {
		node.AddChild(t.ParsePatList())
	}
	return node
}

func (t *Tree) ParsePatList() Node {
	// forming nodes in prefix form:
	//	: case: [CONSOP PATATOM PATLIST]
	//	, case: [ listop patatom ... patatom]
	node := *t.NewNode("patlist")
	patatom := t.ParsePatatom()
	if t.CheckTokenId(":") {
		node.AddChild(*t.NewNode("consop"))
		t.NextToken()
		node.AddChild(patatom)
		node.AddChild(t.ParsePatList())
	} else if t.CheckTokenId(",") {
		node.AddChild(*t.NewNode("listop"))
		node.AddChild(patatom)
		for t.CheckTokenId(",") {
			t.NextToken()
			if t.CheckTokenIdsOr("identifier", "string", "number", "(") {
				node.AddChild(t.ParsePatatom())
			}
		}
	} else {
		node.AddChild(patatom)
	}
	return node
}

func (t *Tree) ParsePatatom() Node {
	node := *t.NewNode("patatom")
	if t.CheckTokenId("identifier") {
		node.AddChild(t.ParseIdentifier())
	} else if t.CheckTokenIdsOr("string", "number") {
		node.AddChild(t.ParseLiteral())
	} else if t.CheckTokenId("(") {
		t.NextToken()
		if t.CheckTokenId(")") {
			// empty list case
			node.AddChild(*t.NewNode("emptylist"))
		} else {
			node.AddChild(t.ParsePatList())
			if !t.CheckTokenId(")") {
				t.ReportError(") expected")
			}
		}
		t.NextToken()
	} else {
		t.ReportError("identifier, literal or ( expected")
	}
	return node
}

func (t *Tree) ParseExpression() Node {
	// forming prefix form of expression
	node := *t.NewNode("expression")
	expression := t.ParseExpr()
	if t.CheckTokenId(",") {
		// list operator
		node.AddChild(*t.NewNode("listop"))
		node.AddChild(expression)
		for t.CheckTokenId(",") {
			t.NextToken()
			if t.CheckTokenIdsOr("identifier", "number", "string", "(", "[") {
				node.AddChild(t.ParseExpr())
			}
		}
	} else if t.CheckTokenId(":") {
		// cons operator
		node.AddChild(*t.NewNode("consop"))
		node.AddChild(expression)
		t.NextToken()
		node.AddChild(t.ParseExpression())
	} else {
		node.AddChild(expression)
	}
	return node
}

func (t *Tree) ParseExpr() Node {
	node := *t.NewNode("expr")
	if t.CheckTokenIdsOr("identifier", "number", "string", "(", "[") {
		node.AddChild(t.ParseAtom())
	} else {
		t.ReportError("identifier, number^ string, [ or ( expected")
	}
	for t.CheckTokenIdsOr("identifier", "number", "string", "(", "[") {
		node.AddChild(t.ParseAtom())
	}
	return node
}

func (t *Tree) ParseAtom() Node {
	node := *t.NewNode("atom")
	if t.CheckTokenId("identifier") {
		node.AddChild(t.ParseIdentifier())
	} else if t.CheckTokenIdsOr("number", "string") {
		node.AddChild(t.ParseLiteral())
	} else if t.CheckTokenId("(") {
		t.NextToken()
		if t.CheckTokenId(")") {
			node.AddChild(*t.NewNode("emptylist"))
		} else {
			node.AddChild(t.ParseExpression())
			if !t.CheckTokenId(")") {
				t.ReportError(") expected")
			}
		}
		// skipping )
		t.NextToken()
	} else if t.CheckTokenId("[") {
		t.ReportError("please, use parentheses instead of brackets, brackets have no actual mean now and disabled.")
	} else {
		t.ReportError("pidentifier, string, number, `(' or `[' expected")
	}
	return node
}

func (t *Tree) ParseRule() Node {
	node := *t.NewNode("rule")
	node.AddChild(t.ParseIdentifier())
	node.AddChild(t.ParsePaterns())

	t.SkipNewlines()
	if t.CheckTokenId("@") {
		t.NextToken()
		node.AddChild(t.ParseDefList())
	}
	t.SkipNewlines()
	if t.CheckTokenId("?") {
		t.NextToken()
		node.AddChild(t.ParseCondlist())
	}
	t.SkipNewlines()
	if t.CheckTokenId("::=") {
		t.NextToken()
		t.SkipNewlines()
		node.AddChild(t.ParseVariants())
	} else {
		t.ReportError("::= expected")
	}
	// record this rule to the rules storage (dictionary)
	name := node.childrens[0].token.Value
	// TODO: More optimal method
	if _, ok := t.Rules[name]; !ok {
		t.Rules[name] = &node
	}
	return node
}

func (t *Tree) ParseVariants() Node {
	node := *t.NewNode("variants")
	for t.CheckTokenId("newline") {
		t.NextToken()
	}
	node.AddChild(t.ParseVariant())
	for t.CheckTokenIdsOr("newline", "|") {
		t.SkipNewlines()
		if t.CheckTokenId("|") {
			t.NextToken()
			t.SkipNewlines()
			node.AddChild(t.ParseVariant())
		}
	}
	return node
}

func (t *Tree) ParseVariant() Node {
	node := *t.NewNode("variant")
	// some "power"
	if t.CheckTokenId("*") {
		t.NextToken()
		if t.CheckTokenId("number") {
			node.AddChild(*t.NewNode("power"))
			t.NextToken()
		} else {
			t.ReportError("number expected")
		}
	}
	if t.CheckTokenId("{") {
		node.AddChild(t.ParseMultilineVariant())
	} else {
		node.AddChild(t.ParseSimpleVariant())
	}
	return node
}

func (t *Tree) ParseSimpleVariant() Node {
	node := *t.NewNode("simplevariant")
	node.AddChild(t.ParseSentence())
	return node
}

func (t *Tree) ParseMultilineVariant() Node {
	node := *t.NewNode("multilinevariant")
	if t.CheckTokenId("{") {
		t.NextToken()
		for t.CheckTokenIdsOr("identifier", "string", "number", "(", "newline") {
			if t.CheckTokenId("newline") {
				node.AddChild(*t.NewNode("newline"))
				t.NextToken()
			} else {
				node.AddChild(t.ParseSentence())
			}
		}
		if t.CheckTokenId("}") {
			t.NextToken()
		} else {
			t.ReportError("} expected")
		}
	} else {
		t.ReportError("{number} expected")
	}
	return node
}

func (t *Tree) ParseSentence() Node {
	node := *t.NewNode("sentence")
	node.AddChild(t.ParseExpression())
	return node
}

func (t *Tree) ParseDefList() Node {
	node := *t.NewNode("deflist")
	t.SkipNewlines()
	node.AddChild(t.ParseDefinition())
	for t.CheckTokenId(",") {
		t.NextToken()
		t.SkipNewlines()
		node.AddChild(t.ParseDefinition())
	}
	return node
}

func (t *Tree) GetRoot() Node {
	return t.root
}

func (t *Tree) ParseDefinition() Node {
	node := *t.NewNode("definition")
	node.AddChild(t.ParsePatatom())
	if t.CheckTokenId("=") {
		t.NextToken()
		node.AddChild(t.ParseExpr())
	} else {
		t.ReportError("expected in definition")
	}
	return node
}

func (t *Tree) ParseCondlist() Node {
	node := *t.NewNode("condlist")
	node.AddChild(t.ParseExpr())
	if t.CheckTokenId(",") {
		t.NextToken()
		t.SkipNewlines()
		node.AddChild(t.ParseExpr())
	}
	return node
}

// Just for test purposes
// func main() {
// 	var dict = map[string]string {
// 		"one" : "value one",
// 		"two" : "value two",
// 	}

// 	reader := frontend.NewReader("./grammar.g")
// 	reader.ParseFile()
// 	for i, n := range reader.Tokens{
// 		fmt.Println(i, n)
// 	}

// 	tree := NewTree(*reader)
// 	tree.ParseGrammar()
// 	tree.PrintGrammar()
// 	for key, value := range tree.Rules {
//     fmt.Println("Key:", key, "Value:", value)
// }

// }
