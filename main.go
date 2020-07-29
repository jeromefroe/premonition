package main

import (
	"fmt"
	"log"
	"strings"
)

const input = `
type_name: Apple
color: Red
---
type_name: Banana
ripe: true
`

func main() {
	objs, err := Decode(strings.NewReader(input))
	if err != nil {
		log.Fatalf("unable to decode objects: %v", err)
	}

	for _, obj := range objs {
		fmt.Printf("%+v\n", obj)
	}
}
