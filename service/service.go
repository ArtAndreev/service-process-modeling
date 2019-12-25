package service

import (
	"os"
	"sync"
	"time"

	"github.com/google/logger"

	"github.com/ArtAndreev/service-process-modeling/client"
	"github.com/ArtAndreev/service-process-modeling/stat"
)

type Service struct {
	Name string
	Stat stat.Stat
	mu   sync.Mutex

	cfg *Config
	lg  *logger.Logger

	In         chan *client.Client
	routeTable map[string]chan<- *client.Client
}

func New(name string, config *Config) *Service {
	return &Service{
		Name: name,
		Stat: stat.Stat{
			Successful: make(map[string]*stat.Next, 10),
			Failed:     make(map[string]*stat.Next, 10),
		},

		cfg: config,
		lg:  logger.Init(name, false, false, os.Stdout),

		In: make(chan *client.Client, config.MaxClientConn),
	}
}

func (s *Service) Run(wg *sync.WaitGroup, routeTable map[string]chan<- *client.Client) {
	s.routeTable = routeTable

	wg.Add(s.cfg.Parallel)

	for i := 0; i < s.cfg.Parallel; i++ {
		go s.runWorker(wg)
	}
}

func (s *Service) runWorker(wg *sync.WaitGroup) {
	defer wg.Done()

	for c := range s.In {
		switch c.GetPathType() {
		case client.PathRequest:
			time.Sleep(s.cfg.RequestProcessTime)
		case client.PathResponse:
			time.Sleep(s.cfg.ResponseProcessTime)
		default:
			panic("unknown path type")
		}

		if p, ok := c.GetNextPath(); ok {
			select {
			case s.routeTable[p.Next] <- c:
			default:
				s.mu.Lock()
				node, ok := s.Stat.Failed[c.GetRequestName()]
				if !ok {
					node = &stat.Next{
						Node: p.Next,
					}
					s.Stat.Failed[c.GetRequestName()] = node
				}
				node.Count++
				s.mu.Unlock()
			}
		} else {
			s.mu.Lock()
			node, ok := s.Stat.Successful[c.GetRequestName()]
			if !ok {
				node = &stat.Next{
					Node: p.Next,
				}
				s.Stat.Successful[c.GetRequestName()] = node
			}
			node.Count++
			s.mu.Unlock()
		}
	}
}
