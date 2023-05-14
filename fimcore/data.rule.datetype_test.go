package fimcore

import (
	"testing"

	"github.com/ThisIsSun/fim/fimapi"
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
	AssertTypeOfPath(t, def, "global_name", fimapi.DataTypeString, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/user_id", fimapi.DataTypeInt, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/username", fimapi.DataTypeString, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/password", fimapi.DataTypeString, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/nickname", fimapi.DataTypeString, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/is_login", fimapi.DataTypeBool, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/rating_core", fimapi.DataTypeFloat, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/country_code", fimapi.DataTypeString, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/area_code", fimapi.DataTypeString, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/phone_number", fimapi.DataTypeString, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login/lastLoginTime[0]", fimapi.DataTypeInt, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix[0]", fimapi.DataTypeFloat, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user", fimapi.DataTypeObject, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk", fimapi.DataTypeObject, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]", fimapi.DataTypeObject, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone", fimapi.DataTypeArray, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login", fimapi.DataTypeObject, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login/lastLoginTime", fimapi.DataTypeArray, fimapi.DataTypeInt)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix", fimapi.DataTypeArray, fimapi.DataTypeFloat)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix[0]", fimapi.DataTypeFloat, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]", fimapi.DataTypeObject, fimapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix", fimapi.DataTypeArray, fimapi.DataTypeUnavailable)
}

func AssertTypeOfPath(t *testing.T, def *DataTypeDefinitions, path string, expectedDateType, expectedPrimArrDataType fimapi.DataType) {
	if dt, pdt, err := def.TypeOfPath(path); err != nil || dt != expectedDateType || pdt != expectedPrimArrDataType {
		t.Fatalf("type of path:%s failed. context: %d %d %s", path, dt, pdt, err)
	}
}

func AssertNonExistOfPath(t *testing.T, def *DataTypeDefinitions, path string) {
	if dt, pdt, err := def.TypeOfPath(path); err == nil || dt != fimapi.DataTypeUnavailable || pdt != fimapi.DataTypeUnavailable {
		t.Fatalf("type of path:%s should not exist. context: %d %d %s", path, dt, pdt, err)
	}
}
