package distribution

import (
	"errors"
	"fmt"
	"time"

	"github.com/FimGroup/fim/fimapi/pluginapi"

	"github.com/FimGroup/logging"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

const (
	NatsFlowStopStatusCode  = "9998"
	NatsFlowErrorStatusCode = "9999"
)

type NatsFlowInvoker struct {
	appName         string
	addr            string
	pipelineMapping map[string]pluginapi.PipelineProcess
	reqTimeoutInSec int

	conn *nats.Conn
	srv  micro.Service

	_logger logging.Logger
}

func (n *NatsFlowInvoker) Metadata() pluginapi.FlowInvokerMeta {
	return pluginapi.NewFlowInvokerMeta("nats", true)
}

func (n *NatsFlowInvoker) AddPipeline(pipelineName string, process pluginapi.PipelineProcess) error {
	if _, ok := n.pipelineMapping[pipelineName]; ok {
		return errors.New("pipeline already exists:" + pipelineName)
	}
	n.pipelineMapping[pipelineName] = process
	return nil
}

func (n *NatsFlowInvoker) Invoke(pipelineFullName string, model pluginapi.Model) error {
	data, err := ModelToData(model)
	if err != nil {
		return err
	}
	reply, err := n.conn.Request(pipelineFullName, data, time.Duration(n.reqTimeoutInSec)*time.Second)
	if err != nil {
		return err
	}

	// handling FlowError/FlowStop/General error
	errorCode := reply.Header.Get(micro.ErrorCodeHeader)
	if errorCode != "" {
		errorCodeDescription := reply.Header.Get(micro.ErrorHeader)
		switch errorCode {
		case NatsFlowStopStatusCode:
			flowStop, err := DataToFlowStop(reply.Data)
			if err != nil {
				return errors.New("unmarshal FlowStop failed")
			}
			return flowStop
		case NatsFlowErrorStatusCode:
			flowError, err := DataToFlowError(reply.Data)
			if err != nil {
				return errors.New("unmarshal FlowError failed")
			}
			return flowError
		default:
			return errors.New(fmt.Sprintf("NatsFlowInvokerError:%s:%s", errorCode, errorCodeDescription))
		}
	}

	replyModel, err := DataToModel(reply.Data)
	if err != nil {
		return err
	}
	if err := TransferModel(replyModel, model); err != nil {
		return err
	}
	return nil
}

func (n *NatsFlowInvoker) StartFlowInvoker() error {
	conn, err := nats.Connect(n.addr)
	if err != nil {
		return err
	}
	n.conn = conn

	srv, err := micro.AddService(conn, micro.Config{
		Name:    n.appName,
		Version: "1.0.0",
	})
	if err != nil {
		return err
	}
	n.srv = srv

	// register pipeline
	{
		for pipelineName, process := range n.pipelineMapping {
			handler := func(pipeline pluginapi.PipelineProcess, pipelineName string) micro.HandlerFunc {
				return func(request micro.Request) {
					// recover from panic
					defer func() {
						if env := recover(); env != nil {
							n._logger.Error("handling panic by pipeline:", env)
							if err := request.Error("500", "recover from panic", nil); err != nil {
								n._logger.Error("nats micro respond error failed:", err)
							}
						}
					}()

					if n._logger.IsDebugEnabled() {
						n._logger.Debug("received request by pipeline:", pipelineName)
					}
					m, err := DataToModel(request.Data())
					if err != nil {
						n._logger.Error("nats micro DataToModel failed:", err)
						if err := request.Error("500", "deserialize request failed", nil); err != nil {
							n._logger.Error("nats micro respond error failed:", err)
						}
						return
					}
					if err := pipeline(m); err != nil {
						switch v := err.(type) {
						case *pluginapi.FlowError:
							// handling flow error
							if data, err := FlowErrorToData(v); err != nil {
								if err := request.Error("500", "FlowError marshalling failed", nil); err != nil {
									n._logger.Error("nats micro respond error failed:", err)
								}
							} else {
								if err := request.Error(NatsFlowErrorStatusCode, "trigger FlowError", data); err != nil {
									n._logger.Error("nats micro respond error failed:", err)
								}
							}
							return
						case *pluginapi.FlowStop:
							// handling flow stop
							if data, err := FlowStopToData(v); err != nil {
								if err := request.Error("500", "FlowStop marshalling failed", nil); err != nil {
									n._logger.Error("nats micro respond error failed:", err)
								}
							} else {
								if err := request.Error(NatsFlowStopStatusCode, "trigger FlowStop", data); err != nil {
									n._logger.Error("nats micro respond error failed:", err)
								}
							}
							return
						}
						n._logger.Error("nats micro pipeline=["+pipelineName+"] handling failed:", err)
						if err := request.Error("500", "handling request failed", nil); err != nil {
							n._logger.Error("nats micro respond error failed:", err)
						}
						return
					} else {
						data, err := ModelToData(m)
						if err != nil {
							n._logger.Error("nats micro ModelToData failed:", err)
							if err := request.Error("500", "serialize response failed", nil); err != nil {
								n._logger.Error("nats micro respond error failed:", err)
							}
							return
						} else {
							if err := request.Respond(data); err != nil {
								n._logger.Error("nats micro respond result failed:", err)
							}
							return
						}
					}
				}
			}(process, pipelineName)
			//Note: endpoint name cannot contain special character like '/' so here use appName instead
			//Note2: when calling the service, subject is used to identify the specific service
			if err := srv.AddEndpoint(n.appName, handler, micro.WithEndpointSubject(pipelineName)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *NatsFlowInvoker) StopFlowInvoker() error {
	n.conn.Close()
	return nil
}

func NewNatsFlowInvoker(appName, addresses string) pluginapi.FlowInvoker {
	return &NatsFlowInvoker{
		appName:         appName,
		addr:            addresses,
		pipelineMapping: map[string]pluginapi.PipelineProcess{},
		reqTimeoutInSec: 10,
		_logger:         logging.GetLoggerManager().GetLogger("FimSupport.Distribution.FlowInvoker.nats"),
	}
}
