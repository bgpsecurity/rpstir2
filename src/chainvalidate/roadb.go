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

// GetChainRoaData 实现 SQLDataSource 的方法
func GetChainRoaData(chainRoaDataCh chan []*ChainCertData, roaWg *sync.WaitGroup) error {
	start := time.Now()
	belogs.Debug("getChainRoaSqlDb(): will select rpki_roa")
	var myCount int64
	sql := `select count(*) as myCount from lab_rpki_roa c`
	found, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("getChainRoaSqlDb():select rpki_roa fail:", err)
		return err
	}
	if !found {
		belogs.Debug("getChainRoaSqlDb():select rpki_roa empty:")
		return nil
	}
	belogs.Debug("getChainRoaSqlDb():select myCount rpki_roa, myCount:", myCount)

	oneSize := int64(conf.Int("chain::getCertsCountInOneSize"))
	sql = `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_roa c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki   
			order by id limit ?,?  `
	selectCount := getChainCertSqlDb(oneSize, myCount, sql, roaWg, chainRoaDataCh)
	belogs.Info("getChainRoaSqlDb(): select rpki_roa limit all, selectCount:", selectCount,
		"  myCount:", myCount, "  time(s):", time.Since(start))
	return nil
}

func getChainRoaSqlDb(chainRoaSqlsCh chan []*ChainCertData, roaWg *sync.WaitGroup) error {

	start := time.Now()
	belogs.Debug("getChainRoaSqlDb(): will select rpki_roa")
	var myCount int64
	sql := `select count(*) as myCount from lab_rpki_roa c`
	found, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("getChainRoaSqlDb():select rpki_roa fail:", err)
		return err
	}
	if !found {
		belogs.Debug("getChainRoaSqlDb():select rpki_roa empty:")
		return nil
	}
	belogs.Debug("getChainRoaSqlDb():select myCount rpki_roa, myCount:", myCount)

	oneSize := int64(conf.Int("chain::getCertsCountInOneSize"))
	sql = `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_roa c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki   
			order by id limit ?,?  `
	/*
		var offset, selectCount int64
		for offset = 0; offset < myCount; offset += oneSize {
			go func(offset int64) {
				startOne := time.Now()
				chainRoaSqls := make([]*ChainCertSql, 0)
				err = xormdb.XormEngine.SQL(sql, offset, oneSize).Find(&chainRoaSqls)
				if err != nil {
					belogs.Error("getChainRoaSqlDb():select rpki_roa limit, fail:", err, " one sql time(s):", time.Since(startOne))
					return
				}
				belogs.Debug("getChainRoaSqlDb(): select rpki_roa limit, len(chainRoaSqls):", len(chainRoaSqls),
					"  offset:", offset, " oneSize:", oneSize,
					"  one sql time(s):", time.Since(startOne))
				atomic.AddInt64(&selectCount, int64(len(chainRoaSqls)))
				roaWg.Add(1)
				chainRoaSqlsCh <- chainRoaSqls
			}(offset)
		}
	*/
	selectCount := getChainCertSqlDb(oneSize, myCount, sql, roaWg, chainRoaSqlsCh)
	belogs.Info("getChainRoaSqlDb(): select rpki_roa limit all, selectCount:", selectCount,
		"  myCount:", myCount, "  time(s):", time.Since(start))
	return nil
}

/*
	func getChainRoaSqlDb(chainRoaSqlCh chan *ChainCertSql, roaWg *sync.WaitGroup) error {
		defer close(chainRoaSqlCh)

		start := time.Now()
		belogs.Debug("getChainRoaSqlDb(): will select lab_rpki_roa")
		chainRoaSql := new(ChainCertSql)
		// if add "order by ***", the sort_mem may not enough
		sql := `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime
				from lab_rpki_roa c
				left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki
				group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime  `
		rows, err := xormdb.XormEngine.SQL(sql).Rows(chainRoaSql)
		if err != nil {
			belogs.Error("getChainRoaSqlDb(): select lab_rpki_roa fail:", err)
			return err
		}
		defer rows.Close()
		var index uint64
		for rows.Next() {
			chainRoaSql = new(ChainCertSql)
			err = rows.Scan(chainRoaSql)
			if err != nil {
				belogs.Error("getChainRoaSqlDb(): Scan fail:", err)
				continue
			}
			roaWg.Add(1)
			belogs.Debug("getChainRoaSqlDb(): roaWg.Add(), chainRoaSql.Id:", chainRoaSql.Id)
			chainRoaSqlCh <- chainRoaSql
			index++
			belogs.Debug("getChainRoaSqlDb(): scan chainRoaSql:", jsonutil.MarshalJson(chainRoaSql), " index:", index)
		}
		belogs.Info("getChainRoaSqlDb(): Scan all, close chainRoaSqlCh, index:", index, "  time(s):", time.Since(start))
		return nil
	}
*/
func getChainRoaSqlsDb() (chainCertSqls []ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_roa c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki   
			group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime  `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("getChainRoaSqlsDb(): lab_rpki_roa id fail:", err)
		return nil, err
	}
	belogs.Info("getChainRoaSqlsDb(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Since(start))
	return chainCertSqls, nil
}

func updateRoaDb(session *xorm.Session, chains *Chains, roaId uint64,
	updateConcurrentCountCh chan int, updateWg *sync.WaitGroup) (err error) {
	defer func() {
		<-updateConcurrentCountCh
		updateWg.Done()
	}()
	start := time.Now()
	chainRoa, err := chains.GetRoaById(roaId)
	if err != nil {
		belogs.Error("updateRoaDb(): GetRoaById fail :", roaId, err)
		return err
	}

	chainDbRoaModel := NewChainDbRoaModel(&chainRoa)
	originModel := model.JudgeOriginByFilePath(chainRoa.FilePath)
	belogs.Debug("updateRoaDb():chainDbRoaModel.id:", chainDbRoaModel.Id,
		"  originModel:", jsonutil.MarshalJson(originModel))

	chainCerts := jsonutil.MarshalJson(*chainDbRoaModel)
	state := jsonutil.MarshalJson(chainRoa.StateModel)
	// will set all roa to valid
	if conf.String("chain::forceValidateAllRoas") == "true" {
		s := model.NewStateModel()
		s.State = "valid"
		state = jsonutil.MarshalJson(s)
		belogs.Debug("updateRoaDb(): forceValidateAllRoas roaId:", roaId, "  state:", state)
	}
	originJson := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateRoaDb():roaId:", roaId, "    chainCerts:", chainCerts,
		"   state:", state, "  originJson:", originJson)
	sqlStr := `UPDATE lab_rpki_roa set chainCerts=?, state=?,  
	origin = (
		case
		  when   origin = ''    then ? 
		  when   origin is null   then  ? 
		else 
		  origin
		end
	)
	where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, originJson, originJson, roaId)
	if err != nil {
		belogs.Error("updateRoaDb(): UPDATE lab_rpki_roa fail :", roaId, err)
		return err
	}
	belogs.Debug("updateRoaDb(): ok, roaId:", roaId, "  time(s):", time.Since(start))
	return nil
}

func UpdateRoas(chains *Chains) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("UpdateRoas(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	roaIds := chains.RoaIds
	belogs.Debug("UpdateRoas():len(roaIds):", len(roaIds))

	var updateWg sync.WaitGroup
	updateConcurrentCountCh := make(chan int, conf.Int("chain::updateConcurrentCount"))
	for _, roaId := range roaIds {
		updateWg.Add(1)
		updateConcurrentCountCh <- 1
		go updateRoaDb(session, chains, roaId, updateConcurrentCountCh, &updateWg)
	}
	updateWg.Wait()
	close(updateConcurrentCountCh)

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateRoas(): CommitSession fail :", err)
		return err
	}
	belogs.Info("UpdateRoas(): ok, len(roaIds):", len(roaIds), "  time(s):", time.Since(start))
	return nil
}
