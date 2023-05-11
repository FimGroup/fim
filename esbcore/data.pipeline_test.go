package esbcore

import (
	"testing"

	"esbconcept/components"
)

var pipelineContent = `
[parameter]
# Declare the variables that are used by the pipeline
"inputs" = [
    "user"
]
"outputs" = [
    "user"
]
# Declare the operations that are used by the pipeline
"pre_outputs" = [
    { "@remove" = "user" }
]

[pipeline]
"source_connectors" = [
    # Source connector will accept inbound traffic and convert it to FlowModel then invoke the following pipeline steps
    # All the connectors here will trigger the same pipeline, which means one pipeline will allow several source connectors
    # @connector should be used to define the type of connector
    # Other key-value pairs are options passing to the connector
    # Note: the option key should follow the format of path format
    # Note2: the connector will also share the same FlowModel
    { "@connector" = "http", "http.listen" = "0.0.0.0:8081", "http.path" = "/hello" }
]
"steps" = [
    # Each step should have @flow as key with flow name as value
    # Other key-value pairs are options passing to the flow generator
    # Note: the option key should follow the format of path format
    #
    # For using target connector(which supports both invoking flow and triggering event)
    # '&' should be used at the beginning of the flow name
    #
    # Invoke Flow(e.g. subflow/module)
    { "@flow" = "register", "example.parameter1" = "value1" },
    # Trigger Event
    # Meaning this step should not have outputs or at least the outputs will be discarded
    # And this step can be invoked in parallel to the other steps
    { "#flow" = "send_register_notification" },
    # target connector
    { "@flow" = "&http", "http.method" = "GET", "http.url" = 'http://www.example.com' },
]

`

func TestNewPipeline(t *testing.T) {
	container := NewContainer()
	if err := components.InitComponent(container); err != nil {
		t.Fatal(err)
	}
	if err := container.LoadFlowModel(flowModelFileContent); err != nil {
		t.Fatal(err)
	}
	if err := container.LoadFlow("register", flowFileContent); err != nil {
		t.Fatal(err)
	}
	if err := container.LoadFlow("send_register_notification", flowFileContent); err != nil {
		t.Fatal(err)
	}

	p, err := NewPipeline(pipelineContent, container)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(p)
	if err := p.setupPipeline(); err != nil {
		t.Fatal(err)
	}

}
