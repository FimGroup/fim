package target

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"esbconcept/esbapi"
	"esbconcept/esbapi/rule"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DatabaseOperationExec  = "exec"
	DatabaseOperationQuery = "query"

	SqlArgParameterPrefix  = "sql.args." // index from 0 (1st arg passing to sql)
	SqlAffectedRowCountKey = "sql.affected_row_count"
)

var dbPgMapping map[string]*pgxpool.Pool

func init() {
	dbPgMapping = map[string]*pgxpool.Pool{}
}

type dbPgConnector struct {
	p *pgxpool.Pool
}

func (d *dbPgConnector) Start() error {
	return nil
}

func (d *dbPgConnector) Stop() error {
	d.p.Close()
	return nil
}

func (d *dbPgConnector) Reload() error {
	return nil
}

func databasePostgresGenerator(options map[string]string, container esbapi.Container, definition *esbapi.MappingDefinition) (*struct {
	esbapi.Connector
	esbapi.ConnectorFlow
	InstanceName string
}, error) {
	dbConnStr, ok := options["database.connect_string"]
	if !ok {
		return nil, errors.New("database.connect_string is not set")
	}
	dbOper, ok := options["database.operation"]
	if !ok {
		return nil, errors.New("database.operation is not set")
	}
	switch dbOper {
	case DatabaseOperationExec:
	case DatabaseOperationQuery:
	default:
		return nil, errors.New("database.operation is invalid")
	}
	dbSql, ok := options["database.sql"]
	if !ok {
		return nil, errors.New("database.sql is not set")
	}

	p, ok := dbPgMapping[dbConnStr]
	if !ok {
		np, err := pgxpool.New(context.Background(), dbConnStr)
		if err != nil {
			return nil, err
		}
		dbPgMapping[dbConnStr] = np
		p = np
	}
	params, err := prepareArgMapping(definition)
	if err != nil {
		return nil, err
	}

	cr := &dbPgConnector{p: p}
	instName := fmt.Sprintf("%s:%s", dbOper, dbSql)
	var f esbapi.ConnectorFlow
	switch dbOper {
	case DatabaseOperationExec:
		f = func(s, d esbapi.Model) error {
			//FIXME allow to configure timeout
			sqlParam := make([]interface{}, len(params))
			for i, v := range params {
				if len(v) == 0 {
					//FIXME no arg in this position
					sqlParam[i] = nil
				} else {
					sqlParam[i] = s.GetFieldUnsafe(v)
				}
			}
			tag, err := p.Exec(context.Background(), dbSql, sqlParam...)
			if err != nil {
				return err
			}
			r := map[string]interface{}{
				SqlAffectedRowCountKey: tag.RowsAffected(),
			}
			if err := convertResponse(definition, d, r); err != nil {
				return err
			}
			return nil
		}
	case DatabaseOperationQuery:
		//FIXME support database query
		return nil, errors.New("not support operation:" + DatabaseOperationQuery)
	}

	return &struct {
		esbapi.Connector
		esbapi.ConnectorFlow
		InstanceName string
	}{Connector: cr, ConnectorFlow: f, InstanceName: instName}, nil
}

func convertResponse(definition *esbapi.MappingDefinition, d esbapi.Model, r map[string]interface{}) error {
	for fp, cp := range definition.Res {
		val, ok := r[cp]
		if !ok {
			continue
		}
		return d.AddOrUpdateField0(rule.SplitFullPath(fp), val)
	}
	return nil
}

func prepareArgMapping(definition *esbapi.MappingDefinition) ([][]string, error) {
	paramIdxMapping := map[int][]string{}
	maxArgIdx := -1
	for fp, cp := range definition.Req {
		if !strings.HasPrefix(cp, SqlArgParameterPrefix) {
			continue
		}
		idxStr := cp[len(SqlArgParameterPrefix):]
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return nil, err
		}
		if idx > maxArgIdx {
			maxArgIdx = idx
		}
		paramIdxMapping[idx] = rule.SplitFullPath(fp)
	}
	if len(paramIdxMapping) != maxArgIdx+1 {
		return nil, errors.New("database argument count doesn't match")
	}
	var r [][]string
	for i := 0; i <= maxArgIdx; i++ {
		r = append(r, paramIdxMapping[i])
	}
	return r, nil
}
