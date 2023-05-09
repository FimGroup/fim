package esbcore

import "log"

var targetConnectorGenMap map[string]TargetConnectorGenerator

type TargetConnectorGenerator func(options map[string]string) (func(source, dest *ModelInst) error, error)

func init() {
	targetConnectorGenMap = map[string]TargetConnectorGenerator{}

	targetConnectorGenMap["&http"] = targetConnectorHttp
}

func targetConnectorHttp(options map[string]string) (func(s, d *ModelInst) error, error) {
	return func(s, d *ModelInst) error {
		//TODO do http request
		log.Println("invoke targetConnectorHttp")
		return nil
	}, nil
}
