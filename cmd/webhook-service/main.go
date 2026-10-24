package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	"webhook-retry-fintech/internal/infrai"
	"webhook-retry-fintech/internal/payment"
)

func main() {
	c, err := infrai.New()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		var e payment.Event
		if json.NewDecoder(r.Body).Decode(&e) != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		n, err := payment.Decide(e)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := c.QueuePublish("payments", n, "payment-"+e.ID); err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(n)
	})
	go func() {
		for {
			time.Sleep(10 * time.Second)
		}
	}()
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8080"
	}
	log.Printf("listening on :%s", addr)
	log.Fatal(http.ListenAndServe(":"+addr, nil))
}
