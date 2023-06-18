package target

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/FimGroup/fim/fimapi/basicapi"
	"github.com/FimGroup/fim/fimapi/pluginapi"
)

const (
	DatabaseOperationExec  = "exec"
	DatabaseOperationQuery = "query"

	SqlArgParameterPrefix                = "sql.args."   // index from 0 (1st arg passing to sql)
	SqlReturnParameterPrefix             = "sql.result." // index from 0 (1st arg passing to sql)
	SqlAffectedRowCountKey               = "sql.affected_row_count"
	SqlReturnArrayParameterKey           = "sql.results"
	SqlReturnArrayElementParameterPrefix = "arg" //argX
)

type dbPgConnector struct {
	instName string
	ref      *struct {
		*pgxpool.Pool
		refCnt int
	}

	operation    string
	sql          string
	reqConverter func(src, dst pluginapi.Model) error
	reqMaxIdx    int
	resConverter func(src, dst pluginapi.Model) error
	definition   *pluginapi.MappingDefinition
	container    pluginapi.Container
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
		local := c.container.NewModel()
		if err := c.reqConverter(s, local); err != nil {
			return err
		}
		//FIXME req mapping can be optimized
		sqlParam := make([]interface{}, c.reqMaxIdx+1)
		for i := 0; i <= c.reqMaxIdx; i++ {
			sqlParam[i] = local.GetFieldUnsafe0([]string{fmt.Sprint(SqlArgParameterPrefix, i)})
		}
		tag, err := c.ref.Pool.Exec(context.Background(), c.sql, sqlParam...)
		if err != nil {
			return err
		}
		r := map[string]interface{}{
			SqlAffectedRowCountKey: tag.RowsAffected(),
		}
		//FIXME res mapping can be optimized
		if err := c.convertResponse(d, r); err != nil {
			return err
		}
		return nil
	case DatabaseOperationQuery:
		local := c.container.NewModel()
		if err := c.reqConverter(s, local); err != nil {
			return err
		}
		//FIXME req mapping can be optimized
		sqlParam := make([]interface{}, c.reqMaxIdx+1)
		for i := 0; i <= c.reqMaxIdx; i++ {
			sqlParam[i] = local.GetFieldUnsafe0([]string{fmt.Sprint(SqlArgParameterPrefix, i)})
		}
		rows, err := c.ref.Pool.Query(context.Background(), c.sql, sqlParam...)
		if err != nil {
			return err
		}
		defer rows.Close()
		//FIXME res mapping can be optimized
		{
			r := map[string]interface{}{}
			var resultArr []interface{}
			rowNum := 0
			for rows.Next() {
				vals, err := rows.Values()
				if err != nil {
					return err
				}
				record := map[string]interface{}{}
				for idx, v := range vals {
					if v == nil {
						continue
					}
					record[fmt.Sprint(SqlReturnArrayElementParameterPrefix, idx)] = basicapi.MustConvertPrimitive(v)
				}
				resultArr = append(resultArr, record)
				// Generate first line result
				if rowNum == 0 {
					for idx, v := range vals {
						if v == nil {
							continue
						}
						r[fmt.Sprint(SqlReturnParameterPrefix, idx)] = basicapi.MustConvertPrimitive(v)
					}
				}
				rowNum++
			}
			r[SqlReturnArrayParameterKey] = resultArr
			if err := c.convertResponse(d, r); err != nil {
				return err
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

func (d *dbPgConnectorGenerator) GenerateTargetConnectorInstance(req pluginapi.TargetConnectorGenerateRequest) (pluginapi.TargetConnector, error) {
	dbConnStr, ok := req.Options["database.connect_string"]
	if !ok {
		return nil, errors.New("database.connect_string is not set")
	}
	dbOper, ok := req.Options["database.operation"]
	if !ok {
		return nil, errors.New("database.operation is not set")
	}
	switch dbOper {
	case DatabaseOperationExec:
	case DatabaseOperationQuery:
	default:
		return nil, errors.New("database.operation is invalid")
	}
	dbSql, ok := req.Options["database.sql"]
	if !ok {
		return nil, errors.New("database.sql is not set")
	}
	dbMaxConnStr, ok := req.Options["database.max_conns"]
	dbMaxConn := 0
	if ok {
		v, err := strconv.Atoi(dbMaxConnStr)
		if err != nil {
			return nil, err
		} else {
			dbMaxConn = v
		}
	}
	dbMinConnStr, ok := req.Options["database.min_conns"]
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
			if int32(dbMaxConn) > 0 {
				config.MaxConns = int32(dbMaxConn)
			}
			if int32(dbMinConn) > 0 {
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
	maxIdx, err := d.prepareArgMapping(req.Definition)
	if err != nil {
		return nil, err
	}
	connector := &dbPgConnector{
		instName:     fmt.Sprintf("%s:%s", dbOper, uuid.Must(uuid.NewV4()).String()),
		ref:          p,
		operation:    dbOper,
		sql:          dbSql,
		reqConverter: req.Definition.ReqConverter,
		reqMaxIdx:    maxIdx,
		resConverter: req.Definition.ResConverter,
		definition:   req.Definition,
		container:    req.Container,
	}

	return connector, nil
}

func (c *dbPgConnector) convertResponse(d pluginapi.Model, r map[string]interface{}) error {
	if m, err := c.container.WrapReadonlyModelFromMap(r); err != nil {
		return err
	} else {
		return c.resConverter(m, d)
	}
}

func (d *dbPgConnectorGenerator) prepareArgMapping(definition *pluginapi.MappingDefinition) (int, error) {
	maxArgIdx := -1
	for _, path := range definition.ReqArgPaths {
		if !strings.HasPrefix(path, SqlArgParameterPrefix) {
			continue
		}
		idxStr := path[len(SqlArgParameterPrefix):]
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return 0, err
		}
		if idx > maxArgIdx {
			maxArgIdx = idx
		}
	}
	return maxArgIdx, nil
}
