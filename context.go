package pubsub

type Publisher interface {
	PublishTo(ctx *Context) error
}

type Subscriber interface {
	SubscribeTo(ctx *Context) error
}

type Context struct {
	Buffer *Buffer
	Done   MarkerChan
	Close  func()
}
