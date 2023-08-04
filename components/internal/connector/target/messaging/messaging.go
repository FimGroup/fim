package messaging

import (
	"encoding/json"
	"errors"

	"github.com/FimGroup/fim/fimapi/pluginapi"

	"github.com/nats-io/nats.go"
)

const (
	HeaderContentType = "Content-Type"

	ContentTypeJson = "application/json"
)

type NatsMessagingTargetConnectorGenerator struct {
}

func NewNatsMessagingTargetConnectorGenerator() pluginapi.TargetConnectorGenerator {
	return &NatsMessagingTargetConnectorGenerator{}
}

func (n *NatsMessagingTargetConnectorGenerator) OriginalGeneratorNames() []string {
	return []string{"&event_nats"}
}

func (n *NatsMessagingTargetConnectorGenerator) GenerateTargetConnectorInstance(req pluginapi.TargetConnectorGenerateRequest) (pluginapi.TargetConnector, error) {
	url, ok := req.Options["nats.url"]
	if !ok {
		return nil, errors.New("nats.url is empty")
	}
	topic, ok := req.Options["nats.topic"]
	if !ok {
		return nil, errors.New("nats.topic is empty")
	}
	container := req.Container
	mapping := req.Definition

	return &NatsMessagingTargetConnector{
		url:       url,
		topic:     topic,
		container: container,
		mapping:   mapping,
	}, nil
}

func (n *NatsMessagingTargetConnectorGenerator) InitializeSubGeneratorInstance(req pluginapi.CommonTargetConnectorGenerateRequest) (pluginapi.TargetConnectorGenerator, error) {
	return nil, errors.New("event_nats does not support sub source generator")
}

func (n *NatsMessagingTargetConnectorGenerator) Startup() error {
	return nil
}

func (n *NatsMessagingTargetConnectorGenerator) Stop() error {
	return nil
}

type NatsMessagingTargetConnector struct {
	url       string
	topic     string
	container pluginapi.Container
	mapping   *pluginapi.MappingDefinition

	conn *nats.Conn
}

func (n *NatsMessagingTargetConnector) Start() error {
	conn, err := nats.Connect(n.url)
	if err != nil {
		return err
	}
	n.conn = conn

	return nil
}

func (n *NatsMessagingTargetConnector) Stop() error {
	if n.conn != nil {
		n.conn.Close()
	}
	return nil
}

func (n *NatsMessagingTargetConnector) Reload() error {
	return nil
}

func (n *NatsMessagingTargetConnector) InvokeFlow(s, d pluginapi.Model) error {
	model := n.container.NewModel()
	if err := n.mapping.ReqConverter(s, model); err != nil {
		return err
	}

	obj := model.ToGeneralObject()
	objMap, ok := obj.(map[string]interface{})
	if !ok {
		return errors.New("object is not map[string]interface{}")
	}
	bodyMap, ok := objMap["body"]
	if !ok {
		bodyMap = map[string]interface{}{}
	}

	// json serialization
	data, err := json.Marshal(bodyMap)
	if err != nil {
		return err
	}

	msg := nats.NewMsg(n.topic)
	msg.Header.Add(HeaderContentType, ContentTypeJson)
	msg.Data = data
	return n.conn.PublishMsg(msg)
}
