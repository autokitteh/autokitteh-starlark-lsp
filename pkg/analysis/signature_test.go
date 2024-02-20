package analysis

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

const fooDoc = `
def foo(a, b, c, d):
  pass

foo%s
`

func TestSignatureHelp(t *testing.T) {
	tt := []struct {
		args   string
		active uint32
	}{
		{args: "(", active: 0},
		{args: "()", active: 0},
		{args: "(1,", active: 1},
		{args: "(1,)", active: 1},
		{args: "(1, 2", active: 1},
		{args: "(1, 2)", active: 1},
		{args: "(1, 2,", active: 2},
		{args: "(1, 2,)", active: 2},
		{args: "(b)", active: 0},
		{args: "(b=", active: 1},
		{args: "(b=)", active: 1},
		{args: "(1, d=", active: 3},
		{args: "(1, d=)", active: 3},
		{args: "(1, d=,)", active: 0},
		{args: "(1, d=True, c=)", active: 2},
	}

	for i, test := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			f := newFixture(t)
			doc := f.MainDoc(fmt.Sprintf(fooDoc, test.args))
			ch := uint32(3 + len(test.args))
			if strings.HasSuffix(test.args, ")") {
				ch -= 1
			}
			pos := protocol.Position{Line: 4, Character: ch}
			help := f.a.SignatureHelp(doc, pos)
			assert.NotNil(t, help)
			if help == nil {
				return
			}
			assert.Equal(t, 1, len(help.Signatures))
			assert.Equal(t, "(a, b, c, d)", help.Signatures[0].Label)
			assert.Equal(t, test.active, help.ActiveParameter)
		})
	}
}

func TestSignatureHelpExtended(t *testing.T) {
	f := newFixture(t)
	f.builtinSymbols()
	f.ParseBuiltins(funcsAndObjectsFixture)

	tests := []struct {
		tData
		expected string
	}{
		// single level
		{tData{doc: "get_s()"}, "(s, ss, sss) -> String"}, // ret is converted str->String
		{tData{doc: "get_d()"}, "(dd: dict) -> Dict"},
		{tData{doc: "get_l()"}, "(lll: list|None) -> List"}, // ret is converted list->List

		// nested chained
		{tData{doc: "get_d().pop()"}, "(key)"},
		{tData{doc: "get_d().keys().append()"}, "(x) -> none"}, // nested calls
		{tData{doc: "get_c1().mm.upper()"}, "() -> String"},    // nested mix - calls and  attributes
		{tData{doc: "get_c2().mm.append()"}, "(x) -> none"},    // ret is converted?, None->none
		{tData{doc: "get_c1().cc.mmm.keys()"}, "() -> List"},
		{tData{doc: "get_c2().cc.mmm.upper()"}, "() -> String"},

		// nested via assignment
		{tData{doc: "s = get_s()\ns.upper()"}, "() -> String"},
		{tData{doc: "s1 = get_s()\ns=s1\ns.upper()"}, "() -> String"},
		{tData{doc: "s1 = get_c1().mm\ns=s1\ns.upper()"}, "() -> String"},
		{tData{doc: "s = get_c2()\ns.mm.append()"}, "(x) -> none"},
		{tData{doc: "r1 = get_c1()\nr=r1.cc\nr.mmm.keys()"}, "() -> List"},
		{tData{doc: "r2 = get_c2().cc\nr1=r2.mmm\nr=r1\nr.upper()"}, "() -> String"},

		{tData{doc: "r1=get_c1()\nr=r1.getT()\nr.upper()"}, "() -> String"},
		{tData{doc: "r1=get_c2().getC()\nr=r1.mmm.keys()\nr.append()"}, "(x) -> none"},

		// negative tests
		{tData{doc: "get_x()"}, ""},
		{tData{doc: "get_d().upper()"}, ""},
		{tData{doc: "get_s().keys()"}, ""},
		{tData{doc: "get_d().keys().upper()"}, ""},
		{tData{doc: "get_c1().mm.append()"}, ""},
		{tData{doc: "get_c1().cc.mmm.upper()"}, ""},
		{tData{doc: "r = get_d()\nr.upper()"}, ""},
		{tData{doc: "r = get_d().keys()\nr.upper()"}, ""},
		{tData{doc: `"".append()`}, ""},
		{tData{doc: `[].keys()`}, ""},
		{tData{doc: `{}.upper()`}, ""},

		// dict, list, string
		{tData{doc: `"aa".startswith()`}, "(prefix) -> bool"},
		{tData{doc: `[].append()`}, "(x) -> none"},
		{tData{doc: `{}.keys()`}, "() -> List"},
		{tData{doc: "r={}\nr.keys()"}, "() -> List"},
		{tData{doc: "r={}\nr.keys().append()"}, "(x) -> none"},
		{tData{doc: "r1={}\nr=r1\nr1.keys().append()"}, "(x) -> none"},
		{tData{doc: "r={1,2}.keys()\nr.append()"}, "(x) -> none"},
	}

	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			doc := f.MainDoc(tt.doc)
			pos := testPosition(tt.tData)
			pos.Character -= 1 // inside the brackets
			help := f.a.SignatureHelp(doc, pos)

			if len(tt.expected) > 0 {
				assert.NotNil(t, help)
				if help == nil {
					return
				}
			} else {
				assert.Nil(t, help)
				return
			}
			assert.Equal(t, 1, len(help.Signatures))
			assert.Equal(t, tt.expected, help.Signatures[0].Label)
			assert.Equal(t, uint32(0), help.ActiveParameter)
		})
	}
}
