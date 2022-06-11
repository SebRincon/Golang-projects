package main

// import "example.com/loadbalancers/roundRobin"
// import "example.com/loadbalancers/roundRobin"
import "example.com/loadbalancers/activeCheck"

func main() {
	// roundRobinLB := RoundRobin.New()
	// roundRobinLB.LbServer()

	activeCheckLb := ActiveCheck.New()
	activeCheckLb.LbServer()

}
