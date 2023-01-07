package main

import (
	"fmt"
	"sync"
)

func generate(data interface{}) <-chan interface{} {
	channel := make(chan interface{})

	go func() {
		for i := 0; i < 3; i++ {
			channel <- data
		}
		close(channel)
	}()

	return channel
}

func multiplex(cs ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup

	out := make(chan interface{})

	send := func(c <-chan interface{}) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go send(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	//Test 1
	c1 := generate(100)
	c2 := generate(2)
	c3 := generate(3333)
	out := multiplex(c1, c2, c3)

	// Test 2
	//c1 := generate("Hello")
	//c2 := generate("There")
	//out := multiplex(c1, c2)

	// Test 3
	//c1 := generate("Hello world")
	//c2 := generate(100)
	//out := multiplex(c1, c2)

	for i := range out {
		fmt.Println(i)
	}
}
