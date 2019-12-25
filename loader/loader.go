package loader

import (
	"math/rand"
	"sync"
	"time"

	"github.com/google/logger"

	"github.com/ArtAndreev/service-process-modeling/client"
	"github.com/ArtAndreev/service-process-modeling/config"
	"github.com/ArtAndreev/service-process-modeling/service"
	"github.com/ArtAndreev/service-process-modeling/stat"
)

type Loader struct {
	cfg *config.Config

	services map[string]*service.Service
}

func New(cfg *config.Config, services map[string]*service.Service) *Loader {
	return &Loader{
		cfg: cfg,

		services: services,
	}
}

func (l *Loader) Run(wg *sync.WaitGroup) {
	loads := initLoads(l.cfg.Loads)

	loadCh := make(chan *client.Client)
	workers := l.runLoadWorkers(wg, loadCh)

	logger.Infof("started loading")

	began := time.Now()

	for {
		if len(loads) == 0 {
			break
		}

		index := rand.Int() % len(loads) // nolint:gosec
		load := loads[index]
		loadCh <- client.New(load.name, load.paths)

		load.cnt--

		if load.cnt <= 0 {
			loads = append(loads[:index], loads[index+1:]...)
		}
	}

	close(loadCh)

	// kostyl, because it's hard to know when all requests have been ended
	time.Sleep(5 * time.Second)

	for _, s := range l.services {
		close(s.In)
	}

	wg.Wait() // no need, because kostyl exists

	logger.Infof("ended loading, elapsed %s", time.Since(began))

	aggregateStats(l.services, workers) // TODO:
}

func (l *Loader) runLoadWorkers(wg *sync.WaitGroup, loadCh <-chan *client.Client) []*worker {
	wg.Add(l.cfg.Parallel)
	workers := make([]*worker, 0, l.cfg.Parallel)

	delay := 1 * time.Second / time.Duration(l.cfg.RPS)

	for i := 0; i < l.cfg.Parallel; i++ {
		w := newWorker(delay, loadCh, l.services)
		go w.Run(wg)

		workers = append(workers, w)
	}

	return workers
}

func aggregateStats(services map[string]*service.Service, workers []*worker) {
	for n, s := range services {
		logger.Infof("--- %s ---", n)
		logger.Infof("Successful:")
		for reqName, sucStat := range s.Stat.Successful {
			logger.Infof("%s: %d times", reqName, sucStat.Count)
		}
		logger.Infof("Failed:")
		for reqName, failStat := range s.Stat.Failed {
			logger.Infof("%s: sent to %s %d times", reqName, failStat.Node, failStat.Count)
		}
	}

	for i, w := range workers {
		logger.Infof("--- worker #%d ---", i)
		logger.Infof("Successful:")
		for reqName, sucStat := range w.Stat.Successful {
			logger.Infof("%s: %d times", reqName, sucStat.Count)
		}
		logger.Infof("Failed:")
		for reqName, failStat := range w.Stat.Failed {
			logger.Infof("%s: sent to %s %d times", reqName, failStat.Node, failStat.Count)
		}
	}
}

type worker struct {
	Stat stat.Stat

	ch       <-chan *client.Client
	services map[string]*service.Service

	delay time.Duration
}

func newWorker(delay time.Duration, ch <-chan *client.Client, services map[string]*service.Service) *worker {
	return &worker{
		Stat: stat.Stat{
			Successful: make(map[string]*stat.Next, 10),
			Failed:     make(map[string]*stat.Next, 10),
		},

		ch:       ch,
		services: services,

		delay: delay,
	}
}

func (w *worker) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for c := range w.ch {
		if p, ok := c.GetNextPath(); !ok {
			node, ok := w.Stat.Successful[c.GetRequestName()]
			if !ok {
				node = new(stat.Next)
				w.Stat.Successful[c.GetRequestName()] = node
			}
			node.Count++
			logger.Warningf("got load with nil path: %s", c.GetRequestName())
		} else {
			select {
			case w.services[p.Next].In <- c:
			default:
				node, ok := w.Stat.Failed[c.GetRequestName()]
				if !ok {
					node = &stat.Next{
						Node: p.Next,
					}
					w.Stat.Failed[c.GetRequestName()] = node
				}
				node.Count++
			}
		}

		time.Sleep(w.delay)
	}
}
