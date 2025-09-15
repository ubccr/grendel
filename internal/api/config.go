package api

import (
	"bytes"

	"github.com/go-fuego/fuego"
	"github.com/spf13/viper"
)

type ConfigGetFileResponse struct {
	Config []byte `json:"config"`
}

type ConfigGetResponse struct {
	Config map[string]string `json:"config"`
}

type ConfigSetRequest struct {
	UpdateConfig map[string]string `json:"update_config"`
}

// TODO: change errors to httperror

func (h *Handler) ConfigGetFile(c fuego.ContextNoBody) (*ConfigGetFileResponse, error) {
	cfgType := c.QueryParam("type")
	if cfgType == "" {
		cfgType = "yaml"
	}

	tmpViper := viper.New()
	cfg, err := h.DB.ReadConfig()
	if err != nil {
		return nil, err
	}
	for key, val := range cfg {
		tmpViper.Set(key, val)
	}
	tmpViper.SetConfigType(cfgType)

	var output bytes.Buffer

	err = tmpViper.WriteConfigTo(&output)
	if err != nil {
		return nil, err
	}
	return &ConfigGetFileResponse{
		Config: output.Bytes(),
	}, nil
}

func (h *Handler) ConfigGet(c fuego.ContextNoBody) (ConfigGetResponse, error) {
	cfg := make(map[string]string, 0)

	keys := c.QueryParamArr("key")
	if len(keys) == 0 {
		keys = viper.AllKeys()
	}

	for _, k := range keys {
		cfg[k] = viper.GetString(k)
	}

	return ConfigGetResponse{Config: cfg}, nil
}

func (h *Handler) ConfigSet(c fuego.ContextWithBody[ConfigSetRequest]) (*GenericResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, err
	}
	tmpViper := viper.New()
	for k, v := range body.UpdateConfig {
		tmpViper.Set(k, v)
	}

	err = viper.MergeConfigMap(tmpViper.AllSettings())
	if err != nil {
		return nil, err
	}

	cfg := make(map[string]string, 0)
	for _, k := range tmpViper.AllKeys() {
		cfg[k] = tmpViper.GetString(k)
	}

	err = h.DB.WriteConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &GenericResponse{
		Title:   "Success",
		Detail:  "Successfully updated config",
		Changed: len(body.UpdateConfig),
	}, nil
}
