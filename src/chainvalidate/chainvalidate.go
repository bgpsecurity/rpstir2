package chainvalidate

import (
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
)

// chainValidateStart 开始证书链验证

// @return nextStep 下一步需要执行的功能
// @return err 返回错误
func chainValidateStart() (nextStep string, err error) {
	start := time.Now()

	belogs.Info("chainValidateStart():")
	// save chain validate starttime to lab_rpki_sync_log
	syncLogId, err := updateRsyncLogChainValidateStateStartDb("chainvalidating")
	if err != nil {
		belogs.Error("chainValidateStart():updateRsyncLogChainValidateStateStartDb fail:", err)
		return "", err
	}
	belogs.Debug("chainValidateStart():updateRsyncLogChainValidateStateStartDb, syncLogId:", syncLogId)
	err = chainValidate(syncLogId)
	if err != nil {
		belogs.Error("chainValidateStart():chainValidate fail, syncLogId:", syncLogId, err)
		return "", err
	}

	// save  chain validate end time
	err = updateSyncLogChainValidateStateEndDb(syncLogId, "chainvalidated")
	if err != nil {
		belogs.Error("chainValidateStart():updateSyncLogChainValidateStateEndDb  fail:", err)
		return "", err
	}

	belogs.Info("chainValidateStart(): end, will call rtr,  time(s):", time.Since(start))
	return "rtr", nil
}

// chainValidate 开始证书链实际验证
//
//	@param syncLogId 当前同步id
//	@return err 返回错误
func chainValidate(syncLogId uint64) (err error) {
	belogs.Debug("chainValidate():syncLogId:", syncLogId)

	chains := NewChains(80000)

	start := time.Now()

	var chainWg sync.WaitGroup
	// get Chains
	chainWg.Add(1)
	go getChainMfts(chains, &chainWg, syncLogId)

	chainWg.Add(1)
	go getChainCrls(chains, &chainWg, syncLogId)

	chainWg.Add(1)
	go getChainCers(chains, &chainWg, syncLogId)

	chainWg.Add(1)
	go getChainRoas(chains, &chainWg, syncLogId)

	chainWg.Add(1)
	go getChainAsas(chains, &chainWg, syncLogId)

	chainWg.Add(1)
	go getChainMoas(chains, &chainWg, syncLogId)

	chainWg.Wait()
	belogs.Info("chainValidate(): GetChains time(s):", time.Since(start))

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

	wgValidate.Add(1)
	go validateMoas(chains, &wgValidate)

	wgValidate.Wait()
	belogs.Info("chainValidate(): after Validates time(s):", time.Since(start))

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
	go updateMfts(chains, &wgUpdate)

	wgUpdate.Add(1)
	go updateCrls(chains, &wgUpdate)

	wgUpdate.Add(1)
	go updateCers(chains, &wgUpdate)

	wgUpdate.Add(1)
	go updateRoas(chains, &wgUpdate)

	wgUpdate.Add(1)
	go updateAsas(chains, &wgUpdate)

	wgUpdate.Add(1)
	go updateMoas(chains, &wgUpdate)

	wgUpdate.Wait()
	belogs.Info("chainValidate(): after updates time(s):", time.Since(start))

	return nil
}

// updateChainByCheckAll  更新证书链，默认关闭
//
//	@param chains 证书链指针
//	@return err 返回错误
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

/*

type DataSource interface {
	GetChainRoaData(chainRoaDataCh chan []*ChainCertData, roaWg *sync.WaitGroup) error
	UpdateRoas(chains *Chains) error

	GetChainCerData(chainCerDataCh chan []*ChainCertData, roaWg *sync.WaitGroup) error
	UpdateCers(chains *Chains) error

	GetChainAsaData(chainAsaDataCh chan []*ChainCertData, roaWg *sync.WaitGroup) error
	UpdateAsas(chains *Chains) error

	GetChainCrlData(chainCrlDataCh chan []*ChainCertData, roaWg *sync.WaitGroup) error
	UpdateCrls(chains *Chains) error

	GetChainMftData(chainMftDataCh chan []*ChainCertData, roaWg *sync.WaitGroup) error
	// 主要是更新state
	UpdateMfts(chains *Chains) error
	GetChainFileHashs(chainMft ChainMft) ([]ChainFileHash, error)
	GetPreviousMft(chainMft ChainMft) (PreviousMft, error)
}
*/
