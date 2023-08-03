package distribution

import (
	"encoding/json"
	"errors"

	"github.com/FimGroup/fim/fimapi/pluginapi"
	"github.com/FimGroup/fim/fimcore/modelinst"
)

func ModelToData(model pluginapi.Model) ([]byte, error) {
	modelEnc, ok := model.(pluginapi.ModelEncoding)
	if !ok {
		return nil, errors.New("err: Model is not ModelEncoding which cannot be used for ModelToData")
	}
	return modelEnc.ToToml()
}

func DataToModel(data []byte) (pluginapi.Model, error) {
	model := modelinst.ModelInstHelper{}.NewInst()
	modelEnc, ok := model.(pluginapi.ModelEncoding)
	if !ok {
		return nil, errors.New("err: Model is not ModelEncoding which cannot be used for DataToModel")
	}
	if err := modelEnc.FromToml(data); err != nil {
		return nil, err
	}
	return model, nil
}

func TransferModel(src, dst pluginapi.Model) error {
	modelCopy, ok := src.(pluginapi.ModelCopy)
	if !ok {
		return errors.New("err: src Model is not ModelCopy which cannot be used for TransferModel")
	}
	return modelCopy.Transfer(dst)
}

func FlowErrorToData(flowError *pluginapi.FlowError) ([]byte, error) {
	return json.Marshal(flowError)
}

func DataToFlowError(data []byte) (*pluginapi.FlowError, error) {
	r := new(pluginapi.FlowError)
	if err := json.Unmarshal(data, r); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func FlowStopToData(flowStop *pluginapi.FlowStop) ([]byte, error) {
	return json.Marshal(flowStop)
}

func DataToFlowStop(data []byte) (*pluginapi.FlowStop, error) {
	r := new(pluginapi.FlowStop)
	if err := json.Unmarshal(data, r); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}
