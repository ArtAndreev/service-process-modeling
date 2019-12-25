package config

import (
	"fmt"
	"time"

	vipercast "github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/ArtAndreev/service-process-modeling/client"
	"github.com/ArtAndreev/service-process-modeling/service"
)

type Config struct {
	Services map[string]*service.Config
	Loads    map[string]*client.Config

	Scale    time.Duration
	Parallel int
	RPS      int
}

func Get() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath("etc")
	v.AddConfigPath(".")

	v.SetDefault("scale", 1*time.Second)
	v.SetDefault("parallel", 1)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	cfg := new(Config)

	var err error

	rawMap, err := vipercast.ToStringMapE(v.Get("services"))
	if err != nil {
		return nil, fmt.Errorf("cannot parse services list: %w", err)
	}

	cfg.Services, err = parseServices(rawMap)
	if err != nil {
		return nil, fmt.Errorf("cannot parse services list: %w", err)
	}

	rawMap, err = vipercast.ToStringMapE(v.Get("load"))
	if err != nil {
		return nil, fmt.Errorf("cannot parse load list: %w", err)
	}

	cfg.Loads, err = parseLoads(rawMap)
	if err != nil {
		return nil, fmt.Errorf("cannot parse load list: %w", err)
	}

	cfg.Scale, err = vipercast.ToDurationE(v.Get("scale"))
	if err != nil {
		return nil, fmt.Errorf("cannot parse scale value: %w", err)
	}

	cfg.Parallel, err = vipercast.ToIntE(v.Get("parallel"))
	if err != nil {
		return nil, fmt.Errorf("cannot parse parallel value: %w", err)
	}

	cfg.RPS, err = vipercast.ToIntE(v.Get("rps"))
	if err != nil {
		return nil, fmt.Errorf("cannot parse rps value: %w", err)
	}

	return cfg, nil
}

func parseServices(raw map[string]interface{}) (map[string]*service.Config, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	cfgs := make(map[string]*service.Config, len(raw))

	for name, scfg := range raw {
		v := viper.New()
		v.SetDefault("max_client_conn", 1024)
		v.SetDefault("request_process_time", 1*time.Millisecond)
		v.SetDefault("response_process_time", 1*time.Millisecond)
		v.SetDefault("parallel", 8)

		rawMap, err := vipercast.ToStringMapE(scfg)
		if err != nil {
			return nil, fmt.Errorf("service config: cannot parse config: %w", err)
		}

		srvCfg := new(service.Config)

		if err = v.MergeConfigMap(rawMap); err != nil {
			return nil, fmt.Errorf("service config: cannot merge map to config: %w", err)
		}

		srvCfg.MaxClientConn, err = vipercast.ToIntE(v.Get("max_client_conn"))
		if err != nil {
			return nil, fmt.Errorf("service config: cannot parse max_client_conn: %w", err)
		}

		srvCfg.RequestProcessTime, err = vipercast.ToDurationE(v.Get("request_process_time"))
		if err != nil {
			return nil, fmt.Errorf("service config: cannot parse request_process_time: %w", err)
		}

		srvCfg.ResponseProcessTime, err = vipercast.ToDurationE(v.Get("response_process_time"))
		if err != nil {
			return nil, fmt.Errorf("service config: cannot parse response_process_time: %w", err)
		}

		srvCfg.Parallel, err = vipercast.ToIntE(v.Get("parallel"))
		if err != nil {
			return nil, fmt.Errorf("service config: cannot parse parallel: %w", err)
		}

		cfgs[name] = srvCfg
	}

	return cfgs, nil
}

func parseLoads(raw map[string]interface{}) (map[string]*client.Config, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	cfgs := make(map[string]*client.Config, len(raw))

	for name, ccfg := range raw {
		v := viper.New()
		v.SetDefault("count", 1)

		rawMap, err := vipercast.ToStringMapE(ccfg)
		if err != nil {
			return nil, fmt.Errorf("client config: cannot parse config: %w", err)
		}

		clientCfg := new(client.Config)

		if err = v.MergeConfigMap(rawMap); err != nil {
			return nil, fmt.Errorf("client config: cannot merge map to config: %w", err)
		}

		rawValue := v.Get("count")
		if rawValue != nil {
			clientCfg.Count, err = vipercast.ToIntE(rawValue)
			if err != nil {
				return nil, fmt.Errorf("client config: cannot parse count: %w", err)
			}
		}

		rawValue = v.Get("request_path")
		if rawValue != nil {
			clientCfg.RequestPath, err = vipercast.ToStringSliceE(rawValue)
			if err != nil {
				return nil, fmt.Errorf("client config: cannot parse request_path: %w", err)
			}
		}

		rawValue = v.Get("response_path")
		if rawValue != nil {
			clientCfg.ResponsePath, err = vipercast.ToStringSliceE(rawValue)
			if err != nil {
				return nil, fmt.Errorf("client config: cannot parse response_path: %w", err)
			}
		}

		cfgs[name] = clientCfg
	}

	return cfgs, nil
}
