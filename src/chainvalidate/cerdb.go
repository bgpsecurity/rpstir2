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

func   GetChainCerData(chainRoaDataCh chan []*ChainCertData, cerWg *sync.WaitGroup) error {

	start := time.Now()
	belogs.Debug("getChainCerSqlDb(): will select rpki_cer")

	var myCount int64
	sql := `select count(*) as myCount from lab_rpki_cer c`
	found, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("getChainCerSqlDb():select rpki_cer fail:", err)
		return err
	}
	if !found {
		belogs.Debug("getChainCerSqlDb():select rpki_cer empty:")
		return nil
	}
	belogs.Debug("getChainCerSqlDb():select myCount rpki_cer, myCount:", myCount)

	oneSize := int64(conf.Int("chain::getCertsCountInOneSize"))
	sql = `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_cer c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.sn and c.aki = v.aki   
			order by id limit ?, ? `
	/*
		var offset, selectCount int
		for offset = 0; offset < myCount; offset += oneSize {
			startOne := time.Now()
			chainCerSqls := make([]*ChainCertSql, 0)
			err = xormdb.XormEngine.SQL(sql, offset, oneSize).Find(&chainCerSqls)
			if err != nil {
				belogs.Error("getChainCerSqlDb():select rpki_cer limit, fail:", err, " one sql time(s):", time.Since(startOne))
				return err
			}
			belogs.Debug("getChainCerSqlDb(): select rpki_cer limit, len(chainCerSqls):", len(chainCerSqls),
				"  offset:", offset, "   oneSize:", oneSize,
				"  one sql time(s):", time.Since(startOne))
			selectCount += len(chainCerSqls)
			cerWg.Add(1)
			chainCerSqlsCh <- chainCerSqls
		}
	*/
	selectCount := getChainCertSqlDb(oneSize, myCount, sql, cerWg, chainRoaDataCh)
	belogs.Info("getChainCerSqlDb(): select rpki_cer limit all, selectCount:", selectCount,
		"  myCount:", myCount, "  time(s):", time.Since(start))
	return nil

}

//func getChainCerSqlDb(chainCerSqlsCh chan []*ChainCertData, cerWg *sync.WaitGroup) error {
//
//	start := time.Now()
//	belogs.Debug("getChainCerSqlDb(): will select rpki_cer")
//
//	var myCount int64
//	sql := `select count(*) as myCount from lab_rpki_cer c`
//	found, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
//	if err != nil {
//		belogs.Error("getChainCerSqlDb():select rpki_cer fail:", err)
//		return err
//	}
//	if !found {
//		belogs.Debug("getChainCerSqlDb():select rpki_cer empty:")
//		return nil
//	}
//	belogs.Debug("getChainCerSqlDb():select myCount rpki_cer, myCount:", myCount)
//
//	oneSize := int64(conf.Int("chain::getCertsCountInOneSize"))
//	sql = `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime
//			from lab_rpki_cer c
//			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.sn and c.aki = v.aki
//			order by id limit ?, ? `
//	/*
//		var offset, selectCount int
//		for offset = 0; offset < myCount; offset += oneSize {
//			startOne := time.Now()
//			chainCerSqls := make([]*ChainCertSql, 0)
//			err = xormdb.XormEngine.SQL(sql, offset, oneSize).Find(&chainCerSqls)
//			if err != nil {
//				belogs.Error("getChainCerSqlDb():select rpki_cer limit, fail:", err, " one sql time(s):", time.Since(startOne))
//				return err
//			}
//			belogs.Debug("getChainCerSqlDb(): select rpki_cer limit, len(chainCerSqls):", len(chainCerSqls),
//				"  offset:", offset, "   oneSize:", oneSize,
//				"  one sql time(s):", time.Since(startOne))
//			selectCount += len(chainCerSqls)
//			cerWg.Add(1)
//			chainCerSqlsCh <- chainCerSqls
//		}
//	*/
//	selectCount := getChainCertSqlDb(oneSize, myCount, sql, cerWg, chainCerSqlsCh)
//	belogs.Info("getChainCerSqlDb(): select rpki_cer limit all, selectCount:", selectCount,
//		"  myCount:", myCount, "  time(s):", time.Since(start))
//	return nil
//}

/*
	func getChainCerSqlDb(chainCerSqlCh chan *ChainCertSql, cerWg *sync.WaitGroup) error {
		defer close(chainCerSqlCh)
		start := time.Now()
		belogs.Debug("getChainCerSqlDb(): will select lab_rpki_cer")
		chainCerSql := new(ChainCertSql)
		// if add "order by ***", the sort_mem may not enough
		sql := `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime
				from lab_rpki_cer c
				left join lab_rpki_crl_revoked_cert_view v on v.sn = c.sn and c.aki = v.aki
				group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime `
		rows, err := xormdb.XormEngine.SQL(sql).Rows(chainCerSql)
		if err != nil {
			belogs.Error("getChainCerSqlDb():select lab_rpki_cer fail:", err)
			return err
		}
		defer rows.Close()
		var index uint64
		for rows.Next() {
			chainCerSql = new(ChainCertSql)
			err = rows.Scan(chainCerSql)
			if err != nil {
				belogs.Error("getChainCerSqlDb(): Scan fail:", err)
				continue
			}
			cerWg.Add(1)
			belogs.Debug("getChainCerSqlDb(): cerWg.Add(), chainCerSql.Id:", chainCerSql.Id)
			chainCerSqlCh <- chainCerSql
			index++
		}
		belogs.Info("getChainCerSqlDb(): Scan all, close chainCerSqlCh, index:", index, "  time(s):", time.Since(start))
		return nil
	}
*/
func getChainCerSqlsDb() (chainCertSqls []ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_cer c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.sn and c.aki = v.aki   
			group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("getChainCerSqlsDb(): lab_rpki_cer id fail:", err)
		return nil, err
	}
	belogs.Info("getChainCerSqlsDb(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Since(start))
	return chainCertSqls, nil
}

func updateCerDb(session *xorm.Session, chains *Chains, cerId uint64,
	updateConcurrentCountCh chan int, updateWg *sync.WaitGroup) (err error) {
	defer func() {
		<-updateConcurrentCountCh
		updateWg.Done()
	}()
	start := time.Now()
	chainCer, err := chains.GetCerById(cerId)
	if err != nil {
		belogs.Error("updateCerDb(): GetCerById fail :", cerId, err)
		return err
	}

	chainDbCerModel := NewChainDbCerModel(&chainCer)
	originModel := model.JudgeOriginByFilePath(chainCer.FilePath)
	belogs.Debug("updateCerDb():chainDbCerModel.id:", chainDbCerModel.Id,
		"  len(chainDbCerModel.ChildChainCers):", len(chainDbCerModel.ChildChainCers),
		"  originModel:", jsonutil.MarshalJson(originModel))

	chainCerts := jsonutil.MarshalJson(*chainDbCerModel)
	state := jsonutil.MarshalJson(chainCer.StateModel)
	originJson := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateCerDb():cerId:", cerId, "    chainCerts", chainCerts,
		"   state:", state, "  originJson:", originJson)
	sqlStr := `UPDATE lab_rpki_cer set chainCerts=?, state=? , 
	origin = (
		case
		  when   origin = ''    then ? 
		  when   origin is null   then  ? 
		else 
		  origin
		end
	)
	where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, originJson, originJson, cerId)
	if err != nil {
		belogs.Error("updateCerDb(): UPDATE lab_rpki_cer fail :", cerId, err)
		return err
	}
	belogs.Debug("updateCerDb(): ok, cerId:", cerId, "  time(s):", time.Since(start))
	return nil
}

func   UpdateCers(chains *Chains) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateCersDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	cerIds := chains.CerIds
	belogs.Debug("updateCersDb():len(cerIds):", len(cerIds))

	var updateWg sync.WaitGroup
	updateConcurrentCountCh := make(chan int, conf.Int("chain::updateConcurrentCount"))
	for _, cerId := range cerIds {
		updateWg.Add(1)
		updateConcurrentCountCh <- 1
		go updateCerDb(session, chains, cerId, updateConcurrentCountCh, &updateWg)
	}
	updateWg.Wait()
	close(updateConcurrentCountCh)

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateCersDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateCersDb(): ok, len(cerIds):", len(cerIds), "  time(s):", time.Since(start))
	return nil
}
