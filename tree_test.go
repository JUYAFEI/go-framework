package go_framework

import (
	"fmt"
	"testing"
)

// 前缀树单元test
func TestTree(t *testing.T) {
	tree := &Tree{
		Name:     "/",
		Children: make([]*Tree, 0),
	}
	tree.Put("/user/get/:id")
	tree.Put("/user/create/hello")
	tree.Put("/user/create/aaa")

	node := tree.Get("/user/get/1")
	fmt.Println(node)

	node = tree.Get("/user/create/hello")
	fmt.Println(node)
}
