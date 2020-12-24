package rsync

import (
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	httpclient "github.com/cpusoft/goutil/httpclient"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"
	randutil "github.com/cpusoft/goutil/randutil"
	rsyncutil "github.com/cpusoft/goutil/rsyncutil"

	"model"
	rsyncmodel "rsync/model"
)

func rsyncByUrl(rsyncModelChan rsyncmodel.RsyncModelChan) {
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
		rpQueue.RsyncResult.FailUrls[rsyncModelChan.Url] = err.Error()
		belogs.Error("RsyncByUrl():RsyncQuiet fail, rsyncModelChan.Url:", rsyncModelChan.Url, "   err:", err, "  time(s):", time.Now().Sub(start).Seconds())
		belogs.Debug("RsyncByUrl():RsyncQuiet fail, before RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		atomic.AddInt64(&rpQueue.RsyncingParsingCount, -1)
		belogs.Debug("RsyncByUrl():RsyncQuiet fail, after RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		return
	}
	belogs.Debug("RsyncByUrl(): rsync.Rsync url:", rsyncModelChan.Url, "   rsyncDestPath:", rsyncDestPath)

	parseModelChan := rsyncmodel.ParseModelChan{FilePathName: rsyncDestPath}
	belogs.Debug("RsyncByUrl():before parseModelChan:", parseModelChan, "   len(rpQueue.ParseModelChan):", len(rpQueue.ParseModelChan))
	belogs.Info("RsyncByUrl(): rsync rsyncModelChan:", rsyncModelChan, "     CurRsyncingCount:", atomic.LoadInt64(&rpQueue.CurRsyncingCount),
		"     rsyncDestPath:", rsyncDestPath, "  time(s):", time.Now().Sub(start).Seconds())

	rpQueue.ParseModelChan <- parseModelChan

}

func parseCerFiles(parseModelChan rsyncmodel.ParseModelChan) {
	defer func() {
		belogs.Debug("parseCerFiles():defer rpQueue.RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		if atomic.LoadInt64(&rpQueue.RsyncingParsingCount) == 0 {
			belogs.Debug("parseCerFiles(): call rsyncmodel.RyncParseEndChan{}, RsyncingParsingCount is 0")
			rpQueue.RsyncParseEndChan <- rsyncmodel.RsyncParseEndChan{}
		}
	}()
	belogs.Debug("parseCerFiles(): parseModelChan:", parseModelChan)

	// if have erorr, should set RsyncingParsingCount -1
	// get all cer files, include subcer
	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	cerFiles, err := osutil.GetAllFilesBySuffixs(parseModelChan.FilePathName, m)
	if err != nil {
		belogs.Error("parseCerFiles():GetAllFilesBySuffixs fail, parseModelChan.FilePathName:", parseModelChan.FilePathName, "   err:", err)
		belogs.Debug("parseCerFiles():GetAllFilesBySuffixs, before RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		atomic.AddInt64(&rpQueue.RsyncingParsingCount, -1)
		belogs.Debug("parseCerFiles():GetAllFilesBySuffixs, after RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		return
	}
	belogs.Debug("parseCerFiles(): len(cerFiles):", len(cerFiles))

	// if there are no cer files, return
	if len(cerFiles) == 0 {
		belogs.Debug("parseCerFiles():len(cerFiles)==0, before RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		atomic.AddInt64(&rpQueue.RsyncingParsingCount, -1)
		belogs.Debug("parseCerFiles():len(cerFiles)==0, after RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
		return
	}

	// foreach every cerfiles to parseCerAndGetSubCaRepositoryUrl
	subCaRepositoryUrls := make([]string, 0, len(cerFiles))
	for _, cerFile := range cerFiles {
		// just trigger sync ,no need save to db
		subCaRepositoryUrl := parseCerAndGetSubCaRepositoryUrl(cerFile)
		if len(subCaRepositoryUrl) > 0 {
			subCaRepositoryUrls = append(subCaRepositoryUrls, subCaRepositoryUrl)
		}
	}
	belogs.Debug("parseCerFiles(): len(subCaRepositoryUrls):", len(subCaRepositoryUrls))

	// check rsync concurrent count, wait some time,
	// the father rsyncingparsingcount -1 ,and the children rsyncingparsingcount + len()
	belogs.Debug("parseCerFiles():will add subCaRepositoryUrls, before RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
	atomic.AddInt64(&rpQueue.RsyncingParsingCount, int64(len(subCaRepositoryUrls)-1))
	belogs.Debug("parseCerFiles():will add subCaRepositoryUrls, after RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
	addSubCaRepositoryUrlsToRpQueue(subCaRepositoryUrls)
}

// call /parsevalidate/parse to parse cert, and save result
func parseCerAndGetSubCaRepositoryUrl(cerFile string) (subCaRepositoryUrl string) {

	// call parse, not need to save body to db
	start := time.Now()
	belogs.Debug("ParseCerAndGetSubCaRepositoryUrl():/parsevalidate/parsefilesimple cerFile:", cerFile)
	// post file, still use http
	resp, body, err := httpclient.PostFile("http", conf.String("rpstir2::serverHost"), conf.Int("rpstir2::serverHttpPort"),
		"/parsevalidate/parsefilesimple", cerFile, "")
	belogs.Debug("ParseCerAndGetSubCaRepositoryUrl():after /parsevalidate/parsefilesimple cerFile:", cerFile, len(body))

	if err != nil {
		rpQueue.RsyncResult.FailParseValidateCerts[cerFile] = err.Error()
		belogs.Error("ParseCerAndGetSubCaRepositoryUrl(): filerepo file connecteds failed:", cerFile, "   err:", err)
		return ""
	}
	defer resp.Body.Close()

	// get parse result
	parseCerSimpleResponse := model.ParseCerSimpleResponse{}
	jsonutil.UnmarshalJson(string(body), &parseCerSimpleResponse)
	belogs.Debug("ParseCerAndGetSubCaRepositoryUrl(): get from parsecert, parseCerSimpleResponse.Result:", parseCerSimpleResponse.Result)
	if parseCerSimpleResponse.HttpResponse.Result != "ok" {
		belogs.Error("ParseCerAndGetSubCaRepositoryUrl(): parsecert file failed:", cerFile, "   err:", parseCerSimpleResponse.HttpResponse.Msg)
		rpQueue.RsyncResult.FailParseValidateCerts[cerFile] = parseCerSimpleResponse.HttpResponse.Msg
		return ""
	}

	// get the sub repo url in cer, and send it to rpqueue
	belogs.Info("ParseCerAndGetSubCaRepositoryUrl(): cerFile:", cerFile, "    caRepository:", parseCerSimpleResponse.ParseCerSimple.CaRepository,
		"  time(s):", time.Now().Sub(start).Seconds())
	return parseCerSimpleResponse.ParseCerSimple.CaRepository

}

func addSubCaRepositoryUrlsToRpQueue(subCaRepositoryUrls []string) {

	rsyncConcurrentCount := conf.Int("rsync::rsyncConcurrentCount")
	belogs.Debug("AddSubCaRepositoryUrlsToRpQueue(): len(rpQueue.RsyncModelChan)+len(subCaRepositoryUrls):", len(rpQueue.RsyncModelChan),
		" + ", len(subCaRepositoryUrls), " compare rsync::rsyncConcurrentCount ", rsyncConcurrentCount)
	for i, subCaRepositoryUrl := range subCaRepositoryUrls {
		belogs.Debug("AddSubCaRepositoryUrlsToRpQueue():will PreCheckRsyncUrl, rpQueue.RsyncingParsingCount: ",
			atomic.LoadInt64(&rpQueue.RsyncingParsingCount),
			"   subCaRepositoryUrl:", subCaRepositoryUrl)
		if !rpQueue.PreCheckRsyncUrl(subCaRepositoryUrl) {
			belogs.Debug("AddSubCaRepositoryUrlsToRpQueue():PreCheckRsyncUrl have exists, before RsyncingParsingCount-1:", subCaRepositoryUrl, atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
			atomic.AddInt64(&rpQueue.RsyncingParsingCount, -1)
			belogs.Debug("AddSubCaRepositoryUrlsToRpQueue():PreCheckRsyncUrl have exists, after RsyncingParsingCount-1:", subCaRepositoryUrl, atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
			continue
		}

		curRsyncingCount := int(atomic.LoadInt64(&rpQueue.CurRsyncingCount))
		if curRsyncingCount <= 2 {
			// when less 2, not need to wait
		} else if curRsyncingCount > 2 && curRsyncingCount <= rsyncConcurrentCount {
			waitForRsyncUrl(1, subCaRepositoryUrl)
		} else {
			belogs.Debug("AddSubCaRepositoryUrlsToRpQueue():waitForRsyncUrl,i + rpQueue.curRsyncingCount: ", curRsyncingCount)
			waitForRsyncUrl(i+curRsyncingCount, subCaRepositoryUrl)
		}
		belogs.Debug("AddSubCaRepositoryUrlsToRpQueue():will AddRsyncUrl subCaRepositoryUrl: ", subCaRepositoryUrl)
		go rpQueue.AddRsyncUrl(subCaRepositoryUrl, conf.VariableString("rsync::destPath")+"/")
	}
}

// will try fail urls  to rsync again
func tryAgainFailRsyncUrls() bool {
	// try again
	belogs.Debug("TryAgainFailRsyncUrls():try fail urls again: len(rpQueue.RsyncResult.FailUrls):", len(rpQueue.RsyncResult.FailUrls),
		"      rpQueue.RsyncResult.FailUrlsTryCount:", rpQueue.RsyncResult.FailUrlsTryCount)
	if len(rpQueue.RsyncResult.FailUrls) > 0 &&
		rpQueue.RsyncResult.FailUrlsTryCount <= uint64(conf.Int("rsync::failRsyncUrlsTryCount")) {
		failRsyncUrls := make([]string, 0, len(rpQueue.RsyncResult.FailUrls))
		for failRsyncUrl := range rpQueue.RsyncResult.FailUrls {
			failRsyncUrls = append(failRsyncUrls, failRsyncUrl)
			// delete saved url ,so can try again
			rpQueue.DelRsyncAddedUrl(failRsyncUrl)
		}
		// clear fail rsync urls
		rpQueue.RsyncResult.FailUrls = make(map[string]string, 200)

		belogs.Debug("TryAgainFailRsyncUrls(): failRysncUrl:", len(failRsyncUrls), failRsyncUrls,
			"   rpQueue.RsyncResult.FailRsyncUrlsTryCount: ", rpQueue.RsyncResult.FailUrlsTryCount)
		atomic.AddUint64(&rpQueue.RsyncResult.FailUrlsTryCount, 1)
		belogs.Debug("TryAgainFailRsyncUrls():after  rpQueue.RsyncResult.FailUrlsTryCount: ", rpQueue.RsyncResult.FailUrlsTryCount)

		// check rsync concurrent count, wait some time,
		rsyncConcurrentCount := conf.Int("rsync::rsyncConcurrentCount")
		atomic.AddInt64(&rpQueue.RsyncingParsingCount, int64(len(failRsyncUrls)))
		belogs.Debug("TryAgainFailRsyncUrls(): len(failRsyncUrls):", len(failRsyncUrls),
			"  failRsyncUrls:", failRsyncUrls)

		for i, failRsyncUrl := range failRsyncUrls {
			curRsyncingCount := int(atomic.LoadInt64(&rpQueue.CurRsyncingCount))
			if curRsyncingCount <= 2 {

			} else if curRsyncingCount > 2 && curRsyncingCount <= rsyncConcurrentCount {
				belogs.Debug("TryAgainFailRsyncUrls():waitForRsyncUrl, i is smaller, i: ", i,
					" , will wait  1:", 1)
				waitForRsyncUrl(1, failRsyncUrl)
			} else {
				belogs.Debug("TryAgainFailRsyncUrls():waitForRsyncUrl, i is bigger, i: ", i,
					" , will wait  curRsyncingCount+1:", curRsyncingCount+1)
				waitForRsyncUrl(1+curRsyncingCount/2, failRsyncUrl)
			}
			go rpQueue.AddRsyncUrl(failRsyncUrl, conf.VariableString("rsync::destPath")+"/")
		}
		return true
	}
	return false
}

//rsync  should wait for some url, because some nic limited access frequency
func waitForRsyncUrl(curRsyncCount int, url string) {

	if curRsyncCount == 0 {
		return
	}
	belogs.Debug("waitForRsyncUrl(): curRsyncCount : ", curRsyncCount, "  will add:", conf.Int("rsync::rsyncDefaultWaitMs"), " 2* runtime.NumCPU():", 2*runtime.NumCPU())
	curRsyncCount = curRsyncCount + conf.Int("rsync::rsyncDefaultWaitMs") + 2*runtime.NumCPU()

	// apnic and afrinic should not visit too often
	if strings.Contains(url, "rpki.apnic.net") {
		curRsyncCount = curRsyncCount * 2
	} else if strings.Contains(url, "rpki.afrinic.net") {
		curRsyncCount = curRsyncCount * 10
	}
	min := uint(conf.Int("rsync::rsyncPerDelayMs") * curRsyncCount)
	randR := uint(conf.Int("rsync::rsyncDelayRandMs"))
	rand := randutil.IntRange(min, randR)
	belogs.Debug("waitForRsyncUrl():after rand, url is :", url, ",  curRsyncCount is:", curRsyncCount, ", will sleep rand ms:", rand)
	time.Sleep(time.Duration(rand) * time.Millisecond)
}
