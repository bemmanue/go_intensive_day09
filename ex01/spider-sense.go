package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func main() {
	ch1 := make(chan string)
	done := make(chan struct{})

	handleSigterm(done)

	go sendURSs(ch1, done)

	ch2 := crawlWeb(ch1, done)
	for i := range ch2 {
		fmt.Println(i)
	}
}

func handleSigterm(done chan struct{}) {
	c := make(chan os.Signal)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(done)
	}()
}

func sendURSs(ch1 chan<- string, done chan struct{}) {
	for i := 10; i <= 200; i++ {
		time.Sleep(1 * time.Second)
		str := "https://rickandmortyapi.com/api/character/" + strconv.Itoa(i)
		select {
		case ch1 <- str:
		case <-done:
			close(ch1)
			os.Exit(1)
		}
	}
	close(ch1)
}

func crawlWeb(ch1 <-chan string, done chan struct{}) chan string {
	ch2 := make(chan string)
	maxGoroutines := 8
	guard := make(chan struct{}, maxGoroutines)

	go func() {
		group := sync.WaitGroup{}

		for i := range ch1 {
			guard <- struct{}{}
			group.Add(1)
			url := i
			go func() {
				getData(url, ch2, &group, done)
				<-guard
			}()
		}
		group.Wait()
		close(ch2)
	}()
	return ch2
}

func getData(url string, ch2 chan string, group *sync.WaitGroup, done chan struct{}) {
	r, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	select {
	case ch2 <- string(body):
	case <-done:
		close(ch2)
		os.Exit(1)
	}

	group.Done()
}
