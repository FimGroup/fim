package esbcore

import "testing"

func TestSyncFlowEngineExample(t *testing.T) {
	flow, def, err := loadFlow()
	if err != nil {
		t.Fatal(err)
	}

	var modelInst = NewModelInst(def)
	if err := modelInst.addOrUpdateField("user/password", "password1"); err != nil {
		t.Fatal(err)
	}

	if modelInst.getField("user/user_id") != nil {
		t.Fatal("should not obtain any value")
	}
	if v, ok := modelInst.getField("user/password").(string); v != "password1" || !ok {
		t.Fatal("value is not expected")
	}

	if err := flow.FlowFn()(modelInst); err != nil {
		t.Fatal(err)
	}

	v := modelInst.getField("user/user_id")
	if vv, ok := v.(int64); vv != 123 || !ok {
		t.Log(v)
		t.Fatal("value is not expected from ModelInst")
	}
	if v := modelInst.getField("user/password"); v != nil {
		t.Log(v)
		t.Fatal("value should not exist")
	}
}

func TestAsyncFlowEngineExample(t *testing.T) {
	panic(_IMPLEMENT_ME)
}
