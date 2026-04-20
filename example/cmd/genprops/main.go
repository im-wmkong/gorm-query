package main

import (
	"log"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/genprops"
)

func main() {
	g := genprops.New()

	if err := g.Generate(&model.User{}); err != nil {
		log.Fatal(err)
	}
}
