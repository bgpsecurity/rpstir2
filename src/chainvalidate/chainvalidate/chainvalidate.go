package chainvalidate

import (
	"sync"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	httpclient "github.com/cpusoft/goutil/httpclient"

	"chainvalidate/db"
	chainmodel "chainvalidate/model"
)

func ChainValidateStart() (err error) {
	belogs.Debug("ChainValidateStart(): start")
	// save chain validate starttime to lab_rpki_sync_log
	labRpkiSyncLogId, err := db.UpdateRsyncLogChainValidateStateStart("chainvalidating")
	if err != nil {
		belogs.Error("ChainValidateStart():UpdateRsyncLogChainValidateStart fail:", err)
		return
	}
	// build and validate chain all cert (include all)
	err = ChainValidate()
	if err != nil {
		belogs.Error("ChainValidateStart():ChainValidateCerts fail:", err)
		return
	}

	// save  chain validate end time
	err = db.UpdateRsyncLogChainValidateStateEnd(labRpkiSyncLogId, "chainvalidated")
	if err != nil {
		belogs.Error("ChainValidateStart():UpdateRsyncLogChainValidateStateEnd fail:", err)
		return
	}
	belogs.Info("ChainValidateStart(): end")

	/////////////////////
	// will call rtr
	go func() {
		httpclient.Post("http", conf.String("rtr::httpserver"), conf.Int("rtr::httpport"),
			"/rtr/update", "")
	}()
	return nil

}

func ChainValidate() (err error) {
	belogs.Info("ChainValidate():start:")

	chains := chainmodel.NewChains(80000)

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
	go db.UpdateMfts(chains, &wgUpdate)

	wgUpdate.Add(1)
	go db.UpdateCrls(chains, &wgUpdate)

	wgUpdate.Add(1)
	go db.UpdateCers(chains, &wgUpdate)

	wgUpdate.Add(1)
	go db.UpdateRoas(chains, &wgUpdate)

	wgUpdate.Wait()
	belogs.Info("ChainValidate(): after updates  time(s):", time.Now().Sub(start).Seconds())

	return nil
}
