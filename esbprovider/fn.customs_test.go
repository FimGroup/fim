package esbprovider

import (
	"testing"

	"esbconcept/components"
	"esbconcept/esbcore"
)

func init() {
	if err := esbcore.RegisterCustomGeneratorFunc("#print_obj", FnPrintObject); err != nil {
		panic(err)
	}
}

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

func loadFlow() (*esbcore.Flow, *esbcore.DataTypeDefinitions, error) {
	container := esbcore.NewContainer()
	if err := components.InitComponent(container); err != nil {
		return nil, nil, err
	}
	def, err := loadDef()
	if err != nil {
		return nil, nil, err
	}

	flow := esbcore.NewFlow(def, container)
	if err := flow.MergeToml(flowFileContent); err != nil {
		return nil, nil, err
	}

	return flow, def, nil
}

const flowFileContent = `
[in]
# flowmodels -> local parameters
# parameters that are not existing here will be ignored
# default value will be given based on the type of left if the left value not exist
"user/user_id" = "user_id"
"user/username" = "username"
"user/password" = "password"
"user/nickname" = "nickname"

[pre_out]
# "@remove" command will remove the given value from the flowmodels
"@remove" = "user"

[out]
# local parameters -> flowmodels
# possible paths: "user/username", "user/phone", "user/phone[1]" which are ALL valid paths(not valid path of definition)
# default value will be  based on the type of right if the left value not exist
"user_id" = "user/user_id"
"username" = "user/username"
"nickname" = "user/nickname"

[flow]
"steps" = [
    # invoke function - "@function" = [ parameter list ]
    { "@assign" = ["user_id", 123] },
    # invoke user defined function
    { "#print_obj" = ["user_id"] },
    # invoke flow(e.g. subflow/module)
    # trigger event
]

`

func loadDef() (*esbcore.DataTypeDefinitions, error) {
	def := esbcore.NewDataTypeDefinitions()
	if err := def.MergeToml(flowModelFileContent); err != nil {
		return nil, err
	}
	return def, nil
}

func TestCustomFn(t *testing.T) {
	flow, def, err := loadFlow()
	if err != nil {
		t.Fatal(err)
	}

	var modelInst = esbcore.NewModelInst(def)

	if err := flow.FlowFn()(modelInst); err != nil {
		t.Fatal(err)
	}
}
