package tools

// ConcurrentMap is a map[string]interface{} with concurrent access
type ConcurrentMap struct {
	req     chan request
	content map[string]interface{}
}

type request struct {
	action    string
	key       string
	content   interface{}
	ioContent chan interface{}
}

// List map content to output channel
func (c *ConcurrentMap) List() <-chan interface{} {
	req := request{action: `list`, ioContent: make(chan interface{}, len(c.content))}
	c.req <- req

	return req.ioContent
}

// Get entry in map
func (c *ConcurrentMap) Get(key string) interface{} {
	req := request{action: `get`, key: key, ioContent: make(chan interface{})}
	c.req <- req

	return <-req.ioContent
}

// Push entry in map
func (c *ConcurrentMap) Push(key string, entry interface{}) {
	req := request{action: `push`, key: key, content: entry}
	c.req <- req
}

// Remove key from map
func (c *ConcurrentMap) Remove(key string) {
	req := request{action: `remove`, key: key}
	c.req <- req
}

// Close stops concurrent map and return map
func (c *ConcurrentMap) Close() map[string]interface{} {
	req := request{action: `stop`, ioContent: make(chan interface{})}
	c.req <- req

	<-req.ioContent
	return c.content
}

// CreateConcurrentMap in a subroutine
func CreateConcurrentMap(contentSize int, channelSize int) *ConcurrentMap {
	concurrentMap := ConcurrentMap{req: make(chan request, channelSize), content: make(map[string]interface{}, contentSize)}

	go func() {
		for request := range concurrentMap.req {
			switch request.action {
			case `get`:
				if entry, ok := concurrentMap.content[request.key]; ok {
					request.ioContent <- entry
				}
				close(request.ioContent)
				break
			case `list`:
				for _, entry := range concurrentMap.content {
					request.ioContent <- entry
				}
				close(request.ioContent)
				break
			case `push`:
				concurrentMap.content[request.key] = request.content
				break
			case `remove`:
				delete(concurrentMap.content, request.key)
				break
			case `stop`:
				close(request.ioContent)
				return
			}
		}
	}()

	return &concurrentMap
}
