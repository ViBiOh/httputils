package router

import (
	"fmt"
	"net/http"
)

type node struct {
	value    map[string]http.Handler
	prefix   string
	children []node
}

func (n node) IsZero() bool {
	return len(n.prefix) == 0
}

func (n node) insert(method string, parts []string, value http.Handler) (node, error) {
	if len(parts) == 0 || len(parts[0]) == 0 {
		if n.value[method] != nil {
			return n, fmt.Errorf("%s already assigned for `%s` method", n.prefix, method)
		}

		n.value[method] = value

		return n, nil
	}

	index := -1

	for i := range n.children {
		if n.children[i].prefix == parts[0] {
			index = i
			break
		}
	}

	if index == -1 {
		n.children = append(n.children, node{
			prefix: parts[0],
			value:  make(map[string]http.Handler),
		})
		index = len(n.children) - 1
	}

	var err error
	n.children[index], err = n.children[index].insert(method, parts[1:], value)
	if err != nil {
		return n, fmt.Errorf("%s/%s", parts[0], err.Error())
	}

	return n, nil
}

func (n node) find(method string, url string, variable bool) (http.Handler, bool) {
	if len(url) == 0 {
		return n.value[method], variable
	}

	remaining, prefix := getURLPart(url)
	if remaining == prefix {
		remaining = ""
	}

	for _, child := range n.children {
		if variable := child.prefix[0] == variablePrefix; variable || child.prefix == prefix {
			if handler, variable := child.find(method, remaining, variable); handler != nil {
				return handler, variable
			}
		}
	}

	for _, child := range n.children {
		if child.prefix[0] == wildcardPrefix {
			if handler, variable := child.find(method, "", true); handler != nil {
				return handler, variable
			}
		}
	}

	return nil, variable
}
