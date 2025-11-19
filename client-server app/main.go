package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

type resultBool struct {
	idx int
	val bool
}

type resultInt struct {
	idx int
	val int
}

type resultString struct {
	idx int
	val string
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, this is a simple handler!")
}

func ex2ProcessString(s string, idx int, wg *sync.WaitGroup, ch chan resultBool) {
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
		ch <- resultBool{idx: idx, val: false}
		return
	}

	// build uint64 number, detect overflow
	var n uint64
	for _, c := range digits {
		d := uint64(c - '0')
		newn := n*10 + d
		if newn < n { // overflow detected
			ch <- resultBool{idx: idx, val: false}
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
		ch <- resultBool{idx: idx, val: true}
	} else {
		ch <- resultBool{idx: idx, val: false}
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
	ch := make(chan resultBool, len(arr))
	var wg sync.WaitGroup
	wg.Add(len(arr))

	for i, s := range arr {
		go ex2ProcessString(s, i, &wg, ch)
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

func ex5ProcessString(s string, idx int, wg *sync.WaitGroup, ch chan resultInt) {
	defer wg.Done()
	for _, c := range s {
		if c != '0' && c != '1' {
			ch <- resultInt{idx: idx, val: -1}
			return
		}
	}
	var n int
	for i, c := range s {
		d := int(c - '0')
		newn := n + d*int(math.Pow(2, float64(len(s)-i-1)))
		if newn < n { // overflow detected
			ch <- resultInt{idx: idx, val: -1}
			return
		}
		n = newn
	}
	ch <- resultInt{idx: idx, val: n}
}

func ex5ArrayHandler(w http.ResponseWriter, r *http.Request) {
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
	ch := make(chan resultInt, len(arr))
	var wg sync.WaitGroup
	wg.Add(len(arr))

	for i, s := range arr {
		go ex5ProcessString(s, i, &wg, ch)
	}

	// Wait for all workers to finish then close the channel and collect results.
	wg.Wait()
	close(ch)

	processed := make([]int, len(arr))
	for res := range ch {
		processed[res.idx] = res.val
	}

	var result []int
	for _, v := range processed {
		if v != -1 {
			result = append(result, v)
		}
	}

	messages = append(messages, fmt.Sprintf("Server sends response to client %s", clientName))

	// Write back the response (including messages)
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{"original": arr, "processed": processed, "count": len(arr), "RESULT": result, "messages": messages}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func ex7ProcessString(s string, idx int, wg *sync.WaitGroup, ch chan resultString) {
	defer wg.Done()
	cnt := -1
	var res string
	for _, r := range s {
		if r >= '0' && r <= '9' {
			cnt = cnt*10 + int(r-'0')
			continue
		}
		for j := 0; j < cnt; j++ {
			res += string(r)
		}
		cnt = 0
	}
	ch <- resultString{idx: idx, val: res}
}

func ex7ArrayHandler(w http.ResponseWriter, r *http.Request) {
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
	ch := make(chan resultString, len(arr))
	var wg sync.WaitGroup
	wg.Add(len(arr))

	for i, s := range arr {
		go ex7ProcessString(s, i, &wg, ch)
	}

	// Wait for all workers to finish then close the channel and collect results.
	wg.Wait()
	close(ch)

	processed := make([]string, len(arr))
	for res := range ch {
		processed[res.idx] = res.val
	}

	messages = append(messages, fmt.Sprintf("Server sends response to client %s", clientName))

	// Write back the response (including messages)
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{"original": arr, "processed": processed, "count": len(arr), "RESULT": "No result value given by the exercise", "messages": messages}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func ex9ProcessString(s string, idx int, wg *sync.WaitGroup, ch chan resultBool) {
	defer wg.Done()
	var res bool = true
	cnt := 0
	for i, c := range s {
		switch c {
		case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
			if i%2 == 0 {
				cnt++
			} else {
				res = false
			}
		}
	}
	if cnt%2 != 0 {
		res = false
	}
	ch <- resultBool{idx: idx, val: res}
}

func ex9ArrayHandler(w http.ResponseWriter, r *http.Request) {
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
	ch := make(chan resultBool, len(arr))
	var wg sync.WaitGroup
	wg.Add(len(arr))

	for i, s := range arr {
		go ex9ProcessString(s, i, &wg, ch)
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

func ex14ProcessString(s string, idx int, wg *sync.WaitGroup, ch chan resultBool) {
	defer wg.Done()
	hasUpper, hasLower, hasDigit, hasSymbol := false, false, false, false

	for _, ch := range s {
		switch {
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		default:
			hasSymbol = true
		}
	}

	if hasUpper && hasLower && hasDigit && hasSymbol {
		ch <- resultBool{idx: idx, val: true} // accepted
	} else {
		ch <- resultBool{idx: idx, val: false} // not accepted
	}
}

func ex14ArrayHandler(w http.ResponseWriter, r *http.Request) {
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
	ch := make(chan resultBool, len(arr))
	var wg sync.WaitGroup
	wg.Add(len(arr))

	for i, s := range arr {
		go ex14ProcessString(s, i, &wg, ch)
	}

	// Wait for all workers to finish then close the channel and collect results.
	wg.Wait()
	close(ch)

	processed := make([]bool, len(arr))
	for res := range ch {
		processed[res.idx] = res.val
	}

	// For result, count the number of true values
	var res []string
	for i, v := range processed {
		if v == true {
			res = append(res, arr[i])
		}
	}

	messages = append(messages, fmt.Sprintf("Server sends response to client %s", clientName))

	// Write back the response (including messages)
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{"original": arr, "processed": processed, "count": len(arr), "RESULT": res, "messages": messages}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/ex2", ex2ArrayHandler)
	http.HandleFunc("/ex5", ex5ArrayHandler)
	http.HandleFunc("/ex7", ex7ArrayHandler)
	http.HandleFunc("/ex9", ex9ArrayHandler)
	http.HandleFunc("/ex14", ex14ArrayHandler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      nil, // default mux
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 4 * time.Second}
	log.Fatal(srv.ListenAndServe())
}
