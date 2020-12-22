// +build ignore

package main

import (
	"log"

	"github.com/facebook/ent/entc"
	"github.com/facebook/ent/entc/gen"
	"github.com/facebook/ent/schema/field"
)

func main() {
	err := entc.Generate("./schema", &gen.Config{
		Header: `
			// Code generated (@generated) by entc, DO NOT EDIT.
		`,
		IDType: &field.TypeInfo{Type: field.TypeInt},
	})
	if err != nil {
		log.Fatal("running ent codegen:", err)
	}
}
