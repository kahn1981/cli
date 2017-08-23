package translatableerror

type RandomPortCombinationError struct {
}

func (RandomPortCombinationError) Error() string {
	return "Cannot specify random-port together with port, hostname, and/or path."
}

func (e RandomPortCombinationError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
