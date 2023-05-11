package source

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"esbconcept/esbapi"

	"github.com/go-chi/chi/v5"
)

var httpServer *HttpServer

func init() {
	httpServer = &HttpServer{
		listenMap: map[string]struct {
			net.Listener
			*http.Server
			Mux *chi.Mux
		}{},
	}
}

type HttpServer struct {
	listenMap map[string]struct {
		net.Listener
		*http.Server
		Mux *chi.Mux
	}
}

func (h *HttpServer) Start() error {
	for _, v := range h.listenMap {
		go func() {
			if err := v.Server.Serve(v.Listener); err != nil {
				log.Println("serving http error:", err)
			}
		}()
	}
	return nil
}

func (h *HttpServer) Stop() error {
	//FIXME shutdown listeners
	return nil
}

func (h *HttpServer) Reload() error {
	//FIXME allow to reload http registrations including start or shutdown listeners
	return nil
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
	method = strings.ToUpper(method)
	if !ok {
		return errors.New("need provide http.method for http")
	}
	//FIXME check path and listen duplication

	lstruct, ok := h.listenMap[ls]
	if !ok {
		r := chi.NewRouter()

		l, err := net.Listen("tcp", ls)
		if err != nil {
			return err
		}
		lstruct = struct {
			net.Listener
			*http.Server
			Mux *chi.Mux
		}{
			Listener: l,
			Server:   &http.Server{Handler: r},
			Mux:      r,
		}
		h.listenMap[ls] = lstruct
	} else {
		// check duplication
		//FIXME should use a better alternative
		if lstruct.Mux.Match(chi.NewRouteContext(), method, path) {
			return errors.New(fmt.Sprintf("duplicated http path:%s method:%s", path, method))
		}
	}

	//FIXME may cause concurrent issue on adding new handler while processing requests
	switch method {
	case "GET":
		lstruct.Mux.Get(path, handleFunc)
	case "POST":
		lstruct.Mux.Post(path, handleFunc)
	case "PUT":
		lstruct.Mux.Put(path, handleFunc)
	case "DELETE":
		lstruct.Mux.Delete(path, handleFunc)
	case "HEAD":
		lstruct.Mux.Head(path, handleFunc)
	case "PATCH":
		lstruct.Mux.Patch(path, handleFunc)
	default:
		return errors.New(fmt.Sprintf("unable to register http path:%s method:%s", path, method))
	}
	return nil
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
