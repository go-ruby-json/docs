// SPDX-License-Identifier: BSD-3-Clause
package main

import (
	"fmt"
	"strings"

	"github.com/go-ruby-json/json"
)

func buildDoc() string {
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < 60; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		ok := "false"
		if i%2 == 0 {
			ok = "true"
		}
		fmt.Fprintf(&b, `{"id":%d,"name":"item-%d","score":%d,"ok":%s,"tags":[%d,%d,%d]}`,
			i, i, (i*37)%100, ok, i, i*2, i*3)
	}
	b.WriteByte(']')
	return b.String()
}

func main() {
	doc := buildDoc()
	tree, _ := json.Parse(doc)
	bench("parse-60obj", 1000, func() { v, _ := json.Parse(doc); sink = v })
	bench("generate-60obj", 1000, func() { v, _ := json.Generate(tree); sink = v })
}
