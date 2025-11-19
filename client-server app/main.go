package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type result struct {
	idx int
	val bool
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, this is a simple handler!")
}

func ex1ProcessString(s string, idx int, wg *sync.WaitGroup, ch chan result) {
	defer wg.Done()
	// extract digits in order
	var digits []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			digits = append(digits, c)
		}
	}

	// no digits -> not a perfect square root
	if len(digits) == 0 {
		ch <- result{idx: idx, val: false}
		return
	}

	// build uint64 number, detect overflow
	var n uint64
	for _, c := range digits {
		d := uint64(c - '0')
		newn := n*10 + d
		if newn < n { // overflow detected
			ch <- result{idx: idx, val: false}
			return
		}
		n = newn
	}

	// integer square root via binary search (avoids floating point)
	var lo, hi uint64 = 0, n
	var root uint64
	for lo <= hi {
		mid := (lo + hi) / 2
		if mid == 0 {
			if n == 0 {
				root = 0
				break
			}
			lo = 1
			continue
		}
		if mid > n/mid { // mid*mid > n (avoid overflow)
			if mid == 0 {
				hi = 0
			} else {
				hi = mid - 1
			}
		} else {
			root = mid
			lo = mid + 1
		}
	}

	// check perfect square
	if root*root == n {
		ch <- result{idx: idx, val: true}
	} else {
		ch <- result{idx: idx, val: false}
	}
}

// arrayHandler accepts a POST request with a JSON array of strings
// and responds with a JSON object acknowledging the received data.
func ex2ArrayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var arr []string
	if err := json.Unmarshal(body, &arr); err != nil {
		http.Error(w, "invalid json: expected array of strings", http.StatusBadRequest)
		return
	}

	// Read client metadata from headers (optional)
	clientName := r.Header.Get("X-Client-Name")
	if clientName == "" {
		clientName = "unknown"
	}
	reqType := r.Header.Get("X-Request-Type")
	if reqType == "" {
		reqType = r.Method
	}

	// Messages exchanged (will be included in the response)
	messages := []string{}
	messages = append(messages, fmt.Sprintf("Server received request from client %s (type=%s) with %d items", clientName, reqType, len(arr)))

	// Process each item concurrently using goroutines.
	ch := make(chan result, len(arr))
	var wg sync.WaitGroup
	wg.Add(len(arr))

	for i, s := range arr {
		go ex1ProcessString(s, i, &wg, ch)
	}

	// Wait for all workers to finish then close the channel and collect results.
	wg.Wait()
	close(ch)

	processed := make([]bool, len(arr))
	for res := range ch {
		processed[res.idx] = res.val
	}

	// For result, count the number of true values
	trueCount := 0
	for _, v := range processed {
		if v == true {
			trueCount++
		}
	}

	messages = append(messages, fmt.Sprintf("Server sends response to client %s", clientName))

	// Write back the response (including messages)
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{"original": arr, "processed": processed, "count": len(arr), "RESULT": trueCount, "messages": messages}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/ex2", ex2ArrayHandler)
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      nil, // default mux
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 4 * time.Second}
	log.Fatal(srv.ListenAndServe())
}
