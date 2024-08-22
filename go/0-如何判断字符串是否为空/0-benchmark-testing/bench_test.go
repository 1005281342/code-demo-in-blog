package demo_test

import "testing"

//go test -bench .
//goos: darwin
//goarch: amd64
//pkg: demo/0-benchmark-testing
//cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
//Benchmark_isEmptyByLength-12            218700448                5.491 ns/op
//Benchmark_isEmptyByStringCompare-12     216776608                5.594 ns/op
//PASS
//ok      demo/0-benchmark-testing        4.282s

func Benchmark_isEmptyByLength(t *testing.B) {
	for i := 0; i < t.N; i++ {
		runEmptyFunc(isEmptyByLength, cases...)
	}
}

func Benchmark_isEmptyByStringCompare(t *testing.B) {
	for i := 0; i < t.N; i++ {
		runEmptyFunc(isEmptyByStringCompare, cases...)
	}
}

func isEmptyByLength(s string) bool {
	return len(s) == 0
}

func isEmptyByStringCompare(s string) bool {
	return s == ""
}

var cases = []string{"", "a", "abc", "abcdefg"}

type isEmptyFunc func(string) bool

func runEmptyFunc(isEmpty isEmptyFunc, cases ...string) {
	for _, caseString := range cases {
		isEmpty(caseString)
	}
}
