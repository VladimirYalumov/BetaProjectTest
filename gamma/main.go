package main

import (
	"BetaProjectTest/response"
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/voting-stats", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(&response.VotingResponse{Result: "OK"})
		return
	})

	log.Fatal(http.ListenAndServe(":8090", nil))
}
