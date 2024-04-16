def get(url, params, headers, body, form_body, form_encoding, json_body, auth):
    """HTTP Get method request

   Args:
     url: str
     params: Optional[str]
     headers: Optional[str]
     raw_body: Optional[str]
     form_body: Optional[str]
     json_body: Optional[str]


   Returns:
     response
       the result of performing a http request
       fields:
         url string
           the url that was ultimately requested (may change after redirects)
         status_code int
           response status code (for example: 200 == OK)
         headers dict
           dictionary of response headers
         encoding string
           transfer encoding. example: "octet-stream" or "application/json"
         body struct
           methods:
             bytes()
               body as bytes
             text()
               body as text
             json()
               attempt to parse resonse body as json, returning a JSON-decoded result
             from()
               body as dictionary of form fields names to values

   Examples:
     get data from http-bin

     ```
     resp = http.get('http://httpbingo.org/get') 
     print(resp.status_code)
     print(resp.body.text())
     ```

     ```
     print(http.get('http://httpbingo.org/json').body.json())
     ```
   """


def get(url, params, headers, body, form_body, form_encoding, json_body, auth):
    """HTTP Get method request
    API:
      https://datatracker.ietf.org/doc/html/rfc9110#name-get
    """
    pass


def head(url, params, headers, body, form_body, form_encoding, json_body, auth):
    """HTTP Head method request
    API:
      https://datatracker.ietf.org/doc/html/rfc9110#name-head"""
    pass


def options(url, params, headers, body, form_body, form_encoding, json_body, auth):
    """HTTP Options method request
    API:
      https://datatracker.ietf.org/doc/html/rfc9110#name-options"""
    pass


def patch(url, params, headers, body, form_body, form_encoding, json_body, auth):
    """HTTP Patch method request
    API:
      https://datatracker.ietf.org/doc/html/rfc5789"""
    pass


def post(args: dict):
    """HTTP Post method request
    API:
      https://datatracker.ietf.org/doc/html/rfc9110#name-post"""
    pass


def put(url, params, headers, body, form_body, form_encoding, json_body, auth):
    """HTTP Put method request
    API:
      https://datatracker.ietf.org/doc/html/rfc9110#name-put"""
    pass
