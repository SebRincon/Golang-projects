package main

// import "example.com/loadbalancers/roundRobin"
import "example.com/loadbalancers/roundRobin"

func main() {
	roundRobinLB := RoundRobin.New()
	roundRobinLB.LbServer()

}
