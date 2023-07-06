package target

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/FimGroup/fim/fimapi/basicapi"
	"github.com/FimGroup/fim/fimapi/pluginapi"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresConnectorGenerator() pluginapi.TargetConnectorGenerator {
	return &postgresConnectorGenerator{
		subConnectorGeneratorMap: map[string]*subPostgresConnectorGenerator{},
	}
}

type postgresConnectorGenerator struct {
	subConnectorGeneratorMap map[string]*subPostgresConnectorGenerator
}

func (p *postgresConnectorGenerator) OriginalGeneratorNames() []string {
	return []string{"database_postgres"}
}

func (p *postgresConnectorGenerator) GenerateTargetConnectorInstance(req pluginapi.TargetConnectorGenerateRequest) (pluginapi.TargetConnector, error) {
	return nil, errors.New("initialize instance is not supported since PostgresConnectorGenerator is an abstract template")
}

func (p *postgresConnectorGenerator) InitializeSubGeneratorInstance(req pluginapi.CommonTargetConnectorGenerateRequest) (pluginapi.TargetConnectorGenerator, error) {
	instanceName := req.InstanceName
	if _, ok := p.subConnectorGeneratorMap[instanceName]; ok {
		return nil, errors.New("already initialized sub generator instance of:" + instanceName)
	}

	inst, err := newSubPostgresConnectorGenerator(req.Options)
	if err != nil {
		return nil, err
	} else {
		p.subConnectorGeneratorMap[instanceName] = inst
		return inst, nil
	}
}

func (p *postgresConnectorGenerator) Startup() error {
	return nil
}

func (p *postgresConnectorGenerator) Stop() error {
	return nil
}

func newSubPostgresConnectorGenerator(options map[string]string) (*subPostgresConnectorGenerator, error) {
	dbConnStr, ok := options["database.connect_string"]
	if !ok {
		return nil, errors.New("database.connect_string is not set")
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

	return &subPostgresConnectorGenerator{
		pool: np,
	}, nil
}

type subPostgresConnectorGenerator struct {
	pool *pgxpool.Pool
}

func (s *subPostgresConnectorGenerator) OriginalGeneratorNames() []string {
	panic("unsupported OriginalGeneratorNames by sub postgres connector generator")
}

func (s *subPostgresConnectorGenerator) GenerateTargetConnectorInstance(req pluginapi.TargetConnectorGenerateRequest) (pluginapi.TargetConnector, error) {
	operation, ok := req.Options["database.operation"]
	if !ok {
		return nil, errors.New("database.operation is not set")
	}
	switch operation {
	case DatabaseOperationExec:
	case DatabaseOperationQuery:
	default:
		return nil, errors.New("database.operation is invalid")
	}
	sql, ok := req.Options["database.sql"]
	if !ok {
		return nil, errors.New("database.sql is not set")
	}

	//FIXME maybe share the database pool rather than connector instance(meaning different instance name)
	//FIXME should/may be shared with instance name since pool minConns and maxConns can be configured individually beyond connection string
	maxIdx, err := s.prepareArgMapping(req.Definition)
	if err != nil {
		return nil, err
	}

	reqConverter := req.Definition.ReqConverter
	reqMaxIdx := maxIdx
	resConverter := req.Definition.ResConverter

	return &pgConnector{
		sql:          sql,
		operation:    operation,
		container:    req.Container,
		pool:         s.pool,
		reqConverter: reqConverter,
		reqMaxIdx:    reqMaxIdx,
		resConverter: resConverter,
	}, nil
}

func (s *subPostgresConnectorGenerator) prepareArgMapping(definition *pluginapi.MappingDefinition) (int, error) {
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

func (s *subPostgresConnectorGenerator) InitializeSubGeneratorInstance(req pluginapi.CommonTargetConnectorGenerateRequest) (pluginapi.TargetConnectorGenerator, error) {
	return nil, errors.New("InitializeSubGeneratorInstance is not supported since SubPostgresConnectorGenerator should produce connector only")
}

func (s *subPostgresConnectorGenerator) Startup() error {
	return nil
}

func (s *subPostgresConnectorGenerator) Stop() error {
	if s.pool != nil {
		s.pool.Close()
	}
	return nil
}

type pgConnector struct {
	operation string
	sql       string
	container pluginapi.Container
	pool      *pgxpool.Pool

	reqConverter func(src, dst pluginapi.Model) error
	reqMaxIdx    int
	resConverter func(src, dst pluginapi.Model) error
}

func (p *pgConnector) Start() error {
	return nil
}

func (p *pgConnector) Stop() error {
	return nil
}

func (p *pgConnector) Reload() error {
	return nil
}

func (p *pgConnector) convertResponse(d pluginapi.Model, r map[string]interface{}) error {
	if m, err := p.container.WrapReadonlyModelFromMap(r); err != nil {
		return err
	} else {
		return p.resConverter(m, d)
	}
}

func (p *pgConnector) InvokeFlow(s, d pluginapi.Model) error {
	//FIXME allow to configure timeout
	switch p.operation {
	case DatabaseOperationExec:
		local := p.container.NewModel()
		if err := p.reqConverter(s, local); err != nil {
			return err
		}
		//FIXME req mapping can be optimized
		sqlParam := make([]interface{}, p.reqMaxIdx+1)
		for i := 0; i <= p.reqMaxIdx; i++ {
			sqlParam[i] = local.GetFieldUnsafe0([]string{fmt.Sprint(SqlArgParameterPrefix, i)})
		}
		tag, err := p.pool.Exec(context.Background(), p.sql, sqlParam...)
		if err != nil {
			return err
		}
		r := map[string]interface{}{
			SqlAffectedRowCountKey: tag.RowsAffected(),
		}
		//FIXME res mapping can be optimized
		if err := p.convertResponse(d, r); err != nil {
			return err
		}
		return nil
	case DatabaseOperationQuery:
		local := p.container.NewModel()
		if err := p.reqConverter(s, local); err != nil {
			return err
		}
		//FIXME req mapping can be optimized
		sqlParam := make([]interface{}, p.reqMaxIdx+1)
		for i := 0; i <= p.reqMaxIdx; i++ {
			sqlParam[i] = local.GetFieldUnsafe0([]string{fmt.Sprint(SqlArgParameterPrefix, i)})
		}
		rows, err := p.pool.Query(context.Background(), p.sql, sqlParam...)
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
			if err := p.convertResponse(d, r); err != nil {
				return err
			}
		}
		return nil
	default:
		return errors.New("unsupported operation:" + p.operation)
	}
}
