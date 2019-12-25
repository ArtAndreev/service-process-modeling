package main

import (
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/google/logger"

	"github.com/ArtAndreev/service-process-modeling/client"
	"github.com/ArtAndreev/service-process-modeling/config"
	"github.com/ArtAndreev/service-process-modeling/loader"
	"github.com/ArtAndreev/service-process-modeling/service"
)

func main() {
	defer logger.Init("model", false, false, os.Stdout).Close()

	cfg, err := config.Get()
	if err != nil {
		logger.Fatalf("cannot get config: %s", err)
	}

	if len(cfg.Loads) == 0 {
		logger.Warningf("load list is empty")
		return
	}

	wg := new(sync.WaitGroup)

	services := runServices(wg, cfg.Services)

	rand.Seed(time.Now().UnixNano())

	l := loader.New(cfg, services)
	l.Run(wg)
}

func runServices(wg *sync.WaitGroup, cfgs map[string]*service.Config) map[string]*service.Service {
	services := make(map[string]*service.Service, len(cfgs))
	routeTable := make(map[string]chan<- *client.Client, len(cfgs))
	for name, s := range cfgs {
		srv := service.New(name, s)

		services[name] = srv
		routeTable[name] = srv.In
	}

	for _, s := range services {
		s.Run(wg, routeTable)
	}

	return services
}
