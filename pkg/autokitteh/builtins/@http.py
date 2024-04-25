class HTTPResponseBody:
    def bytes():
        "Returns response body as bytes"
        pass

    def text():
        "Returns response body as text"
        pass

    def json():
        "Parses response body as json, returning JSON-decoded result"
        pass


class HTTPResponse:
    "The result of performing HTTP request"
    url: string  # the url that was ultimately requested (may change after redirects)
    status_code: int  # response status code (for example: 200 == OK)
    headers: dict  # dictionary of response headers
    encoding: string  # transfer encoding. example: "octet-stream" or "application/json"
    body: HTTPResponseBody


def delete(url: str, params: dict|None, headers: dict|None, data: string|bytes|list|dict|None, json: str|dict|None) -> HTTPResponse:
    """HTTP Delete method request

   Args:
     url: str
     params: Optional[dict]
     headers: Optional[dict]
     data: string|bytes|list|dict|None
     json: string|dict|None

   Returns:
     HTTPresponse - the result of performing a HTTP request
   """
    pass

def get(url: str, params: str|None, headers: str|None) -> HTTPResponse:
    """HTTP Get method request

   Args:
     url: str
     params: Optional[str]
     headers: Optional[str]

   Returns:
     HTTPresponse - the result of performing a HTTP request

   Examples:
     ```python
     resp = http.get('http://httpbin.org/get')
     print(resp.status_code)
     print(resp.body.text())

     print(http.get('http://httpbin.org/json').body.json())
     ```
   """
    pass

def head(url: str, params: dict|None, headers: dict|None, data: string|bytes|list|dict|None, json: str|dict|None) -> HTTPResponse:
    """HTTP Head method request

   Args:
     url: str
     params: Optional[dict]
     headers: Optional[dict]
     data: string|bytes|list|dict|None
     json: string|dict|None

   Returns:
     HTTPresponse - the result of performing a HTTP request
   """
    pass

def options(url: str, params: dict|None, headers: dict|None, data: string|bytes|list|dict|None, json: str|dict|None) -> HTTPResponse:
    """HTTP Options method request

   Args:
     url: str
     params: Optional[dict]
     headers: Optional[dict]
     data: string|bytes|list|dict|None
     json: string|dict|None

   Returns:
     HTTPresponse - the result of performing a HTTP request
   """
    pass

def patch(url: str, params: dict|None, headers: dict|None, data: string|bytes|list|dict|None, json: str|dict|None) -> HTTPResponse:
    """HTTP Patch method request

   Args:
     url: str
     params: Optional[dict]
     headers: Optional[dict]
     data: string|bytes|list|dict|None
     json: string|dict|None

   Returns:
     HTTPresponse - the result of performing a HTTP request
   """
    pass

def post(url: str, params: dict|None, headers: dict|None, data: string|bytes|list|dict|None, json: str|dict|None) -> HTTPResponse:
    """HTTP Post method request

   Args:
     url: str
     params: Optional[dict]
     headers: Optional[dict]
     data: string|bytes|list|dict|None
     json: string|dict|None

   Returns:
     HTTPresponse - the result of performing a HTTP request

   Examples:
     ```python
     # POST URL + params. Print as bytes
     print(http.put("http://httpbin.org/post?foo=bar", data="meow")).body.bytes()

     # POST as form
     print(http.put("http://httpbin.org/post", data={"foo": "bar"})).body.text()

     # POST as JSON
     print(http.put("http://httpbin.org/post", data={"foo": "bar"}, headers={"Content-Type": "application/json"})).body.text()

     # POST as JSON
     print(http.put("http://httpbin.org/post", json={"foo": "bar"})).body.json()
     ```
   """
    pass

def put(url: str, params: dict|None, headers: dict|None, data: string|bytes|list|dict|None, json: str|dict|None) -> HTTPResponse:
    """HTTP Put method request

   Args:
     url: str
     params: Optional[dict]
     headers: Optional[dict]
     data: string|bytes|list|dict|None
     json: string|dict|None

   Returns:
     HTTPresponse - the result of performing a HTTP request
   """
    pass
