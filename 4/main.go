package main

import (
	"errors"
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("All ok"))
}

func main() {
	var e2 error = fmt.Errorf("some error")
	var e error = fmt.Errorf("It is %w", e2)
	fmt.Print(errors.Is(e, e2))

	

}
