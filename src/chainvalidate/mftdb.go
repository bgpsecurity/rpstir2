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

func GetChainMftData(chainMftDataCh chan []*ChainCertData, mftWg *sync.WaitGroup) error {

	start := time.Now()
	belogs.Debug("getChainMftSqlDb(): will select rpki_mft")
	var myCount int64
	sql := `select count(*) as myCount from lab_rpki_mft c`
	found, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("getChainMftSqlDb():select rpki_mft fail:", err)
		return err
	}
	if !found {
		belogs.Debug("getChainMftSqlDb():select rpki_mft empty:")
		return nil
	}
	belogs.Debug("getChainMftSqlDb():select myCount rpki_mft, myCount:", myCount)

	oneSize := int64(conf.Int("chain::getCertsCountInOneSize"))
	sql = `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_mft c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki   
			order by id limit ?,? `
	/*
		var offset, selectCount int
		for offset = 0; offset < myCount; offset += oneSize {
			startOne := time.Now()
			chainMftSqls := make([]*ChainCertSql, 0)
			err = xormdb.XormEngine.SQL(sql, offset, oneSize).Find(&chainMftSqls)
			if err != nil {
				belogs.Error("getChainMftSqlDb():select rpki_mft limit,fail:", err, " one sql time(s):", time.Since(startOne))
				return err
			}
			belogs.Debug("getChainMftSqlDb(): select rpki_mft limit, len(chainMftSqls):", len(chainMftSqls),
				"  offset:", offset, "   oneSize:", oneSize,
				"  one sql time(s):", time.Since(startOne))
			selectCount += len(chainMftSqls)
			mftWg.Add(1)
			chainMftSqlsCh <- chainMftSqls
		}
	*/
	selectCount := getChainCertSqlDb(oneSize, myCount, sql, mftWg, chainMftDataCh)
	belogs.Info("getChainMftSqlDb(): select rpki_mft limit all, selectCount:", selectCount,
		"  myCount:", myCount, "  time(s):", time.Since(start))
	return nil
}

/*
	func getChainMftSqlDb(chainMftSqlCh chan *ChainCertSql, mftWg *sync.WaitGroup) error {
		defer close(chainMftSqlCh)

		start := time.Now()
		belogs.Debug("getChainMftSqlDb(): will select lab_rpki_mft")
		chainMftSql := new(ChainCertSql)
		// if add "order by ***", the sort_mem may not enough
		sql := `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime
				from lab_rpki_mft c
				left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki
				group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime `
		rows, err := xormdb.XormEngine.SQL(sql).Rows(chainMftSql)
		if err != nil {
			belogs.Error("getChainMftSqlDb(): select lab_rpki_mft fail:", err)
			return err
		}
		defer rows.Close()
		var index uint64
		for rows.Next() {
			chainMftSql = new(ChainCertSql)
			err = rows.Scan(chainMftSql)
			if err != nil {
				belogs.Error("getChainMftSqlDb(): Scan fail:", err)
				continue
			}
			mftWg.Add(1)
			belogs.Debug("getChainMftSqlDb(): mftWg.Add(), chainMftSql.Id:", chainMftSql.Id)
			chainMftSqlCh <- chainMftSql
			index++
			belogs.Debug("getChainMftSqlDb(): Scan index:", index, "  time(s):", time.Since(start))
		}
		belogs.Info("getChainMftSqlDb(): Scan all, close chainMftSqlCh, index:", index, "  time(s):", time.Since(start))
		return nil
	}
*/
func getChainMftSqlsDb() (chainCertSqls []ChainCertSql, err error) {
	start := time.Now()
	chainCertSqls = make([]ChainCertSql, 0, 50000)
	// if add "order by ***", the sort_mem may not enough
	sql := `select c.id, c.syncLogId, c.jsonAll, c.state, v.fileName as crlFileName, v.revocationTime 
			from lab_rpki_mft c 
			left join lab_rpki_crl_revoked_cert_view v on v.sn = c.jsonAll->>'$.eeCertModel.sn' and c.aki = v.aki   
			group by c.id, c.jsonAll, c.state, v.fileName, v.revocationTime `
	err = xormdb.XormEngine.SQL(sql).Find(&chainCertSqls)
	if err != nil {
		belogs.Error("getChainMftSqlsDb(): lab_rpki_mft id fail:", err)
		return nil, err
	}
	belogs.Info("getChainMftSqlsDb(): len(chainCertSqls):", len(chainCertSqls), "  time(s):", time.Since(start))
	return chainCertSqls, nil
}

func GetChainFileHashs(chainMft ChainMft) (chainFileHashs []ChainFileHash, err error) {
	start := time.Now()
	mftId := chainMft.Id
	belogs.Debug("SQLDataSource.GetChainFileHashs(): will select lab_rpki_mft_file_hash_view, mftId:", mftId)
	/*
		sql := `select  v.file as file, v.hash as hash,CONCAT(IFNULL(cer.filePath,''),IFNULL(crl.filePath,''),IFNULL(roa.filePath,'')) as path
				from lab_rpki_mft_file_hash_view v
					left join lab_rpki_cer cer on cer.aki=v.aki and cer.fileName=v.file
					left join lab_rpki_crl crl on crl.aki=v.aki and crl.fileName=v.file
					left join lab_rpki_roa roa on roa.aki=v.aki and roa.fileName=v.file
				where v.mftId=?
				order by v.mftFileHashId `
		err = xormdb.XormEngine.SQL(sql, mftId).Find(&chainFileHashs)
		if err != nil {
			belogs.Error("getChainFileHashsDb(): lab_rpki_mft_file_hash fail:", err, "  time(s):", time.Since(start))
			return chainFileHashs, err
		}
	*/
	chainFileHashs = make([]ChainFileHash, 0)

	asaSql := `select h.file, h.hash, c.filepath as path from lab_rpki_mft_file_hash h ,lab_rpki_mft m, lab_rpki_asa c 
			where  m.id = h.mftId and c.aki = m.aki  and c.filename = h.file and m.id = ? 
			order by m.id,h.file`
	asaFs := make([]ChainFileHash, 0)
	err = xormdb.XormEngine.SQL(asaSql, mftId).Find(&asaFs)
	if err != nil {
		belogs.Error("SQLDataSource.GetChainFileHashs(): mft_file_hash and asa fail, mftId:", mftId, err, "  time(s):", time.Since(start))
		return nil, err
	}
	belogs.Debug("SQLDataSource.GetChainFileHashs():mft_file_hash and asa, mftId:", mftId, "    len(asaFs):", len(asaFs), "  time(s):", time.Since(start))
	if len(asaFs) > 0 {
		chainFileHashs = append(chainFileHashs, asaFs...)
		belogs.Debug("SQLDataSource.GetChainFileHashs():mft_file_hash and len(asaFs), mftId:", mftId, "    len(asaFs):", len(asaFs), "  time(s):", time.Since(start))
	}

	cerSql := `select h.file, h.hash, c.filepath as path from lab_rpki_mft_file_hash h ,lab_rpki_mft m, lab_rpki_cer c 
			where  m.id = h.mftId and c.aki = m.aki  and c.filename = h.file and m.id = ? 
			order by m.id,h.file`
	cerFs := make([]ChainFileHash, 0)
	err = xormdb.XormEngine.SQL(cerSql, mftId).Find(&cerFs)
	if err != nil {
		belogs.Error("SQLDataSource.GetChainFileHashs(): mft_file_hash and cer fail, mftId:", mftId, err, "  time(s):", time.Since(start))
		return nil, err
	}
	belogs.Debug("SQLDataSource.GetChainFileHashs():mft_file_hash and cer, mftId:", mftId, "    len(cerFs):", len(cerFs), "  time(s):", time.Since(start))
	if len(cerFs) > 0 {
		chainFileHashs = append(chainFileHashs, cerFs...)
		belogs.Debug("SQLDataSource.GetChainFileHashs():mft_file_hash and len(cerFs), mftId:", mftId, "    len(cerFs):", len(cerFs), "  time(s):", time.Since(start))
	}

	crlSql := `select  h.file, h.hash, c.filepath as path from lab_rpki_mft_file_hash h ,lab_rpki_mft m, lab_rpki_crl c 
				where  m.id = h.mftId and c.aki = m.aki  and c.filename = h.file and m.id = ? 
				order by m.id,h.file`
	crlFs := make([]ChainFileHash, 0)
	err = xormdb.XormEngine.SQL(crlSql, mftId).Find(&crlFs)
	if err != nil {
		belogs.Error("SQLDataSource.GetChainFileHashs(): mft_file_hash and crl fail, mftId:", mftId, err, "  time(s):", time.Since(start))
		return nil, err
	}
	belogs.Debug("SQLDataSource.GetChainFileHashs():mft_file_hash and crl, mftId:", mftId, "    len(crlFs):", len(crlFs), "  time(s):", time.Since(start))
	if len(crlFs) > 0 {
		chainFileHashs = append(chainFileHashs, crlFs...)
		belogs.Debug("SQLDataSource.GetChainFileHashs():mft_file_hash and len(crlFs), mftId:", mftId, "    len(crlFs):", len(crlFs), "  time(s):", time.Since(start))
	}

	roaSql := `select  h.file, h.hash, c.filepath as path from lab_rpki_mft_file_hash h ,lab_rpki_mft m, lab_rpki_roa c 
			where  m.id = h.mftId and c.aki = m.aki  and c.filename = h.file and m.id = ? 
			order by m.id,h.file`
	roaFs := make([]ChainFileHash, 0)
	err = xormdb.XormEngine.SQL(roaSql, mftId).Find(&roaFs)
	if err != nil {
		belogs.Error("SQLDataSource.GetChainFileHashs(): mft_file_hash and roa fail, mftId:", mftId, err, "  time(s):", time.Since(start))
		return nil, err
	}
	belogs.Debug("SQLDataSource.GetChainFileHashs():mft_file_hash and roa, mftId:", mftId, "    len(roaFs):", len(roaFs), "  time(s):", time.Since(start))
	if len(roaFs) > 0 {
		chainFileHashs = append(chainFileHashs, roaFs...)
		belogs.Debug("SQLDataSource.GetChainFileHashs():mft_file_hash and len(roaFs), mftId:", mftId, "    len(roaFs):", len(roaFs), "  time(s):", time.Since(start))
	}

	moaSql := `select  h.file, h.hash, c.filepath as path from lab_rpki_mft_file_hash h ,lab_rpki_mft m, lab_rpki_moa c 
			where  m.id = h.mftId and c.aki = m.aki  and c.filename = h.file and m.id = ? 
			order by m.id,h.file`
	moaFs := make([]ChainFileHash, 0)
	err = xormdb.XormEngine.SQL(moaSql, mftId).Find(&moaFs)
	if err != nil {
		belogs.Error("SQLDataSource.GetChainFileHashs(): mft_file_hash and moa fail, mftId:", mftId, err, "  time(s):", time.Since(start))
		return nil, err
	}
	belogs.Debug("SQLDataSource.GetChainFileHashs():mft_file_hash and moa, mftId:", mftId, "    len(moaFs):", len(moaFs), "  time(s):", time.Since(start))
	if len(moaFs) > 0 {
		chainFileHashs = append(chainFileHashs, moaFs...)
		belogs.Debug("SQLDataSource.GetChainFileHashs():mft_file_hash and len(moaFs), mftId:", mftId, "    len(moaFs):", len(moaFs), "  time(s):", time.Since(start))
	}
	belogs.Info("SQLDataSource.GetChainFileHashs():get from mftId:", mftId,
		"    chainFileHashs:", chainFileHashs, "  time(s):", time.Since(start))

	return chainFileHashs, nil
}

func GetPreviousMft(chainMft ChainMft) (previousMft PreviousMft, err error) {
	start := time.Now()
	mftId := chainMft.Id
	/* // because using json directly, it will cause ' Out of sort memory, consider increasing server sort buffer size'
	sql := `select f.jsonAll->>'$.mftNumber' as mftNumber,f.jsonAll->>'$.thisUpdate' as thisUpdate, f.jsonAll->>'$.nextUpdate' as nextUpdate
	        from  lab_rpki_sync_log_file f ,
				( select m.mftNumber,m.filePath,m.fileName,m.syncLogId from lab_rpki_mft m where m.id = ? ) t
			where f.filePath = t.filePath and f.fileName=t.fileName and f.syncLogId < t.syncLogId
			order by f.syncLogId desc limit 1  `
	*/
	// get id ,then select from json
	sql := `select  ff.jsonAll->>'$.mftNumber' as mftNumber,ff.jsonAll->>'$.thisUpdate' as thisUpdate, ff.jsonAll->>'$.nextUpdate' as nextUpdate 
			from lab_rpki_sync_log_file ff where ff.id = ( 
				select f.id from  lab_rpki_sync_log_file f , ( select m.mftNumber,m.filePath,m.fileName,m.syncLogId from lab_rpki_mft m where m.id = ? ) t 
				where f.filePath = t.filePath and f.fileName=t.fileName and f.syncLogId < t.syncLogId 
                order by f.syncLogId desc limit 1 
	     	)`
	found, err := xormdb.XormEngine.SQL(sql, mftId).Get(&previousMft)
	if err != nil {
		belogs.Error("SQLDataSource.GetPreviousMft(): lab_rpki_sync_log_file fail, mftId:", mftId, err, "  time(s):", time.Since(start))
		return previousMft, err
	}
	previousMft.Found = found
	belogs.Debug("SQLDataSource.GetPreviousMft():mftId:", mftId, "   previousMft:", previousMft,
		"  time(s):", time.Since(start))
	return previousMft, nil
}

func updateMftDb(session *xorm.Session, chains *Chains, mftId uint64,
	updateConcurrentCountCh chan int, updateWg *sync.WaitGroup) (err error) {
	defer func() {
		<-updateConcurrentCountCh
		updateWg.Done()
	}()
	start := time.Now()
	chainMft, err := chains.GetMftById(mftId)
	if err != nil {
		belogs.Error("updateMftDb(): GetMftById fail :", mftId, err)
		return err
	}

	chainDbMftModel := NewChainDbMftModel(&chainMft)
	originModel := model.JudgeOriginByFilePath(chainMft.FilePath)
	belogs.Debug("updateMftDb():chainDbMftModel.id:", chainDbMftModel.Id,
		"  originModel:", jsonutil.MarshalJson(originModel))

	chainCerts := jsonutil.MarshalJson(*chainDbMftModel)
	state := jsonutil.MarshalJson(chainMft.StateModel)
	originJson := jsonutil.MarshalJson(originModel)
	belogs.Debug("updateMftDb():mftId:", mftId, "   chainCerts:", chainCerts,
		"   state:", state, "  originJson:", originJson)
	sqlStr := `UPDATE lab_rpki_mft set chainCerts=?, state=?,  
	origin = (
		case
		  when   origin = ''    then ? 
		  when   origin is null   then  ? 
		else 
		  origin
		end
	)
	where id=? `
	_, err = session.Exec(sqlStr, chainCerts, state, originJson, originJson, mftId)
	if err != nil {
		belogs.Error("updateMftDb(): UPDATE lab_rpki_mft fail :", mftId, err)
		return err
	}
	belogs.Debug("updateMftDb(): ok, mftId:", mftId, "  time(s):", time.Since(start))
	return nil
}

func UpdateMfts(chains *Chains) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("UpdateCrls(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	mftIds := chains.MftIds
	belogs.Debug("UpdateMfts():len(mftIds):", len(mftIds))

	var updateWg sync.WaitGroup
	updateConcurrentCountCh := make(chan int, conf.Int("chain::updateConcurrentCount"))
	for _, mftId := range mftIds {
		updateWg.Add(1)
		updateConcurrentCountCh <- 1
		go updateMftDb(session, chains, mftId, updateConcurrentCountCh, &updateWg)
	}
	updateWg.Wait()
	close(updateConcurrentCountCh)

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateMfts(): CommitSession fail :", err)
		return err
	}
	belogs.Info("UpdateMfts(): ok, len(mftIds):", len(mftIds), "  time(s):", time.Since(start))
	return nil
}
