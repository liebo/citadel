package cluster

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/citadel/citadel"
)

var (
	ErrEngineNotConnected = errors.New("engine is not connected to docker's REST API")
)

type Cluster struct {
	mux sync.Mutex

	Containers      map[*citadel.Engine][]*citadel.Container
	engines         map[string]*citadel.Engine
	schedulers      map[string]citadel.Scheduler
	resourceManager citadel.ResourceManager
}

func New(manager citadel.ResourceManager, update time.Duration, engines ...*citadel.Engine) (*Cluster, error) {
	c := &Cluster{
		Containers:      make(map[*citadel.Engine][]*citadel.Container),
		engines:         make(map[string]*citadel.Engine),
		schedulers:      make(map[string]citadel.Scheduler),
		resourceManager: manager,
	}

	for _, e := range engines {
		if !e.IsConnected() {
			return nil, ErrEngineNotConnected
		}

		c.engines[e.ID] = e
	}

	go func() {
		for {
			c.Update()
			time.Sleep(update)
		}
	}()

	return c, nil
}

func (c *Cluster) Update() error {
	for _, e := range c.engines {
		state, err := e.State()
		if err != nil {
			return err
		}
		c.Containers[state.Engine] = state.Containers
	}
	return nil
}

func (c *Cluster) Events(handler citadel.EventHandler) error {
	for _, e := range c.engines {
		if err := e.Events(handler); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) RegisterScheduler(tpe string, s citadel.Scheduler) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.schedulers[tpe] = s

	return nil
}

func (c *Cluster) AddEngine(e *citadel.Engine) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.engines[e.ID] = e

	return nil
}

func (c *Cluster) RemoveEngine(e *citadel.Engine) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	delete(c.engines, e.ID)

	return nil
}

// ListContainers returns all the containers running in the cluster
func (c *Cluster) ListContainers(all bool) ([]*citadel.Container, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	out := []*citadel.Container{}

	for _, e := range c.engines {
		containers, err := e.ListContainers(all)
		if err != nil {
			return nil, err
		}

		out = append(out, containers...)
	}

	return out, nil
}

func (c *Cluster) Kill(container *citadel.Container, sig int) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	engine := c.engines[container.Engine.ID]
	if engine == nil {
		return fmt.Errorf("engine with id %s is not in cluster", container.Engine.ID)
	}

	return engine.Kill(container, sig)
}

func (c *Cluster) Stop(container *citadel.Container) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	engine := c.engines[container.Engine.ID]
	if engine == nil {
		return fmt.Errorf("engine with id %s is not in cluster", container.Engine.ID)
	}

	return engine.Stop(container)
}

func (c *Cluster) Restart(container *citadel.Container, timeout int) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	engine := c.engines[container.Engine.ID]
	if engine == nil {
		return fmt.Errorf("engine with id %s is not in cluster", container.Engine.ID)
	}

	return engine.Restart(container, timeout)
}

func (c *Cluster) Remove(container *citadel.Container) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	engine := c.engines[container.Engine.ID]
	if engine == nil {
		return fmt.Errorf("engine with id %s is not in cluster", container.Engine.ID)
	}

	return engine.Remove(container)
}

func (c *Cluster) Start(image *citadel.Image, pull bool) (*citadel.Container, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	var (
		accepted  = []*citadel.State{}
		scheduler = c.schedulers[image.Type]
	)

	if scheduler == nil {
		return nil, fmt.Errorf("no scheduler for type %s", image.Type)
	}

	for _, e := range c.engines {
		canrun, err := scheduler.Schedule(image, e)
		if err != nil {
			return nil, err
		}

		if canrun {
			state, err := e.State()
			if err != nil {
				return nil, err
			}

			accepted = append(accepted, state)
		}
	}

	if len(accepted) == 0 {
		return nil, fmt.Errorf("no eligible engines to run image")
	}

	container := &citadel.Container{
		Image: image,
		Name:  image.ContainerName,
	}

	s, err := c.resourceManager.PlaceContainer(container, accepted)
	if err != nil {
		return nil, err
	}

	if err := s.Engine.Start(container, pull); err != nil {
		return nil, err
	}

	return container, nil
}

// Engines returns the engines registered in the cluster
func (c *Cluster) Engines() []*citadel.Engine {
	c.mux.Lock()
	defer c.mux.Unlock()

	out := []*citadel.Engine{}

	for _, e := range c.engines {
		out = append(out, e)
	}

	return out
}

// Info returns information about the cluster
func (c *Cluster) ClusterInfo() (*citadel.ClusterInfo, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	containerCount := 0
	imageCount := 0
	engineCount := len(c.engines)
	totalCpu := 0.0
	totalMemory := 0.0
	reservedCpus := 0.0
	reservedMemory := 0.0
	for _, e := range c.engines {
		c, err := e.ListContainers(false)
		if err != nil {
			return nil, err
		}
		for _, cnt := range c {
			reservedCpus += cnt.Image.Cpus
			reservedMemory += cnt.Image.Memory
		}
		i, err := e.ListImages()
		if err != nil {
			return nil, err
		}
		containerCount += len(c)
		imageCount += len(i)
		totalCpu += e.Cpus
		totalMemory += e.Memory
	}

	return &citadel.ClusterInfo{
		Cpus:           totalCpu,
		Memory:         totalMemory,
		ContainerCount: containerCount,
		ImageCount:     imageCount,
		EngineCount:    engineCount,
		ReservedCpus:   reservedCpus,
		ReservedMemory: reservedMemory,
	}, nil
}

// Close signals to the cluster that no other actions will be applied
func (c *Cluster) Close() error {
	return nil
}
