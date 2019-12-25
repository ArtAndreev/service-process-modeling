package loader

import (
	"math/rand"

	"github.com/google/logger"

	"github.com/ArtAndreev/service-process-modeling/client"
)

type load struct {
	name  string
	paths []client.Path
	cnt   int
}

func initLoads(loadParams map[string]*client.Config) []*load {
	loads := make([]*load, 0, 10)

	for n, l := range loadParams {
		if l.Count != 0 {
			paths := make([]client.Path, 0, len(l.RequestPath)+len(l.ResponsePath))
			for _, next := range l.RequestPath {
				paths = append(paths, client.Path{
					PathType: client.PathRequest,
					Next:     next,
				})
			}

			for _, next := range l.ResponsePath {
				paths = append(paths, client.Path{
					PathType: client.PathResponse,
					Next:     next,
				})
			}

			loads = append(loads, &load{
				name:  n,
				paths: paths,
				cnt:   l.Count,
			})
		} else {
			logger.Warningf("got load cfg with 0 load count: %+v", l)
		}
	}

	// For better random component
	rand.Shuffle(len(loads), func(i, j int) { loads[i], loads[j] = loads[j], loads[i] })

	return loads
}
