package tools

// ConcurrentMap is a map[string]MapContent with concurrent access
type ConcurrentMap struct {
	req     chan request
	content map[string]MapContent
}

// MapContent stored in map
type MapContent interface {
	GetID() string
}

type simpleMapContent struct {
	ID string
}

func (c simpleMapContent) GetID() string {
	return c.ID
}

type request struct {
	action    string
	key       string
	ioContent chan MapContent
}

// Get entry in map
func (c *ConcurrentMap) Get(key string) MapContent {
	req := request{action: `get`, key: key, ioContent: make(chan MapContent)}
	c.req <- req

	return <-req.ioContent
}

// Push entry in map
func (c *ConcurrentMap) Push(entry MapContent) {
	req := request{action: `push`, ioContent: make(chan MapContent)}
	c.req <- req

	req.ioContent <- entry
}

// Remove key from map
func (c *ConcurrentMap) Remove(ID string) {
	req := request{action: `remove`, ioContent: make(chan MapContent)}
	c.req <- req

	req.ioContent <- simpleMapContent{ID}
}

// List map content to output channel
func (c *ConcurrentMap) List() <-chan MapContent {
	req := request{action: `list`, ioContent: make(chan MapContent, len(c.content))}
	c.req <- req

	return req.ioContent
}

// Close stops concurrent map and return map
func (c *ConcurrentMap) Close() map[string]MapContent {
	req := request{action: `stop`, ioContent: make(chan MapContent)}
	c.req <- req

	<-req.ioContent
	return c.content
}

// CreateConcurrentMap in a subroutine
func CreateConcurrentMap(contentSize int, channelSize int) *ConcurrentMap {
	concurrentMap := ConcurrentMap{req: make(chan request, channelSize), content: make(map[string]MapContent, contentSize)}

	go func() {
		for request := range concurrentMap.req {
			if request.action == `get` {
				if entry, ok := concurrentMap.content[request.key]; ok {
					request.ioContent <- entry
				}
			} else if request.action == `push` {
				entry := <-request.ioContent
				concurrentMap.content[entry.GetID()] = entry
			} else if request.action == `remove` {
				entry := <-request.ioContent
				delete(concurrentMap.content, entry.GetID())
			} else if request.action == `list` {
				for _, entry := range concurrentMap.content {
					request.ioContent <- entry
				}
			} else if request.action == `stop` {
				close(request.ioContent)
				return
			}

			close(request.ioContent)
		}
	}()

	return &concurrentMap
}
