package poker

import "testing"

func BenchmarkEvalInfo(b *testing.B) {
	T := 0
	for n := 0; n < b.N; n++ {
		ei := makeEvalInfo()
		T += int(ei.slowRankToPacked[0])
	}
}
