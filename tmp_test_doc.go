package main

import (
	"fmt"
	_ "github.com/lever-dev/padel-backend/docs"
	"github.com/swaggo/swag"
)

func main() {
	doc, err := swag.ReadDoc()
	fmt.Println("err=", err)
	fmt.Println(len(doc))
	fmt.Println(doc[:100])
}
