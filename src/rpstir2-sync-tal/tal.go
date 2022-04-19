package tal

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/base64util"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/rrdputil"
	"github.com/cpusoft/goutil/rsyncutil"
	"github.com/cpusoft/goutil/urlutil"
)

func getTals() (passTalModels []model.TalModel, err error) {
	start := time.Now()
	belogs.Debug("getTals():")

	// get all tal files
	talPath := conf.VariableString("sync::talPath")
	files, err := getAllTalFiles(talPath)
	if err != nil {
		belogs.Error("getTals(): GetAllTalFile failed:", err)
		return
	}
	belogs.Debug("getTals(): tal path is ", talPath, ", all tal files are ", files)

	// parse tal files
	talModels, err := parseTalFiles(files)
	if err != nil {
		belogs.Error("getTals(): GetAllTalFile failed:", err)
		return
	}
	belogs.Info("getTals(): files:", jsonutil.MarshalJson(files), "     talModels:", jsonutil.MarshalJson(talModels))

	// save tal files to local temp dir, and judge sync style(rrdp/rsync), and verify using subjectpublickeyinfo
	passTalModels, err = syncToLocalAndParseValidateCers(talModels)
	if err != nil {
		belogs.Error("getTals(): syncToLocal failed:", err)
		return
	}

	belogs.Info("getTals(): passTalModels:", jsonutil.MarshalJson(passTalModels), "  time(s):", time.Now().Sub(start).Seconds())
	return passTalModels, nil
}

func getAllTalFiles(talPath string) ([]string, error) {

	belogs.Debug("getAllTalFiles():input tal file or path :", talPath)

	// get all tail files in tal path
	isDir, err := osutil.IsDir(talPath)
	if err != nil {
		belogs.Error("getAllTalFiles():IsDir err:", talPath, err)
		return nil, err
	}
	if isDir {
		suffixs := make(map[string]string)
		suffixs[".tal"] = ".tal"
		files, err := osutil.GetAllFilesBySuffixs(talPath, suffixs)
		if err != nil {
			belogs.Error("getAllTalFiles():GetAllFilesBySuffixs err:", talPath, err)
			return nil, err
		}
		belogs.Debug("getAllTalFiles(): tal path is ", talPath, ", all tal files are ", files)
		return files, nil
	} else {
		belogs.Error("getAllTalFiles():talPath is not dir:", talPath)
		return nil, errors.New("talPath is not dir")
	}

}

func parseTalFiles(files []string) (talModels []model.TalModel, err error) {
	belogs.Debug("parseTalFiles(): files:", files)

	talModels = make([]model.TalModel, 0)
	for _, file := range files {
		talModel, err := parseTalFile(file)
		if err != nil {
			belogs.Error("parseTalFiles():tal file err, will continue to next: ", file, err)
			continue
		}
		talModels = append(talModels, talModel)
	}
	belogs.Debug("parseTalFiles(): files: ", files, ", talModels:", jsonutil.MarshalJson(talModels))
	return talModels, nil
}

func parseTalFile(file string) (talModel model.TalModel, err error) {
	belogs.Debug("parseTalFile(): file:", file)

	f, err := os.Open(file)
	if err != nil {
		belogs.Error("parseTalFile(): file Open err:", file, err)
		return talModel, err
	}

	input := bufio.NewScanner(f)
	var buffer bytes.Buffer
	talSyncUrls := make([]model.TalSyncUrl, 0)
	for input.Scan() { // when  meet "\n" or \r\n
		tmp := strings.TrimSpace(input.Text())
		if strings.HasPrefix(tmp, "#") || len(tmp) == 0 {
			continue
		}
		tmp = strings.Replace(tmp, "\r", "", -1)
		tmp = strings.Replace(tmp, "\n", "", -1)
		if strings.HasPrefix(tmp, "https:") || strings.HasPrefix(tmp, "rsync:") {
			talSyncUrl := model.TalSyncUrl{}
			if strings.HasPrefix(tmp, "https:") {
				talSyncUrl.SupportRrdp = true
			}
			if strings.HasPrefix(tmp, "rsync:") {
				talSyncUrl.SupportRsync = true
			}
			if talSyncUrl.SupportRrdp == false && talSyncUrl.SupportRsync == false {
				belogs.Error("parseTalFile(): not support url:", file, tmp)
				return talModel, errors.New("not support url:" + file)
			}
			talSyncUrl.TalUrl = tmp
			talSyncUrls = append(talSyncUrls, talSyncUrl)
		} else {
			buffer.WriteString(tmp)
		}
	}
	talModel.TalSyncUrls = talSyncUrls
	talModel.SubjectPublicKeyInfo = buffer.String()
	belogs.Debug("parseTalFile():file:", file, "  talModel:", jsonutil.MarshalJson(talModel))
	return talModel, nil
}

func syncToLocalAndParseValidateCers(talModels []model.TalModel) (passTalModels []model.TalModel, err error) {
	start := time.Now()
	belogs.Debug("syncToLocalAndParseValidateCers(): talModels:", jsonutil.MarshalJson(talModels))

	// will save cer to local temp dir
	tmpDir, err := ioutil.TempDir("", "tal") // temp file
	if err != nil {
		belogs.Error("syncToLocalAndParseValidateCers(): TempDir err:", err)
		return nil, err
	}
	var wg sync.WaitGroup
	for i := range talModels {
		for j := range talModels[i].TalSyncUrls {
			talUrl := talModels[i].TalSyncUrls[j].TalUrl
			belogs.Debug("syncToLocalAndParseValidateCers(): talUrl:", talUrl)

			// save to localfile,and parse and verify
			wg.Add(1)
			go syncToLocalAndParseValidateCer(tmpDir, talUrl, talModels[i].SubjectPublicKeyInfo, &talModels[i].TalSyncUrls[j], &wg)

		}
	}
	wg.Wait()
	os.RemoveAll(tmpDir)

	// remove error syncUrl
	// if all url is error, then rm this talModel
	belogs.Debug("syncToLocalAndParseValidateCers(): before remove error, all talModels:", len(talModels), jsonutil.MarshalJson(talModels))
	for i := range talModels {
		passTalModel := model.TalModel{}
		passTalModel.TalSyncUrls = make([]model.TalSyncUrl, 0)
		for j := range talModels[i].TalSyncUrls {
			if talModels[i].TalSyncUrls[j].Error == "" {
				passTalModel.TalSyncUrls = append(passTalModel.TalSyncUrls, talModels[i].TalSyncUrls[j])
			}
		}
		if len(passTalModel.TalSyncUrls) > 0 {
			passTalModels = append(passTalModels, talModels[i])
		}
	}
	belogs.Debug("syncToLocalAndParseValidateCers():  after remove error,  passTalModels:", len(passTalModels), jsonutil.MarshalJson(passTalModels),
		"  time(s):", time.Now().Sub(start).Seconds())
	return passTalModels, nil
}

func syncToLocalAndParseValidateCer(tmpDir, talUrl, subjectPublicKeyInfo string, talSyncUrl *model.TalSyncUrl, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	belogs.Debug("syncToLocalAndParseValidateCer(): tmpDir,  talUrl:", tmpDir, talUrl)
	talSyncUrl.Error = ""
	// get file name
	_, _, file, err := urlutil.HostAndPathAndFile(talUrl)
	if err != nil {
		belogs.Error("syncToLocalAndParseValidateCer(): HostAndPathAndFile err:", talUrl, err)
		talSyncUrl.Error = err.Error()
		return
	}
	belogs.Debug("syncToLocalAndParseValidateCer(): HostAndPathAndFile, file:", file)

	// if url is https, then get cer file by http; if url is rsync, then get cer file by rsnc
	if strings.HasPrefix(talUrl, "https:") {
		// should verify https
		resp, body, err := httpclient.GetHttpsVerify(talUrl, true)
		if err != nil {
			belogs.Error("syncToLocalAndParseValidateCer(): GetHttpsVerify fail, err:", talUrl, err)
			talSyncUrl.Error = err.Error()
			talSyncUrl.SupportRrdp = false
			return
		}
		if resp.StatusCode != 200 {
			belogs.Error("syncToLocalAndParseValidateCer(): GetHttpsVerify StatusCode != 200 :", talUrl, resp.StatusCode)
			talSyncUrl.Error = "http status code is not 200"
			talSyncUrl.SupportRrdp = false
			return
		}
		// save to localFile
		talSyncUrl.LocalFile = osutil.JoinPathFile(tmpDir, file)
		err = fileutil.WriteBytesToFile(talSyncUrl.LocalFile, []byte(body))
		if err != nil {
			belogs.Error("syncToLocalAndParseValidateCer(): WriteBytesToFile fail, err:", talUrl, err)
			talSyncUrl.Error = err.Error()
			talSyncUrl.SupportRrdp = false
			return
		}

	} else if strings.HasPrefix(talUrl, "rsync:") {
		// rsycn to local file
		rsyncDestPath, _, err := rsyncutil.RsyncQuiet(talUrl, tmpDir)
		if err != nil {
			belogs.Error("syncToLocalAndParseValidateCer(): RsyncQuiet fail, url, tmpDir, err:", talUrl, tmpDir, err)
			talSyncUrl.Error = err.Error()
			talSyncUrl.SupportRsync = false
			return
		}
		talSyncUrl.LocalFile = osutil.JoinPathFile(rsyncDestPath, file)

	} else {
		talSyncUrl.Error = "talUrl is not supported:" + talUrl
		talSyncUrl.SupportRrdp = false
		talSyncUrl.SupportRsync = false
		return
	}

	// parse to get rsync style, and check cer, using subjectpublickeyinfo
	err = parseAndValidateCer(talUrl, subjectPublicKeyInfo, tmpDir, talSyncUrl)
	if err != nil {
		belogs.Error("syncToLocalAndParseValidateCer(): parseAndValidateCer err:", talSyncUrl.LocalFile, err)
		talSyncUrl.Error = err.Error()
		return
	}

	belogs.Debug("syncToLocalAndParseValidateCer(): syncUrl:", jsonutil.MarshalJson(talSyncUrl), "  time(s):", time.Now().Sub(start).Seconds())
	return

}

func parseAndValidateCer(talUrl, subjectPublicKeyInfo, tmpDir string, talSyncUrl *model.TalSyncUrl) (err error) {
	start := time.Now()
	belogs.Debug("parseAndValidateCer(): talUrl:", talUrl, "   subjectPublicKeyInfo:", subjectPublicKeyInfo, "  tmpDir:", tmpDir,
		"  talSyncUrl:", jsonutil.MarshalJson(talSyncUrl))

	// parse by /parsevalidate/parsefilesimple
	// post file, still use http
	parseCerSimple := model.ParseCerSimple{}
	err = httpclient.PostFileAndUnmarshalResponseModel("http://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpPort")+
		"/parsevalidate/parsefilesimple", talSyncUrl.LocalFile, "file", false, &parseCerSimple)
	if err != nil {
		belogs.Error("parseAndValidateCer(): PostFileAndUnmarshalResponseModel fail:", talSyncUrl.LocalFile, "   err:", err)
		return err
	}
	belogs.Debug("parseAndValidateCer(): PostFileAndUnmarshalResponseModel file :", talSyncUrl.LocalFile,
		"   parseCerSimple:", jsonutil.MarshalJson(parseCerSimple))

	if conf.String("sync::supportTestCer") == "true" {
		if strings.Contains(talUrl, "zdns.cn") && strings.Contains(talSyncUrl.LocalFile, "zdns-test.cer") {
			parseCerSimple.RpkiNotify = "https://rpki-rrdp-test.zdns.cn/notification.xml"
			parseCerSimple.CaRepository = ""
			belogs.Info("parseAndValidateCer():supportTestCer, talUrl, talSyncUrl.LocalFile:", talUrl, talSyncUrl.LocalFile,
				"    parseCerSimple:", jsonutil.MarshalJson(parseCerSimple))
		}
	}

	// check rpkiNotify(rrdp)
	if len(parseCerSimple.RpkiNotify) > 0 {
		start := time.Now()
		belogs.Debug("parseAndValidateCer(): test rrdp is ok:", parseCerSimple.RpkiNotify)
		_, err = rrdputil.GetRrdpNotification(parseCerSimple.RpkiNotify)
		if err != nil {
			belogs.Error("GetRrdpSnapshot(): rrdputil.GetRrdpNotification fail:", parseCerSimple.RpkiNotify,
				"  time(s):", time.Now().Sub(start).Seconds(), err)
			talSyncUrl.SupportRrdp = false
		} else {
			talSyncUrl.SupportRrdp = true
			talSyncUrl.RrdpUrl = parseCerSimple.RpkiNotify
		}
	} else {
		talSyncUrl.SupportRrdp = false
	}
	belogs.Debug("parseAndValidateCer():after check rpkiNotify(rrdp), parseCerSimple.RpkiNotify:", parseCerSimple.RpkiNotify,
		"  talUrl:", talUrl, "  talSyncUrl:", jsonutil.MarshalJson(talSyncUrl))

	// check caRepository(rsync), but just test talUrl
	if len(parseCerSimple.CaRepository) > 0 {
		// must start with "rsync", otherwise root cer cannot  download by rsync
		if strings.HasPrefix(talSyncUrl.TalUrl, "rsync:") {
			belogs.Debug("parseAndValidateCer(): test rsync is ok:", talSyncUrl.TalUrl)
			_, _, err := rsyncutil.RsyncQuiet(talUrl, tmpDir)
			if err != nil {
				belogs.Error("parseAndValidateCer(): RsyncQuiet fail, url,err:", talSyncUrl.TalUrl, err)
				talSyncUrl.SupportRsync = false
			} else {
				talSyncUrl.SupportRsync = true
				talSyncUrl.RsyncUrl = talSyncUrl.TalUrl //it is root cer, is not caRepository,
			}
		} else {
			belogs.Debug("parseAndValidateCer(): have CaRepository, but not start with 'rsync':", talSyncUrl.TalUrl, err)
			talSyncUrl.SupportRsync = false
		}
	} else {
		talSyncUrl.SupportRsync = false
	}
	belogs.Debug("parseAndValidateCer():after check caRepository(rsync), parseCerSimple.CaRepository:", parseCerSimple.CaRepository,
		"  talUrl:", talUrl, "  talSyncUrl:", jsonutil.MarshalJson(talSyncUrl))

	// validate, using public key info
	subjectPublicKeyInfoInCer := base64util.EncodeBase64(parseCerSimple.SubjectPublicKeyInfo)
	belogs.Debug("parseAndValidateCer(): localFile:", talSyncUrl.LocalFile,
		"  subjectPublicKeyInfoInCer:\r\n", subjectPublicKeyInfoInCer, "\r\n   subjectPublicKeyInfo:\r\n", subjectPublicKeyInfo)
	if subjectPublicKeyInfoInCer != subjectPublicKeyInfo {
		belogs.Error("parseAndValidateCer(): subjectInfo is not equal:", talSyncUrl.LocalFile,
			"  subjectPublicKeyInfoInCer:\r\n", subjectPublicKeyInfoInCer, "\r\n   subjectPublicKeyInfo:\r\n", subjectPublicKeyInfo)
		return errors.New("subjectInfo is not equal")
	}
	belogs.Info("parseAndValidateCer(): pass, talUrl:", talUrl, "   subjectPublicKeyInfo:", subjectPublicKeyInfo, "  tmpDir:", tmpDir,
		"  talSyncUrl:", jsonutil.MarshalJson(talSyncUrl), "   time(s):", time.Now().Sub(start))
	return nil
}
