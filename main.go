package main

import (
	"fmt"
	"gomodest/pkg"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Print(err.Error())
		os.Exit(1)
	}

	fmt.Println("Listening on http://localhost:4000")
	http.ListenAndServe(":4000", pkg.NewRouter())
}
