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

func GetChainCrlData(chainCrlDataCh chan []*ChainCertData, crlWg *sync.WaitGroup) error {

	start := time.Now()
	belogs.Debug("getChainCrlSqlDb(): will select rpki_crl")
	var myCount int64
	sql := `select count(*) as myCount from lab_rpki_crl c`
	found, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("getChainCrlSqlDb():select rpki_crl fail:", err)
		return err
	}
	if !found {
		belogs.Debug("getChainCrlSqlDb():select rpki_crl empty:")
		return nil
	}
	belogs.Debug("getChainCrlSqlDb():select myCount rpki_crl, myCount:", myCount)

	oneSize := int64(conf.Int("chain::getCertsCountInOneSize"))
	sql = `select c.id, c.syncLogId, c.jsonAll, c.state ,cer.cerFiles,roa.roaFiles, mft.mftFiles ,asa.asaFiles 
	from lab_rpki_crl c  
	left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as cerFiles , v.id as crlId from lab_rpki_cer c, lab_rpki_crl_revoked_cert_view v 
		 where c.sn = v.sn and c.aki =v.aki 
		 group by v.id) cer on cer.crlId = c.id	
	left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as roaFiles , v.id as crlId from lab_rpki_roa c, lab_rpki_crl_revoked_cert_view v 
		where c.jsonAll->>'$.eeCertModel.sn' = v.sn and c.aki =v.aki 
		 group by v.id) roa on roa.crlId = c.id 
	left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as mftFiles , v.id as crlId from lab_rpki_mft c, lab_rpki_crl_revoked_cert_view v 
		 where c.jsonAll->>'$.eeCertModel.sn' = v.sn and c.aki =v.aki 
		 group by v.id) mft on mft.crlId = c.id	
	left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as asaFiles , v.id as crlId from lab_rpki_asa c, lab_rpki_crl_revoked_cert_view v 
		 where c.jsonAll->>'$.eeCertModel.sn' = v.sn and c.aki =v.aki 
		 group by v.id) asa on asa.crlId = c.id		
	order by id limit ?,?  `
	/*
		var offset, selectCount int
		for offset = 0; offset < myCount; offset += oneSize {
			startOne := time.Now()
			chainCrlSqls := make([]*ChainCertSql, 0)
			err = xormdb.XormEngine.SQL(sql, offset, oneSize).Find(&chainCrlSqls)
			if err != nil {
				belogs.Error("getChainCrlSqlDb():select rpki_crl limit,fail:", err, " one sql time(s):", time.Since(startOne))
				return err
			}
			belogs.Debug("getChainCrlSqlDb(): select rpki_crl limit, len(chainCrlSqls):", len(chainCrlSqls),
				"  offset:", offset, "  oneSize:", oneSize,
				"  one sql time(s):", time.Since(startOne))
			selectCount += len(chainCrlSqls)
			crlWg.Add(1)
			chainCrlSqlsCh <- chainCrlSqls
		}
	*/
	selectCount := getChainCertSqlDb(oneSize, myCount, sql, crlWg, chainCrlDataCh)
	belogs.Info("getChainCrlSqlDb(): select rpki_crl limit all, selectCount:", selectCount,
		"  myCount:", myCount, "  time(s):", time.Since(start))
	return nil
}

/*
	func getChainCrlSqlDb(chainCrlSqlCh chan *ChainCertSql, crlWg *sync.WaitGroup) error {
		defer close(chainCrlSqlCh)
		start := time.Now()
		belogs.Debug("getChainCrlSqlDb(): will select lab_rpki_crl")
		chainCrlSql := new(ChainCertSql)
		// if add "order by ***", the sort_mem may not enough
		sql := `select c.id, c.syncLogId, c.jsonAll, c.state ,cer.cerFiles,roa.roaFiles, mft.mftFiles
			from lab_rpki_crl c
			left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as cerFiles , v.id as crlId from lab_rpki_cer c, lab_rpki_crl_revoked_cert_view v
			 	where c.sn = v.sn and c.aki =v.aki
			 	group by v.id) cer on cer.crlId = c.id
			left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as roaFiles , v.id as crlId from lab_rpki_roa c, lab_rpki_crl_revoked_cert_view v
				where c.jsonAll->>'$.eeCertModel.sn' = v.sn and c.aki =v.aki
			 	group by v.id) roa on roa.crlId = c.id
			left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as mftFiles , v.id as crlId from lab_rpki_mft c, lab_rpki_crl_revoked_cert_view v
			 	where c.jsonAll->>'$.eeCertModel.sn' = v.sn and c.aki =v.aki
			 	group by v.id) mft on mft.crlId = c.id	 `
		rows, err := xormdb.XormEngine.SQL(sql).Rows(chainCrlSql)
		if err != nil {
			belogs.Error("getChainCrlSqlDb(): select lab_rpki_crl fail:", err)
			return err
		}
		defer rows.Close()
		var index uint64
		for rows.Next() {
			chainCrlSql = new(ChainCertSql)
			err = rows.Scan(chainCrlSql)
			if err != nil {
				belogs.Error("getChainCrlSqlDb(): Scan fail:", err)
				continue
			}
			crlWg.Add(1)
			belogs.Debug("getChainCrlSqlDb(): crlWg.Add(), chainCrlSql.Id:", chainCrlSql.Id)
			chainCrlSqlCh <- chainCrlSql
			index++
		}
		belogs.Info("getChainCrlSqlDb(): Scan all, close chainCrlSqlCh, index:", index, "  time(s):", time.Since(start))
		return nil
	}
*/
func getChainCrlSqlsDb() (chainCertSqls []ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.syncLogId, c.jsonAll, c.state ,cer.cerFiles,roa.roaFiles, mft.mftFiles 
		from lab_rpki_crl c  
		left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as cerFiles , v.id as crlId from lab_rpki_cer c, lab_rpki_crl_revoked_cert_view v 
			 where c.sn = v.sn and c.aki =v.aki 
			 group by v.id) cer on cer.crlId = c.id	
		left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as roaFiles , v.id as crlId from lab_rpki_roa c, lab_rpki_crl_revoked_cert_view v 
			 where c.jsonAll->>'$.eeCertModel.sn' = v.sn and c.aki =v.aki 
			 group by v.id) roa on roa.crlId = c.id 
		left join (select GROUP_CONCAT(CONCAT(c.filePath,c.fileName) SEPARATOR  ',') as mftFiles , v.id as crlId from lab_rpki_mft c, lab_rpki_crl_revoked_cert_view v 
			 where c.jsonAll->>'$.eeCertModel.sn' = v.sn and c.aki =v.aki 
			 group by v.id) mft on mft.crlId = c.id	 `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("getChainCrlSqlsDb(): lab_rpki_crl id fail:", err)
		return nil, err
	}
	belogs.Info("getChainCrlSqlsDb(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Since(start))
	return chainCertSqls, nil
}

func updateCrlDb(session *xorm.Session, chains *Chains, crlId uint64,
	updateConcurrentCountCh chan int, updateWg *sync.WaitGroup) (err error) {
	defer func() {
		<-updateConcurrentCountCh
		updateWg.Done()
	}()

	start := time.Now()
	chainCrl, err := chains.GetCrlById(crlId)
	if err != nil {
		belogs.Error("updateCrlDb(): GetCrlById fail :", crlId, err)
		return err
	}

	chainDbCrlModel := NewChainDbCrlModel(&chainCrl)
	originModel := model.JudgeOriginByFilePath(chainCrl.FilePath)
	belogs.Debug("updateCrlDb():chainDbCrlModel.id:", chainDbCrlModel.Id,
		"  originModel:", jsonutil.MarshalJson(originModel))

	chainCerts := jsonutil.MarshalJson(*chainDbCrlModel)
	state := jsonutil.MarshalJson(chainCrl.StateModel)
	originJson := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateCrlDb():crlId:", crlId, "   chainCerts:", chainCerts,
		"   state:", state, "  originJson:", originJson)
	sqlStr := `UPDATE lab_rpki_crl set chainCerts=?, state=?, 
	origin = (
		case
		  when   origin = ''    then ? 
		  when   origin is null   then  ? 
		else 
		  origin
		end
	)
	where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, originJson, originJson, crlId)
	if err != nil {
		belogs.Error("updateCrlDb(): UPDATE lab_rpki_crl fail :", crlId, err)
		return err
	}
	belogs.Debug("updateCrlDb(): ok, crlId:", crlId, "  time(s):", time.Since(start))
	return nil
}

func UpdateCrls(chains *Chains) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("UpdateCrls(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	crlIds := chains.CrlIds
	belogs.Debug("UpdateCrls():len(crlIds):", len(crlIds))

	var updateWg sync.WaitGroup
	updateConcurrentCountCh := make(chan int, conf.Int("chain::updateConcurrentCount"))
	for _, crlId := range crlIds {
		updateWg.Add(1)
		updateConcurrentCountCh <- 1
		go updateCrlDb(session, chains, crlId, updateConcurrentCountCh, &updateWg)
	}
	updateWg.Wait()
	close(updateConcurrentCountCh)

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateCrls(): CommitSession fail :", err)
		return err
	}
	belogs.Info("UpdateCrls(): ok, len(crlIds):", len(crlIds), "  time(s):", time.Since(start))
	return nil
}
