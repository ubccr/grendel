package storetest

import (
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/internal/store"
)

// TODO: benchmark different batch sizes
var batchSizes []int = []int{5000}

type BenchTestSuite interface {
	SetFile(string)
	SetT(*testing.T)
	SetupTest()
	BenchmarkWriteNodes(size int, b *testing.B)
	BenchmarkWriteSingleNode(size int, b *testing.B)
	BenchmarkReadAll(size int, b *testing.B)
	BenchmarkFind(size int, b *testing.B)
	BenchmarkRandomReads(size int, b *testing.B)
	BenchmarkRandomWrites(size int, b *testing.B)
	BenchmarkResolveIP(size int, b *testing.B)
	BenchmarkReverseResolve(size int, b *testing.B)
}

func tempfile(b *testing.B) string {
	return path.Join(b.TempDir(), "grendel-benchmark.db")
}

// reportCPU runs a benchmark body and records what it cost in CPU, as custom
// metrics alongside ns/op.
//
// Under RunParallel, ns/op is wall time, which says nothing about how many
// cores the work consumes. That distinction is the whole story of the 2026-08
// DNS incident: the server was burning nine of ten cores to serve ninety five
// queries a second, and it was almost entirely system time. A benchmark that
// only reports wall time cannot show that, and cannot show it coming back.
//
// Metrics emitted:
//
//	cpu-us/op   CPU microseconds per operation, the efficiency number
//	%cpu        cores consumed, 100 per core
//	%sys        share of CPU spent in the kernel
func reportCPU(b *testing.B, fn func()) {
	startUser, startSys, ok := cpuTime()
	start := time.Now()

	fn()

	wall := time.Since(start)
	endUser, endSys, endOK := cpuTime()
	if !ok || !endOK {
		return
	}

	total := (endUser - startUser) + (endSys - startSys)
	sys := endSys - startSys

	if b.N > 0 {
		b.ReportMetric(float64(total.Microseconds())/float64(b.N), "cpu-us/op")
	}
	if wall > 0 {
		b.ReportMetric(100*total.Seconds()/wall.Seconds(), "%cpu")
	}
	if total > 0 {
		b.ReportMetric(100*sys.Seconds()/total.Seconds(), "%sys")
	}
}

func Run(bt BenchTestSuite, b *testing.B) {
	bt.SetT(&testing.T{})
	for _, size := range batchSizes {
		b.Run(fmt.Sprintf("test=WriteNodes/size=%d", size), func(b *testing.B) {
			file := tempfile(b)
			bt.SetFile(file)
			bt.SetupTest()
			bt.BenchmarkWriteNodes(size, b)
		})

		b.Run(fmt.Sprintf("test=WriteSingleNode/size=%d", size), func(b *testing.B) {
			file := tempfile(b)
			bt.SetFile(file)
			bt.SetupTest()
			bt.BenchmarkWriteSingleNode(size, b)
		})

		b.Run(fmt.Sprintf("test=ReadAll/size=%d", size), func(b *testing.B) {
			file := tempfile(b)
			bt.SetFile(file)
			bt.SetupTest()
			bt.BenchmarkReadAll(size, b)
		})

		b.Run(fmt.Sprintf("test=Find/size=%d", size), func(b *testing.B) {
			file := tempfile(b)
			bt.SetFile(file)
			bt.SetupTest()
			bt.BenchmarkFind(size, b)
		})

		b.Run(fmt.Sprintf("test=RandomReads/size=%d", size), func(b *testing.B) {
			file := tempfile(b)
			bt.SetFile(file)
			bt.SetupTest()
			bt.BenchmarkRandomReads(size, b)
		})

		b.Run(fmt.Sprintf("test=ResolveIP/size=%d", size), func(b *testing.B) {
			file := tempfile(b)
			bt.SetFile(file)
			bt.SetupTest()
			bt.BenchmarkResolveIP(size, b)
		})

		b.Run(fmt.Sprintf("test=ReverseResolve/size=%d", size), func(b *testing.B) {
			file := tempfile(b)
			bt.SetFile(file)
			bt.SetupTest()
			bt.BenchmarkReverseResolve(size, b)
		})
	}
}

func BenchmarkStores(b *testing.B) {
	// Every SetupTest opens a store, which logs at info. That output interleaves
	// with the benchmark result lines and makes them unparseable by benchstat.
	store.Log.Logger.SetLevel(logrus.ErrorLevel)

	tests := map[string]BenchTestSuite{
		"Sqlstore": new(SqlStoreTestSuite),
	}

	for name, ts := range tests {
		b.Run("store="+name, func(b *testing.B) {
			Run(ts, b)
		})
	}
}
