package pluginapi

import "fmt"

type FlowError struct {
	Key     string
	Message string
}

func (f FlowError) Error() string {
	return fmt.Sprint(f.Key, "::", f.Message)
}

func (f FlowError) ErrorKey() string {
	return f.Key
}

func (f FlowError) ErrorMessage() string {
	return f.Message
}
