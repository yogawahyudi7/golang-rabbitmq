//go:build ignore
// +build ignore

// This file demonstrates why for-range approach is NOT recommended
// for ticker monitoring with context cancellation.
//
// To run this example:
// go run -tags ignore alternative_approaches.go

package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	fmt.Println("🔍 Demonstrating for-range vs select approaches...")

	// Simulate quick cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()

	// Start both approaches
	monitorWithForRange(ctx)
	monitorWithSelect(ctx)

	// Wait and then cancel
	time.Sleep(1 * time.Second)
	fmt.Println("🏁 Demo complete")
}

// ❌ PROBLEMATIC: for range approach
func monitorWithForRange(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		fmt.Println("🔄 Starting for-range monitoring...")

		for tick := range ticker.C {
			select {
			case <-ctx.Done():
				fmt.Println("🛑 for-range: DELAYED cancellation detected")
				return
			default:
				fmt.Printf("📊 for-range tick at: %v\n", tick.Format("15:04:05.000"))
			}
		}
	}()
}

// ✅ BETTER: select-first approach
func monitorWithSelect(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		fmt.Println("✅ Starting select-based monitoring...")

		for {
			select {
			case <-ctx.Done():
				fmt.Println("🛑 select: IMMEDIATE cancellation detected")
				return
			case tick := <-ticker.C:
				fmt.Printf("📊 select tick at: %v\n", tick.Format("15:04:05.000"))
			}
		}
	}()
}

/*
Expected output shows that select-based approach responds
to cancellation immediately, while for-range has delay.

🔄 Starting for-range monitoring...
✅ Starting select-based monitoring...
📊 select tick at: 15:04:05.500
📊 for-range tick at: 15:04:05.500
🛑 select: IMMEDIATE cancellation detected
📊 for-range tick at: 15:04:06.000
🛑 for-range: DELAYED cancellation detected
🏁 Demo complete
*/
