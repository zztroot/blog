package config

import (
"github.com/zztroot/rconfig"
)

type ConfigRead struct {
	RConfig *rconfig.ConfigStruct
	err error
}

func (c *ConfigRead) RConfigConn() (*rconfig.ConfigStruct, error){
	c.RConfig, c.err = rconfig.OpenConfig("config/config.ini")
	if c.err != nil {
		return nil, c.err
	}
	return c.RConfig, nil
}
