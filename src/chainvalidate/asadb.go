package chainvalidate

import (
	"sync"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

func GetChainAsaData(chainRoaDataCh chan []*ChainCertData, asaWg *sync.WaitGroup) error {
	start := time.Now()
	belogs.Debug("getChainAsaSqlDb(): will select rpki_asa")

	var myCount int64
	sql := `select count(*) as myCount from lab_rpki_asa c`
	found, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("getChainAsaSqlDb():select rpki_asa fail:", err)
		return err
	}
	if !found {
		belogs.Debug("getChainAsaSqlDb():select rpki_asa empty:")
		return nil
	}
	belogs.Debug("getChainAsaSqlDb():select myCount rpki_asa, myCount:", myCount)

	oneSize := int64(conf.Int("chain::getCertsCountInOneSize"))
	sql = `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime
			from lab_rpki_asa c
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki
			order by id limit ?, ?  `
	/*
		        var offset, selectCount int
				 for offset = 0; offset < myCount; offset += oneSize {
				startOne := time.Now()
				chainAsaSqls := make([]*ChainCertSql, 0)
				err = xormdb.XormEngine.SQL(sql, offset, oneSize).Find(&chainAsaSqls)
				if err != nil {
					belogs.Error("getChainAsaSqlDb():select rpki_asa limit, fail:", err, " one sql time(s):", time.Since(startOne))
					return err
				}
				belogs.Debug("getChainAsaSqlDb(): select rpki_asa limit, len(chainAsaSqls):", len(chainAsaSqls),
					"  offset:", offset, "  oneSize:", oneSize,
					"  one sql time(s):", time.Since(startOne))
				selectCount += len(chainAsaSqls)
				asaWg.Add(1)
				chainAsaSqlsCh <- chainAsaSqls
			}
	*/
	selectCount := getChainCertSqlDb(oneSize, myCount, sql, asaWg, chainRoaDataCh)
	belogs.Info("getChainAsaSqlDb(): select rpki_asa limit all, selectCount:", selectCount,
		"  myCount:", myCount, "  time(s):", time.Since(start))
	return nil

}

func getChainAsaSqlDb(chainAsaSqlsCh chan []*ChainCertData, asaWg *sync.WaitGroup) error {

	start := time.Now()
	belogs.Debug("getChainAsaSqlDb(): will select rpki_asa")

	var myCount int64
	sql := `select count(*) as myCount from lab_rpki_asa c`
	found, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("getChainAsaSqlDb():select rpki_asa fail:", err)
		return err
	}
	if !found {
		belogs.Debug("getChainAsaSqlDb():select rpki_asa empty:")
		return nil
	}
	belogs.Debug("getChainAsaSqlDb():select myCount rpki_asa, myCount:", myCount)

	oneSize := int64(conf.Int("chain::getCertsCountInOneSize"))
	sql = `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime
			from lab_rpki_asa c
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki
			order by id limit ?, ?  `
	/*
		        var offset, selectCount int
				 for offset = 0; offset < myCount; offset += oneSize {
				startOne := time.Now()
				chainAsaSqls := make([]*ChainCertSql, 0)
				err = xormdb.XormEngine.SQL(sql, offset, oneSize).Find(&chainAsaSqls)
				if err != nil {
					belogs.Error("getChainAsaSqlDb():select rpki_asa limit, fail:", err, " one sql time(s):", time.Since(startOne))
					return err
				}
				belogs.Debug("getChainAsaSqlDb(): select rpki_asa limit, len(chainAsaSqls):", len(chainAsaSqls),
					"  offset:", offset, "  oneSize:", oneSize,
					"  one sql time(s):", time.Since(startOne))
				selectCount += len(chainAsaSqls)
				asaWg.Add(1)
				chainAsaSqlsCh <- chainAsaSqls
			}
	*/
	selectCount := getChainCertSqlDb(oneSize, myCount, sql, asaWg, chainAsaSqlsCh)
	belogs.Info("getChainAsaSqlDb(): select rpki_asa limit all, selectCount:", selectCount,
		"  myCount:", myCount, "  time(s):", time.Since(start))
	return nil
}

/*
	func getChainAsaSqlDb(chainAsaSqlCh chan *ChainCertSql, asaWg *sync.WaitGroup) error {
		defer close(chainAsaSqlCh)
		start := time.Now()
		belogs.Debug("getChainAsaSqlDb(): will select lab_rpki_asa")
		chainAsaSql := new(ChainCertSql)
		// if add "order by ***", the sort_mem may not enough
		sql := `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime
				from lab_rpki_asa c
				left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki
				group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime  `
		rows, err := xormdb.XormEngine.SQL(sql).Rows(chainAsaSql)
		if err != nil {
			belogs.Error("getChainAsaSqlDb(): select lab_rpki_asa fail:", err)
			return err
		}
		defer rows.Close()
		var index uint64
		for rows.Next() {
			chainAsaSql = new(ChainCertSql)
			err = rows.Scan(chainAsaSql)
			if err != nil {
				belogs.Error("getChainAsaSqlDb(): Scan fail:", err)
				continue
			}
			asaWg.Add(1)
			belogs.Debug("getChainAsaSqlDb(): asaWg.Add(), chainAsaSql.Id:", chainAsaSql.Id)
			chainAsaSqlCh <- chainAsaSql
			index++
		}
		belogs.Info("getChainAsaSqlDb(): Scan all, close chainAsaSqlCh, index:", index, "  time(s):", time.Since(start))
		return nil
	}
*/
func getChainAsaSqlsDb() (chainCertSqls []ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_asa c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki   
			group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime  `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("getChainAsaSqlsDb(): lab_rpki_asa id fail:", err)
		return nil, err
	}
	belogs.Info("getChainAsaSqlsDb(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Since(start))
	return chainCertSqls, nil
}

func UpdateAsas(chains *Chains) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("UpdateAsas(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	asaIds := chains.AsaIds
	var updateWg sync.WaitGroup
	updateConcurrentCountCh := make(chan int, conf.Int("chain::updateConcurrentCount"))
	for _, asaId := range asaIds {
		updateWg.Add(1)
		updateConcurrentCountCh <- 1
		go updateAsaDb(session, chains, asaId, updateConcurrentCountCh, &updateWg)
	}
	updateWg.Wait()
	close(updateConcurrentCountCh)

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateAsas(): CommitSession fail :", err)
		return err
	}
	belogs.Info("UpdateAsas(): ok, len(asaIds):", len(asaIds), "  time(s):", time.Since(start))
	return nil
}

func updateAsaDb(session *xorm.Session, chains *Chains, asaId uint64, updateConcurrentCountCh chan int, wg *sync.WaitGroup) {
	defer func() {
		<-updateConcurrentCountCh
		wg.Done()
	}()
	start := time.Now()
	chainAsa, err := chains.GetAsaById(asaId)
	if err != nil {
		belogs.Error("updateAsaDb(): GetAsaById fail :", asaId, err)
		return
	}

	chainDbAsaModel := NewChainDbAsaModel(&chainAsa)
	originModel := model.JudgeOriginByFilePath(chainAsa.FilePath)
	belogs.Debug("updateAsaDb():chainDbAsaModel.id:", chainDbAsaModel.Id,
		"  originModel:", jsonutil.MarshalJson(originModel))

	chainCerts := jsonutil.MarshalJson(*chainDbAsaModel)
	state := jsonutil.MarshalJson(chainAsa.StateModel)
	originJson := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateAsaDb():asaId:", asaId, "    chainCerts:", chainCerts,
		"   state:", state, "    originJson:", originJson)
	sqlStr := `UPDATE lab_rpki_asa set chainCerts=?, state=?, 
	origin = (
		case
		  when   origin = ''    then ? 
		  when   origin is null   then  ? 
		else 
		  origin
		end
	)
	where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, originJson, originJson, asaId)
	if err != nil {
		belogs.Error("updateAsaDb(): UPDATE lab_rpki_asa fail :", asaId, err)
		return
	}
	belogs.Debug("updateAsaDb(): ok, asaId:", asaId, "  time(s):", time.Since(start))
}
