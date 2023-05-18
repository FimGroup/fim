package pluginapi

import "fmt"

type FlowError struct {
	Key     string
	Message string
}

func (f FlowError) Error() string {
	return fmt.Sprint(f.Key, "::", f.Message)
}

type FlowStop struct {
	Key     string
	Message string
}

func (f FlowStop) Error() string {
	return fmt.Sprint(f.Key, "::", f.Message)
}
