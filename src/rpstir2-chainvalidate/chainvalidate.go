package chainvalidate

import (
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
)

func chainValidateStart() (nextStep string, err error) {
	start := time.Now()

	belogs.Info("chainValidateStart(): start")
	// save chain validate starttime to lab_rpki_sync_log
	labRpkiSyncLogId, err := updateRsyncLogChainValidateStateStartDb("chainvalidating")
	if err != nil {
		belogs.Error("chainValidateStart():updateRsyncLogChainValidateStateStartDb fail:", err)
		return "", err
	}

	// build and validate chain all cert (include all)
	err = chainValidate()
	if err != nil {
		belogs.Error("chainValidateStart():chainValidate fail:", err)
		return "", err
	}

	// save  chain validate end time
	err = updateRsyncLogChainValidateStateEndDb(labRpkiSyncLogId, "chainvalidated")
	if err != nil {
		belogs.Error("chainValidateStart():updateRsyncLogChainValidateStateEndDb fail:", err)
		return "", err
	}

	belogs.Info("chainValidateStart(): end, will call rtr,  time(s):", time.Since(start))
	return "rtr", nil
}

func chainValidate() (err error) {
	belogs.Info("chainValidate():start:")

	chains := NewChains(80000)

	start := time.Now()
	var wgGetChain sync.WaitGroup
	// get Chains
	wgGetChain.Add(1)
	go getChainMfts(chains, &wgGetChain)

	wgGetChain.Add(1)
	go getChainCrls(chains, &wgGetChain)

	wgGetChain.Add(1)
	go getChainCers(chains, &wgGetChain)

	wgGetChain.Add(1)
	go getChainRoas(chains, &wgGetChain)

	wgGetChain.Add(1)
	go getChainAsas(chains, &wgGetChain)

	wgGetChain.Wait()
	belogs.Info("chainValidate(): GetChains  time(s):", time.Since(start))

	// validate
	start = time.Now()
	var wgValidate sync.WaitGroup
	wgValidate.Add(1)
	go validateMfts(chains, &wgValidate)

	wgValidate.Add(1)
	go validateCrls(chains, &wgValidate)

	wgValidate.Add(1)
	go validateCers(chains, &wgValidate)

	wgValidate.Add(1)
	go validateRoas(chains, &wgValidate)

	wgValidate.Add(1)
	go validateAsas(chains, &wgValidate)

	wgValidate.Wait()
	belogs.Info("chainValidate(): after Validates  time(s):", time.Since(start))

	// will check all certs in chain: mft invalid --> crl/roa/cer invalid
	err = updateChainByCheckAll(chains)
	if err != nil {
		belogs.Error("chainValidate():updateChainByCheckAll fail:", err)
		return err
	}

	// save
	start = time.Now()
	var wgUpdate sync.WaitGroup
	wgUpdate.Add(1)
	go updateMftsDb(chains, &wgUpdate)

	wgUpdate.Add(1)
	go updateCrlsDb(chains, &wgUpdate)

	wgUpdate.Add(1)
	go updateCersDb(chains, &wgUpdate)

	wgUpdate.Add(1)
	go updateRoasDb(chains, &wgUpdate)

	wgUpdate.Add(1)
	go updateAsasDb(chains, &wgUpdate)

	wgUpdate.Wait()
	belogs.Info("chainValidate(): after updates  time(s):", time.Since(start))

	return nil
}

func updateChainByCheckAll(chains *Chains) (err error) {
	// after all, will check again:
	// if mft is invalid,may effect roa/crl/cer --> ignore/warning/invalid, not found mft
	invalidMftEffect := conf.String("policy::invalidMftEffect")
	if invalidMftEffect == "warning" || invalidMftEffect == "invalid" {
		err = updateChainByMft(chains, invalidMftEffect)
		if err != nil {
			belogs.Error("updateChainByCheckAll():updateChainByMft fail:", err)
			//return err
		}
	}
	return nil
}
