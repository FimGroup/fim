package fimcore

import "github.com/ThisIsSun/fim/components"

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
]

`

func loadFlow() (*Flow, *DataTypeDefinitions, error) {
	container := NewContainer()
	if err := components.InitComponent(container); err != nil {
		return nil, nil, err
	}
	def, err := loadDef()
	if err != nil {
		return nil, nil, err
	}

	flow := NewFlow(def, container)
	if err := flow.MergeToml(flowFileContent); err != nil {
		return nil, nil, err
	}

	return flow, def, nil
}
