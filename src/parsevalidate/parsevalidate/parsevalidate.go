package parsevalidate

import (
	"errors"
	"strings"
	"sync"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"

	"model"
	db "parsevalidate/db"
	parsevalidatemodel "parsevalidate/model"
)

// ParseValidateStart: start
func ParseValidateStart() (nextStep string, err error) {

	start := time.Now()
	belogs.Info("ParseValidateStart(): start")
	// save starttime to lab_rpki_sync_log
	labRpkiSyncLogId, err := db.UpdateRsyncLogParseValidateStart("parsevalidating")
	if err != nil {
		belogs.Error("ParseValidateStart():InsertRsyncLogRsyncStat fail:", err)
		return "", err
	}
	belogs.Debug("ParseValidateStart(): labRpkiSyncLogId:", labRpkiSyncLogId)

	// get all need rsyncLogFile
	syncLogFileModels, err := db.GetSyncLogFileModelsBySyncLogId(labRpkiSyncLogId)
	if err != nil {
		belogs.Error("ParseValidateStart():GetSyncLogFileModelsBySyncLogId fail:", labRpkiSyncLogId, err)
		return "", err
	}
	belogs.Debug("ParseValidateStart(): GetSyncLogFileModelsBySyncLogId, syncLogFileModels.SyncLogId:", labRpkiSyncLogId, syncLogFileModels.SyncLogId)

	//process "del" and "update" rsyncLogFile
	err = DelCertByDelAndUpdate(syncLogFileModels)
	if err != nil {
		belogs.Error("ParseValidateStart():DelCertByDelAndUpdate fail:", err)
		return "", err
	}
	belogs.Debug("ParseValidateStart(): after DelCertByDelAndUpdate, syncLogFileModels.SyncLogId:", syncLogFileModels.SyncLogId)

	// process "add" and "update" rsyncLogFile
	err = InsertCertByAddAndUpdate(syncLogFileModels)
	if err != nil {
		belogs.Error("ParseValidateStart():InsertCertByInsertAndUpdate fail:", err)
		return "", err
	}
	// save to db
	err = db.UpdateRsyncLogParseValidateStateEnd(labRpkiSyncLogId, "parsevalidated", make([]string, 0))
	if err != nil {
		belogs.Debug("ParseValidateStart(): UpdateRsyncLogAndCert fail: ", err)
		return "", err
	}

	belogs.Info("ParseValidateStart(): end, will call chainvalidate,  time(s):", time.Now().Sub(start).Seconds())
	return "chainvalidate", nil
}

// get del;
// get update, because "update" should del first
func DelCertByDelAndUpdate(syncLogFileModels *parsevalidatemodel.SyncLogFileModels) (err error) {
	start := time.Now()

	belogs.Debug("DelCertByDelAndUpdate(): syncLogFileModels.SyncLogId.:", syncLogFileModels.SyncLogId)

	var wg sync.WaitGroup

	// get "del" and "update" cer synclog files to del
	belogs.Debug("DelCertByDelAndUpdate(): len(syncLogFileModels.DelCerSyncLogFileModels):", len(syncLogFileModels.DelCerSyncLogFileModels),
		"       len(syncLogFileModels.UpdateCerSyncLogFileModels):", len(syncLogFileModels.UpdateCerSyncLogFileModels))
	if len(syncLogFileModels.DelCerSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateCerSyncLogFileModels) > 0 {
		wg.Add(1)
		go db.DelCers(syncLogFileModels.DelCerSyncLogFileModels, syncLogFileModels.UpdateCerSyncLogFileModels, &wg)
	}

	// get "del" and "update" crl synclog files to del
	belogs.Debug("DelCertByDelAndUpdate(): len(syncLogFileModels.DelCrlSyncLogFileModels):", len(syncLogFileModels.DelCrlSyncLogFileModels),
		"       len(syncLogFileModels.UpdateCrlSyncLogFileModels):", len(syncLogFileModels.UpdateCrlSyncLogFileModels))
	if len(syncLogFileModels.DelCrlSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateCrlSyncLogFileModels) > 0 {
		wg.Add(1)
		go db.DelCrls(syncLogFileModels.DelCrlSyncLogFileModels, syncLogFileModels.UpdateCrlSyncLogFileModels, &wg)
	}

	// get "del" and "update" mft synclog files to del
	belogs.Debug("DelCertByDelAndUpdate(): len(syncLogFileModels.DelMftSyncLogFileModels):", len(syncLogFileModels.DelMftSyncLogFileModels),
		"       len(syncLogFileModels.UpdateMftSyncLogFileModels):", len(syncLogFileModels.UpdateMftSyncLogFileModels))
	if len(syncLogFileModels.DelMftSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateMftSyncLogFileModels) > 0 {
		wg.Add(1)
		go db.DelMfts(syncLogFileModels.DelMftSyncLogFileModels, syncLogFileModels.UpdateMftSyncLogFileModels, &wg)
	}

	// get "del" and "update" roa synclog files to del
	belogs.Debug("DelCertByDelAndUpdate(): len(syncLogFileModels.DelRoaSyncLogFileModels):", len(syncLogFileModels.DelRoaSyncLogFileModels),
		"       len(syncLogFileModels.UpdateRoaSyncLogFileModels):", len(syncLogFileModels.UpdateRoaSyncLogFileModels))
	if len(syncLogFileModels.DelRoaSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateRoaSyncLogFileModels) > 0 {
		wg.Add(1)
		go db.DelRoas(syncLogFileModels.DelRoaSyncLogFileModels, syncLogFileModels.UpdateRoaSyncLogFileModels, &wg)
	}

	wg.Wait()
	belogs.Info("DelCertByDelAndUpdate(): end,  time(s):", time.Now().Sub(start).Seconds())
	return nil

}

// InsertCertByAddAndUpdate :  use update, because "update" should add
func InsertCertByAddAndUpdate(syncLogFileModels *parsevalidatemodel.SyncLogFileModels) (err error) {

	start := time.Now()
	belogs.Debug("InsertCertByInsertAndUpdate(): syncLogFileModels.SyncLogId:", syncLogFileModels.SyncLogId)

	var wg sync.WaitGroup

	// add/update crl
	belogs.Debug("InsertCertByInsertAndUpdate():len(syncLogFileModels.UpdateCerSyncLogFileModels):", len(syncLogFileModels.UpdateCerSyncLogFileModels))
	if len(syncLogFileModels.UpdateCerSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCerts(syncLogFileModels.UpdateCerSyncLogFileModels, "cer", &wg)
	}

	// add/update crl
	belogs.Debug("InsertCertByInsertAndUpdate():len(syncLogFileModels.UpdateCrlSyncLogFileModels):", len(syncLogFileModels.UpdateCrlSyncLogFileModels))
	if len(syncLogFileModels.UpdateCrlSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCerts(syncLogFileModels.UpdateCrlSyncLogFileModels, "crl", &wg)
	}

	// add/update mft
	belogs.Debug("InsertCertByInsertAndUpdate():len(syncLogFileModels.UpdateMftSyncLogFileModels):", len(syncLogFileModels.UpdateMftSyncLogFileModels))
	if len(syncLogFileModels.UpdateMftSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCerts(syncLogFileModels.UpdateMftSyncLogFileModels, "mft", &wg)
	}

	// add/update roa
	belogs.Debug("InsertCertByInsertAndUpdate():len(syncLogFileModels.UpdateRoaSyncLogFileModels):", len(syncLogFileModels.UpdateRoaSyncLogFileModels))
	if len(syncLogFileModels.UpdateRoaSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCerts(syncLogFileModels.UpdateRoaSyncLogFileModels, "roa", &wg)
	}

	wg.Wait()
	belogs.Info("InsertCertByInsertAndUpdate(): end,  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func parseValidateAndAddCerts(syncLogFileModels []parsevalidatemodel.SyncLogFileModel, fileType string, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()

	// parsevalidate
	belogs.Debug("parseValidateAndAddCerts(): len(syncLogFileModels):", len(syncLogFileModels), "  fileType:", fileType)
	var parseValidateWg sync.WaitGroup
	parseValidateCh := make(chan int, conf.Int("parse::parseConcurrentCount"))
	for i := range syncLogFileModels {
		parseValidateWg.Add(1)
		parseValidateCh <- 1
		go parseValidateCert(&syncLogFileModels[i], &parseValidateWg, parseValidateCh)
	}
	parseValidateWg.Wait()
	close(parseValidateCh)

	belogs.Info("parseValidateAndAddCerts():end parseValidate, len(syncLogFileModels):", len(syncLogFileModels), "  fileType:", fileType, "  fileType:", fileType, "  time(s):", time.Now().Sub(start).Seconds())

	// add to db
	switch fileType {
	case "cer":
		db.AddCers(syncLogFileModels)
	case "crl":
		db.AddCrls(syncLogFileModels)
	case "mft":
		db.AddMfts(syncLogFileModels)
	case "roa":
		db.AddRoas(syncLogFileModels)
	}
	belogs.Info("parseValidateAndAddCerts():end add***(), len(syncLogFileModels):", len(syncLogFileModels), "  fileType:", fileType, "  time(s):", time.Now().Sub(start).Seconds())
}

func parseValidateCert(syncLogFileModel *parsevalidatemodel.SyncLogFileModel,
	wg *sync.WaitGroup, parseValidateCh chan int) (parseFailFile string, err error) {
	defer func() {
		wg.Done()
		<-parseValidateCh
	}()

	start := time.Now()
	belogs.Debug("parseValidateCert(): syncLogFileModel :", jsonutil.MarshalJson(syncLogFileModel))
	file := osutil.JoinPathFile(syncLogFileModel.FilePath, syncLogFileModel.FileName)
	belogs.Debug("parseValidateCert(): file :", file)
	_, certModel, stateModel, err := ParseValidateFile(file)
	if err != nil {
		belogs.Error("parseValidateCer(): parseValidateFile fail: ", file, err)
		return file, err
	}
	syncLogFileModel.CertModel = certModel
	syncLogFileModel.StateModel = stateModel
	belogs.Debug("parseValidateCert(): parseValidateFile file :", file,
		"   syncType:", syncLogFileModel.SyncType, "  time(s):", time.Now().Sub(start).Seconds())

	return "", nil

}

/*
MFT: Manifests for the Resource Public Key Infrastructure (RPKI)
https://datatracker.ietf.org/doc/rfc6486/?include_text=1

ROA: A Profile for Route Origin Authorizations (ROAs)
https://datatracker.ietf.org/doc/rfc6482/?include_text=1

CRL: Internet X.509 Public Key Infrastructure Certificate and Certificate Revocation List (CRL) Profile
https://datatracker.ietf.org/doc/rfc5280/?include_text=1

EE: Signed Object Template for the Resource Public Key Infrastructure (RPKI)
https://datatracker.ietf.org/doc/rfc6488/?include_text=1

CER: IP/AS:  X.509 Extensions for IP Addresses and AS Identifiers
https://datatracker.ietf.org/doc/rfc3779/?include_text=1

CER: A Profile for X.509 PKIX Resource Certificates
https://datatracker.ietf.org/doc/rfc6487/?include_text=1



A Profile for X.509 PKIX Resource Certificates
https://datatracker.ietf.org/doc/rfc6487/?include_text=1


A Profile for Route Origin Authorizations (ROAs)
https://datatracker.ietf.org/doc/rfc6482/?include_text=1

Signed Object Template for the Resource Public Key Infrastructure (RPKI)
https://datatracker.ietf.org/doc/rfc6488/?include_text=1

X.509 Extensions for IP Addresses and AS Identifiers
https://datatracker.ietf.org/doc/rfc3779/?include_text=1


Internet X.509 Public Key Infrastructure Certificate and Certificate Revocation List (CRL) Profile
https://datatracker.ietf.org/doc/rfc5280/?include_text=1
*/
// upload file to parse
func ParseValidateFile(certFile string) (certType string, certModel interface{}, stateModel model.StateModel, err error) {
	belogs.Debug("parseValidateFile(): parsevalidate start:", certFile)

	if strings.HasSuffix(certFile, ".cer") {
		cerModel, stateModel, err := ParseValidateCer(certFile)
		belogs.Debug("parseValidateFile():  after ParseCer():certFile, stateModel:", certFile, stateModel, "  err:", err)
		return "cer", cerModel, stateModel, err
	} else if strings.HasSuffix(certFile, ".crl") {
		crlModel, stateModel, err := ParseValidateCrl(certFile)
		belogs.Debug("parseValidateFile(): after ParseCrl(): certFile,stateModel:", certFile, stateModel, "  err:", err)
		return "crl", crlModel, stateModel, err
	} else if strings.HasSuffix(certFile, ".mft") {
		mftModel, stateModel, err := ParseValidateMft(certFile)
		belogs.Debug("parseValidateFile(): after ParseMft():certFile,stateModel:", certFile, stateModel, "  err:", err)
		return "mft", mftModel, stateModel, err
	} else if strings.HasSuffix(certFile, ".roa") {
		roaModel, stateModel, err := ParseValidateRoa(certFile)
		belogs.Debug("parseValidateFile():after ParseRoa(): certFile,stateModel:", certFile, stateModel, "  err:", err)
		return "roa", roaModel, stateModel, err
	} else {
		return "", nil, stateModel, errors.New("unknown file type")
	}
}

func ParseFile(certFile string) (certModel interface{}, err error) {
	belogs.Debug("parseValidateFile(): parsevalidate start:", certFile)
	if strings.HasSuffix(certFile, ".cer") {
		cerModel, _, err := ParseValidateCer(certFile)
		if err != nil {
			belogs.Error("parseValidateFile(): ParseValidateCer:", certFile, "  err:", err)
			return nil, err
		}
		cerModel.FilePath = ""
		belogs.Debug("parseValidateFile(): certFile,cerModel:", certFile, cerModel)
		return cerModel, nil

	} else if strings.HasSuffix(certFile, ".crl") {
		crlModel, _, err := ParseValidateCrl(certFile)
		if err != nil {
			belogs.Error("parseValidateFile(): ParseValidateCrl:", certFile, "  err:", err)
			return nil, err
		}
		crlModel.FilePath = ""
		belogs.Debug("parseValidateFile(): certFile, crlModel:", certFile, crlModel)
		return crlModel, nil

	} else if strings.HasSuffix(certFile, ".mft") {
		mftModel, _, err := ParseValidateMft(certFile)
		if err != nil {
			belogs.Error("parseValidateFile(): ParseValidateMft:", certFile, "  err:", err)
			return nil, err
		}
		mftModel.FilePath = ""
		belogs.Debug("parseValidateFile(): certFile, mftModel:", certFile, mftModel)
		return mftModel, nil

	} else if strings.HasSuffix(certFile, ".roa") {
		roaModel, _, err := ParseValidateRoa(certFile)
		if err != nil {
			belogs.Error("parseValidateFile(): ParseValidateRoa:", certFile, "  err:", err)
			return nil, err
		}
		roaModel.FilePath = ""
		belogs.Debug("parseValidateFile(): certFile, roaModel:", certFile, roaModel)
		return roaModel, nil

	} else {
		return nil, errors.New("unknown file type")
	}
}

// only parse cer to get ca repository/rpkiNotify, raw subjct public key info
func ParseFileSimple(certFile string) (parseCerSimple model.ParseCerSimple, err error) {
	if strings.HasSuffix(certFile, ".cer") {
		return ParseCerSimple(certFile)
	}
	return parseCerSimple, errors.New("unknown file type")
}
