package analysis

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

const fileInclude = "autokitteh-starlark.include"

func TestBasicDefinition(t *testing.T) {
	for _, tc := range []struct {
		name string
		line uint32
		char uint32
		// leave this empty to indicate the case expects 0 definition locations
		expectedFile  string
		expectedRange protocol.Range
	}{
		{"func definition", 1, 5, "", protocol.Range{}},
		{"func call", 7, 1, fixtureFileName, protocol.Range{Start: protocol.Position{
			Line:      1,
			Character: 0,
		}, End: protocol.Position{
			Line:      3,
			Character: 7,
		}}},
		{"var definition", 5, 0, "", protocol.Range{}},
		{"var reference", 7, 4, fixtureFileName, protocol.Range{Start: protocol.Position{
			Line:      5,
			Character: 0,
		}, End: protocol.Position{
			Line:      5,
			Character: 5,
		}}},
		{"unknown var", 8, 2, "", protocol.Range{}},
		{"out-of-scope var", 9, 6, "", protocol.Range{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)

			doc := f.MainDoc(`
def foo():
  print('hello')
  y = 5

x = 3

foo(x)
asdf
print(y)
`)

			result := f.a.Definition(f.ctx, doc, protocol.Position{Character: tc.char, Line: tc.line})
			if tc.expectedFile == "" {
				require.Len(t, result, 0)
			} else {
				require.Len(t, result, 1)
				l := result[0]
				require.Equal(t, protocol.Location{
					URI:   uri.File(tc.expectedFile),
					Range: tc.expectedRange,
				}, l)
			}
		})
	}
}

func TestCrossFileDefinition(t *testing.T) {
	f := newFixture(t)

	f.Document(fileInclude, `
print('hi')

def foo():
  print('foo')
`)

	doc := f.MainDoc(`
load('` + fileInclude + `', 'foo')

foo()
`)

	result := f.a.Definition(f.ctx, doc, protocol.Position{Line: 3, Character: 1})
	require.Len(t, result, 1)
	require.Equal(t, protocol.Location{
		URI: uri.File(fileInclude),
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      3,
				Character: 0,
			},
			End: protocol.Position{
				Line:      4,
				Character: 14,
			},
		},
	}, result[0])
}

func TestBuiltinDefinition(t *testing.T) {
	f := newFixture(t)
	f.AddSymbol("k8s_resource", "foo")
	doc := f.MainDoc("k8s_resource()")
	result := f.a.Definition(f.ctx, doc, protocol.Position{Character: 3})
	require.Len(t, result, 0)
}
