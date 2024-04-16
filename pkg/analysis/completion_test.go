package analysis

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"

	"github.com/autokitteh/starlark-lsp/pkg/query"
)

func (f *fixture) builtinSymbols() {
	_ = WithStarlarkBuiltins()(f.a)

	allStringFuncs = []string{}
	for _, method := range f.a.builtins.Types["String"].Methods {
		allStringFuncs = append(allStringFuncs, method.Name)
	}

	allDictFuncs = []string{}
	for _, method := range f.a.builtins.Types["Dict"].Methods {
		allDictFuncs = append(allDictFuncs, method.Name)
	}

	allListFuncs = []string{}
	for _, method := range f.a.builtins.Types["List"].Methods {
		allListFuncs = append(allListFuncs, method.Name)
	}
}

func (f *fixture) builtinTypeMembers(t string) []string {
	members := []string{}
	for _, member := range f.a.builtins.Types[t].Members {
		members = append(members, member.Name)
	}
	return members
}

func (f *fixture) osSysSymbols() {
	f.Symbols("os", "sys")
	f.builtins.Symbols[0].Children = []query.Symbol{
		f.Symbol("environ"),
		f.Symbol("name"),
	}
	f.builtins.Symbols[1].Children = []query.Symbol{
		f.Symbol("argv"),
		f.Symbol("executable"),
	}
}

func assertCompletionResult(t *testing.T, names []string, result *protocol.CompletionList) {
	labels := make([]string, len(result.Items))
	for i, item := range result.Items {
		labels[i] = item.Label
	}
	assert.ElementsMatch(t, names, labels)
}

type tData struct { // testData
	doc        string
	line, char uint32
}

func testPosition(td tData) protocol.Position {
	if td.line == 0 && td.char == 0 { // last position if not specified
		lines := strings.Split(td.doc, "\n")
		lastLineIdx := len(lines) - 1
		return protocol.Position{Line: uint32(lastLineIdx), Character: uint32(len(lines[lastLineIdx]))}
	}
	return protocol.Position{Line: td.line, Character: td.char}
}

const docWithMultiplePlaces = `
def f1():
    pass

s = "a string"

def f2():
    # <- position 2
    return False

# <- position 1

if True:
    # <- position 3
    pass

t = 1234
`

const docWithErrorNode = `
def foo():
  pass

f(

def quux():
  pass
`

var (
	allDictFuncs   = []string{"clear", "get", "items", "keys", "pop", "popitem", "setdefault", "update", "values"}
	allListFuncs   = []string{"append", "clear", "extend", "index", "insert", "pop", "remove"}
	allStringFuncs = []string{"elem_ords", "capitalize", "codepoint_ords", "count", "endswith", "find", "format", "index", "isalnum", "isalpha", "isdigit", "islower", "isspace", "istitle", "isupper", "join", "lower", "lstrip", "partition", "removeprefix", "removesuffix", "replace", "rfind", "rindex", "rpartition", "rsplit", "rstrip", "split", "elems", "codepoints", "splitlines", "startswith", "strip", "title", "upper"}
)

func TestSimpleCompletion(t *testing.T) {
	f := newFixture(t)
	f.Symbols("foo", "bar", "baz")

	tests := []struct {
		tData
		expected []string
	}{
		{tData: tData{doc: ""}, expected: []string{"foo", "bar", "baz"}},
		{tData: tData{doc: "ba"}, expected: []string{"bar", "baz"}},
	}
	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			doc := f.MainDoc(tt.doc)
			result := f.a.Completion(doc, testPosition(tt.tData))
			assertCompletionResult(t, tt.expected, result)
		})
	}
}

func TestCompletions(t *testing.T) {
	tests := []struct {
		d              tData
		expected       []string
		osSys, builtin bool
		insertDot      int
	}{
		{d: tData{doc: ""}, expected: []string{"os", "sys"}, osSys: true},
		{d: tData{doc: "os."}, expected: []string{"environ", "name"}, osSys: true},
		{d: tData{doc: "os.e"}, expected: []string{"environ"}, osSys: true},

		// position 1
		{d: tData{doc: docWithMultiplePlaces, line: 10}, expected: []string{"f1", "s", "f2", "t", "os", "sys"}, osSys: true},

		// position 2
		{d: tData{doc: docWithMultiplePlaces, line: 7, char: 4}, expected: []string{"f1", "s", "f2", "t", "os", "sys"}, osSys: true},

		// position 3
		{d: tData{doc: docWithMultiplePlaces, line: 13, char: 4}, expected: []string{"f1", "s", "f2", "t", "os", "sys"}, osSys: true},
		{d: tData{doc: docWithErrorNode, line: 4, char: 1}, expected: []string{"foo"}, osSys: true},

		// inside string
		{d: tData{doc: `f = "abc123"`, char: 5}, expected: []string{}, osSys: true},

		// inside comment
		{d: tData{doc: `f = true # abc123`, char: 12}, expected: []string{}, osSys: true},

		// builtins
		{d: tData{doc: `f`}, expected: []string{"float", "fail"}, builtin: true},
		{d: tData{doc: `N`}, expected: []string{"None"}, builtin: true},
		{d: tData{doc: `T`}, expected: []string{"True"}, builtin: true},
		{d: tData{doc: `F`}, expected: []string{"False"}, builtin: true},

		// inside function body
		{d: tData{doc: "def fn():\n  \nx = True", line: 1, char: 2}, expected: []string{"fn", "x", "os", "sys"}, osSys: true},
		{d: tData{doc: "def fn():\n  a = 1\n  \n  \b  b = 2\n  return b\nx = True", line: 2, char: 2}, expected: []string{"a", "fn", "os", "sys", "x"}, osSys: true},

		// inside a list
		// {d: tData{doc: "x = [os.]", char: 8}, expected: []string{"environ", "name"}, osSys: true},
		// FIXME: due to changes in resolving ERROR node

		// inside a binary expression
		{d: tData{doc: "x = 'foo' + \nprint('')", char: 12}, expected: []string{"x", "os", "sys"}, osSys: true},
		{d: tData{doc: "x = 'foo' + os.\nprint('')", char: 15}, expected: []string{"environ", "name"}, osSys: true},

		// inside function argument lists
		{d: tData{doc: `foo()`, char: 4}, expected: []string{"os", "sys"}, osSys: true},
		{d: tData{doc: `foo(1, )`, char: 7}, expected: []string{"os", "sys"}, osSys: true},

		// inside condition of a conditional
		{d: tData{doc: "if :\n  pass\n", char: 3}, expected: []string{"os", "sys"}, osSys: true},
		//{d: tData{doc: "if os.:\n  pass\n", char: 6}, expected: []string{"environ", "name"}, osSys: true},
		//{d: tData{doc: "if flag and os.:\n  pass\n", char: 15}, expected: []string{"environ", "name"}, osSys: true},
		// FIXME: 2 last due to changes in resolving ERROR node

		// other edge cases
		// - because this gets parsed as an ERROR node at the top level, there's
		//   no assignment expression and the variable `flag` will not be in
		//   scope
		{d: tData{doc: "flag = "}, expected: []string{"os", "sys"}, osSys: true},
		{d: tData{doc: "flag = os."}, expected: []string{"environ", "name"}, osSys: true},

		// These should not trigger completion since the attribute expression is
		// anchored to a function call
		{d: tData{doc: "flag = len(os).", char: 15}, expected: []string{}, osSys: true},
		{d: tData{doc: "flag = len(os).sys", char: 15}, expected: []string{}, osSys: true},
	}

	for _, tt := range tests {
		t.Run(tt.d.doc, func(t *testing.T) {
			f := newFixture(t)
			if tt.builtin {
				f.builtinSymbols()
			}
			if tt.osSys {
				f.osSysSymbols()
			}

			doc := f.MainDoc(tt.d.doc)
			result := f.a.Completion(doc, testPosition(tt.d))
			assertCompletionResult(t, tt.expected, result)
		})
	}
}

func TestIdentifierCompletion(t *testing.T) {
	tests := []struct {
		tData
		expected []string
	}{
		// FIXME: fix resolving last dot. due to changes in ERROR node
		{tData{doc: ""}, []string{""}},
		{tData{doc: "os"}, []string{"os"}},
		{tData{doc: "os."}, []string{"os", ""}},
		{tData{doc: "os.e"}, []string{"os", "e"}},
		{tData{doc: "os.path."}, []string{"os", "path", ""}},
		{tData{doc: "os.path.e"}, []string{"os", "path", "e"}},
		{tData{doc: "[os]", char: 3}, []string{"os"}},
		//{tData{doc: "[os.]", char: 4}, []string{"os", ""}},
		{tData{doc: "[os.e]", char: 5}, []string{"os", "e"}},
		{tData{doc: "x = [os]", char: 7}, []string{"os"}},
		//{tData{doc: "x = [os.]", char: 8}, []string{"os", ""}},
		{tData{doc: "x = [os.e]", char: 9}, []string{"os", "e"}},
		// {tData{doc: "x = [os.path.]", char: 13}, []string{"os", "path", ""}},
		{tData{doc: "x = [os.path.e]", char: 14}, []string{"os", "path", "e"}},
		{tData{doc: "x = "}, []string{""}},
		{tData{doc: "if x and : pass", char: 9}, []string{""}},
		//{tData{doc: "if x and os.: pass", char: 12}, []string{"os", ""}},
		{tData{doc: "foo().bar.", char: 12}, []string{"foo", "bar", ""}},
		{tData{doc: `foo(11, True, "aa").bar.baz.`}, []string{"foo", "bar", "baz", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			f := newFixture(t)
			doc := f.MainDoc(tt.doc)
			pos := testPosition(tt.tData)
			nodes, ok := f.a.nodesAtPointForCompletion(doc, query.PositionToPoint(pos))
			assert.True(t, ok)
			ids := query.ExtractIdentifiers(doc, nodes, nil)
			assert.ElementsMatch(t, tt.expected, ids)
		})
	}
}

const functionFixture = `
def docker_build(ref: str,
                 context: str,
                 build_args: Dict[str, str] = {},
                 dockerfile: str = "Dockerfile",
                 dockerfile_contents: Union[str, Blob] = "",
                 live_update: List[LiveUpdateStep]=[],
                 match_in_env_vars: bool = False,
                 ignore: Union[str, List[str]] = [],
                 only: Union[str, List[str]] = [],
                 entrypoint: Union[str, List[str]] = [],
                 target: str = "",
                 ssh: Union[str, List[str]] = "",
                 network: str = "",
                 secret: Union[str, List[str]] = "",
                 extra_tag: Union[str, List[str]] = "",
                 container_args: List[str] = None,
                 cache_from: Union[str, List[str]] = [],
                 pull: bool = False,
                 platform: str = "") -> None:
    pass

def local(command: Union[str, List[str]],
          quiet: bool = False,
          command_bat: Union[str, List[str]] = "",
          echo_off: bool = False,
          env: Dict[str, str] = {},
          dir: str = "") -> Blob:
    pass
`

const customFn = `
def fn(a, b, c):
  pass

fn()
fn(b=1,)
`

func TestKeywordArgCompletion(t *testing.T) {
	// FIXME: fix commented out tests
	tests := []struct {
		doc        string
		line, char uint32
		expected   []string
	}{
		{doc: "local(c", char: 7, expected: []string{"command=", "command_bat="}},
		{doc: "local()", char: 6, expected: []string{"command=", "quiet=", "command_bat=", "echo_off=", "env=", "dir=", "docker_build", "local"}},
		//{doc: "local(", char: 6, expected: []string{"command=", "quiet=", "command_bat=", "echo_off=", "env=", "dir=", "docker_build", "local"}},
		{doc: "docker_build()", char: 13, expected: []string{"ref=", "context=", "build_args=", "dockerfile=", "dockerfile_contents=", "live_update=", "match_in_env_vars=", "ignore=", "only=", "entrypoint=", "target=", "ssh=", "network=", "secret=", "extra_tag=", "container_args=", "cache_from=", "pull=", "platform=", "docker_build", "local"}},

		// past first arg, exclude `command`
		//{doc: "local('echo',", char: 13, expected: []string{"quiet=", "command_bat=", "echo_off=", "env=", "dir=", "docker_build", "local"}},
		// past second arg, exclude `ref` and `context`
		{doc: "docker_build(ref, context,)", char: 26, expected: []string{"build_args=", "dockerfile=", "dockerfile_contents=", "live_update=", "match_in_env_vars=", "ignore=", "only=", "entrypoint=", "target=", "ssh=", "network=", "secret=", "extra_tag=", "container_args=", "cache_from=", "pull=", "platform=", "docker_build", "local"}},
		// used several kwargs
		{
			doc: "docker_build(ref='image:latest', context='.', dockerfile='Dockerfile.test', build_args={'DEBUG':'1'},)", char: 101,
			expected: []string{"dockerfile_contents=", "live_update=", "match_in_env_vars=", "ignore=", "only=", "entrypoint=", "target=", "ssh=", "network=", "secret=", "extra_tag=", "container_args=", "cache_from=", "pull=", "platform=", "docker_build", "local"},
		},

		// used `command` by position, `env` by keyword
		{doc: "local('echo $MESSAGE', env={'MESSAGE':'HELLO'},)", char: 47, expected: []string{"quiet=", "command_bat=", "echo_off=", "dir=", "docker_build", "local"}},

		// didn't use any positional arguments, but `quiet` is used
		{doc: "local(quiet=True,)", char: 17, expected: []string{"command=", "command_bat=", "echo_off=", "env=", "dir=", "docker_build", "local"}},

		// started to complete a keyword argument
		//{doc: "local(quiet=True,command)", char: 24, expected: []string{"command=", "command_bat="}},

		// not in an argument context
		//{doc: "local(quiet=True,command=)", char: 25, expected: []string{"docker_build", "local"}},
		{doc: "local(quiet=True,command=c)", char: 25, expected: []string{}},

		//{doc: customFn, line: 4, char: 3, expected: []string{"a=", "b=", "c=", "fn", "docker_build", "local"}},
		//{doc: customFn, line: 5, char: 7, expected: []string{"a=", "c=", "fn", "docker_build", "local"}},
	}

	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			f := newFixture(t)
			f.ParseBuiltins(functionFixture)

			doc := f.MainDoc(tt.doc)
			result := f.a.Completion(doc, protocol.Position{Line: tt.line, Character: tt.char})
			assertCompletionResult(t, tt.expected, result)
		})
	}
}

func TestMemberCompletion(t *testing.T) {
	f := newFixture(t)
	f.builtinSymbols()
	tests := []struct {
		tData
		expected []string
	}{
		{tData{doc: "pr"}, []string{"print"}},
		{tData{doc: "def print2():\n  print3=print2()\n  pr"}, []string{"print", "print2", "print3"}},
		{tData{doc: "pr.endswith"}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			doc := f.MainDoc(tt.doc)
			pos := testPosition(tt.tData)
			result := f.a.Completion(doc, pos)
			assertCompletionResult(t, tt.expected, result)
		})
	}
}

func TestKnownImmediateTypesCompletion(t *testing.T) {
	f := newFixture(t)
	f.builtinSymbols()

	f.builtins.Functions["foo"] = query.Signature{
		Name:       "foo",
		ReturnType: "str",
	}
	f.builtins.Functions["bar"] = query.Signature{
		Name:       "bar",
		ReturnType: "None",
	}
	f.builtins.Functions["baz"] = query.Signature{
		Name:       "baz",
		ReturnType: "dict",
	}

	tests := []struct {
		tData
		expected []string
	}{
		//
		//zero char/only dot completion

		// chained
		{tData{doc: `"".`}, allStringFuncs},
		{tData{doc: `[].`}, allListFuncs},
		{tData{doc: `{}.`}, allDictFuncs},

		// assignment
		{tData{doc: "s = \"\"\ns."}, allStringFuncs},
		{tData{doc: "l = []\nl."}, allListFuncs},
		{tData{doc: "s = {}\ns."}, allDictFuncs},

		// type propagation
		{tData{doc: "s1 = \"\"\ns=s1\ns."}, allStringFuncs},
		{tData{doc: "l1 = []\nl=l1\nl."}, allListFuncs},
		{tData{doc: "d1 = {}\nd=d1\nd."}, allDictFuncs},

		//
		// members

		// chained
		{tData{doc: `"aa".u`}, []string{"upper"}},
		{tData{doc: `"".keys`}, []string{}},
		{tData{doc: `"".append`}, []string{}},

		{tData{doc: `{}.k`}, []string{"keys"}},
		{tData{doc: `{}.upper`}, []string{}},
		{tData{doc: `{}.append`}, []string{}},

		{tData{doc: `[].a`}, []string{"append"}},
		{tData{doc: `[].keys`}, []string{}},
		{tData{doc: `[].upper`}, []string{}},

		// assignment
		{tData{doc: "s = \"\"\ns.u"}, []string{"upper"}},
		{tData{doc: "l = []\nl.a"}, []string{"append"}},
		{tData{doc: "d = {}\nd.k"}, []string{"keys"}},

		// type propagation
		{tData{doc: "s1 = \"\"\ns=s1\ns.u"}, []string{"upper"}},
		{tData{doc: "l1 = []\nl=l1\nl.a"}, []string{"append"}},
		{tData{doc: "d1 = {}\nd=d1\nd.k"}, []string{"keys"}},

		//
		// nested

		// chained
		{tData{doc: `"".upper().u`}, []string{"upper"}},
		{tData{doc: `"".upper().upper().u`}, []string{"upper"}},
		{tData{doc: `{}.keys().u`}, []string{}}, // list, not string
		{tData{doc: `{}.keys().a`}, []string{"append"}},

		{tData{doc: "s=\"\"\ns.upper().u"}, []string{"upper"}},
		{tData{doc: "s1=\"\"\ns=s1.upper()\ns.u"}, []string{"upper"}},
		{tData{doc: "s=\"\".upper()\ns.u"}, []string{"upper"}},
	}

	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			doc := f.MainDoc(tt.doc)
			pos := testPosition(tt.tData)
			result := f.a.Completion(doc, pos)
			assertCompletionResult(t, tt.expected, result)
		})
	}
}

const funcsAndObjectsFixture = `
class C1:
	i1: int
	mm: str
	cc: C4
	
	def getT(arg: str) -> str:
	    pass

	def getC(c1_getC_1: int) -> C3:
		"C1 -> GET C3"
		pass
	
class C2:
	i2: int
	mm: list
	cc: C3

	def getT() -> list:
	    pass

	def getC() -> C4:
		"C2 -> GET C4"
		pass

class C3:
	b3: bool
	mmm: str

class C4:
	b4: bool
	mmm: dict

def get_c1(s: str) -> C1:
	"GET C1"
	return C1()

def get_c2(i: int) -> C2:
	"GET C2"
	return C2()

def get_s(s, ss, sss) -> str:
	pass

def get_d(dd: dict) -> Dict:
	pass

def get_l(lll: list|None) -> list:
	pass
`

func TestBuiltinCompletionByIdentifiers(t *testing.T) {
	f := newFixture(t)
	f.builtinSymbols()
	f.ParseBuiltins(funcsAndObjectsFixture)

	tests := []struct {
		ids      []string
		expected []string
	}{
		{[]string{}, []string{}},     // no identifiers
		{[]string{""}, []string{}},   // empty id
		{[]string{"ak"}, []string{}}, // no such func/type

		{[]string{"get_c"}, []string{"get_c1", "get_c2"}}, // builtin funcs, partial match
		{[]string{"get_c1"}, []string{"get_c1"}},          // builtin func, exact match

		// builtin func + (1 level) +  all members
		{[]string{"get_c1", ""}, []string{"i1", "mm", "cc", "getT", "getC"}}, // builtin func, members
		{[]string{"get_s", ""}, allStringFuncs},
		{[]string{"get_d", ""}, allDictFuncs},
		{[]string{"get_l", ""}, allListFuncs},

		// builtin func + (1 level) + parial members
		{[]string{"get_c1", "get"}, []string{"getT", "getC"}},

		// builtin func + (2 levels) + all members
		{[]string{"get_c1", "i1", ""}, []string{}},
		{[]string{"get_c1", "mm", ""}, allStringFuncs},
		{[]string{"get_c1", "cc", ""}, []string{"b4", "mmm"}},

		// builtin func + (2 level) + parial members
		{[]string{"get_c1", "getT", "u"}, []string{"upper"}}, // via func
		{[]string{"get_c1", "getT", "keys"}, []string{}},
		{[]string{"get_c1", "mm", "u"}, []string{"upper"}}, // via member
		{[]string{"get_c1", "mm", "keys"}, []string{}},
		{[]string{"get_c1", "getC", "b"}, []string{"b3"}}, // via func
		{[]string{"get_c1", "cc", "b"}, []string{"b4"}},   // via member

		// builtin func + (3 levels) + all members
		{[]string{"get_c1", "getC", "mmm", ""}, allStringFuncs},
		{[]string{"get_c1", "getC", "b3", ""}, []string{}},
		{[]string{"get_c1", "cc", "mmm", ""}, allDictFuncs},
		{[]string{"get_c1", "cc", "b4", ""}, []string{}},

		// builtin func + (3 level) + parial members
		{[]string{"get_c1", "getC", "mmm", "u"}, []string{"upper"}}, // via func
		{[]string{"get_c1", "getC", "mmm", "keys"}, []string{}},
		{[]string{"get_c1", "cc", "mmm", "u"}, []string{"update"}}, // via member
		{[]string{"get_c1", "cc", "mmm", "k"}, []string{"keys"}},

		// builtin types
		{[]string{"String"}, []string{"String"}},

		{[]string{"String", ""}, allStringFuncs},
		{[]string{"String", "u"}, []string{"upper"}},
		{[]string{"String", "keys"}, []string{}},

		{[]string{"Dict", ""}, allDictFuncs},
		{[]string{"Dict", "u"}, []string{"update"}},
		{[]string{"Dict", "keys"}, []string{"keys"}},

		{[]string{"List", ""}, allListFuncs},
		{[]string{"List", "up"}, []string{}},
		{[]string{"List", "a"}, []string{"append"}},

		{[]string{"C1", ""}, []string{"i1", "mm", "cc", "getT", "getC"}},
		{[]string{"C1", "get"}, []string{"getT", "getC"}},

		{[]string{"C1", "mm", ""}, allStringFuncs},
		{[]string{"C1", "cc", ""}, []string{"b4", "mmm"}},

		{[]string{"C1", "mm", "u"}, []string{"upper"}},
		{[]string{"C1", "cc", "b"}, []string{"b4"}},

		{[]string{"C1", "cc", "mmm", ""}, allDictFuncs},
		{[]string{"C1", "cc", "mmm", "u"}, []string{"update"}},
		{[]string{"C1", "cc", "mmm", "k"}, []string{"keys"}},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			doc := f.MainDoc("")
			result, _ := f.a.builtinsCompletion(doc, tt.ids)
			names := query.SymbolNames(result)
			assert.ElementsMatch(t, names, tt.expected)
		})
	}
}

func TestResolveSymbolIdentifiers(t *testing.T) {
	f := newFixture(t)
	f.builtinSymbols()
	f.ParseBuiltins(funcsAndObjectsFixture)

	tests := []struct {
		tData
		expected []string
	}{
		{tData{doc: "r = get_d()"}, []string{"get_d"}},
		{tData{doc: `r = get_d().k`}, []string{"get_d", "k"}},
		{tData{doc: `r = get_d(1,2).k`}, []string{"get_d", "k"}},      // with args
		{tData{doc: `r = get_d(1, bar()).k`}, []string{"get_d", "k"}}, // ..
		{tData{doc: `r = get_d(1, bar()).k`}, []string{"get_d", "k"}}, // ..

		{tData{doc: "r1 = get_d()\nr = r1"}, []string{"get_d"}}, // indirect
		{tData{doc: "r1 = get_d()\nr = r1.keys"}, []string{"get_d", "keys"}},

		{tData{doc: "r1 = get_d().keys()\nr = r1.upper"}, []string{"get_d", "keys", "upper"}},
		{tData{doc: "r1 = get_C(1)\nr2 = r1.mm\nr=r2.u"}, []string{"get_C", "mm", "u"}},
		{tData{doc: "r1 = get_C(1,2).mm\nr2 = r1.upper()\nr=r2"}, []string{"get_C", "mm", "upper"}},
		{tData{doc: "r1 = get_C(bar(1)).cc\nr2 = r1.mmm()\nr=r2.k"}, []string{"get_C", "cc", "mmm", "k"}},

		{tData{doc: "r = {}"}, []string{"Dict"}},
		{tData{doc: "r = {}.keys"}, []string{"Dict", "keys"}},
		{tData{doc: "r1 = {}\nr = r1 "}, []string{"Dict"}},
		{tData{doc: "r1 = {}\nr = r1.k "}, []string{"Dict", "k"}},
		{tData{doc: "r1 = {}\nr2 = r1.keys()\nr3=r2.upper()\nr=r3"}, []string{"Dict", "keys", "upper"}},
	}
	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			doc := f.MainDoc(tt.doc)
			symbols := doc.Symbols()
			assert.NotEmpty(t, symbols)
			sym := SymbolMatching(symbols, "r")
			identifiers := f.a.resolveSymbolIdentifiers(symbols, sym)
			assert.ElementsMatch(t, tt.expected, identifiers)
		})
	}
}

func TestBuiltinNestedChainedAccess(t *testing.T) {
	f := newFixture(t)
	f.builtinSymbols()
	f.ParseBuiltins(funcsAndObjectsFixture)

	tests := []struct {
		tData
		expected []string
	}{
		// TODO: add get_s(bar()) and rewrite extractNodes for completion and extractIdentifiers
		// right now bar is identifier

		{tData{doc: `get_s("s").`}, allStringFuncs},
		{tData{doc: `get_d(1, d).`}, allDictFuncs},
		{tData{doc: `get_l(["a"], arg=l).`}, allListFuncs},
		{tData{doc: `get_c1().`}, []string{"i1", "mm", "cc", "getC", "getT"}},
		{tData{doc: `get_c2().`}, []string{"i2", "mm", "cc", "getC", "getT"}},

		// member is starlark builtin type
		{tData{doc: `get_c1().m`}, []string{"mm"}},
		{tData{doc: `get_c2().m`}, []string{"mm"}},
		{tData{doc: `get_c1().mm.up`}, []string{"upper"}}, // C1.mm is string
		{tData{doc: `get_c1().mm.append`}, []string{}},    //
		{tData{doc: `get_c2().mm.upper`}, []string{}},     // C2.mm is list
		{tData{doc: `get_c2().mm.a`}, []string{"append"}},
		{tData{doc: `get_c1().mm.`}, allStringFuncs},
		{tData{doc: `get_c2().mm.`}, allListFuncs},

		// member is a custom type
		{tData{doc: `get_c1().c`}, []string{"cc"}},
		{tData{doc: `get_c2().c`}, []string{"cc"}},
		{tData{doc: `get_c1().cc.b`}, []string{"b4"}},
		{tData{doc: `get_c2().cc.b`}, []string{"b3"}},
		{tData{doc: `get_c1().cc.m`}, []string{"mmm"}},
		{tData{doc: `get_c2().cc.m`}, []string{"mmm"}},
		{tData{doc: `get_c1().cc.mmm.upper`}, []string{}}, // C1.cc is C4, C4.mmm is dict
		{tData{doc: `get_c1().cc.mmm.k`}, []string{"keys"}},
		{tData{doc: `get_c2().cc.mmm.u`}, []string{"upper"}}, // C2.cc is C3, C3.mmm is str
		{tData{doc: `get_c1().cc.`}, []string{"b4", "mmm"}},
		{tData{doc: `get_c2().cc.`}, []string{"b3", "mmm"}},
		{tData{doc: `get_c1().cc.mmm.`}, allDictFuncs},
		{tData{doc: `get_c2().cc.mmm.`}, allStringFuncs},

		{tData{doc: `get_c1().get`}, []string{"getC", "getT"}},
		{tData{doc: `get_c2().get`}, []string{"getC", "getT"}},
		{tData{doc: `get_c1().getC().b`}, []string{"b3"}},
		{tData{doc: `get_c2().getC().b`}, []string{"b4"}},
		{tData{doc: `get_c1().getC().m`}, []string{"mmm"}},
		{tData{doc: `get_c2().getC().m`}, []string{"mmm"}},

		{tData{doc: `get_c1().getC().mmm.u`}, []string{"upper"}}, // C1.getC() is C3, C3.mmm is str
		{tData{doc: `get_c1().getC().mmm.k`}, []string{}},
		{tData{doc: `get_c2().getC().mmm.upper`}, []string{}},      // C2.cc is C4, C4.mmm is dict
		{tData{doc: `get_c2().getC().mmm.up`}, []string{"update"}}, // up-> update, since C2.mmm is dict
		{tData{doc: `get_c2().getC().mmm.k`}, []string{"keys"}},

		{tData{doc: `get_c1().getC().`}, []string{"b3", "mmm"}},
		{tData{doc: `get_c2().getC().`}, []string{"b4", "mmm"}},

		{tData{doc: `get_c1().getT().u`}, []string{"upper"}}, // C1.getT() is str
		{tData{doc: `get_c1().getT().append`}, []string{}},
		{tData{doc: `get_c2().getT().upper`}, []string{}}, // C2.getT() is list
		{tData{doc: `get_c2().getT().a`}, []string{"append"}},
		{tData{doc: `get_c1().getT().`}, allStringFuncs},
		{tData{doc: `get_c2().getT().`}, allListFuncs},
	}
	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			doc := f.MainDoc(tt.doc)
			pos := testPosition(tt.tData)
			result := f.a.Completion(doc, pos)
			assertCompletionResult(t, tt.expected, result)
		})
	}
}

func TestBuiltinNestedAssignment(t *testing.T) {
	// similar to TestBuiltinNestedChainedAccess, but with assignments
	f := newFixture(t)
	f.builtinSymbols()
	f.ParseBuiltins(funcsAndObjectsFixture)

	tests := []struct {
		tData
		expected []string
	}{
		{tData{doc: "r = get_s(`s`)\nr."}, allStringFuncs},
		{tData{doc: "r1 = get_d(1,d)\nr=r1\nr."}, allDictFuncs},
		{tData{doc: "r1 = get_l([`l`], arg=l)\nr2=r1\nr=r2\nr."}, allListFuncs},
		{tData{doc: "r=get_c1()\nr."}, []string{"i1", "mm", "cc", "getC", "getT"}},
		{tData{doc: "r1=get_c2()\nr=r1\nr."}, []string{"i2", "mm", "cc", "getC", "getT"}},

		// member is starlark builtin type
		{tData{doc: "r=get_c1()\nr.m"}, []string{"mm"}},
		{tData{doc: "r1=get_c2()\nr=r1\nr.m"}, []string{"mm"}},

		{tData{doc: "r=get_c1().mm\nr.up"}, []string{"upper"}}, // via identifier node
		{tData{doc: "r=get_c1()\nr.mm.up"}, []string{"upper"}}, // via attribute node
		{tData{doc: "r=get_c2()\nr.mm.upper"}, []string{}},     // C2.mm is list

		{tData{doc: "r1=get_c2()\nr=r1\nr.mm.ap"}, []string{"append"}}, // the same with one more propagation
		{tData{doc: "r1=get_c2().mm\nr=r1\nr.ap"}, []string{"append"}},
		{tData{doc: "r1=get_c2()\nr=r1.mm\nr.ap"}, []string{"append"}},
		{tData{doc: "r1=get_c1()\nr=r1.mm\nr.ap"}, []string{}}, // C1.mm is string

		{tData{doc: "r = get_c1().mm\nr."}, allStringFuncs},
		{tData{doc: "r = get_c1()\nr.mm."}, allStringFuncs},
		{tData{doc: "r1 = get_c2().mm\nr=r1\nr."}, allListFuncs},
		{tData{doc: "r1 = get_c2()\nr=r1.mm\nr."}, allListFuncs},
		{tData{doc: "r1 = get_c2()\nr=r1\nr.mm."}, allListFuncs},

		// member is a custom type
		{tData{doc: "r=get_c1().cc.mmm\nr.u"}, []string{"update"}},
		{tData{doc: "r=get_c1().cc\nr.mmm.u"}, []string{"update"}},
		{tData{doc: "r=get_c1()\nr.cc.mmm.u"}, []string{"update"}},
		{tData{doc: "r=get_c2()\nr.cc.mmm.u"}, []string{"upper"}},

		{tData{doc: "r2=get_c2()\nr1=r2.cc\nr=r1.mmm\nr.u"}, []string{"upper"}},
		{tData{doc: "r1=get_c2().cc\nr=r1\nr.mmm.u"}, []string{"upper"}},

		{tData{doc: "r=get_c2().getC()\nr.mmm.upper"}, []string{}},
		{tData{doc: "r=get_c2().getC()\nr.mmm.k"}, []string{"keys"}},
	}

	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			doc := f.MainDoc(tt.doc)
			pos := testPosition(tt.tData)
			result := f.a.Completion(doc, pos)
			assertCompletionResult(t, tt.expected, result)
		})
	}
}

// test various use cases/bugs fixes
func TestMix(t *testing.T) {

	f := newFixture(t)
	f.osSysSymbols()
	tests := []struct {
		tData
		expected []string
	}{
		{tData{doc: "def foo():\n print()\n os."}, []string{"environ", "name"}}, // ENG-687
	}

	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			doc := f.MainDoc(tt.doc)
			pos := testPosition(tt.tData)
			result := f.a.Completion(doc, pos)
			assertCompletionResult(t, tt.expected, result)
		})
	}
}
