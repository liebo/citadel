package main

import (
	"log"
	"time"

	"github.com/citadel/citadel"
	"github.com/citadel/citadel/api"
	"github.com/citadel/citadel/cluster"
	"github.com/citadel/citadel/scheduler"
)

type logHandler struct {
}

func (l *logHandler) Handle(e *citadel.Event) error {
	log.Printf("type: %s time: %s image: %s container: %s\n",
		e.Type, e.Time.Format(time.RubyDate), e.Container.Image.Name, e.Container.ID)

	return nil
}

func main() {
	cluster1 := &citadel.Engine{
		ID:     "cluster-1",
		Addr:   "http://ec2-54-68-133-155.us-west-2.compute.amazonaws.com:4242",
		Memory: 2048,
		Cpus:   1,
	}
	cluster2 := &citadel.Engine{
		ID:     "cluster-2",
		Addr:   "http://ec2-54-69-225-30.us-west-2.compute.amazonaws.com:4242",
		Memory: 2048,
		Cpus:   1,
	}

	if err := cluster1.Connect(nil); err != nil {
		log.Fatalf("cluster1.Connect: %v", err)
	}
	if err := cluster2.Connect(nil); err != nil {
		log.Fatalf("cluster2.Connect: %v", err)
	}

	c, err := cluster.New(scheduler.NewResourceManager(), 2*time.Second, cluster1, cluster2)
	if err != nil {
		log.Fatalf("cluster.New: %v", err)
	}
	defer c.Close()

	if err := c.RegisterScheduler("service", &scheduler.LabelScheduler{}); err != nil {
		log.Fatalf("c.RegisterScheduler: %v", err)
	}

	if err := c.Events(&logHandler{}); err != nil {
		log.Fatalf("c.Events: %v", err)
	}

	/*	for {
			fmt.Println("")
			for _, containers := range c.Containers {
				for _, container := range containers {
					fmt.Println(container)
				}
			}
			time.Sleep(2 * time.Second)
		}
	*/
	/*
		image := &citadel.Image{
			Name:   "redis",
			Memory: 256,
			Cpus:   0.4,
			Type:   "service",
		}

		for i := 0; i < 2; i++ {
			container, err := c.Start(image, true)
			if err != nil {
				log.Fatalf("c.Start: %v", err)
			}

			log.Printf("ran container %s\n", container.ID)
		}
		containers, err := c.ListContainers(false)
		if err != nil {
			log.Fatal(err)
		}

	*/

	log.Fatal(api.ListenAndServe(c, ":4243"))
}
