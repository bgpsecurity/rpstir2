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

	//process "del" and "update" rsyncLogFile
	updateCerSyncLogFileModels, updateCrlSyncLogFileModels,
		updateMftSyncLogFileModels, updateRoaSyncLogFileModels, err := DelCertByDelAndUpdate(labRpkiSyncLogId)
	if err != nil {
		belogs.Error("ParseValidateStart():InsertRsyncLogRsyncStat fail:", err)
		return
	}
	belogs.Debug("ParseValidateStart(): after del Certs, will Insert Certs, labRpkiSyncLogId:", labRpkiSyncLogId)

	// process "add" and "update" rsyncLogFile
	err = InsertCertByAddAndUpdate(labRpkiSyncLogId, updateCerSyncLogFileModels, updateCrlSyncLogFileModels,
		updateMftSyncLogFileModels, updateRoaSyncLogFileModels)
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
func DelCertByDelAndUpdate(labRpkiSyncLogId uint64) (updateCerSyncLogFileModels, updateCrlSyncLogFileModels,
	updateMftSyncLogFileModels, updateRoaSyncLogFileModels []model.SyncLogFileModel, err error) {
	start := time.Now()

	belogs.Debug("DelCertByDelAndUpdate(): labRpkiSyncLogId:", labRpkiSyncLogId)

	var wg sync.WaitGroup

	// get "del" and "update" cer synclog files to del
	delCerSyncLogFileModels, updateCerSyncLogFileModels, err := getDelAndUpdateSyncLogFileModels(labRpkiSyncLogId, "cer")
	if err != nil {
		belogs.Error("DelCertByDelAndUpdate(): getDelAndUpdateSyncLogFileModels cer fail: ", labRpkiSyncLogId, err)
		return nil, nil, nil, nil, err
	}
	belogs.Debug("DelCertByDelAndUpdate(): len(delCerSyncLogFileModels):", len(delCerSyncLogFileModels),
		"       len(updateCerSyncLogFileModels):", len(updateCerSyncLogFileModels))
	if len(delCerSyncLogFileModels) > 0 || len(updateCerSyncLogFileModels) > 0 {
		wg.Add(1)
		go db.DelCers(append(delCerSyncLogFileModels, updateCerSyncLogFileModels...), &wg)
	}

	// get "del" and "update" crl synclog files to del
	delCrlSyncLogFileModels, updateCrlSyncLogFileModels, err := getDelAndUpdateSyncLogFileModels(labRpkiSyncLogId, "crl")
	if err != nil {
		belogs.Error("DelCertByDelAndUpdate(): getDelAndUpdateSyncLogFileModels crl fail: ", labRpkiSyncLogId, err)
		return nil, nil, nil, nil, err
	}
	belogs.Debug("DelCertByDelAndUpdate(): len(delCrlSyncLogFileModels):", len(delCrlSyncLogFileModels),
		"       len(updateCrlSyncLogFileModels):", len(updateCrlSyncLogFileModels))
	if len(delCrlSyncLogFileModels) > 0 || len(updateCrlSyncLogFileModels) > 0 {
		wg.Add(1)
		go db.DelCrls(append(delCrlSyncLogFileModels, updateCrlSyncLogFileModels...), &wg)
	}

	// get del synclog files --> mft
	delMftSyncLogFileModels, updateMftSyncLogFileModels, err := getDelAndUpdateSyncLogFileModels(labRpkiSyncLogId, "mft")
	if err != nil {
		belogs.Error("DelCertByDelAndUpdate(): getDelAndUpdateSyncLogFileModels mft fail: ", labRpkiSyncLogId, err)
		return nil, nil, nil, nil, err
	}
	belogs.Debug("DelCertByDelAndUpdate(): len(delMftSyncLogFileModels):", len(delMftSyncLogFileModels),
		"       len(updateMftSyncLogFileModels):", len(updateMftSyncLogFileModels))
	if len(delMftSyncLogFileModels) > 0 || len(updateMftSyncLogFileModels) > 0 {
		wg.Add(1)
		go db.DelMfts(append(delMftSyncLogFileModels, updateMftSyncLogFileModels...), &wg)
	}

	// get del synclog files --> roa
	delRoaSyncLogFileModels, updateRoaSyncLogFileModels, err := getDelAndUpdateSyncLogFileModels(labRpkiSyncLogId, "roa")
	if err != nil {
		belogs.Error("DelCertByDelAndUpdate(): getDelAndUpdateSyncLogFileModels roa fail: ", labRpkiSyncLogId, err)
		return nil, nil, nil, nil, err
	}
	belogs.Debug("DelCertByDelAndUpdate(): len(delRoaSyncLogFileModels):", len(delRoaSyncLogFileModels),
		"       len(updateRoaSyncLogFileModels):", len(updateRoaSyncLogFileModels))
	if len(delRoaSyncLogFileModels) > 0 || len(updateRoaSyncLogFileModels) > 0 {
		wg.Add(1)
		go db.DelRoas(append(delRoaSyncLogFileModels, updateRoaSyncLogFileModels...), &wg)
	}

	// update "del" rsync_log_file
	if len(delCerSyncLogFileModels) > 0 || len(delCerSyncLogFileModels) > 0 ||
		len(delCerSyncLogFileModels) > 0 || len(delCerSyncLogFileModels) > 0 {

		wg.Add(1)
		syncLogFileModels := make([]model.SyncLogFileModel, 0,
			len(delCerSyncLogFileModels)+len(delCrlSyncLogFileModels)+
				len(delMftSyncLogFileModels)+len(delRoaSyncLogFileModels))
		syncLogFileModels = append(syncLogFileModels, delCerSyncLogFileModels...)
		syncLogFileModels = append(syncLogFileModels, delCrlSyncLogFileModels...)
		syncLogFileModels = append(syncLogFileModels, delMftSyncLogFileModels...)
		syncLogFileModels = append(syncLogFileModels, delRoaSyncLogFileModels...)
		go db.UpdateSyncLogFilesJsonAllAndState(syncLogFileModels, &wg)

	}
	wg.Wait()
	belogs.Info("DelCertByDelAndUpdate(): end,  time(s):", time.Now().Sub(start).Seconds())
	return updateCerSyncLogFileModels, updateCrlSyncLogFileModels,
		updateMftSyncLogFileModels, updateRoaSyncLogFileModels, nil

}

// get add;
// use update, because "update" should add
func InsertCertByAddAndUpdate(labRpkiSyncLogId uint64, updateCerSyncLogFileModels, updateCrlSyncLogFileModels,
	updateMftSyncLogFileModels, updateRoaSyncLogFileModels []model.SyncLogFileModel) (err error) {

	start := time.Now()
	belogs.Debug("InsertCertByInsertAndUpdate(): labRpkiSyncLogId:", labRpkiSyncLogId)

	var wg sync.WaitGroup

	// get "add" and "update" cer synclog files to add
	addCerSyncLogFileModels, err := getAddSyncLogFileModels(labRpkiSyncLogId, "cer")
	if err != nil {
		belogs.Error("InsertCertByInsertAndUpdate(): getAddSyncLogFileModels cer fail: ", labRpkiSyncLogId, err)
		return err
	}
	belogs.Debug("InsertCertByInsertAndUpdate(): len(addCerSyncLogFileModels):", len(addCerSyncLogFileModels),
		"       len(updateCerSyncLogFileModels):", len(updateCerSyncLogFileModels))
	if len(addCerSyncLogFileModels) > 0 || len(updateCerSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCers(append(addCerSyncLogFileModels, updateCerSyncLogFileModels...), &wg)
	}

	// get "add" and "update" crl synclog files to add
	addCrlSyncLogFileModels, err := getAddSyncLogFileModels(labRpkiSyncLogId, "crl")
	if err != nil {
		belogs.Error("InsertCertByInsertAndUpdate(): getAddSyncLogFileModels crl fail: ", labRpkiSyncLogId, err)
		return err
	}
	belogs.Debug("InsertCertByInsertAndUpdate(): len(addCrlSyncLogFileModels):", len(addCrlSyncLogFileModels),
		"       len(updateCrlSyncLogFileModels):", len(updateCrlSyncLogFileModels))
	if len(addCrlSyncLogFileModels) > 0 || len(updateCrlSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCrls(append(addCrlSyncLogFileModels, updateCrlSyncLogFileModels...), &wg)
	}

	// get "add" and "update" mft synclog files to add
	addMftSyncLogFileModels, err := getAddSyncLogFileModels(labRpkiSyncLogId, "mft")
	if err != nil {
		belogs.Error("InsertCertByInsertAndUpdate(): getAddSyncLogFileModels mft fail: ", labRpkiSyncLogId, err)
		return err
	}
	belogs.Debug("InsertCertByInsertAndUpdate(): len(addMftSyncLogFileModels):", len(addMftSyncLogFileModels),
		"       len(updateMftSyncLogFileModels):", len(updateMftSyncLogFileModels))
	if len(addMftSyncLogFileModels) > 0 || len(updateMftSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddMfts(append(addMftSyncLogFileModels, updateMftSyncLogFileModels...), &wg)
	}

	// get "add" and "update" cer synclog files to add
	addRoaSyncLogFileModels, err := getAddSyncLogFileModels(labRpkiSyncLogId, "roa")
	if err != nil {
		belogs.Error("InsertCertByInsertAndUpdate(): getAddSyncLogFileModels roa fail: ", labRpkiSyncLogId, err)
		return err
	}
	belogs.Debug("InsertCertByInsertAndUpdate(): len(addRoaSyncLogFileModels):", len(addCerSyncLogFileModels),
		"       len(updateRoaSyncLogFileModels):", len(updateRoaSyncLogFileModels))
	if len(addRoaSyncLogFileModels) > 0 || len(updateRoaSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddRoas(append(addRoaSyncLogFileModels, updateRoaSyncLogFileModels...), &wg)
	}

	// update "del" rsync_log_file
	if len(addCerSyncLogFileModels) > 0 || len(updateCerSyncLogFileModels) > 0 ||
		len(addCrlSyncLogFileModels) > 0 || len(updateCrlSyncLogFileModels) > 0 ||
		len(addMftSyncLogFileModels) > 0 || len(updateMftSyncLogFileModels) > 0 ||
		len(addRoaSyncLogFileModels) > 0 || len(updateRoaSyncLogFileModels) > 0 {

		wg.Add(1)
		syncLogFileModels := make([]model.SyncLogFileModel, 0,
			len(addCerSyncLogFileModels)+len(updateCerSyncLogFileModels)+
				len(addCrlSyncLogFileModels)+len(updateCrlSyncLogFileModels)+
				len(addMftSyncLogFileModels)+len(updateMftSyncLogFileModels)+
				len(addRoaSyncLogFileModels)+len(updateRoaSyncLogFileModels))
		syncLogFileModels = append(syncLogFileModels, addCerSyncLogFileModels...)
		syncLogFileModels = append(syncLogFileModels, updateCerSyncLogFileModels...)
		syncLogFileModels = append(syncLogFileModels, addCrlSyncLogFileModels...)
		syncLogFileModels = append(syncLogFileModels, updateCrlSyncLogFileModels...)
		syncLogFileModels = append(syncLogFileModels, addMftSyncLogFileModels...)
		syncLogFileModels = append(syncLogFileModels, updateMftSyncLogFileModels...)
		syncLogFileModels = append(syncLogFileModels, addRoaSyncLogFileModels...)
		syncLogFileModels = append(syncLogFileModels, updateRoaSyncLogFileModels...)
		go db.UpdateSyncLogFilesJsonAllAndState(syncLogFileModels, &wg)

	}

	wg.Wait()
	belogs.Info("InsertCertByInsertAndUpdate(): end,  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func parseValidateAndAddCers(syncLogFileModels []model.SyncLogFileModel, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	parseValidateCerts(syncLogFileModels)
	db.AddCers(syncLogFileModels)
	belogs.Info("parseValidateAndAddCers(): len(cers):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
}
func parseValidateAndAddCrls(syncLogFileModels []model.SyncLogFileModel, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	parseValidateCerts(syncLogFileModels)
	db.AddCrls(syncLogFileModels)
	belogs.Info("parseValidateAndAddCrls(): len(crls), ", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
}
func parseValidateAndAddMfts(syncLogFileModels []model.SyncLogFileModel, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	parseValidateCerts(syncLogFileModels)
	db.AddMfts(syncLogFileModels)
	belogs.Info("parseValidateAndAddMfts(): len(mfts), ", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
}
func parseValidateAndAddRoas(syncLogFileModels []model.SyncLogFileModel, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	parseValidateCerts(syncLogFileModels)
	db.AddRoas(syncLogFileModels)
	belogs.Info("parseValidateAndAddRoas(): len(roas), ", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
}
func parseValidateCerts(syncLogFileModels []model.SyncLogFileModel) {

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

func parseValidateCert(syncLogFileModel *model.SyncLogFileModel) (parseFailFile string, err error) {

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
func getDelAndUpdateSyncLogFileModels(labRpkiSyncLogId uint64, fileType string) (delSyncLogFileModels, updateSyncLogFileModels []model.SyncLogFileModel,
	err error) {
	// get del synclog files
	delSyncLogFileModels, err = db.GetSyncLogFileModels(labRpkiSyncLogId, "del", fileType)
	if err != nil {
		belogs.Error("GetDelAndUpdateSyncLogFileModels():del syncLogFileDelModels fail: ", labRpkiSyncLogId, fileType, err)
		return nil, nil, err
	}
	updateSyncLogFileModels, err = db.GetSyncLogFileModels(labRpkiSyncLogId, "update", fileType)
	if err != nil {
		belogs.Error("GetDelAndUpdateSyncLogFileModels():update syncLogFileDelModels fail: ", labRpkiSyncLogId, fileType, err)
		return nil, nil, err
	}
	return delSyncLogFileModels, updateSyncLogFileModels, err

}

func getAddSyncLogFileModels(labRpkiSyncLogId uint64, fileType string) (addSyncLogFileModels []model.SyncLogFileModel,
	err error) {
	// get del synclog files
	addSyncLogFileModels, err = db.GetSyncLogFileModels(labRpkiSyncLogId, "add", fileType)
	if err != nil {
		belogs.Error("getAddSyncLogFileModels():add syncLogFileDelModels fail: ", labRpkiSyncLogId, fileType, err)
		return nil, err
	}

	return addSyncLogFileModels, err

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
