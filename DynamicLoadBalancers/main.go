package main

// import "example.com/loadbalancers/roundRobin"
// import "example.com/loadbalancers/roundRobin"
// import "example.com/loadbalancers/activeCheck"
import "example.com/loadbalancers/passiveCheck"

func main() {
	// roundRobinLB := RoundRobin.New()
	// roundRobinLB.LbServer()

	// activeCheckLb := ActiveCheck.New()
	// activeCheckLb.LbServer()

	passiveCheckLb := PassiveCheck.New()
	passiveCheckLb.LbServer()
}
