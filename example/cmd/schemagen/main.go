package main

import (
	"log"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/schemagen"
)

func main() {
	g := schemagen.New()

	if err := g.Generate(&model.User{}, &model.Profile{}); err != nil {
		log.Fatal(err)
	}
}
