# starlib.re
# defines regular expression functions,
# it's intended to be a drop-in subset of python's re module for starlark: https://docs.python.org/3/library/re.html


## Types

class Pattern:
    def match(text: str, flags: int = 0 ) -> str:
        pass

    def findall(text: str, flags : int = 0 ) -> list[str]:
        pass

    def split(text: str, maxsplit: int = 0, flags : int = 0 ) -> list[str]:
        pass

    def sub(repl: str, text: str, count : int = 0, flags: int = 0) -> str:
        pass


## Functions

def compile(pattern: str) -> Pattern:
    """Compile a regular expression pattern into a regular expression object, which can be used for matching using its match(), search() and other methods.

    Args:
      pattern : string - regular expression pattern string
    """
    pass

def findall(pattern: str, text: str, flags: int = 0) -> list[str]:
    """Returns all non-overlapping matches of pattern in string, as a list of strings. The string is scanned left-to-right, and matches are returned in the order found. If one or more groups are present in the pattern, return a list of groups; this will be a list of tuples if the pattern has more than one group. Empty matches are included in the result.

    Args:
      pattern : string - regular expression pattern string
      text : string - string to find within
      flags : int - integer flags to control regex behaviour. reserved for future use
    """
    pass

def split(pattern: str, text: str, maxsplit: int = 0, flags: int = 0) -> list[str]:
    """Split text by the occurrences of pattern. If capturing parentheses are used in pattern, then the text of all groups in the pattern are also returned as part of the resulting list. If maxsplit is nonzero, at most maxsplit splits occur, and the remainder of the string is returned as the final element of the list

    Args:
      pattern : string - regular expression pattern string
      text : string - input string to split
      maxsplit : int - maximum number of splits to make. default 0 splits all matches
      flags : int - integer flags to control regex behaviour. reserved for future use
    """
    pass

def sub(pattern: str, repl: str, text: str, count: int = 0, flags: int = 0) -> str:
    """Return the string obtained by replacing the leftmost non-overlapping occurrences of pattern in string by the replacement repl. If the pattern isn’t found, string is returned unchanged. repl can be a string or a function; if it is a string, any backslash escapes in it are processed. That is, \n is converted to a single newline character, \r is converted to a carriage return, and so forth.

    Args:
      pattern : string - regular expression pattern string
      repl : string - string to replace matches with
      text : string - input string to replace
      count : int - number of replacements to make, default 0 means replace all matches
      flags : int - integer flags to control regex behaviour. reserved for future use
    """
    pass

def match(pattern: str, string: str, flags: int = 0 ) -> str:
    """If zero or more characters at the beginning of string match the regular expression pattern, return a corresponding match string tuple. Return None if the string does not match the pattern

    Args:
      pattern : string - regular expression pattern string
      string : string - input string to match
    """
    pass
