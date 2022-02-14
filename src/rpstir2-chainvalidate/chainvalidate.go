package chainvalidate

import (
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
)

func chainValidateStart() (nextStep string, err error) {
	start := time.Now()

	belogs.Debug("chainValidateStart(): start")
	// save chain validate starttime to lab_rpki_sync_log
	labRpkiSyncLogId, err := UpdateRsyncLogChainValidateStateStart("chainvalidating")
	if err != nil {
		belogs.Error("chainValidateStart():UpdateRsyncLogChainValidateStart fail:", err)
		return "", err
	}

	// build and validate chain all cert (include all)
	err = ChainValidate()
	if err != nil {
		belogs.Error("chainValidateStart():ChainValidateCerts fail:", err)
		return "", err
	}

	// save  chain validate end time
	err = UpdateRsyncLogChainValidateStateEnd(labRpkiSyncLogId, "chainvalidated")
	if err != nil {
		belogs.Error("chainValidateStart():UpdateRsyncLogChainValidateStateEnd fail:", err)
		return "", err
	}

	belogs.Info("chainValidateStart(): end, will call rtr,  time(s):", time.Now().Sub(start).Seconds())
	return "rtr", nil
}

func ChainValidate() (err error) {
	belogs.Info("ChainValidate():start:")

	chains := NewChains(80000)

	start := time.Now()
	var wgGetChain sync.WaitGroup
	// get Chains
	wgGetChain.Add(1)
	go GetChainMfts(chains, &wgGetChain)

	wgGetChain.Add(1)
	go GetChainCrls(chains, &wgGetChain)

	wgGetChain.Add(1)
	go GetChainCers(chains, &wgGetChain)

	wgGetChain.Add(1)
	go GetChainRoas(chains, &wgGetChain)

	wgGetChain.Wait()
	belogs.Info("ChainValidate(): GetChains  time(s):", time.Now().Sub(start).Seconds())

	// validate
	start = time.Now()
	var wgValidate sync.WaitGroup
	wgValidate.Add(1)
	go ValidateMfts(chains, &wgValidate)

	wgValidate.Add(1)
	go ValidateCrls(chains, &wgValidate)

	wgValidate.Add(1)
	go ValidateCers(chains, &wgValidate)

	wgValidate.Add(1)
	go ValidateRoas(chains, &wgValidate)

	wgValidate.Wait()
	belogs.Info("ChainValidate(): after Validates  time(s):", time.Now().Sub(start).Seconds())

	// save
	start = time.Now()
	var wgUpdate sync.WaitGroup
	wgUpdate.Add(1)
	go UpdateMfts(chains, &wgUpdate)

	wgUpdate.Add(1)
	go UpdateCrls(chains, &wgUpdate)

	wgUpdate.Add(1)
	go UpdateCers(chains, &wgUpdate)

	wgUpdate.Add(1)
	go UpdateRoas(chains, &wgUpdate)

	wgUpdate.Wait()
	belogs.Info("ChainValidate(): after updates  time(s):", time.Now().Sub(start).Seconds())

	return nil
}
