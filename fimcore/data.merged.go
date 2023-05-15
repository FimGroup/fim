package fimcore

import (
	"bytes"

	"github.com/pelletier/go-toml/v2"
)

type MergedDefinition struct {
	Pipelines map[string]*Pipeline     `toml:"pipelines"`
	Flows     map[string]*templateFlow `toml:"flows"`
}

func LoadMergedDefinition(content string) (*MergedDefinition, error) {
	r := new(MergedDefinition)
	if err := toml.NewDecoder(bytes.NewBufferString(content)).DisallowUnknownFields().Decode(r); err != nil {
		return nil, err
	}
	return r, nil
}
