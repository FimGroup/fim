package esbcore

import (
	"testing"

	"esbconcept/esbapi"
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
	AssertTypeOfPath(t, def, "global_name", esbapi.DataTypeString, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/user_id", esbapi.DataTypeInt, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/username", esbapi.DataTypeString, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/password", esbapi.DataTypeString, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/nickname", esbapi.DataTypeString, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/is_login", esbapi.DataTypeBool, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/rating_core", esbapi.DataTypeFloat, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/country_code", esbapi.DataTypeString, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/area_code", esbapi.DataTypeString, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/phone_number", esbapi.DataTypeString, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login/lastLoginTime[0]", esbapi.DataTypeInt, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix[0]", esbapi.DataTypeFloat, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user", esbapi.DataTypeObject, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk", esbapi.DataTypeObject, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]", esbapi.DataTypeObject, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone", esbapi.DataTypeArray, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login", esbapi.DataTypeObject, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login/lastLoginTime", esbapi.DataTypeArray, esbapi.DataTypeInt)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix", esbapi.DataTypeArray, esbapi.DataTypeFloat)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix[0]", esbapi.DataTypeFloat, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]", esbapi.DataTypeObject, esbapi.DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix", esbapi.DataTypeArray, esbapi.DataTypeUnavailable)
}

func AssertTypeOfPath(t *testing.T, def *DataTypeDefinitions, path string, expectedDateType, expectedPrimArrDataType esbapi.DataType) {
	if dt, pdt, err := def.TypeOfPath(path); err != nil || dt != expectedDateType || pdt != expectedPrimArrDataType {
		t.Fatalf("type of path:%s failed. context: %d %d %s", path, dt, pdt, err)
	}
}

func AssertNonExistOfPath(t *testing.T, def *DataTypeDefinitions, path string) {
	if dt, pdt, err := def.TypeOfPath(path); err == nil || dt != esbapi.DataTypeUnavailable || pdt != esbapi.DataTypeUnavailable {
		t.Fatalf("type of path:%s should not exist. context: %d %d %s", path, dt, pdt, err)
	}
}
