package esbcore

import (
	"errors"
	"log"
	"net"
	"net/http"
)

type SourceConnectorGenerator func(options map[string]string, container *Container) (func(PipelineProcess) error, error)

var sourceConnectorGenMap map[string]SourceConnectorGenerator

func init() {
	sourceConnectorGenMap = map[string]SourceConnectorGenerator{}

	sourceConnectorGenMap["http"] = sourceConnectorHttp
}

func sourceConnectorHttp(options map[string]string, container *Container) (func(PipelineProcess) error, error) {

	ls, ok := options["http.listen"]
	if !ok {
		return nil, errors.New("need provide http.listen for http")
	}
	path, ok := options["http.path"]
	if !ok {
		return nil, errors.New("need provide http.path for http")
	}

	return func(fn PipelineProcess) error {
		// mux
		mux := http.NewServeMux()
		mux.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
			m := NewModelInst(container.flowModel)
			// run process
			if err := fn(m); err != nil {
				log.Fatalln(err)
			}
		})

		l, err := net.Listen("tcp", ls)
		if err != nil {
			return err
		}

		return http.Serve(l, mux)
	}, nil
}
