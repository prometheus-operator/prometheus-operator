package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.ListenAndServe(":5001", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Alertmanager Notification Payload Received")
	}))
}
