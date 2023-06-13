package source

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/providers"
	"github.com/FimGroup/fim/fimapi/rule"
	"github.com/FimGroup/fim/fimsupport/logging"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

const (
	ParamHttpBodyPrefix        = "http/body/"
	ParamHttpQueryStringPrefix = "http/query_string/"
	ParamHttpHeaderPrefix      = "http/header/"
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

type accessLogger struct {
	loggerManager providers.LoggerManager
	logger        providers.Logger
}

func (a *accessLogger) Print(v ...interface{}) {
	a.logger.Info(v...)
}

func newAccessLogger() *accessLogger {
	lm, err := logging.NewLoggerManager("logs/http_access", 7, 10*1024*1024, 5, logrus.InfoLevel, false, false)
	if err != nil {
		panic(err)
	}
	return &accessLogger{
		loggerManager: lm,
		logger:        lm.GetLogger("FimGroup.Component.HttpAccessLog"),
	}
}

func NewHttpRestServerGenerator() pluginapi.SourceConnectorGenerator {
	accessLog := newAccessLogger()
	return &HttpRestServerGenerator{
		listenMap: map[string]struct {
			net.Listener
			*http.Server
			Mux *chi.Mux
		}{},

		_logger:       logging.GetLoggerManager().GetLogger("FimGroup.Component.HttpRestServerConnector"),
		_accessLogger: accessLog,
	}
}

type HttpRestServerGenerator struct {
	listenMap map[string]struct {
		net.Listener
		*http.Server
		Mux *chi.Mux
	}

	_logger       providers.Logger
	_accessLogger *accessLogger
}

func (h *HttpRestServerGenerator) GeneratorNames() []string {
	return []string{"http_rest"}
}

func (h *HttpRestServerGenerator) Start() error {
	for _, v := range h.listenMap {
		go func() {
			if err := v.Server.Serve(v.Listener); err != nil {
				h._logger.Error("serving http error:", err)
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

		// middlewares
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		// http access log
		r.Use(func(handler http.Handler) http.Handler {
			format := &middleware.DefaultLogFormatter{Logger: h._accessLogger, NoColor: true}
			fn := func(w http.ResponseWriter, r *http.Request) {
				entry := format.NewLogEntry(r)
				ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
				t1 := time.Now()
				defer func() {
					entry.Write(ww.Status(), ww.BytesWritten(), ww.Header(), time.Since(t1), nil)
				}()
				handler.ServeHTTP(ww, middleware.WithLogEntry(r, entry))
			}
			return http.HandlerFunc(fn)
		})
		// middlewares
		r.Use(middleware.Recoverer)
		// Set a timeout value on the request context (ctx), that will signal
		// through ctx.Done() that the request has timed out and further
		// processing should be stopped.
		//r.Use(middleware.Timeout(300 * time.Second))

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
		errSimpleMapping := map[string]map[string]string{}
		for _, v := range mappingDef.ErrSimple {
			key, ok := v["error_key"]
			if !ok {
				continue
			}
			errSimpleMapping[key] = v
		}

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
			contextModel := container.NewModel()
			if err := h.convertQueryStringAndJsonRequestModel(request, body, contextModel, mappingDef, container); err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			// run process
			if err := fn(contextModel); err != nil {
				// handling error simple
				if flowErr, ok := err.(*pluginapi.FlowError); ok {
					errMapping, ok := errSimpleMapping[flowErr.Key]
					if ok {
						r := map[string]interface{}{}
						messagePath, ok := errMapping["error_message"]
						if ok {
							if strings.HasPrefix(messagePath, ParamHttpBodyPrefix) {
								destPaths := rule.SplitFullPath(messagePath[len(ParamHttpBodyPrefix):])
								m := r
								for _, p := range destPaths[:len(destPaths)-1] {
									//FIXME need support the following data types: array
									nm, ok := m[p]
									if !ok {
										nm = map[string]interface{}{}
										m[p] = nm
									}
									m = nm.(map[string]interface{})
								}
								lastPath := destPaths[len(destPaths)-1]
								m[lastPath] = flowErr.Message
							} else {
								//FIXME support more data access, e.g. headers
							}
						}
						status, ok := errMapping["http/status"]
						if ok {
							code, err := strconv.Atoi(status)
							if err != nil {
								h._logger.Error("error processing error simple status code:", err)
								writer.WriteHeader(http.StatusInternalServerError)
								return
							}
							writer.Header().Add("Content-Type", "application/json")
							writer.WriteHeader(code)
						} else {
							writer.Header().Add("Content-Type", "application/json")
							writer.WriteHeader(http.StatusInternalServerError)
						}
						data, err := json.Marshal(r)
						if err == nil {
							_, err := writer.Write(data)
							if err != nil {
								h._logger.Error("write response error:", err)
							}
							return
						} else {
							h._logger.Error("json marshal failed:", err)
						}
					}
				}
				h._logger.Error("error processing:", err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			// convert response
			if data, err := h.convertJsonResponseModel(contextModel, mappingDef, container); err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				writer.Header().Add("Content-Type", "application/json")
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write(data)
				if err != nil {
					h._logger.Error("write response error:", err)
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

func (h *HttpRestServerGenerator) convertJsonResponseModel(m pluginapi.Model, def *pluginapi.MappingDefinition, container pluginapi.Container) ([]byte, error) {
	res := container.NewModel()
	if err := def.ResConverter(m, res); err != nil {
		return nil, err
	}
	obj := res.ToGeneralObject()
	if obj == nil {
		return nil, nil
	}
	objMap, ok := obj.(map[string]interface{})
	if !ok {
		return nil, errors.New("response object is not a map[string]interface{}")
	}
	httpObj, ok := objMap["http"]
	if !ok {
		return nil, nil
	}
	httpMap, ok := httpObj.(map[string]interface{})
	if !ok {
		return nil, errors.New("http object is not a map[string]interface{}")
	}
	bodyObject, ok := httpMap["body"]
	if !ok {
		return nil, nil
	}
	return json.Marshal(bodyObject)
}

func (h *HttpRestServerGenerator) convertQueryStringAndJsonRequestModel(request *http.Request, body []byte, m pluginapi.Model, def *pluginapi.MappingDefinition, container pluginapi.Container) error {
	httpObj := map[string]interface{}{}
	src := map[string]interface{}{
		"http": httpObj,
	}

	// prepare body
	{
		var b interface{}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &b); err != nil {
				h._logger.Error(err)
			}
			httpObj["body"] = b
		}
	}
	// prepare query string
	{
		parsed, err := url.ParseRequestURI(request.RequestURI)
		if err != nil {
			return err
		}
		qparsed, err := url.ParseQuery(parsed.RawQuery)
		if err != nil {
			return err
		}
		if len(qparsed) > 0 {
			queryStringMap := map[string]interface{}{}
			for k, v := range qparsed {
				if len(v) > 0 {
					queryStringMap[k] = v[0]
				}
			}
			httpObj["query_string"] = queryStringMap
		}
	}
	// prepare headers
	{
		header := request.Header
		headerMap := map[string]interface{}{}
		for k, v := range header {
			if len(v) > 0 {
				headerMap[k] = v[0]
			}
		}
		httpObj["header"] = headerMap
	}
	// convert
	srcModel, err := container.WrapReadonlyModelFromMap(src)
	if err != nil {
		return err
	}
	return def.ReqConverter(srcModel, m)
}
