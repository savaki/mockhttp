# mockhttp
Tools to simplify testing golang http http.Handler

## JSON Example

``` go
type Login struct {
  Username string 
  Password string
}

app := mockhttp.New(handler)

resp := app.POST("/api/foo", Login{
  Username: "foo",
  Password: "bar",
})

// resp contains accessors to the response
```
