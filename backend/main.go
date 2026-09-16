package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Multi-Window Media Sequencer Backend")
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	http.ListenAndServe(":8080", nil)
}
