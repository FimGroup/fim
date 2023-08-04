package messaging

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/FimGroup/fim/fimapi/pluginapi"

	"github.com/FimGroup/logging"

	"github.com/nats-io/nats.go"
)

const (
	HeaderContentType = "Content-Type"

	ContentTypeJson = "application/json"
)

type NatsMessagingSourceConnectorGenereator struct {
	_logger logging.Logger
}

func NewNatsMessagingSourceConnectorGenereator() pluginapi.SourceConnectorGenerator {
	return &NatsMessagingSourceConnectorGenereator{
		_logger: logging.GetLoggerManager().GetLogger("FimGroup.Component.MessagingSourceConnector"),
	}
}

func (n *NatsMessagingSourceConnectorGenereator) OriginalGeneratorNames() []string {
	return []string{"event_nats"}
}

func (n *NatsMessagingSourceConnectorGenereator) GenerateSourceConnectorInstance(req pluginapi.SourceConnectorGenerateRequest) (pluginapi.SourceConnector, error) {
	url, ok := req.Options["nats.url"]
	if !ok {
		return nil, errors.New("nats.url is empty")
	}
	topic, ok := req.Options["nats.topic"]
	if !ok {
		return nil, errors.New("nats.topic is empty")
	}
	group, ok := req.Options["nats.group"]
	if !ok {
		return nil, errors.New("nats.group is empty")
	}
	container := req.Container
	mapping := req.Definition

	return &NatsMessagingSourceConnector{
		url:       url,
		topic:     topic,
		group:     group,
		container: container,
		mapping:   mapping,
		_logger:   n._logger,
	}, nil
}

func (n *NatsMessagingSourceConnectorGenereator) InitializeSubGeneratorInstance(req pluginapi.CommonSourceConnectorGenerateRequest) (pluginapi.SourceConnectorGenerator, error) {
	return nil, errors.New("event_nats does not support sub source generator")
}

func (n *NatsMessagingSourceConnectorGenereator) Startup() error {
	return nil
}

func (n *NatsMessagingSourceConnectorGenereator) Stop() error {
	return nil
}

type NatsMessagingSourceConnector struct {
	url   string
	topic string
	group string

	container pluginapi.Container
	mapping   *pluginapi.MappingDefinition

	pipeline pluginapi.PipelineProcess

	conn *nats.Conn
	sub  *nats.Subscription

	_logger logging.Logger
}

func (n *NatsMessagingSourceConnector) Start() error {
	conn, err := nats.Connect(n.url)
	if err != nil {
		return err
	}
	n.conn = conn

	sub, err := conn.QueueSubscribe(n.topic, n.group, func(msg *nats.Msg) {
		// handle panic
		defer func() {
			if env := recover(); env != nil {
				n._logger.Error(fmt.Sprintf("handling panic/error by topic:[%s] group:[%s]", n.topic, n.group))
			}
		}()

		if n._logger.IsDebugEnabled() {
			n._logger.DebugF("handling event_nats by topic:[%s] group:[%s]", n.topic, n.group)
		}

		// deserialization
		contentType := msg.Header.Get(HeaderContentType)
		if contentType == "" {
			contentType = ContentTypeJson
		}
		var obj interface{}
		switch contentType {
		case ContentTypeJson:
			// json type
			if err := json.Unmarshal(msg.Data, &obj); err != nil {
				panic(err)
			}
		default:
			panic(errors.New("unknown content type:" + contentType))
		}
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			panic(errors.New("object is not map[string]interface{}"))
		}
		{
			newObjMap := make(map[string]interface{})
			newObjMap["body"] = objMap
			objMap = newObjMap
		}
		model, err := n.container.WrapReadonlyModelFromMap(objMap)
		if err != nil {
			panic(err)
		}
		dstModel := n.container.NewModel()
		if err := n.mapping.ReqConverter(model, dstModel); err != nil {
			panic(err)
		}

		// trigger pipeline process
		if err := n.pipeline(dstModel); err != nil {
			panic(err)
		}
	})
	if err != nil {
		return err
	}
	n.sub = sub

	return nil
}

func (n *NatsMessagingSourceConnector) Stop() error {
	if n.conn != nil {
		n.conn.Close()
	}
	return nil
}

func (n *NatsMessagingSourceConnector) Reload() error {
	return nil
}

func (n *NatsMessagingSourceConnector) BindPipeline(process pluginapi.PipelineProcess) error {
	n.pipeline = process
	return nil
}
