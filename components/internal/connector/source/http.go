package source

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/ThisIsSun/fim/fimapi/pluginapi"
	"github.com/ThisIsSun/fim/fimapi/rule"

	"github.com/go-chi/chi/v5"
)

const (
	ParamHttpBodyPrefix = "http/body/"
)

type httpRestServerConnector struct {
	instName  string
	generator *HttpRestServerGenerator

	entryPoint func(process pluginapi.PipelineProcess, definition *pluginapi.MappingDefinition) error
}

func (h *httpRestServerConnector) InvokeProcess(process pluginapi.PipelineProcess, definition *pluginapi.MappingDefinition) error {
	return h.entryPoint(process, definition)
}

func (h *httpRestServerConnector) Start() error {
	return h.generator.Start()
}

func (h *httpRestServerConnector) Stop() error {
	return h.generator.Stop()
}

func (h *httpRestServerConnector) Reload() error {
	return h.generator.Reload()
}

func (h *httpRestServerConnector) ConnectorName() string {
	return h.instName
}

func NewHttpRestServerGenerator() pluginapi.SourceConnectorGenerator {
	return &HttpRestServerGenerator{listenMap: map[string]struct {
		net.Listener
		*http.Server
		Mux *chi.Mux
	}{}}
}

type HttpRestServerGenerator struct {
	listenMap map[string]struct {
		net.Listener
		*http.Server
		Mux *chi.Mux
	}
}

func (h *HttpRestServerGenerator) GeneratorNames() []string {
	return []string{"http_rest"}
}

func (h *HttpRestServerGenerator) Start() error {
	for _, v := range h.listenMap {
		go func() {
			if err := v.Server.Serve(v.Listener); err != nil {
				log.Println("serving http error:", err)
			}
		}()
	}
	return nil
}

func (h *HttpRestServerGenerator) Stop() error {
	//FIXME shutdown listeners
	return nil
}

func (h *HttpRestServerGenerator) Reload() error {
	//FIXME allow to reload http registrations including start or shutdown listeners
	return nil
}

func (h *HttpRestServerGenerator) addHandler(options map[string]string, handleFunc http.HandlerFunc) error {
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

func (h *HttpRestServerGenerator) GenerateSourceConnectorInstance(options map[string]string, container pluginapi.Container) (pluginapi.SourceConnector, error) {
	entryPoint := func(fn pluginapi.PipelineProcess, mappingDef *pluginapi.MappingDefinition) error {
		f := func(writer http.ResponseWriter, request *http.Request) {

			body, err := io.ReadAll(request.Body)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			if len(body) > 0 {
				contentType := request.Header.Get("Content-Type")
				if !strings.HasPrefix(contentType, "application/json") {
					writer.WriteHeader(http.StatusBadRequest)
					return
				}
			}

			// convert request
			m := container.NewModel()
			if err := h.convertJsonRequestModel(request, body, m, mappingDef); err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			// run process
			if err := fn(m); err != nil {
				log.Println("error processing:", err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			// convert response
			if data, err := h.convertJsonResponseModel(m, mappingDef); err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				writer.Header().Add("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write(data)
				if err != nil {
					log.Println("write response error:", err)
				}
				return
			}
		}
		if err := h.addHandler(options, f); err != nil {
			return err
		}
		if err := h.Reload(); err != nil {
			return err
		}
		return nil
	}

	return &httpRestServerConnector{
		instName:   "http_rest",
		generator:  h,
		entryPoint: entryPoint,
	}, nil
}

func (h *HttpRestServerGenerator) convertJsonResponseModel(m pluginapi.Model, def *pluginapi.MappingDefinition) ([]byte, error) {
	r := map[string]interface{}{}
	for fp, cp := range def.Res {
		val := m.GetFieldUnsafe(rule.SplitFullPath(fp))
		if val == nil {
			continue
		}
		if strings.HasPrefix(cp, ParamHttpBodyPrefix) {
			destPaths := rule.SplitFullPath(cp[len(ParamHttpBodyPrefix):])
			m := r
			for _, p := range destPaths[:len(destPaths)-1] {
				//FIXME need support the following data types: array
				nm, ok := m[p]
				if !ok {
					nm = map[string]interface{}{}
					m[p] = nm
				}
				if nmv, ok := nm.(map[string]interface{}); !ok {
					return nil, errors.New("data type is not object")
				} else {
					m = nmv
				}
			}
			lastPath := destPaths[len(destPaths)-1]
			m[lastPath] = val
		} else {
			//FIXME support more data access, e.g. headers
		}
	}
	return json.Marshal(r)
}

func (h *HttpRestServerGenerator) convertJsonRequestModel(request *http.Request, body []byte, m pluginapi.Model, def *pluginapi.MappingDefinition) error {
	var b interface{}
	if err := json.Unmarshal(body, &b); err != nil {
		log.Println(err)
	}

	for fp, cp := range def.Req {
		if strings.HasPrefix(cp, ParamHttpBodyPrefix) {
			// http body
			val, err := h.traverseRetrievingFromGenericJson(b, rule.SplitFullPath(cp[len(ParamHttpBodyPrefix):]))
			if err != nil {
				return err
			}
			if err := m.AddOrUpdateField0(rule.SplitFullPath(fp), val); err != nil {
				return err
			}
		} else {
			//FIXME support more data access, e.g. headers
		}
	}
	return nil
}

func (h *HttpRestServerGenerator) traverseRetrievingFromGenericJson(o interface{}, paths []string) (interface{}, error) {
	if o == nil {
		return nil, nil
	}
	//FIXME need support the following data types: array
	val, ok := o.(map[string]interface{})[paths[0]]
	if !ok {
		return nil, nil
	}
	if len(paths) == 1 {
		switch val.(type) {
		case map[string]interface{}:
			return nil, errors.New("source object is not a primitive type but got object")
		case []interface{}:
			return nil, errors.New("source object is not a primitive type but got array")
		}
		//FIXME should check data type
		return val, nil
	} else {
		return h.traverseRetrievingFromGenericJson(val, paths[1:])
	}
}
