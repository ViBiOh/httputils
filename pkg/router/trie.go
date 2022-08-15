package router

import (
	"fmt"
	"net/http"
)

type node struct {
	value                 map[string]http.Handler
	prefix                string
	children              []node
	hasDescendingVariable bool
}

func (n node) IsZero() bool {
	return len(n.prefix) == 0
}

func (n node) insert(method, url string, variable bool, value http.Handler) (node, error) {
	if len(url) == 0 {
		if n.value[method] != nil {
			return n, fmt.Errorf(" is already assigned for `%s` method", method)
		}

		n.value[method] = value

		return n, nil
	}

	index := -1

	remaining, prefix := getURLPart(url)
	if remaining == prefix {
		remaining = ""
	}

	for i := range n.children {
		if n.children[i].prefix == prefix {
			index = i

			break
		}
	}

	if index == -1 {
		n.children = append(n.children, node{
			prefix:                prefix,
			hasDescendingVariable: variable || prefix[0] == variablePrefix,
			value:                 make(map[string]http.Handler),
		})
		index = len(n.children) - 1
	}

	var err error
	node := n.children[index]
	node, err = node.insert(method, remaining, node.hasDescendingVariable, value)
	if err != nil {
		return n, fmt.Errorf("%s/%s", prefix, err.Error())
	}

	n.children[index] = node

	return n, nil
}

func (n node) find(method, url string) (http.Handler, bool) {
	if len(url) == 0 {
		return n.value[method], n.hasDescendingVariable
	}

	remaining, prefix := getURLPart(url)
	if remaining == prefix {
		remaining = ""
	}

	for _, child := range n.children {
		if child.prefix[0] == variablePrefix || child.prefix == prefix {
			if handler, variable := child.find(method, remaining); handler != nil {
				return handler, variable
			}
		}
	}

	for _, child := range n.children {
		if child.prefix[0] == wildcardPrefix {
			if handler, variable := child.find(method, ""); handler != nil {
				return handler, variable
			}
		}
	}

	return nil, false
}

func (n node) extractVariable(method, url string, params map[string]string) bool {
	if len(url) == 0 {
		return n.value[method] != nil
	}

	remaining, prefix := getURLPart(url)
	if remaining == prefix {
		remaining = ""
	}

	for _, child := range n.children {
		if variable := child.prefix[0] == variablePrefix; variable || child.prefix == prefix {
			if variable {
				params[child.prefix[1:]] = prefix
			}

			if child.extractVariable(method, remaining, params) {
				return true
			}
		}
	}

	for _, child := range n.children {
		if child.prefix[0] == wildcardPrefix {
			params[child.prefix[1:]] = url

			if child.extractVariable(method, "", params) {
				return true
			}
		}
	}

	return false
}
