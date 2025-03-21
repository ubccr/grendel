package storetest

import (
	"fmt"
	"path"
	"testing"
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
	tests := map[string]BenchTestSuite{
		"Sqlstore": new(SqlStoreTestSuite),
	}

	for name, ts := range tests {
		b.Run("store="+name, func(b *testing.B) {
			Run(ts, b)
		})
	}
}
