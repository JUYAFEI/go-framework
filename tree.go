package go_framework

import (
	"strings"
)

type Tree struct {
	Name       string
	Children   []*Tree
	RouterName string
	IsEnd      bool
}

// put path： /user/get/:id

func (t *Tree) Put(path string) {
	root := t
	strs := strings.Split(path, "/")
	for index, name := range strs {
		if index == 0 {
			continue
		}
		children := t.Children
		isMatch := false
		for _, node := range children {
			if node.Name == name {
				isMatch = true
				t = node
				break
			}
		}
		if !isMatch {
			isEnd := false
			if index == len(strs)-1 {
				isEnd = true
			}
			node := &Tree{
				Name:     name,
				Children: make([]*Tree, 0),
				IsEnd:    isEnd,
			}
			children = append(children, node)
			t.Children = children
			t = node
		}
	}
	t = root
}

//get path: /user/get/1

func (t *Tree) Get(path string) *Tree {
	strs := strings.Split(path, "/")
	routerName := ""
	for index, name := range strs {
		if index == 0 {
			continue
		}
		children := t.Children
		isMatch := false
		for _, node := range children {
			if node.Name == name || node.Name == "*" || strings.Contains(node.Name, ":") {
				isMatch = true
				routerName += "/" + node.Name
				node.RouterName = routerName
				t = node
				if index == len(strs)-1 {
					return node
				}
				break
			}
		}
		if !isMatch {
			for _, node := range children {
				if node.Name == "*" {
					return node
				}
			}
		}

	}
	return nil
}
