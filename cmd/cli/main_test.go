package main

import "testing"

// Benchmark for Profile-Guided Optimization (PGO).
// Run with: just profile
// This generates cpuprofile.pprof for optimized builds.

func Benchmark_Main_With_No_Args_Should_Return_Without_Error(b *testing.B) {
	for b.Loop() {
		main()
	}
}
