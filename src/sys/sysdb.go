package sys

import (
	"errors"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/randutil"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

// initResetDb 初始化数据库
//
//	@param sysStyle  初始化类型
//	@return error 返回错误
func initResetDb(sysStyle SysStyle) error {
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("initResetDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	//truncate all table
	err = initResetImplDb(session, sysStyle)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "initResetDb(): initResetImplDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "initResetDb(): CommitSession fail", err)
	}
	return nil
}

// initResetImplDb 实际执行初始化数据库命令, 其中注意要初始化sessionId, 生成新值
//
//	@param session
//	@param sysStyle
//	@return error
func initResetImplDb(session *xorm.Session, sysStyle SysStyle) error {
	defer func(session1 *xorm.Session) {
		sql := `set foreign_key_checks=1;`
		if _, err := session1.Exec(sql); err != nil {
			belogs.Error("initResetImplDb(): SET foreign_key_checks=1 fail", err)
			xormdb.RollbackAndLogError(session, "initResetImplDb():SET foreign_key_checks=1 fail", err)
		}
	}(session)

	start := time.Now()
	sql := `set foreign_key_checks=0;`
	if _, err := session.Exec(sql); err != nil {
		belogs.Error("initResetImplDb(): SET foreign_key_checks=0 fail", err)
		return xormdb.RollbackAndLogError(session, "initResetImplDb():SET foreign_key_checks=0 fail: ", err)
	}
	belogs.Debug("initResetImplDb():foreign_key_checks=0; time(s):", time.Since(start))

	// delete rtr_session
	sqls := make([]string, 0, 200)
	switch sysStyle.SysStyle {
	case "init":
		sqls = append(sqls, dropSqls...)
		sqls = append(sqls, createSqls...)
	case "fullsync":
		sqls = append(sqls, truncateSqls...)
		sqls = append(sqls, optimizeSqls...)
	}
	//belogs.Debug("initResetImplDb():will Exec sqls:", jsonutil.MarshalJson(sqls))
	belogs.Debug("initResetImplDb():will Exec len(sqls):", len(sqls))
	for _, sq := range sqls {
		if _, err := session.Exec(sq); err != nil {
			belogs.Error("initResetImplDb():  "+sq+" fail", err)
			return xormdb.RollbackAndLogError(session, "initResetImplDb():sql fail: "+sq, err)
		}
	}
	belogs.Info("initResetImplDb(): len(sqls):", len(sqls), "  time(s):", time.Since(start))

	// generate new session random, insert lab_rpki_rtr_session
	rtrSession := model.LabRpkiRtrSession{}
	rtrSession.SessionId = uint64(randutil.IntRange(99, 999))
	rtrSession.CreateTime = time.Now()
	belogs.Info("initResetImplDb():insert rtr_session, rtrSession:", jsonutil.MarshalJson(rtrSession))
	if _, err := session.Insert(&rtrSession); err != nil {
		belogs.Error("initResetImplDb():insert rtr_session fail", err)
		return xormdb.RollbackAndLogError(session, "initResetImplDb():insert rtr_session fail", err)
	}

	// insert _conf
	sql = `insert lab_rpki_conf ( section, myKey, myValue, defaultMyValue, updateTime) 
			values(?,?,?,?,?) `
	_, err := session.Exec(sql, "rpOperate", "cacheUpdateType", "manual", "manual", time.Now())
	if err != nil {
		belogs.Error("initResetImplDb(): insert lab_rpki_conf fail", err)
		return xormdb.RollbackAndLogError(session, "initResetImplDb():insert lab_rpki_conf fail", err)
	}

	belogs.Info("initResetImplDb(): all are done, len(sqls):", len(sqls), "  time(s):", time.Since(start))
	return nil
}

func getResultsDb() (results CertResults, err error) {
	results.CerResult, err = getResultDb("lab_rpki_cer", "cer")
	if err != nil {
		belogs.Error("getResultsDb():select lab_rpki_cer, fail:", err)
		return results, err
	}
	results.CrlResult, err = getResultDb("lab_rpki_crl", "crl")
	if err != nil {
		belogs.Error("getResultsDb():select lab_rpki_crl , fail:", err)
		return results, err
	}
	results.MftResult, err = getResultDb("lab_rpki_mft", "mft")
	if err != nil {
		belogs.Error("getResultsDb():select lab_rpki_mft, fail:", err)
		return results, err
	}
	results.RoaResult, err = getResultDb("lab_rpki_roa", "roa")
	if err != nil {
		belogs.Error("getResultsDb():select lab_rpki_roa, fail:", err)
		return results, err
	}
	results.AsaResult, err = getResultDb("lab_rpki_asa", "asa")
	if err != nil {
		belogs.Error("getResultsDb():select lab_rpki_asa, fail:", err)
		return results, err
	}
	results.MoaResult, err = getResultDb("lab_rpki_moa", "moa")
	if err != nil {
		belogs.Error("getResultsDb():select lab_rpki_moa, fail:", err)
		return results, err
	}
	return results, nil
}

func getResultDb(table, fileType string) (result CertResult, err error) {
	sql :=
		`select al.count as allCount, va.count as validCount, wa.count as warningCount, ia.count as invalidCount , '` + fileType + `' as fileType  from 
		(select count(*) as count from ` + table + ` c) al,
		(select count(*) as count from ` + table + ` c where c.state->>"$.state" ='valid' ) va,
		(select count(*) as count from ` + table + ` c where c.state->>"$.state" ='warning') wa,
		(select count(*) as count from ` + table + ` c where c.state->>"$.state" ='invalid') ia`
	has, err := xormdb.XormEngine.SQL(sql).Get(&result)
	if err != nil {
		belogs.Error("getResultDb():select count, fail:", table, err)
		return result, err
	}
	if !has {
		belogs.Error("getResultDb(): not get count, fail:", table)
		return result, errors.New("not get count")
	}
	belogs.Debug("getResultDb():result :", jsonutil.MarshalJson(result))
	return result, nil
}

func exportRoasDb() (exportRoas []ExportRoa, err error) {
	sql :=
		`select asn, addressPrefix, maxLength, rir, repo 
		from lab_rpki_roa_ipaddress_view v
		order by rir, repo,addressPrefix,maxLength,asn`
	err = xormdb.XormEngine.SQL(sql).Find(&exportRoas)
	if err != nil {
		belogs.Error("exportRoasDb():Find, fail:", err)
		return nil, err
	}

	belogs.Debug("exportRoasDb():len(exportRoas):", len(exportRoas))
	return exportRoas, nil
}

func exportRtrForManrsDb() (rtrForManrss []RtrForManrs, err error) {
	rtrForManrss = make([]RtrForManrs, 0)
	sql :=
		`select asn, address, prefixLength,maxLength as max_length 
		from lab_rpki_rtr_full order by id `
	err = xormdb.XormEngine.SQL(sql).Find(&rtrForManrss)
	if err != nil {
		belogs.Error("exportRtrForManrsDb():Find, fail:", err)
		return nil, err
	}

	belogs.Debug("exportRtrForManrsDb():len(rtrForManrss):", len(rtrForManrss))
	return rtrForManrss, nil
}

func getChainCertsDb() (chainCertIds []ChainCertId, certIdRepos []CertIdRepo, err error) {
	chainCertIdSql := `
		select id, chainCerts->'$.parentChainCers' as parentcers ,'cer' as filetype from  lab_rpki_cer					 
		union
		select id, chainCerts->'$.parentChainCers' as parentcers,'crl' as filetype  from  lab_rpki_crl				
		union
		select id, chainCerts->'$.parentChainCers' as parentcers,'mft' as filetype  from  lab_rpki_mft
		union
		select id, chainCerts->'$.parentChainCers' as parentcers,'roa' as filetype  from  lab_rpki_roa
		`
	chainCertIds = make([]ChainCertId, 0)
	err = xormdb.XormEngine.SQL(chainCertIdSql).Find(&chainCertIds)
	if err != nil {
		belogs.Error("getChainCertsDb():Find chainCertIds, fail:", err)
		return nil, nil, err
	}

	certIdReposSql := `
		select c.id, SUBSTRING_INDEX(SUBSTRING_INDEX(l.sourceurl, '/', 3), '//', -1) as repo, 'cer' as filetype from lab_rpki_sync_log_file l, lab_rpki_cer c where c.synclogFileId = l.id  
		union
		select c.id, SUBSTRING_INDEX(SUBSTRING_INDEX(l.sourceurl, '/', 3), '//', -1) as repo, 'crl' as filetype from lab_rpki_sync_log_file l, lab_rpki_crl c where c.synclogFileId = l.id  
		union
		select c.id, SUBSTRING_INDEX(SUBSTRING_INDEX(l.sourceurl, '/', 3), '//', -1) as repo, 'mft' as filetype from lab_rpki_sync_log_file l, lab_rpki_mft c where c.synclogFileId = l.id  
		union
		select c.id, SUBSTRING_INDEX(SUBSTRING_INDEX(l.sourceurl, '/', 3), '//', -1) as repo, 'roa' as filetype from lab_rpki_sync_log_file l, lab_rpki_roa c where c.synclogFileId = l.id  
		`
	certIdRepos = make([]CertIdRepo, 0)
	err = xormdb.XormEngine.SQL(certIdReposSql).Find(&certIdRepos)
	if err != nil {
		belogs.Error("getChainCertsDb():Find certIdRepos, fail:", err)
		return nil, nil, err
	}
	return chainCertIds, certIdRepos, nil
}
