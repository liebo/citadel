package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"time"

	"github.com/citadel/citadel"
	"github.com/citadel/citadel/api"
	"github.com/citadel/citadel/cluster"
	"github.com/citadel/citadel/discovery"
	"github.com/citadel/citadel/scheduler"
)

// temporary register 2 nodes
func registerTestSlaves() {
	// Ubuntu boxes
	discovery.RegisterSlave("http://discovery.crosbymichael.com", "citadel_test", "cluster", "node1", "http://ec2-54-68-133-155.us-west-2.compute.amazonaws.com:4242")
	discovery.RegisterSlave("http://discovery.crosbymichael.com", "citadel_test", "cluster", "node2", "http://ec2-54-69-225-30.us-west-2.compute.amazonaws.com:4242")
	// Fedora
	discovery.RegisterSlave("http://discovery.crosbymichael.com", "citadel_test", "cluster", "node3", "http://ec2-54-69-11-29.us-west-2.compute.amazonaws.com:4242")
}

func main() {

	go func() {
		for {
			time.Sleep(25 * time.Second)
			registerTestSlaves()
		}
	}()

	registerTestSlaves()
	nodes, err := discovery.FetchSlaves("http://discovery.crosbymichael.com", "citadel_test", "cluster")
	if err != nil {
		log.Fatal(err)
	}

	var engines []*citadel.Engine
	for _, node := range nodes {
		engine := citadel.NewEngine(fmt.Sprintf("node-%x", md5.Sum([]byte(node))), node, 2048, 1)
		if err := engine.Connect(nil); err != nil {
			log.Fatalf("node.Connect: %v", err)
		}
		engines = append(engines, engine)
	}

	c, err := cluster.New(scheduler.NewResourceManager(), 2*time.Second, engines...)
	if err != nil {
		log.Fatalf("cluster.New: %v", err)
	}
	defer c.Close()

	scheduler := scheduler.NewMultiScheduler(&scheduler.LabelScheduler{}, &scheduler.HostScheduler{})

	if err := c.RegisterScheduler("service", scheduler); err != nil {
		log.Fatalf("c.RegisterScheduler: %v", err)
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
		containers, err := c.ListContainers()
		if err != nil {
			log.Fatal(err)
		}

	*/

	log.Fatal(api.ListenAndServe(c, ":4243"))
}
