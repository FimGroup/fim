package target

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimapi/rule"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DatabaseOperationExec  = "exec"
	DatabaseOperationQuery = "query"

	SqlArgParameterPrefix    = "sql.args."   // index from 0 (1st arg passing to sql)
	SqlReturnParameterPrefix = "sql.result." // index from 0 (1st arg passing to sql)
	SqlAffectedRowCountKey   = "sql.affected_row_count"
)

type dbPgConnector struct {
	instName string
	ref      *struct {
		*pgxpool.Pool
		refCnt int
	}

	operation     string
	sql           string
	params        [][]string
	returnMapping map[int][]string
	definition    *pluginapi.MappingDefinition
}

func (c *dbPgConnector) Start() error {
	c.ref.refCnt++
	return nil
}

func (c *dbPgConnector) Stop() error {
	c.ref.refCnt--
	if c.ref.refCnt == 0 {
		c.ref.Pool.Close()
	}
	return nil
}

func (c *dbPgConnector) Reload() error {
	return nil
}

func (c *dbPgConnector) ConnectorName() string {
	return c.instName
}

func (c *dbPgConnector) InvokeFlow(s, d pluginapi.Model) error {
	//FIXME allow to configure timeout
	switch c.operation {
	case DatabaseOperationExec:
		sqlParam := make([]interface{}, len(c.params))
		for i, v := range c.params {
			if len(v) == 0 {
				//FIXME no arg in this position
				sqlParam[i] = nil
			} else {
				sqlParam[i] = s.GetFieldUnsafe0(v)
			}
		}
		tag, err := c.ref.Pool.Exec(context.Background(), c.sql, sqlParam...)
		if err != nil {
			return err
		}
		r := map[string]interface{}{
			SqlAffectedRowCountKey: tag.RowsAffected(),
		}
		if err := c.convertResponse(c.definition, d, r); err != nil {
			return err
		}
		return nil
	case DatabaseOperationQuery:
		sqlParam := make([]interface{}, len(c.params))
		for i, v := range c.params {
			if len(v) == 0 {
				//FIXME no arg in this position
				sqlParam[i] = nil
			} else {
				sqlParam[i] = s.GetFieldUnsafe0(v)
			}
		}
		rows, err := c.ref.Pool.Query(context.Background(), c.sql, sqlParam...)
		if err != nil {
			return err
		}
		defer rows.Close()
		//FIXME need support multiple rows
		for rows.Next() {
			vals, err := rows.Values()
			if err != nil {
				return err
			}
			for idx, val := range vals {
				paths, ok := c.returnMapping[idx]
				if ok {
					if err := d.AddOrUpdateField0(paths, val); err != nil {
						return err
					}
				}
			}
		}
		return nil
	default:
		return errors.New("unsupported operation:" + c.operation)
	}
}

func NewDatabasePostgresGenerator() pluginapi.TargetConnectorGenerator {
	return &dbPgConnectorGenerator{dbPoolMapping: map[string]*struct {
		*pgxpool.Pool
		refCnt int
	}{}}
}

type dbPgConnectorGenerator struct {
	dbPoolMapping map[string]*struct {
		*pgxpool.Pool
		refCnt int
	}
}

func (d *dbPgConnectorGenerator) GeneratorNames() []string {
	return []string{"database_postgres"}
}

func (d *dbPgConnectorGenerator) GenerateTargetConnectorInstance(options map[string]string, container pluginapi.Container, definition *pluginapi.MappingDefinition) (pluginapi.TargetConnector, error) {
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
	dbMaxConnStr, ok := options["database.max_conns"]
	dbMaxConn := 0
	if ok {
		v, err := strconv.Atoi(dbMaxConnStr)
		if err != nil {
			return nil, err
		} else {
			dbMaxConn = v
		}
	}
	dbMinConnStr, ok := options["database.min_conns"]
	dbMinConn := 0
	if ok {
		v, err := strconv.Atoi(dbMinConnStr)
		if err != nil {
			return nil, err
		} else {
			dbMinConn = v
		}
	}

	//FIXME maybe share the database pool rather than connector instance(meaning different instance name)
	//FIXME should/may be shared with instance name since pool minConns and maxConns can be configured individually beyond connection string
	p, ok := d.dbPoolMapping[dbConnStr]
	if !ok {
		config, err := pgxpool.ParseConfig(dbConnStr)
		if err != nil {
			return nil, err
		}
		{
			// custom configuration
			if dbMaxConn > 0 {
				config.MaxConns = int32(dbMaxConn)
			}
			if dbMinConn > 0 {
				config.MinConns = int32(dbMinConn)
			}
		}
		np, err := pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			return nil, err
		}
		n := &struct {
			*pgxpool.Pool
			refCnt int
		}{
			Pool:   np,
			refCnt: 0,
		}
		d.dbPoolMapping[dbConnStr] = n
		p = n
	}
	params, err := d.prepareArgMapping(definition)
	if err != nil {
		return nil, err
	}
	returnMapping, err := d.prepareReturnMapping(definition)
	if err != nil {
		return nil, err
	}
	connector := &dbPgConnector{
		instName:      fmt.Sprintf("%s:%s", dbOper, uuid.Must(uuid.NewV4()).String()),
		ref:           p,
		operation:     dbOper,
		sql:           dbSql,
		params:        params,
		returnMapping: returnMapping,
		definition:    definition,
	}

	return connector, nil
}

func (c *dbPgConnector) convertResponse(definition *pluginapi.MappingDefinition, d pluginapi.Model, r map[string]interface{}) error {
	for _, paramPair := range definition.Res {
		if len(paramPair) != 2 {
			return errors.New("paramPair should contains 2 params")
		}
		fp := paramPair[0]
		cp := paramPair[1]
		val, ok := r[cp]
		if !ok {
			continue
		}
		return d.AddOrUpdateField0(rule.SplitFullPath(fp), val)
	}
	return nil
}

func (d *dbPgConnectorGenerator) prepareArgMapping(definition *pluginapi.MappingDefinition) ([][]string, error) {
	paramIdxMapping := map[int][]string{}
	maxArgIdx := -1
	for _, paramPair := range definition.Req {
		if len(paramPair) != 2 {
			return nil, errors.New("paramPair should contains 2 params")
		}
		fp := paramPair[0]
		cp := paramPair[1]
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

func (d *dbPgConnectorGenerator) prepareReturnMapping(definition *pluginapi.MappingDefinition) (map[int][]string, error) {
	responseMapping := map[int][]string{}

	for _, paramPair := range definition.Res {
		if len(paramPair) != 2 {
			return nil, errors.New("paramPair should contains 2 params")
		}
		fp := paramPair[0]
		cp := paramPair[1]
		//FIXME currently only support return parameter index mapping
		if !strings.HasPrefix(cp, SqlReturnParameterPrefix) {
			continue
		}
		idxStr := cp[len(SqlReturnParameterPrefix):]
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return nil, err
		}
		responseMapping[idx] = rule.SplitFullPath(fp)
	}

	return responseMapping, nil
}
