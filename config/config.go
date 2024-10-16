package config

import (
	"encoding/json"
	"net"
	"path/filepath"
	"strconv"
)

var DefaultConfig = Config{
	Output: ".",
}

type Config struct {
	Bind     string      `json:"bind,omitempty"`
	Port     uint16      `json:"port,omitempty"`
	TlsCert  string      `json:"tls_cert,omitempty"`
	TlsKey   string      `json:"tls_key,omitempty"`
	Output   string      `json:"output,omitempty"`
	Reversed Value[bool] `json:"reversed,omitempty"`
}

type Value[T any] struct {
	Value T
	IsSet bool
}

func (v Value[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}

func (v *Value[T]) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &v.Value); err != nil {
		return err
	}
	v.IsSet = true
	return nil
}

func (v Value[T]) Override(v2 Value[T]) Value[T] {
	if v2.IsSet {
		return v2
	}
	return v
}

func (c *Config) ByKey(key string, value ...any) any {
	switch key {
	case "bind":
		return set(&c.Bind, value...)
	case "port":
		return set(&c.Port, value...)
	case "tls_cert":
		return set(&c.TlsCert, value...)
	case "tls_key":
		return set(&c.TlsKey, value...)
	case "output":
		return set(&c.Output, value...)
	case "reversed":
		return set(&c.Reversed, value...)
	}
	panic("unknown config key: " + key)
}

func set[T any](dst *T, value ...any) T {
	if len(value) > 0 {
		var ok bool
		*dst, ok = value[0].(T)
		if !ok {
			panic("invalid type")
		}
	}
	return *dst
}

func (c Config) Addr() string {
	return net.JoinHostPort(c.Bind, strconv.FormatUint(uint64(c.Port), 10))
}

func (cfg *Config) Override(overrideCfg Config) {
	cfg.Port = override(cfg.Port, overrideCfg.Port)
	cfg.Bind = override(cfg.Bind, overrideCfg.Bind)
	cfg.TlsKey = override(cfg.TlsKey, overrideCfg.TlsKey)
	cfg.TlsCert = override(cfg.TlsCert, overrideCfg.TlsCert)
	cfg.Output = override(cfg.Output, overrideCfg.Output)
	cfg.Reversed = cfg.Reversed.Override(overrideCfg.Reversed)
}

func override[T comparable](v, o T) T {
	var empty T
	if o != empty {
		return o
	}
	return v
}

func (cfg *Config) Normalize() {
	if cfg.Output == "" || filepath.IsAbs(cfg.Output) {
		return
	}
	absout, err := filepath.Abs(cfg.Output)
	if err != nil {
		panic(err)
	}
	cfg.Output = absout
}
