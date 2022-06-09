package main

import "roundRobin/roundRobin"

func main() {
	RoundRobin := roundRobin.RoundRobinLoadbalancer.New()
	RoundRobin.rrServer()
}
