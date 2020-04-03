package db

import (
	"errors"
	"time"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	statisticmodel "statistic/model"
	"statistic/util"
)

func GetLatestSync() (syncModel statisticmodel.SyncModel, err error) {
	start := time.Now()
	sql := `
		select id as syncLogId, syncStyle,
	        rsyncState->>'$.startTime' as rsyncStartTime, rsyncState->>'$.endTime' as rsyncEndTime,
	        rrdpState->>'$.startTime' as rrdpStartTime, rrdpState->>'$.endTime' as rrdpEndTime
		from lab_rpki_sync_log order by id desc limit 1`
	has, err := xormdb.XormEngine.Sql(sql).Get(&syncModel)
	if err != nil {
		belogs.Error("GetLatestSync():select id as syncLogId from lab_rpki_sync_log, fail:", err)
		return syncModel, err
	}
	if !has {
		belogs.Error("GetLatestSync(): latest sync is no exist :")
		return syncModel, errors.New("latest sync is no exist")
	}
	belogs.Debug("GetLatestSync():syncModel :", jsonutil.MarshalJson(syncModel), "    time(s):", time.Now().Sub(start).Seconds())
	return syncModel, nil
}

func GetRirFileCountDb() (rirFileCountDbs []statisticmodel.RirFileCountDb, err error) {
	/*
		sql := `select c.filePath, c.fileName,  c.state,c.origin, f.syncTime, f.syncType,f.syncLogId
		        from ` + table + ` c, lab_rpki_sync_log_file f
		        where c.state->'$.state' in ('invalid' , 'warning') and c.syncLogFileId = f.id order by c.id`
		belogs.Debug("GetStatisticDb():statisticsDbs  sql:", sql)
		err = xormdb.XormEngine.SQL(sql).Find(&statisticsDbs)
		if err != nil {
			belogs.Error("GetStatisticDb(): statisticsDbs sql fail :", sql, err)
			return nil, nil, err
		}
	*/
	start := time.Now()
	sql := `
		select c.origin->>'$.rir' as rir, count(*) as count ,  c.state->>'$.state' as state  ,'cer' as fileType 
	        from lab_rpki_cer c group by c.origin->>'$.rir', c.state->>'$.state' 
		union all
		select c.origin->>'$.rir' as rir, count(*) as count ,  c.state->>'$.state' as state  ,'crl' as fileType 
	        from lab_rpki_crl c group by c.origin->>'$.rir', c.state->>'$.state' 
		union all
		select c.origin->>'$.rir' as rir, count(*) as count ,  c.state->>'$.state' as state  ,'mft' as fileType 
		    from lab_rpki_mft c group by c.origin->>'$.rir', c.state->>'$.state' 
		union all
		select c.origin->>'$.rir' as rir, count(*) as count ,  c.state->>'$.state' as state  ,'roa' as fileType 
	        from lab_rpki_roa c group by c.origin->>'$.rir', c.state->>'$.state'
		order by rir, filetype 	`
	belogs.Debug("GetRirFileCountDb():rirFileCountDbs sql:", sql)
	err = xormdb.XormEngine.SQL(sql).Find(&rirFileCountDbs)
	if err != nil {
		belogs.Error("GetRirFileCountDb():rirFileCountDbs fail :", sql, err)
		return nil, err
	}
	belogs.Debug("GetRirFileCountDb():len(rirFileCountDbs):", len(rirFileCountDbs), "    time(s):", time.Now().Sub(start).Seconds())
	return rirFileCountDbs, nil
}

func GetRepoFileCountDb(rir string) (repoFileCountDbs []statisticmodel.RepoFileCountDb, err error) {
	start := time.Now()
	sql := `
	    select  c.origin->>'$.repo' as repo,count(*) as count ,  c.state->>'$.state' as state  ,'cer' as fileType 
	        from lab_rpki_cer c where c.origin->>'$.rir'='` + rir + `' group by c.origin->>'$.rir',c.origin->>'$.repo' , c.state->>'$.state' 
		union all
		select  c.origin->>'$.repo' as repo,count(*) as count ,  c.state->>'$.state' as state  ,'crl' as fileType 
	        from lab_rpki_crl c  where c.origin->>'$.rir'='` + rir + `' group by c.origin->>'$.rir', c.origin->>'$.repo' ,c.state->>'$.state' 
		union all
		select  c.origin->>'$.repo' as repo,count(*) as count ,  c.state->>'$.state' as state  ,'mft' as fileType 
		    from lab_rpki_mft c  where c.origin->>'$.rir'='` + rir + `' group by c.origin->>'$.rir',c.origin->>'$.repo' , c.state->>'$.state' 
		union all
		select  c.origin->>'$.repo' as repo,count(*) as count ,  c.state->>'$.state' as state  ,'roa' as fileType 
	        from lab_rpki_roa c  where c.origin->>'$.rir'='` + rir + `' group by c.origin->>'$.rir',c.origin->>'$.repo' , c.state->>'$.state'
		order by repo,filetype`
	belogs.Debug("GetRepoFileCountDb():repoFileCountDbs sql:", sql)
	err = xormdb.XormEngine.SQL(sql).Find(&repoFileCountDbs)
	if err != nil {
		belogs.Error("GetRepoFileCountDb():repoFileCountDbs fail :", sql, err)
		return nil, err
	}
	belogs.Debug("GetRepoFileCountDb():len(repoFileCountDbs):", len(repoFileCountDbs), "    time(s):", time.Now().Sub(start).Seconds())
	return repoFileCountDbs, nil
}

func GetFileState(rir, repo string) (fileStates []statisticmodel.FileState, err error) {
	start := time.Now()
	sql := `
	    select c.origin->>'$.rir' as rir, c.origin->>'$.repo' as repo, c.filePath, c.fileName, 'cer' as fileType, c.state->>'$.state' as state, 
		    c.state as stateFailDetails from lab_rpki_cer c 
		    where c.state->>'$.state' in ('invalid','warning') and c.origin->>'$.rir'='` + rir + `' and c.origin->>'$.repo'='` + repo + `'
		union all
		select c.origin->>'$.rir' as rir, c.origin->>'$.repo' as repo, c.filePath, c.fileName,'mft' as fileType, c.state->>'$.state' as state, 
			c.state as stateFailDetails from lab_rpki_mft c 
			where c.state->>'$.state' in ('invalid','warning') and c.origin->>'$.rir'='` + rir + `' and c.origin->>'$.repo'='` + repo + `' 
		union all
		select c.origin->>'$.rir' as rir, c.origin->>'$.repo' as repo, c.filePath, c.fileName,'roa' as fileType, c.state->>'$.state' as state, 
			c.state as stateFailDetails from lab_rpki_roa c 
			where c.state->>'$.state' in ('invalid','warning') and c.origin->>'$.rir'='` + rir + `' and c.origin->>'$.repo'='` + repo + `'
		union all
		select c.origin->>'$.rir' as rir, c.origin->>'$.repo' as repo, c.filePath, c.fileName,'crl' as fileType, c.state->>'$.state' as state, 
			c.state as stateFailDetails from lab_rpki_crl c 
			where c.state->>'$.state' in ('invalid','warning') and c.origin->>'$.rir'='` + rir + `' and c.origin->>'$.repo'='` + repo + `'
		order by fileType,state, filePath, fileName `
	belogs.Debug("GetFileState():fileStates sql:", sql)
	err = xormdb.XormEngine.SQL(sql).Find(&fileStates)
	if err != nil {
		belogs.Error("GetFileState():fileStates fail :", sql, err)
		return nil, err
	}
	for i, _ := range fileStates {
		fileStates[i].Url = util.GetUrlByFilePathName(fileStates[i].FilePath, fileStates[i].FileName)
		fileStates[i].FailDetails = util.GetFailDetailsByState(fileStates[i].StateFailDetails)
	}
	belogs.Debug("GetFileState():len(fileStates):", len(fileStates), "    time(s):", time.Now().Sub(start).Seconds(),
		"   rir, repo:", rir, repo, jsonutil.MarshalJson(fileStates))
	return fileStates, nil
}

// update
func UpdateStatistic(rirs []statisticmodel.RirStatisticModel) (err error) {
	if len(rirs) == 0 {
		return nil
	}
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	syncLogId := rirs[0].SyncModel.SyncLogId

	// delete same synclogId
	_, err = session.Exec("delete from lab_rpki_statistic  where sync->>'$.syncLogId' = ?", syncLogId)
	if err != nil {
		belogs.Error("UpdateStatistic():delete  from lab_rpki_statistic failed, syncLogId:", syncLogId, err)
		return xormdb.RollbackAndLogError(session, "UpdateStatistic(): delete  from lab_rpki_statistic failed: "+
			convert.ToString(syncLogId), err)
	}
	sql := `insert into lab_rpki_statistic
			  (rir, cerFileCount, crlFileCount, mftFileCount, roaFileCount,      repos, sync) 
		values(?,?,?,?,?,   ?,?)`
	for i, _ := range rirs {
		rir := rirs[i].Rir
		cerFileCount := jsonutil.MarshalJson(rirs[i].CerFileCount)
		crlFileCount := jsonutil.MarshalJson(rirs[i].CrlFileCount)
		mftFileCount := jsonutil.MarshalJson(rirs[i].MftFileCount)
		roaFileCount := jsonutil.MarshalJson(rirs[i].RoaFileCount)
		repos := jsonutil.MarshalJson(rirs[i].RepoStatisticModels)
		sync := jsonutil.MarshalJson(rirs[i].SyncModel)
		_, err = session.Exec(sql, rir, cerFileCount, crlFileCount, mftFileCount, roaFileCount, repos, sync)
		if err != nil {
			belogs.Error("UpdateStatistic():insert lab_rpki_statistic failed:", jsonutil.MarshalJson(rirs[i]), err)
			return xormdb.RollbackAndLogError(session, "SaveCertCount(): insert lab_rpki_statistic failed failed: "+
				jsonutil.MarshalJson(rirs[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateStatistic(): CommitSession fail :", err)
		return err
	}
	belogs.Debug("UpdateStatistic():  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
