package parsevalidate

import (
	"errors"
	"strings"
	"sync"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	httpclient "github.com/cpusoft/goutil/httpclient"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"

	"model"
	db "parsevalidate/db"
	parsevalidatemodel "parsevalidate/model"
)

func ParseValidateStart() {

	start := time.Now()
	belogs.Info("ParseValidateStart(): start")
	// save starttime to lab_rpki_sync_log
	labRpkiSyncLogId, err := db.UpdateRsyncLogParseValidateStart("parsevalidating")
	if err != nil {
		belogs.Error("ParseValidateStart():InsertRsyncLogRsyncStat fail:", err)
		return
	}
	belogs.Debug("ParseValidateStart(): labRpkiSyncLogId:", labRpkiSyncLogId)

	// get all need rsyncLogFile
	syncLogFileModels, err := db.GetSyncLogFileModelsBySyncLogId(labRpkiSyncLogId)
	if err != nil {
		belogs.Error("ParseValidateStart():GetSyncLogFileModelsBySyncLogId fail:", err)
		return
	}
	belogs.Debug("ParseValidateStart(): GetSyncLogFileModelsBySyncLogId, syncLogFileModels.SyncLogId:", syncLogFileModels.SyncLogId)

	//process "del" and "update" rsyncLogFile
	err = DelCertByDelAndUpdate(syncLogFileModels)
	if err != nil {
		belogs.Error("ParseValidateStart():DelCertByDelAndUpdate fail:", err)
		return
	}
	belogs.Debug("ParseValidateStart(): after DelCertByDelAndUpdate, syncLogFileModels.SyncLogId:", syncLogFileModels.SyncLogId)

	// process "add" and "update" rsyncLogFile
	err = InsertCertByAddAndUpdate(syncLogFileModels)
	if err != nil {
		belogs.Error("ParseValidateStart():InsertCertByInsertAndUpdate fail:", err)
		return
	}
	// save to db
	err = db.UpdateRsyncLogParseValidateStateEnd(labRpkiSyncLogId, "parsevalidated", make([]string, 0))
	if err != nil {
		belogs.Debug("ParseValidateStart(): UpdateRsyncLogAndCert fail: ", err)
		return
	}

	belogs.Info("ParseValidateStart(): end, will call chainvalidate,  time(s):", time.Now().Sub(start).Seconds())
	// will call ChainValidate
	go func() {
		httpclient.Post("http", conf.String("rpstir2::chainvalidateserver"), conf.Int("rpstir2::httpport"),
			"/chainvalidate/start", "")
	}()
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
		go db.DelCrls(syncLogFileModels.UpdateCrlSyncLogFileModels, syncLogFileModels.UpdateCrlSyncLogFileModels, &wg)
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

// get add;
// use update, because "update" should add
func InsertCertByAddAndUpdate(syncLogFileModels *parsevalidatemodel.SyncLogFileModels) (err error) {

	start := time.Now()
	belogs.Debug("InsertCertByInsertAndUpdate(): syncLogFileModels.SyncLogId:", syncLogFileModels.SyncLogId)

	var wg sync.WaitGroup

	// add/update crl
	belogs.Debug("InsertCertByInsertAndUpdate(): len(syncLogFileModels.AddCerSyncLogFileModels):", len(syncLogFileModels.AddCerSyncLogFileModels),
		"       len(syncLogFileModels.UpdateCerSyncLogFileModels):", len(syncLogFileModels.UpdateCerSyncLogFileModels))
	if len(syncLogFileModels.AddCerSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateCerSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCers(append(syncLogFileModels.AddCerSyncLogFileModels, syncLogFileModels.UpdateCerSyncLogFileModels...), &wg)
	}

	// add/update crl
	belogs.Debug("InsertCertByInsertAndUpdate(): len(syncLogFileModels.AddCrlSyncLogFileModels):", len(syncLogFileModels.AddCrlSyncLogFileModels),
		"       len(syncLogFileModels.UpdateCrlSyncLogFileModels):", len(syncLogFileModels.UpdateCrlSyncLogFileModels))
	if len(syncLogFileModels.AddCrlSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateCrlSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCrls(append(syncLogFileModels.AddCrlSyncLogFileModels, syncLogFileModels.UpdateCrlSyncLogFileModels...), &wg)
	}

	// add/update mft
	belogs.Debug("InsertCertByInsertAndUpdate(): len(syncLogFileModels.AddMftSyncLogFileModels):", len(syncLogFileModels.AddMftSyncLogFileModels),
		"       len(syncLogFileModels.UpdateMftSyncLogFileModels):", len(syncLogFileModels.UpdateMftSyncLogFileModels))
	if len(syncLogFileModels.AddMftSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateMftSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddMfts(append(syncLogFileModels.AddMftSyncLogFileModels, syncLogFileModels.UpdateMftSyncLogFileModels...), &wg)
	}

	// add/update roa
	belogs.Debug("InsertCertByInsertAndUpdate(): len(syncLogFileModels.AddRoaSyncLogFileModels):", len(syncLogFileModels.AddCerSyncLogFileModels),
		"       len(syncLogFileModels.UpdateRoaSyncLogFileModels):", len(syncLogFileModels.UpdateRoaSyncLogFileModels))
	if len(syncLogFileModels.AddRoaSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateRoaSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddRoas(append(syncLogFileModels.AddRoaSyncLogFileModels, syncLogFileModels.UpdateRoaSyncLogFileModels...), &wg)
	}

	wg.Wait()
	belogs.Info("InsertCertByInsertAndUpdate(): end,  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func parseValidateAndAddCers(syncLogFileModels []parsevalidatemodel.SyncLogFileModel, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	parseValidateCerts(syncLogFileModels)
	db.AddCers(syncLogFileModels)

	belogs.Info("parseValidateAndAddCers(): len(cers):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
}
func parseValidateAndAddCrls(syncLogFileModels []parsevalidatemodel.SyncLogFileModel, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	parseValidateCerts(syncLogFileModels)
	db.AddCrls(syncLogFileModels)
	belogs.Info("parseValidateAndAddCrls(): len(crls):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
}
func parseValidateAndAddMfts(syncLogFileModels []parsevalidatemodel.SyncLogFileModel, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	parseValidateCerts(syncLogFileModels)
	db.AddMfts(syncLogFileModels)
	belogs.Info("parseValidateAndAddMfts(): len(mfts):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
}
func parseValidateAndAddRoas(syncLogFileModels []parsevalidatemodel.SyncLogFileModel, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	parseValidateCerts(syncLogFileModels)
	db.AddRoas(syncLogFileModels)
	belogs.Info("parseValidateAndAddRoas(): len(roas):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
}
func parseValidateCerts(syncLogFileModels []parsevalidatemodel.SyncLogFileModel) {

	belogs.Debug("parseValidateCerts(): len(syncLogFileModels):", len(syncLogFileModels))
	for i, _ := range syncLogFileModels {
		parseFailFile, err := parseValidateCert(&syncLogFileModels[i])
		if err != nil {
			belogs.Error("parseValidateCerts(): parseValidateCert fail: ",
				syncLogFileModels[i].FilePath, syncLogFileModels[i].FileName, parseFailFile, err)
			// not return err
			continue
		}
	}
}

func parseValidateCert(syncLogFileModel *parsevalidatemodel.SyncLogFileModel) (parseFailFile string, err error) {

	start := time.Now()
	belogs.Debug("parseValidateCert(): syncLogFileModel :", jsonutil.MarshalJson(syncLogFileModel))
	file := osutil.JoinPathFile(syncLogFileModel.FilePath, syncLogFileModel.FileName)
	belogs.Debug("parseValidateCert(): file :", file)
	_, certModel, stateModel, jsonAll, err := ParseValidateFile(file)
	if err != nil {
		belogs.Error("parseValidateCer(): parseValidateFile fail: ", file, err)
		return file, err
	}
	syncLogFileModel.CertModel = certModel
	syncLogFileModel.StateModel = stateModel
	syncLogFileModel.JsonAll = jsonAll
	belogs.Info("parseValidateCert(): parseValidateFile file :", file,
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
func ParseValidateFile(certFile string) (certType string, certModel interface{}, stateModel model.StateModel, jsonAll string, err error) {
	belogs.Debug("parseValidateFile(): parsevalidate start:", certFile)

	if strings.HasSuffix(certFile, ".cer") {
		cerModel, stateModel, err := ParseValidateCer(certFile)
		belogs.Debug("parseValidateFile():  after ParseCer():certFile, stateModel:", certFile, stateModel, "  err:", err)
		return "cer", cerModel, stateModel, jsonutil.MarshalJson(cerModel), err
	} else if strings.HasSuffix(certFile, ".crl") {
		crlModel, stateModel, err := ParseValidateCrl(certFile)
		belogs.Debug("parseValidateFile(): after ParseCrl(): certFile,stateModel:", certFile, stateModel, "  err:", err)
		return "crl", crlModel, stateModel, jsonutil.MarshalJson(crlModel), err
	} else if strings.HasSuffix(certFile, ".mft") {
		mftModel, stateModel, err := ParseValidateMft(certFile)
		belogs.Debug("parseValidateFile(): after ParseMft():certFile,stateModel:", certFile, stateModel, "  err:", err)
		return "mft", mftModel, stateModel, jsonutil.MarshalJson(mftModel), err
	} else if strings.HasSuffix(certFile, ".roa") {
		roaModel, stateModel, err := ParseValidateRoa(certFile)
		belogs.Debug("parseValidateFile():after ParseRoa(): certFile,stateModel:", certFile, stateModel, "  err:", err)
		return "roa", roaModel, stateModel, jsonutil.MarshalJson(roaModel), err
	} else {
		return "", nil, stateModel, "", errors.New("unknown file type")
	}
}

// only parse cer to get ca repository
func ParseValidateFileRepo(certFile string) (caRepository string, err error) {
	if strings.HasSuffix(certFile, ".cer") {
		return ParseValidateCerRepo(certFile)
	} else {
		return "", errors.New("unknown file type")
	}
}
