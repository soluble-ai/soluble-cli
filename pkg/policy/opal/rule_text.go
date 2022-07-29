package opal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
)

type textRange struct {
	start, end int
}

type ruleText struct {
	path        string
	text        []byte
	packageDecl *textRange
	regoMetaDoc *textRange
	inputType   string
}

func readRuleText(path string) (*ruleText, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	t := &ruleText{
		path: path,
		text: dat,
	}
	return t.parse()
}

func (t *ruleText) parse() (*ruleText, error) {
	parser := ast.NewParser().WithFilename(t.path).WithReader(bytes.NewReader(t.text))
	stmts, _, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	// we only need to look at top-level statements
	for _, stmt := range stmts {
		// fmt.Printf("%T %s\n", stmt, stmt)
		switch s := stmt.(type) {
		case *ast.Package:
			if len(s.Path) < 2 {
				return nil, fmt.Errorf("%s:%d: invalid package declaration", s.Location.File, s.Location.Row)
			}
			if !ast.Var("data").Equal(s.Path[0].Value) || !ast.String("rules").Equal(s.Path[1].Value) {
				return nil, fmt.Errorf("%s:%d: the package for a rule must start with \"rules\"",
					s.Location.File, s.Location.Row)
			}
			last := s.Path[len(s.Path)-1]
			t.packageDecl = &textRange{
				start: s.Location.Offset,
				end:   last.Location.Offset + len(last.Location.Text),
			}
		case ast.Body:
			for _, expr := range s {
				// fmt.Printf("expr %T %s\n", expr, expr)
				if expr.IsAssignment() {
					terms := expr.Terms.([]*ast.Term)
					if len(terms) == 3 {
						switch {
						case ast.Var("__rego__metadoc__").Equal(terms[1].Value):
							t.regoMetaDoc = &textRange{
								start: expr.Location.Offset,
								end:   expr.Location.Offset + len(expr.Location.Text),
							}
						case ast.Var("input_type").Equal(terms[1].Value):
							val, ok := terms[2].Value.(ast.String)
							if ok {
								t.inputType = string(val)
							}
						}
					}
				}
			}
		}
	}
	if t.packageDecl == nil {
		return t, fmt.Errorf("a package declaration in %s is required", t.path)
	}
	return t, nil
}

func (t *ruleText) write(w io.Writer, metadata policy.Metadata) error {
	var (
		tail    int
		newline bool
	)
	switch {
	case t.regoMetaDoc != nil:
		// if we have __rego__metadoc__ then replace it
		if _, err := w.Write(t.text[0:t.regoMetaDoc.start]); err != nil {
			return err
		}
		tail = t.regoMetaDoc.end + 1
	case t.packageDecl != nil:
		// we've only got a package declaration, so we'll write
		// the metadoc just after that
		if _, err := w.Write(t.text[0:t.packageDecl.end]); err != nil {
			return err
		}
		newline = true
		tail = t.packageDecl.end + 1
	default:
		return fmt.Errorf("rule is missing package declaration")
	}
	if _, err := w.Write(toRegoMetaDoc(newline, metadata)); err != nil {
		return err
	}
	_, err := w.Write(t.text[tail:])
	return err
}

func toRegoMetaDoc(newline bool, metadata policy.Metadata) []byte {
	buf := &bytes.Buffer{}
	if newline {
		buf.WriteString("\n\n")
	}
	buf.WriteString("__rego__metadoc__ := {")
	first := true
	keys := []string{}
	for k := range metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
loop:
	for _, k := range keys {
		var regoKey, regoValue string
		switch {
		case k == "sid":
			regoKey = "id"
			regoValue = regoQuote(metadata.GetString(k))
		case k == "title" || k == "description":
			regoKey = k
			regoValue = regoQuote(metadata.GetString(k))
		case k == "severity":
			regoKey = "custom"
			regoValue = fmt.Sprintf(`{ "severity": %s }`, regoQuote(metadata.GetString(k)))
		default:
			continue loop
		}
		if first {
			first = false
		} else {
			buf.WriteRune(',')
		}
		fmt.Fprintf(buf, "\n  \"%s\": %s", regoKey, regoValue)
	}
	buf.WriteString("\n}\n")
	return buf.Bytes()
}

func regoQuote(s string) string {
	if !strings.ContainsAny(s, "\"\n") {
		return fmt.Sprintf(`"%s"`, s)
	}
	b := &strings.Builder{}
	b.WriteRune('"')
	for _, ch := range s {
		switch ch {
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		default:
			b.WriteRune(ch)
		}
	}
	b.WriteRune('"')
	return b.String()
}
