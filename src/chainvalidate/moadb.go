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

func GetChainMoaData(chainRoaDataCh chan []*ChainCertData, moaWg *sync.WaitGroup) error {
	start := time.Now()
	belogs.Debug("getChainMoaSqlDb(): will select rpki_moa")

	var myCount int64
	sql := `select count(*) as myCount from lab_rpki_moa c`
	found, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("getChainMoaSqlDb():select rpki_moa fail:", err)
		return err
	}
	if !found {
		belogs.Debug("getChainMoaSqlDb():select rpki_moa empty:")
		return nil
	}
	belogs.Debug("getChainMoaSqlDb():select myCount rpki_moa, myCount:", myCount)

	oneSize := int64(conf.Int("chain::getCertsCountInOneSize"))
	sql = `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime
			from lab_rpki_moa c
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki
			order by id limit ?, ?  `
	/*
		        var offset, selectCount int
				 for offset = 0; offset < myCount; offset += oneSize {
				startOne := time.Now()
				chainMoaSqls := make([]*ChainCertSql, 0)
				err = xormdb.XormEngine.SQL(sql, offset, oneSize).Find(&chainMoaSqls)
				if err != nil {
					belogs.Error("getChainMoaSqlDb():select rpki_moa limit, fail:", err, " one sql time(s):", time.Since(startOne))
					return err
				}
				belogs.Debug("getChainMoaSqlDb(): select rpki_moa limit, len(chainMoaSqls):", len(chainMoaSqls),
					"  offset:", offset, "  oneSize:", oneSize,
					"  one sql time(s):", time.Since(startOne))
				selectCount += len(chainMoaSqls)
				moaWg.Add(1)
				chainMoaSqlsCh <- chainMoaSqls
			}
	*/
	selectCount := getChainCertSqlDb(oneSize, myCount, sql, moaWg, chainRoaDataCh)
	belogs.Info("getChainMoaSqlDb(): select rpki_moa limit all, selectCount:", selectCount,
		"  myCount:", myCount, "  time(s):", time.Since(start))
	return nil

}

func getChainMoaSqlDb(chainMoaSqlsCh chan []*ChainCertData, moaWg *sync.WaitGroup) error {

	start := time.Now()
	belogs.Debug("getChainMoaSqlDb(): will select rpki_moa")

	var myCount int64
	sql := `select count(*) as myCount from lab_rpki_moa c`
	found, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("getChainMoaSqlDb():select rpki_moa fail:", err)
		return err
	}
	if !found {
		belogs.Debug("getChainMoaSqlDb():select rpki_moa empty:")
		return nil
	}
	belogs.Debug("getChainMoaSqlDb():select myCount rpki_moa, myCount:", myCount)

	oneSize := int64(conf.Int("chain::getCertsCountInOneSize"))
	sql = `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime
			from lab_rpki_moa c
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki
			order by id limit ?, ?  `
	/*
		        var offset, selectCount int
				 for offset = 0; offset < myCount; offset += oneSize {
				startOne := time.Now()
				chainMoaSqls := make([]*ChainCertSql, 0)
				err = xormdb.XormEngine.SQL(sql, offset, oneSize).Find(&chainMoaSqls)
				if err != nil {
					belogs.Error("getChainMoaSqlDb():select rpki_moa limit, fail:", err, " one sql time(s):", time.Since(startOne))
					return err
				}
				belogs.Debug("getChainMoaSqlDb(): select rpki_moa limit, len(chainMoaSqls):", len(chainMoaSqls),
					"  offset:", offset, "  oneSize:", oneSize,
					"  one sql time(s):", time.Since(startOne))
				selectCount += len(chainMoaSqls)
				moaWg.Add(1)
				chainMoaSqlsCh <- chainMoaSqls
			}
	*/
	selectCount := getChainCertSqlDb(oneSize, myCount, sql, moaWg, chainMoaSqlsCh)
	belogs.Info("getChainMoaSqlDb(): select rpki_moa limit all, selectCount:", selectCount,
		"  myCount:", myCount, "  time(s):", time.Since(start))
	return nil
}

/*
	func getChainMoaSqlDb(chainMoaSqlCh chan *ChainCertSql, moaWg *sync.WaitGroup) error {
		defer close(chainMoaSqlCh)
		start := time.Now()
		belogs.Debug("getChainMoaSqlDb(): will select lab_rpki_moa")
		chainMoaSql := new(ChainCertSql)
		// if add "order by ***", the sort_mem may not enough
		sql := `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime
				from lab_rpki_moa c
				left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki
				group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime  `
		rows, err := xormdb.XormEngine.SQL(sql).Rows(chainMoaSql)
		if err != nil {
			belogs.Error("getChainMoaSqlDb(): select lab_rpki_moa fail:", err)
			return err
		}
		defer rows.Close()
		var index uint64
		for rows.Next() {
			chainMoaSql = new(ChainCertSql)
			err = rows.Scan(chainMoaSql)
			if err != nil {
				belogs.Error("getChainMoaSqlDb(): Scan fail:", err)
				continue
			}
			moaWg.Add(1)
			belogs.Debug("getChainMoaSqlDb(): moaWg.Add(), chainMoaSql.Id:", chainMoaSql.Id)
			chainMoaSqlCh <- chainMoaSql
			index++
		}
		belogs.Info("getChainMoaSqlDb(): Scan all, close chainMoaSqlCh, index:", index, "  time(s):", time.Since(start))
		return nil
	}
*/
func getChainMoaSqlsDb() (chainCertSqls []ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_moa c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki   
			group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime  `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("getChainMoaSqlsDb(): lab_rpki_moa id fail:", err)
		return nil, err
	}
	belogs.Info("getChainMoaSqlsDb(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Since(start))
	return chainCertSqls, nil
}

func UpdateMoas(chains *Chains) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("UpdateMoas(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	moaIds := chains.MoaIds
	var updateWg sync.WaitGroup
	updateConcurrentCountCh := make(chan int, conf.Int("chain::updateConcurrentCount"))
	for _, moaId := range moaIds {
		updateWg.Add(1)
		updateConcurrentCountCh <- 1
		go updateMoaDb(session, chains, moaId, updateConcurrentCountCh, &updateWg)
	}
	updateWg.Wait()
	close(updateConcurrentCountCh)

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateMoas(): CommitSession fail :", err)
		return err
	}
	belogs.Info("UpdateMoas(): ok, len(moaIds):", len(moaIds), "  time(s):", time.Since(start))
	return nil
}

func updateMoaDb(session *xorm.Session, chains *Chains, moaId uint64, updateConcurrentCountCh chan int, wg *sync.WaitGroup) {
	defer func() {
		<-updateConcurrentCountCh
		wg.Done()
	}()
	start := time.Now()
	chainMoa, err := chains.GetMoaById(moaId)
	if err != nil {
		belogs.Error("updateMoaDb(): GetMoaById fail :", moaId, err)
		return
	}

	chainDbMoaModel := NewChainDbMoaModel(&chainMoa)
	originModel := model.JudgeOriginByFilePath(chainMoa.FilePath)
	belogs.Debug("updateMoaDb():chainDbMoaModel.id:", chainDbMoaModel.Id,
		"  originModel:", jsonutil.MarshalJson(originModel))

	chainCerts := jsonutil.MarshalJson(*chainDbMoaModel)
	state := jsonutil.MarshalJson(chainMoa.StateModel)
	originJson := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateMoaDb():moaId:", moaId, "    chainCerts:", chainCerts,
		"   state:", state, "    originJson:", originJson)
	sqlStr := `UPDATE lab_rpki_moa set chainCerts=?, state=?, 
	origin = (
		case
		  when   origin = ''    then ? 
		  when   origin is null   then  ? 
		else 
		  origin
		end
	)
	where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, originJson, originJson, moaId)
	if err != nil {
		belogs.Error("updateMoaDb(): UPDATE lab_rpki_moa fail :", moaId, err)
		return
	}
	belogs.Debug("updateMoaDb(): ok, moaId:", moaId, "  time(s):", time.Since(start))
}
