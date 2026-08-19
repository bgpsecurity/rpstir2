package chainvalidate

import (
	"sync"
	"testing"
)

func Test_getChainCrls(t *testing.T) {
	type args struct {
		chains    *Chains
		chainWg   *sync.WaitGroup
		syncLogId uint64
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getChainCrls(tt.args.chains, tt.args.chainWg, tt.args.syncLogId)
		})
	}
}
