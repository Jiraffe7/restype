package restype

// Default is a convenience struct for embedding
// that provides default (nil) implementation of optional Request[T] methods.
type Default struct{}

func (Default) PathParams() map[string]string {
	return nil
}

func (Default) QueryParams() map[string]string {
	return nil
}

func (Default) Headers() map[string]string {
	return nil
}

func (Default) Body() ([]byte, error) {
	return nil, nil
}

func (Default) ResponseFromBytes([]byte) (any, error) {
	return nil, nil
}
