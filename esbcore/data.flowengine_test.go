package esbcore

import "testing"

func TestSyncFlowEngineExample(t *testing.T) {
	flow, def, err := loadFlow()
	if err != nil {
		t.Fatal(err)
	}

	var modelInst = NewModelInst(def)

	if err := flow.FlowFn()(modelInst); err != nil {
		t.Fatal(err)
	}
}

func TestAsyncFlowEngineExample(t *testing.T) {
	panic(_IMPLEMENT_ME)
}
