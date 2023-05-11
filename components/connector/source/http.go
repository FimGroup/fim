package source

import (
	"errors"
	"log"
	"net/http"

	"esbconcept/esbapi"
)

var httpServer *HttpServer

func init() {
	httpServer = &HttpServer{
		mux:     http.NewServeMux(),
		tempMux: http.NewServeMux(),
	}
}

type HttpServer struct {
	mux     *http.ServeMux
	tempMux *http.ServeMux
}

func (h *HttpServer) Start() error {
	//TODO implement me
	panic(esbapi.IMPLEMENT_ME)
}

func (h *HttpServer) Stop() error {
	//TODO implement me
	panic(esbapi.IMPLEMENT_ME)
}

func (h *HttpServer) Reload() error {
	//TODO implement me
	panic(esbapi.IMPLEMENT_ME)
}

func (h *HttpServer) addHandler(options map[string]string, handleFunc http.HandlerFunc) error {
	ls, ok := options["http.listen"]
	if !ok {
		return errors.New("need provide http.listen for http")
	}
	path, ok := options["http.path"]
	if !ok {
		return errors.New("need provide http.path for http")
	}
	method, ok := options["http.method"]
	if !ok {
		return errors.New("need provide http.method for http")
	}
	//FIXME check path and listen duplication

	var _ = ls
	var _ = path
	var _ = method
	panic(esbapi.IMPLEMENT_ME)
}

func sourceConnectorHttpRest(options map[string]string, container esbapi.Container) (struct {
	esbapi.Connector
	esbapi.ConnectorProcessEntryPoint
	InstanceName string
}, error) {

	entryPoint := func(fn esbapi.PipelineProcess, mappingDef *esbapi.MappingDefinition) error {
		f := func(writer http.ResponseWriter, request *http.Request) {
			m := container.NewModel()
			// run process
			if err := fn(m); err != nil {
				log.Fatalln(err)
			}
		}
		if err := httpServer.addHandler(options, f); err != nil {
			return err
		}
		if err := httpServer.Reload(); err != nil {
			return err
		}
		return nil
	}

	return struct {
		esbapi.Connector
		esbapi.ConnectorProcessEntryPoint
		InstanceName string
	}{
		Connector:                  httpServer,
		ConnectorProcessEntryPoint: entryPoint,
		InstanceName:               "http_rest",
	}, nil
}
