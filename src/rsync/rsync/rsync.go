package rsync

import (
	"os"
	"sync/atomic"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	_ "github.com/cpusoft/goutil/httpclient"
	httpclient "github.com/cpusoft/goutil/httpclient"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	_ "github.com/cpusoft/goutil/logs"
	osutil "github.com/cpusoft/goutil/osutil"
	rsyncutil "github.com/cpusoft/goutil/rsyncutil"
	talutil "github.com/cpusoft/goutil/talutil"

	"model"
	"rsync/db"
	rsyncmodel "rsync/model"
)

var rpQueue *rsyncmodel.RsyncParseQueue

// start to rsync
func Start() {
	belogs.Info("Start():rsync")
	// get all tal files
	files, err := talutil.GetAllTalFile(conf.VariableString("rsync::talpath"))
	if err != nil {
		belogs.Error("Start(): GetAllTalFile failed:", err)
		return
	}
	belogs.Notice("Start(): GetAllTalFile:", files)

	// parse tal local
	talInfos, err := talutil.ParseTalInfos(files)
	if err != nil {
		belogs.Error("Start(): GetAllTalFile failed:", err)
		return
	}
	belogs.Debug("Start(): ParseTalInfoList:", talInfos)

	// save starttime to lab_rpki_sync_log
	labRpkiSyncLogId, err := db.InsertRsyncLogRsyncStateStart("rsyncing", "rsync")
	if err != nil {
		belogs.Error("Start():InsertRsyncLogRsyncStat fail:", err)
		return
	}
	//start rpQueue and rsyncForSelect
	rpQueue = rsyncmodel.NewQueue()
	go startRsyncServer()

	rpQueue.LabRpkiSyncLogId = labRpkiSyncLogId
	belogs.Debug("Start(): rpQueue:", jsonutil.MarshalJson(rpQueue))

	// start to rsync by sync url in tal, to get root cer
	// first: remove all root cer, so can will rsync download and will trigger parse all cer files.
	// otherwise, will have to load all root file manually
	os.RemoveAll(conf.VariableString("rsync::destpath") + "/root/")
	os.MkdirAll(conf.VariableString("rsync::destpath")+"/root/", os.ModePerm)
	atomic.AddInt64(&rpQueue.RsyncingParsingCount, int64(len(talInfos)))
	belogs.Debug("Start():after RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
	for _, talInfo := range talInfos {
		url := talInfo.SyncUrl
		go rpQueue.AddRsyncUrl(url, conf.VariableString("rsync::destpath")+"/root/")
	}

}

// start server ,wait input channel
func startRsyncServer() {
	belogs.Info("startRsyncServer():start")

	for {
		select {
		case rsyncModelChan := <-rpQueue.RsyncModelChan:
			belogs.Debug("startRsyncServer(): rsyncModelChan:", rsyncModelChan,
				"  len(rsyncrpQueue.RsyncModelChan):", len(rpQueue.RsyncModelChan),
				"  receive rsyncModelChan rpQueue.RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
			go RsyncByUrl(rsyncModelChan)
		case parseModelChan := <-rpQueue.ParseModelChan:
			belogs.Debug("startRsyncServer(): parseModelChan:", parseModelChan,
				"  receive parseModelChan rpQueue.RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
			go ParseCerFiles(parseModelChan)
		case rsyncParseEndChan := <-rpQueue.RsyncParseEndChan:
			belogs.Debug("startRsyncServer():rsyncParseEndChan:", rsyncParseEndChan, "  rpQueue.RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
			//if !WaitForRsyncParseEnd() {
			//	continue
			//}

			// try again the fail urls
			belogs.Debug("startRsyncServer():try fail urls again: len(rpQueue.RsyncMisc.FailRsyncUrls):", len(rpQueue.RsyncMisc.FailRsyncUrls))
			if TryAgainFailRsyncUrls() {
				belogs.Debug("startRsyncServer(): TryAgainFailRsyncUrls continue")
				continue
			}

			// save endtime to lab_rpki_sync_log
			labRpkiSyncLogId := rpQueue.LabRpkiSyncLogId
			err := db.UpdateRsyncLogRsyncStateEnd(labRpkiSyncLogId, "rsynced", &rpQueue.RsyncMisc)
			if err != nil {
				belogs.Error("startRsyncServer():UpdateRsyncLogRsyncState fail:", err)
				return
			}

			belogs.Info("startRsyncServer():end this rsync sucess: len(rpQueue.RsyncMisc.FailRsyncUrls):",
				len(rpQueue.RsyncMisc.FailRsyncUrls))
			// close rpQueue
			rpQueue.Close()

			// call FoundDiffFiles
			belogs.Info("startRsyncServer():call FoundDiffFiles,  labRpkiSyncLogId:", labRpkiSyncLogId)
			go FoundDiffFiles(labRpkiSyncLogId)
			// return out of the for
			belogs.Info("startRsyncServer():end ")
			return
		}
	}
}

func RsyncByUrl(rsyncModelChan rsyncmodel.RsyncModelChan) {
	defer func() {
		belogs.Debug("RsyncByUrl():defer rpQueue.RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		if atomic.LoadInt64(&rpQueue.RsyncingParsingCount) == 0 {
			belogs.Debug("RsyncByUrl(): call rsyncmodel.RsyncParseEndChan{}, RsyncingParsingCount is 0")
			rpQueue.RsyncParseEndChan <- rsyncmodel.RsyncParseEndChan{}
		}
	}()

	// start rsync and check err
	// if have error, should set RsyncingParsingCount -1
	start := time.Now()

	// CurRsyncingCount should +1 and then -1
	atomic.AddInt64(&rpQueue.CurRsyncingCount, 1)
	belogs.Debug("RsyncByUrl(): before rsync, rsyncModelChan:", rsyncModelChan, "    CurRsyncingCount:", atomic.LoadInt64(&rpQueue.CurRsyncingCount))
	rsyncDestPath, _, err := rsyncutil.RsyncQuiet(rsyncModelChan.Url, rsyncModelChan.Dest)
	atomic.AddInt64(&rpQueue.CurRsyncingCount, -1)
	belogs.Debug("RsyncByUrl(): rsync rsyncModelChan:", rsyncModelChan, "     CurRsyncingCount:", atomic.LoadInt64(&rpQueue.CurRsyncingCount),
		"     rsyncDestPath:", rsyncDestPath)
	if err != nil {
		rpQueue.RsyncMisc.FailRsyncUrls[rsyncModelChan.Url] = err.Error()
		belogs.Error("RsyncByUrl():RsyncQuiet fail, rsyncModelChan.Url:", rsyncModelChan.Url, "   err:", err)
		belogs.Debug("RsyncByUrl():RsyncQuiet fail, before RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		atomic.AddInt64(&rpQueue.RsyncingParsingCount, -1)
		belogs.Debug("RsyncByUrl():RsyncQuiet fail, after RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		return
	}

	atomic.AddUint64(&rpQueue.RsyncMisc.OkRsyncUrlLen, 1)
	belogs.Debug("RsyncByUrl(): rsync.Rsync url:", rsyncModelChan.Url, "   rsyncDestPath:", rsyncDestPath)

	parseModelChan := rsyncmodel.ParseModelChan{FilePathName: rsyncDestPath}
	belogs.Debug("RsyncByUrl():before parseModelChan:", parseModelChan, "   len(rpQueue.ParseModelChan):", len(rpQueue.ParseModelChan))
	belogs.Info("RsyncByUrl(): rsync rsyncModelChan:", rsyncModelChan, "     CurRsyncingCount:", atomic.LoadInt64(&rpQueue.CurRsyncingCount),
		"     rsyncDestPath:", rsyncDestPath, "  time(s):", time.Now().Sub(start).Seconds())

	rpQueue.ParseModelChan <- parseModelChan

}
func ParseCerFiles(parseModelChan rsyncmodel.ParseModelChan) {
	defer func() {
		belogs.Debug("ParseCerFiles():defer rpQueue.RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		if atomic.LoadInt64(&rpQueue.RsyncingParsingCount) == 0 {
			belogs.Debug("ParseCerFiles(): call rsyncmodel.RyncParseEndChan{}, RsyncingParsingCount is 0")
			rpQueue.RsyncParseEndChan <- rsyncmodel.RsyncParseEndChan{}
		}
	}()
	belogs.Debug("ParseCerFiles(): parseModelChan:", parseModelChan)

	// if have erorr, should set RsyncingParsingCount -1
	// get all cer files, include subcer
	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	cerFiles, err := osutil.GetAllFilesBySuffixs(parseModelChan.FilePathName, m)
	if err != nil {
		belogs.Error("ParseCerFiles():GetAllFilesBySuffixs fail, parseModelChan.FilePathName:", parseModelChan.FilePathName, "   err:", err)
		belogs.Debug("ParseCerFiles():GetAllFilesBySuffixs, before RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		atomic.AddInt64(&rpQueue.RsyncingParsingCount, -1)
		belogs.Debug("ParseCerFiles():GetAllFilesBySuffixs, after RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		return
	}
	belogs.Debug("ParseCerFiles(): len(cerFiles):", len(cerFiles))

	// if there are no cer files, return
	if len(cerFiles) == 0 {
		belogs.Debug("ParseCerFiles():len(cerFiles)==0, before RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		atomic.AddInt64(&rpQueue.RsyncingParsingCount, -1)
		belogs.Debug("ParseCerFiles():len(cerFiles)==0, after RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		return
	}

	// foreach every cerfiles to ParseCerAndGetSubCaRepositoryUrl
	subCaRepositoryUrls := make([]string, 0, len(cerFiles))
	for _, cerFile := range cerFiles {
		// just trigger sync ,no need save to db
		subCaRepositoryUrl := ParseCerAndGetSubCaRepositoryUrl(cerFile)
		if len(subCaRepositoryUrl) > 0 {
			subCaRepositoryUrls = append(subCaRepositoryUrls, subCaRepositoryUrl)
		}
	}
	belogs.Debug("ParseCerFiles(): len(subCaRepositoryUrls):", len(subCaRepositoryUrls))

	// check rsync concurrent count, wait some time,
	// the father rsyncingparsingcount -1 ,and the children rsyncingparsingcount + len()
	belogs.Debug("ParseCerFiles():will add subCaRepositoryUrls, before RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
	atomic.AddInt64(&rpQueue.RsyncingParsingCount, int64(len(subCaRepositoryUrls)-1))
	belogs.Debug("ParseCerFiles():will add subCaRepositoryUrls, after RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
	AddSubCaRepositoryUrlsToRpQueue(subCaRepositoryUrls)
}

// call /parsevalidate/parse to parse cert, and save result
func ParseCerAndGetSubCaRepositoryUrl(cerFile string) (subCaRepositoryUrl string) {

	// call parse, not need to save body to db
	start := time.Now()
	belogs.Debug("ParseCerAndGetSubCaRepositoryUrl():/parsevalidate/file cerFile:", cerFile)
	resp, body, err := httpclient.PostFile("http", conf.String("rpstir2::parsevalidateserver"), conf.Int("rpstir2::httpport"), "/parsevalidate/filerepo",
		cerFile, "")
	belogs.Debug("ParseCerAndGetSubCaRepositoryUrl():after /parsevalidate/filerepo cerFile:", cerFile, len(body))

	if err != nil {
		rpQueue.RsyncMisc.FailParseValidateCerts[cerFile] = err.Error()
		belogs.Error("ParseCerAndGetSubCaRepositoryUrl(): filerepo file connecteds failed:", cerFile, "   err:", err)
		return ""
	}
	defer resp.Body.Close()

	// get parse result
	parseCertRepoResponse := model.ParseCertRepoResponse{}
	jsonutil.UnmarshalJson(string(body), &parseCertRepoResponse)
	belogs.Debug("ParseCerAndGetSubCaRepositoryUrl(): get from parsecert, parseCertRepoResponse.Result:", parseCertRepoResponse.Result)
	if parseCertRepoResponse.HttpResponse.Result != "ok" {
		belogs.Error("ParseCerAndGetSubCaRepositoryUrl(): parsecert file failed:", cerFile, "   err:", parseCertRepoResponse.HttpResponse.Msg)
		rpQueue.RsyncMisc.FailParseValidateCerts[cerFile] = parseCertRepoResponse.HttpResponse.Msg
		return ""
	}

	// get the sub repo url in cer, and send it to rpqueue
	belogs.Info("ParseCerAndGetSubCaRepositoryUrl(): cerFile:", cerFile, "    CaRepository:", parseCertRepoResponse.CaRepository,
		"  time(s):", time.Now().Sub(start).Seconds())
	return parseCertRepoResponse.CaRepository

}
