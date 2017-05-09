package frontend

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	TokenRes = map[string]*regexp.Regexp{
		"identifier": regexp.MustCompile(`(?P<value>\A[-+_a-zA-Z][-+_a-zA-Z0-9]*)`),
		"number":     regexp.MustCompile(`(?P<value>\A[0-9]+)`),
		"string":     regexp.MustCompile(`\A"(?P<value>(?:[^"\\]|\\.)*)"`),
		"$(":         regexp.MustCompile(`\A$\(`),
		"`":          regexp.MustCompile("\\A`"),
		"(":          regexp.MustCompile(`\A\(`),
		")":          regexp.MustCompile(`\A\)`),
		"[":          regexp.MustCompile(`\A\[`),
		"]":          regexp.MustCompile(`\A\]`),
		"{":          regexp.MustCompile(`\A{`),
		"}":          regexp.MustCompile(`\A}`),
		",":          regexp.MustCompile(`\A,`),
		":":          regexp.MustCompile(`\A:`),
		"|":          regexp.MustCompile(`\A\|`),
		"@":          regexp.MustCompile(`\A@`),
		"?":          regexp.MustCompile(`\A\?`),
		"=":          regexp.MustCompile(`\A=`),
		"::=":        regexp.MustCompile(`\A::=`),
		"*":          regexp.MustCompile(`\A\*`),
		"%":          regexp.MustCompile(`\A%`),
		"$":          regexp.MustCompile(`\A$`),
		"^":          regexp.MustCompile(`\A^`),
		"newline":    regexp.MustCompile(`\A\n(P<value>\\s*)`),
		"comment":    regexp.MustCompile(`\A#`),
	}
)

// Token class - represents syntax token
type Token struct {
	TokenIds     map[string]bool
	TokenPaterns map[string]string
	TokenRes     map[string]*regexp.Regexp

	Value string // Token string value
	NLine int    // Number of line with token
	NChar int    // Number of char with token
	Id    string // Token id
}

func NewToken(Id string, NLine int, NChar int, Value_opt ...string) *Token { //(string, error){
	Value := "dontcare"
	if len(Value_opt) > 0 {
		Value = Value_opt[0]
	}

	t := new(Token)

	t.TokenIds = map[string]bool{
		"identifier": true, // word
		"number":     true, // usual number
		"string":     true, // "x\"xx" 'x\'xx' `x\'xx`
		"(":          true,
		")":          true,
		"[":          true,
		"]":          true,
		"{":          true,
		"}":          true,
		",":          true, // ,
		":":          true, // :
		"|":          true, // |
		"@":          true, // where or @
		"=":          true, // =
		"::=":        true, // ::=
		"*":          true, // *
		"?":          true, // ?
		"%":          true, // %
		"$":          true, // $
		"$(":         true, // $(
		"`":          true, // `
		"^":          true, // ^
		"newline":    true,
		"endoffile":  true,
	}

	t.Value = Value
	t.NLine = NLine + 1
	t.NChar = NChar + 1
	t.Id = Id
	if !t.TokenIds[Id] {
		fmt.Printf("Wrong token id `%s' given", Id)
	}

	return t
}

func (t *Token) CheckId(Id string) bool {
	if !t.TokenIds[Id] {
		fmt.Printf("Wrong token id `%s' given", Id)
	}
	return t.Id == Id
}

func (t *Token) CheckValue(Value string) bool {
	return t.Value == Value
}

func (t *Token) String() string {
	return fmt.Sprintf("(\"%s\" at (%d:%d) of type '%s')",
		t.Value,
		t.NLine,
		t.NChar,
		t.Id)
}

// Reader error
type ReaderError struct {
	message string
}

func (e *ReaderError) Error() string {
	return e.message
}

func NewReaderError(message string) error {
	return &ReaderError{message}
}

// End of file error handling
type EndOfFile struct {
	s string
}

func (e *EndOfFile) Error() string {
	return e.s
}

func NewEndOfFile(text string) error {
	return &EndOfFile{text}
}

// Reader contain all stuff for reading tokens from FILE
type Reader struct {
	Filename string
	Lines    []string
	Tokens   []*Token
	NLine    int
	NChar    int
	Cur      string
	Indent   string
}

func NewReader(Filename string) *Reader {
	return &Reader{Filename: Filename}
}

func (r *Reader) ParseFile() error {
	file, err := os.Open(r.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		r.Lines = append(r.Lines, scanner.Text())
	}

	if len(r.Lines) == 0 {
		return NewReaderError(fmt.Sprintf("empty file '%s' given", r.Filename))
	}

	r.Cur = r.Lines[0]

	for r.NLine < len(r.Lines) {
		t, err := r.getToken()
		if err != nil {
			if t == nil {
				fmt.Println(err)
				os.Exit(1)
			}
			r.Tokens = append(r.Tokens, t)
			r.Tokens = append(r.Tokens, NewToken("endoffile", r.NLine, r.NChar, ""))
			r.Lines = append(r.Lines, "")
			return nil
		}
		r.Tokens = append(r.Tokens, t)
	}
	return nil
}

// Removing comments and returns leading spaces and some text
func (r *Reader) cleanLine(Line string) (string, string) {
	nspaces := 0
	for _, c := range Line {
		if c == ' ' || c == '\t' {
			nspaces += 1
		} else {
			break
		}
	}
	return Line[:nspaces], Line[nspaces:]
}

// Set cusor to the next line
func (r *Reader) nextLine() error {
	r.NLine += 1

	if r.NLine >= len(r.Lines) {
		return NewEndOfFile(fmt.Sprintf("End of '%s' file", r.Filename))
	}

	spaces, text := r.cleanLine(r.Lines[r.NLine])
	r.NChar = len(spaces)
	r.Cur = strings.TrimSpace(text)
	r.Indent = spaces
	return nil
}

// Matching Token
func (r *Reader) matchToken(Type string) (bool, map[string]string) {
	m := TokenRes[Type].MatchString(r.Cur)
	groups := map[string]string{}

	if m {
		// creating maps with catched
		names := TokenRes[Type].SubexpNames()
		// fmt.Println("------------------------")
		// fmt.Println(Type, TokenRes[Type], TokenRes[Type].FindAllStringSubmatch(r.Cur, -1)[0])
		// fmt.Println("'",r.Cur, "'")
		// fmt.Println(m)
		// fmt.Println("++++++++++++++++++++++++")
		res := TokenRes[Type].FindAllStringSubmatch(r.Cur, -1)[0]
		for i, n := range res {
			groups[names[i]] = n
		}

		mIndexes := TokenRes[Type].FindStringIndex(r.Cur)
		tokenLen := mIndexes[1] - mIndexes[0] // TODO: Potential bad
		r.Cur = r.Cur[tokenLen:]
		r.NChar += tokenLen
	}

	return m, groups
}

// Return next token
func (r *Reader) getToken() (*Token, error) {
	// handling leading spaces
	var spaces string
	spaces, r.Cur = r.cleanLine(r.Cur)
	r.NChar += len(spaces)
	nLine, nChar := r.NLine, r.NChar
	var m bool
	var groups map[string]string

	// here we go!
	//types := []string{"identifier", "number"}
	for _, n := range []string{"identifier", "number"}{//types {
		m, groups = r.matchToken(n)
		if m {
			return NewToken(n, nLine, nChar, groups["value"]), nil
		}
	}

	// Strings
	m, groups = r.matchToken("string")
	if m {
		// replace all special symbols
		// fmt.Println(">>>>",groups["value"])
		s := strings.Replace(groups["value"], "\\\"", "\"", -1)
		s = strings.Replace(s, "\\", "\\", -1)
		s = strings.Replace(s, "\n", "\n", -1)
		// fmt.Println("<<<<",s)
		// TODO Add literally replacing for ", \\, \n
		return NewToken("string", nLine, nChar, s), nil
	}

	types := []string{
		"(", ")", "[", "]", "{", "}", ",", "::=", "=", "*", "@", "|", ":", "?"}
	for _, n := range types {
		m, groups := r.matchToken(n)
		if m {
			return NewToken(n, nLine, nChar, groups[""]), nil
		}
	}

	m, groups = r.matchToken("comment")
	if m || r.Cur == "" || r.Cur == "\n" || r.Cur == "\r\n" {
		err := r.nextLine()
		return NewToken("newline", nLine, nChar, r.Indent), err
	}

	//generate error messages
	r.ReportError("unexpected token")
	return nil, NewReaderError("unexpected token")
}

func (r *Reader) ReportError(message string) {
	panic(
		fmt.Sprintf("%s:%d(%d): error: %s\n%s%s\n%s^\n",
			r.Filename, r.NLine+1, r.NChar+1, message, strings.Repeat(" ", 8),
			r.Lines[r.NLine], strings.Repeat(" ", (8+r.NChar))))
}
