package target

const (
	DatabaseOperationExec  = "exec"
	DatabaseOperationQuery = "query"

	SqlArgParameterPrefix                = "sql.args."   // index from 0 (1st arg passing to sql)
	SqlReturnParameterPrefix             = "sql.result." // index from 0 (1st arg passing to sql)
	SqlAffectedRowCountKey               = "sql.affected_row_count"
	SqlReturnArrayParameterKey           = "sql.results"
	SqlReturnArrayElementParameterPrefix = "arg" //argX
)
