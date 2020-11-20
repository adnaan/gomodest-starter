package main

import (
	"fmt"
	"gomodest/pkg"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("Listening on http://localhost:4000")
	http.ListenAndServe(":4000", pkg.NewRouter())
}
