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

	"github.com/FimGroup/logging"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

const (
	ParamHttpBodyPrefix        = "http/body/"
	ParamHttpQueryStringPrefix = "http/query_string/"
	ParamHttpHeaderPrefix      = "http/header/"

	TypeHttpRest     = "http_rest"
	TypeHttpTemplate = "http_template"
)

type httpRestServerConnector struct {
	instName  string
	generator *HttpRestServerGenerator

	entryPoint func(process pluginapi.PipelineProcess) error
}

func (h *httpRestServerConnector) BindPipeline(process pluginapi.PipelineProcess) error {
	return h.entryPoint(process)
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

func (h *HttpRestServerGenerator) InitializeSubGeneratorInstance(req pluginapi.CommonSourceConnectorGenerateRequest) (pluginapi.SourceConnectorGenerator, error) {
	return nil, errors.New("InitializeSubGeneratorInstance is not supported by http source connector")
}

func (h *HttpRestServerGenerator) Startup() error {
	//TODO implement me
	panic("implement me")
}

func (h *HttpRestServerGenerator) OriginalGeneratorNames() []string {
	return []string{TypeHttpRest, TypeHttpTemplate}
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

func (h *HttpRestServerGenerator) addRestHandler(req pluginapi.SourceConnectorGenerateRequest, handleFunc http.HandlerFunc) error {
	ls, ok := req.Options["http.listen"]
	if !ok {
		return errors.New("need provide http.listen for http")
	}
	path, ok := req.Options["http.path"]
	if !ok {
		return errors.New("need provide http.path for http")
	}
	method, ok := req.Options["http.method"]
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

func (h *HttpRestServerGenerator) addTemplateHandler(req pluginapi.SourceConnectorGenerateRequest, fn pluginapi.PipelineProcess, def *pluginapi.MappingDefinition, errSimpleMapping map[string]map[string]string) error {
	ls, ok := req.Options["http.listen"]
	if !ok {
		return errors.New("need provide http.listen for http")
	}
	path, ok := req.Options["http.path"]
	if !ok {
		return errors.New("need provide http.path for http")
	}
	method, ok := req.Options["http.method"]
	method = strings.ToUpper(method)
	if !ok {
		return errors.New("need provide http.method for http")
	}
	resourceManagerName, ok := req.Options["http.resource_manager"]
	if !ok {
		return errors.New("no resource manager found")
	}
	templatePath, ok := req.Options["http.template_path"]
	if !ok {
		return errors.New("no template path found")
	}
	fileMgr := req.Application.GetFileResourceManager(resourceManagerName)
	if fileMgr == nil {
		return errors.New("cannot find file resource manager for http template loading:" + resourceManagerName)
	}
	tr, err := loadTemplate(templatePath, fileMgr)
	if err != nil {
		return err
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

	sendHtml := func(writer http.ResponseWriter, status int, obj interface{}) {
		writer.Header().Add("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(status)
		if err := tr.Render(writer, obj); err != nil {
			h._logger.Error("render and write response error:", err)
		}
	}
	sendError := func(writer http.ResponseWriter, status int) {
		writer.WriteHeader(status)
	}

	f := func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(body) > 0 {
			//FIXME need support more content-types
			contentType := request.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		// convert request
		contextModel := req.Container.NewModel()
		if err := h.convertQueryStringAndJsonRequestModel(request, body, contextModel, def, req.Container); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		// run process
		if err := fn(contextModel); err != nil {
			//FIXME handling error simple
			//FIXME need support template error rendering
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
						writer.Header().Set("Content-Type", "application/json; charset=utf-8")
						writer.WriteHeader(code)
					} else {
						writer.Header().Set("Content-Type", "application/json; charset=utf-8")
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
		if obj, err := h.convertTemplateObjectModel(contextModel, def, req.Container); err != nil {
			sendError(writer, http.StatusInternalServerError)
			return
		} else {
			sendHtml(writer, http.StatusOK, obj)
			return
		}
	}

	//FIXME may cause concurrent issue on adding new handler while processing requests
	switch method {
	case "GET":
		lstruct.Mux.Get(path, f)
	case "POST":
		lstruct.Mux.Post(path, f)
	case "PUT":
		lstruct.Mux.Put(path, f)
	case "DELETE":
		lstruct.Mux.Delete(path, f)
	case "HEAD":
		lstruct.Mux.Head(path, f)
	case "PATCH":
		lstruct.Mux.Patch(path, f)
	default:
		return errors.New(fmt.Sprintf("unable to register http path:%s method:%s", path, method))
	}
	return nil
}

func (h *HttpRestServerGenerator) GenerateSourceConnectorInstance(req pluginapi.SourceConnectorGenerateRequest) (pluginapi.SourceConnector, error) {
	entryPoint := func(fn pluginapi.PipelineProcess) error {
		mappingDef := req.Definition

		errSimpleMapping := map[string]map[string]string{}
		for _, v := range mappingDef.ErrSimple {
			key, ok := v["error_key"]
			if !ok {
				continue
			}
			errSimpleMapping[key] = v
		}

		// handling http template
		if strings.HasSuffix(req.Options["@connector"], TypeHttpTemplate) {
			if err := h.addTemplateHandler(req, fn, mappingDef, errSimpleMapping); err != nil {
				return err
			}
		} else if strings.HasSuffix(req.Options["@connector"], TypeHttpRest) {
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
				contextModel := req.Container.NewModel()
				if err := h.convertQueryStringAndJsonRequestModel(request, body, contextModel, mappingDef, req.Container); err != nil {
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
								writer.Header().Set("Content-Type", "application/json; charset=utf-8")
								writer.WriteHeader(code)
							} else {
								writer.Header().Set("Content-Type", "application/json; charset=utf-8")
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
				if data, err := h.convertJsonResponseModel(contextModel, mappingDef, req.Container); err != nil {
					writer.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					writer.Header().Set("Content-Type", "application/json; charset=utf-8")
					writer.WriteHeader(http.StatusOK)
					_, err := writer.Write(data)
					if err != nil {
						h._logger.Error("write response error:", err)
					}
					return
				}
			}
			if err := h.addRestHandler(req, f); err != nil {
				return err
			}
		} else {
			return errors.New("unknown connector type of http source connector:" + req.Options["@connector"])
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

func (h *HttpRestServerGenerator) convertTemplateObjectModel(m pluginapi.Model, def *pluginapi.MappingDefinition, container pluginapi.Container) (interface{}, error) {
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
	return bodyObject, nil
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
