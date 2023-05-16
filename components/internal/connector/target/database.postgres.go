package target

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ThisIsSun/fim/fimapi/pluginapi"
	"github.com/ThisIsSun/fim/fimapi/rule"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DatabaseOperationExec  = "exec"
	DatabaseOperationQuery = "query"

	SqlArgParameterPrefix  = "sql.args." // index from 0 (1st arg passing to sql)
	SqlAffectedRowCountKey = "sql.affected_row_count"
)

type dbPgConnector struct {
	instName string
	p        *pgxpool.Pool
	ref      *struct {
		connector *dbPgConnector
		refCnt    int
	}

	operation  string
	sql        string
	params     [][]string
	definition *pluginapi.MappingDefinition
}

func (c *dbPgConnector) Start() error {
	c.ref.refCnt++
	return nil
}

func (c *dbPgConnector) Stop() error {
	c.ref.refCnt--
	if c.ref.refCnt == 0 {
		c.p.Close()
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
	switch c.operation {
	case DatabaseOperationExec:
		//FIXME allow to configure timeout
		sqlParam := make([]interface{}, len(c.params))
		for i, v := range c.params {
			if len(v) == 0 {
				//FIXME no arg in this position
				sqlParam[i] = nil
			} else {
				sqlParam[i] = s.GetFieldUnsafe(v)
			}
		}
		tag, err := c.p.Exec(context.Background(), c.sql, sqlParam...)
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
		//FIXME support database query
		return errors.New("not support operation:" + DatabaseOperationQuery)
	default:
		return errors.New("unsupported operation:" + c.operation)
	}
}

func NewDatabasePostgresGenerator() pluginapi.TargetConnectorGenerator {
	return &dbPgConnectorGenerator{dbPgMapping: map[string]*struct {
		connector *dbPgConnector
		refCnt    int
	}{}}
}

type dbPgConnectorGenerator struct {
	dbPgMapping map[string]*struct {
		connector *dbPgConnector
		refCnt    int
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

	//FIXME maybe share the database pool rather than connector instance(meaning different instance name)
	p, ok := d.dbPgMapping[dbConnStr]
	if !ok {
		np, err := pgxpool.New(context.Background(), dbConnStr)
		if err != nil {
			return nil, err
		}
		n := &struct {
			connector *dbPgConnector
			refCnt    int
		}{
			connector: &dbPgConnector{
				p:        np,
				instName: fmt.Sprintf("%s:%s", dbOper, dbSql),
			},
			refCnt: 0,
		}
		n.connector.ref = n
		d.dbPgMapping[dbConnStr] = n
		p = n
	}
	params, err := d.prepareArgMapping(definition)
	if err != nil {
		return nil, err
	}
	p.connector.sql = dbSql
	p.connector.params = params
	p.connector.definition = definition
	p.connector.operation = dbOper

	return p.connector, nil
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
