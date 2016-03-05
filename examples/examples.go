package examples

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io"
)

type GreetingIn struct {
	Name string
}

type GreetingOut struct {
	Message string
}

func Greeting(w http.ResponseWriter, req *http.Request) {
	input := GreetingIn{}
	json.NewDecoder(req.Body).Decode(&input)
	defer req.Body.Close()

	output := GreetingOut{
		Message: fmt.Sprintf("Hello %v", input.Name),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)
}

func Echo(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	message := req.FormValue("q")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, message)
}

func Router() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/greeting", Greeting)
	router.HandleFunc("/echo", Echo)
	return router
}
