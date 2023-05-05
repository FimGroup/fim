package esbcore

import "testing"

var flowModelFileContent = `
[model]
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

func TestNewDataTypeDefinitions(t *testing.T) {
	def := NewDataTypeDefinitions()
	if err := def.MergeToml(flowModelFileContent); err != nil {
		t.Fatal(err)
	}
	t.Log(def)

	AssertNonExistOfPath(t, def, "non_exist")
	AssertTypeOfPath(t, def, "global_name", DataTypeString, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/user_id", DataTypeInt, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/username", DataTypeString, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/password", DataTypeString, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/nickname", DataTypeString, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/is_login", DataTypeBool, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/rating_core", DataTypeFloat, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/country_code", DataTypeString, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/area_code", DataTypeString, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]/phone_number", DataTypeString, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login/lastLoginTime[0]", DataTypeInt, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix[0]", DataTypeFloat, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user", DataTypeObject, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk", DataTypeObject, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone[0]", DataTypeObject, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/phone", DataTypeArray, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login", DataTypeObject, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/login/lastLoginTime", DataTypeArray, DataTypeInt)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix", DataTypeArray, DataTypeFloat)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]/sub_matrix[0]", DataTypeFloat, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix[0]", DataTypeArray, DataTypeUnavailable)
	AssertTypeOfPath(t, def, "user/risk/matrix", DataTypeArray, DataTypeUnavailable)
}

func AssertTypeOfPath(t *testing.T, def *DataTypeDefinitions, path string, expectedDateType, expectedPrimArrDataType DataType) {
	if dt, pdt, err := def.TypeOfPath(path); err != nil || dt != expectedDateType || pdt != expectedPrimArrDataType {
		t.Fatalf("type of path:%s failed. context: %d %d %s", path, dt, pdt, err)
	}
}

func AssertNonExistOfPath(t *testing.T, def *DataTypeDefinitions, path string) {
	if dt, pdt, err := def.TypeOfPath(path); err == nil || dt != DataTypeUnavailable || pdt != DataTypeUnavailable {
		t.Fatalf("type of path:%s should not exist. context: %d %d %s", path, dt, pdt, err)
	}
}
