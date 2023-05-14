package fimcore

import (
	"testing"

	"github.com/ThisIsSun/fim/fimapi/pluginapi"
)

var flowModelFileContent = `
[model]
# path -> type(primitive type only)
# primitive type default value
# * string = ""
# * int = 0
# * bool = false
# * float = 0.0
# compound type default value
# * object = (not exist)
# * array = (not exist)
"global_name" = "string"
"user/user_id" = "int"
"user/username" = "string"
"user/password" = "string"
"user/nickname" = "string"
"user/is_login" = "bool"
"user/risk/rating_core" = "float"
"user/phone[]/country_code" = "string"
"user/phone[]/area_code" = "string"
"user/phone[]/phone_number" = "string"
"user/login/lastLoginTime[]" = "int"
"user/risk/matrix[]/sub_matrix[]" = "float"

`

func loadDef() (*DataTypeDefinitions, error) {
	def := NewDataTypeDefinitions()
	if err := def.MergeToml(flowModelFileContent); err != nil {
		return nil, err
	}
	return def, nil
}

func TestNewDataTypeDefinitions(t *testing.T) {
	def, err := loadDef()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(def)

	AssertNonExistOfPath(t, def, "non_exist")
	AssertTypeOfPath(t, def, "global_name", pluginapi.DataTypeString, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/user_id", pluginapi.DataTypeInt, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/username", pluginapi.DataTypeString, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/password", pluginapi.DataTypeString, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/nickname", pluginapi.DataTypeString, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/is_login", pluginapi.DataTypeBool, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/rating_core", pluginapi.DataTypeFloat, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/country_code", pluginapi.DataTypeString, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/area_code", pluginapi.DataTypeString, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/phone_number", pluginapi.DataTypeString, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login/lastLoginTime[0]", pluginapi.DataTypeInt, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix[0]", pluginapi.DataTypeFloat, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user", pluginapi.DataTypeObject, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk", pluginapi.DataTypeObject, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]", pluginapi.DataTypeObject, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone", pluginapi.DataTypeArray, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login", pluginapi.DataTypeObject, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login/lastLoginTime", pluginapi.DataTypeArray, pluginapi.DataTypeInt)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix", pluginapi.DataTypeArray, pluginapi.DataTypeFloat)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix[0]", pluginapi.DataTypeFloat, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]", pluginapi.DataTypeObject, pluginapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix", pluginapi.DataTypeArray, pluginapi.DataTypeUnavailable)
}

func AssertTypeOfPath(t *testing.T, def *DataTypeDefinitions, path string, expectedDateType, expectedPrimArrDataType pluginapi.DataType) {
	if dt, pdt, err := def.TypeOfPath(path); err != nil || dt != expectedDateType || pdt != expectedPrimArrDataType {
		t.Fatalf("type of path:%s failed. context: %d %d %s", path, dt, pdt, err)
	}
}

func AssertNonExistOfPath(t *testing.T, def *DataTypeDefinitions, path string) {
	if dt, pdt, err := def.TypeOfPath(path); err == nil || dt != pluginapi.DataTypeUnavailable || pdt != pluginapi.DataTypeUnavailable {
		t.Fatalf("type of path:%s should not exist. context: %d %d %s", path, dt, pdt, err)
	}
}
