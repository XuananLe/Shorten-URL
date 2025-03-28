package main

import (
	"fmt"
	"math/rand"
	"time"
	"os"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	readTarget := vegeta.Target{
		Method: "GET",
		URL:    "https://httpbin.org/get",
	}

	postTarget := vegeta.Target{
		Method: "POST",
		URL:    "https://httpbin.org/post",
		Body:   []byte(`{"key":"value"}`),
	}

	targeter := func(tgt *vegeta.Target) error {
		if rand.Float32() < 0.8 {
			*tgt = readTarget
		} else {
			*tgt = postTarget
		}
		return nil
	}

	rate := vegeta.Rate{Freq: 5000, Per: time.Second} 
	duration := 5 * time.Second
	resultsFile, _ := os.Create("results.bin")

	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics

	encoder := vegeta.NewEncoder(resultsFile)
	for res := range attacker.Attack(targeter, rate, duration, "80-20-load-test") {
		metrics.Add(res)
		if err := encoder.Encode(res); err != nil {
			panic(err)
		}
	}

	metrics.Close()

	fmt.Println("Results saved to results.bin")
	fmt.Printf("Mean latency: %s\n", metrics.Latencies.Mean)
	fmt.Printf("Success rate: %.2f%%\n", metrics.Success*100)
	fmt.Printf("Status codes: %v\n", metrics.StatusCodes)
}