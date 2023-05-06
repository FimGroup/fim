package esbcore

const flowFileContent = `
[in]
"user/user_id" = "user_id"
"user/username" = "username"
"user/password" = "password"
"user/nickname" = "nickname"

[pre_out]
"@remove" = "user"

[out]
"user_id" = "user/user_id"
"username" = "user/username"
"nickname" = "user/nickname"

[flow]
"steps" = [
    { "@assign" = ["user_id", 123] }
]
`

func loadFlow() (*Flow, *DataTypeDefinitions, error) {
	def, err := loadDef()
	if err != nil {
		return nil, nil, err
	}

	flow := NewFlow(def)
	if err := flow.MergeToml(flowFileContent); err != nil {
		return nil, nil, err
	}

	return flow, def, nil
}
