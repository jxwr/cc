package streams

var (
	NodeStateStream = make(chan interface{}, 1024)
)
