# mockhttp
Tools to simplify testing golang http.Handler

## JSON Example

``` go
type Login struct {
  Username string 
  Password string
}

app := mockhttp.New(handler)

resp, err := app.POST("/api/foo", Login{
  Username: "foo",
  Password: "bar",
})

// resp contains accessors to the response
```

## Remote example

In addition to testing local interfaces, mockhttp can now also be used to test remote apis

``` go
app := mockhttp.New(nil, mockhttp.Codebase("http://example.com"))
resp, err := app.GET("/")
```