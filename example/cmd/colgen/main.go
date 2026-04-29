package main

import (
	"log"

	"github.com/im-wmkong/gorm-query/colgen"
	"github.com/im-wmkong/gorm-query/example/model"
)

func main() {
	g := colgen.New()

	if err := g.Generate(&model.User{}, &model.Profile{}); err != nil {
		log.Fatal(err)
	}
}
