package fimcore

import (
	"testing"

	"github.com/ThisIsSun/fim/fimapi/pluginapi/rule"
)

func TestSyncFlowEngineExample(t *testing.T) {
	flow, def, err := loadFlow()
	if err != nil {
		t.Fatal(err)
	}

	var modelInst = NewModelInst(def)
	if err := modelInst.addOrUpdateField(rule.SplitFullPath("user/password"), "password1"); err != nil {
		t.Fatal(err)
	}

	if modelInst.getField(rule.SplitFullPath("user/user_id")) != nil {
		t.Fatal("should not obtain any value")
	}
	if v, ok := modelInst.getField(rule.SplitFullPath("user/password")).(string); v != "password1" || !ok {
		t.Fatal("value is not expected")
	}

	if err := flow.FlowFn()(modelInst); err != nil {
		t.Fatal(err)
	}

	v := modelInst.getField(rule.SplitFullPath("user/user_id"))
	if vv, ok := v.(int64); vv != 123 || !ok {
		t.Log(v)
		t.Fatal("value is not expected from ModelInst")
	}
	if v := modelInst.getField(rule.SplitFullPath("user/password")); v != nil {
		t.Log(v)
		t.Fatal("value should not exist")
	}
}

func BenchmarkSyncFlowEngineExample(b *testing.B) {
	flow, def, err := loadFlow()
	if err != nil {
		b.Fatal(err)
	}
	var modelInst = NewModelInst(def)
	if err := modelInst.addOrUpdateField(rule.SplitFullPath("user/password"), "password1"); err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		if err := flow.FlowFn()(modelInst); err != nil {
			b.Fatal(err)
		}
	}
}
