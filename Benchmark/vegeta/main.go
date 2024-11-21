package main

import (
	"fmt"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {
	rate := vegeta.Rate{Freq: 100, Per: time.Second}
	duration := 4 * time.Second
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    "https://go.dev/learn/",
	})
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	latencyPoints := make(plotter.XYs, 0) // Collect latencies over time

	startTime := time.Now()
	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		metrics.Add(res)

		// Add latency as a point for plotting
		elapsed := res.Timestamp.Sub(startTime).Seconds()
		latency := float64(res.Latency.Milliseconds())
		latencyPoints = append(latencyPoints, plotter.XY{X: elapsed, Y: latency})
	}
	metrics.Close()

	fmt.Printf("Max latencies: %s\n", metrics.Latencies.Max)
	fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)

	if err := plotLatencies(latencyPoints); err != nil {
		fmt.Printf("Error plotting latencies: %v\n", err)
	}
}

func plotLatencies(points plotter.XYs) error {
	p := plot.New()
	
	p.Title.Text = "Latency Over Time"
	p.X.Label.Text = "Time (seconds)"
	p.Y.Label.Text = "Latency (ms)"

	line, err := plotter.NewLine(points)
	if err != nil {
		return fmt.Errorf("could not create line plot: %v", err)
	}
	p.Add(line)

	outputFile := "latency_plot.png"
	if err := p.Save(10*vg.Inch, 4*vg.Inch, outputFile); err != nil {
		return fmt.Errorf("could not save plot: %v", err)
	}

	fmt.Printf("Latency plot saved to %s\n", outputFile)
	return nil
}
