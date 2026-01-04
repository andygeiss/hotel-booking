package main

import "testing"

// Every benchmark should follow the Benchmark_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the benchmarks to be easy to read and understand.

// We use the following benchmarks to create a baseline for our application.
// This will be used for Profile Guided Optimization (PGO) later.
// You can run these benchmarks with `just profile` and writes the results to `.cpuprofile.out`.

// During the build process, we will use the results of these benchmarks to optimize our application.
// You can run the build process with `just build` and it will use the -pgo flag to optimize the application.

func Benchmark_Main_With_Inbound_And_Outbound_Adapters_Should_Run_Efficiently(b *testing.B) {
	for b.Loop() {
		main()
	}
}
