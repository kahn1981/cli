package translatableerror

type TCPPortCombinationError struct {
}

func (TCPPortCombinationError) Error() string {
	return "Cannot specify port together with hostname and/or path."
}

func (e TCPPortCombinationError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
