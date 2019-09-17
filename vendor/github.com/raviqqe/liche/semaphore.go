package main

type semaphore struct {
	channel chan bool
}

func newSemaphore(n int) semaphore {
	return semaphore{make(chan bool, n)}
}

func (s semaphore) Request() {
	s.channel <- true
}

func (s semaphore) Release() {
	<-s.channel
}
