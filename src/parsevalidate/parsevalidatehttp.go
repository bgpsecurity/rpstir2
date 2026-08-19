package parsevalidate

import (
	"os"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	parsevalidatecore "github.com/bgpsecurity/rpstir2/parsevalidate-core"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

func ParseValidateStart(c *gin.Context) {
	belogs.Debug("ParseValidateStart(): start: ")

	//check serviceState
	rpstir2RpUrl := "https://" + conf.String("rpstir2-rp::serverHost") + ":" + conf.String("rpstir2-rp::serverHttpsPort")
	_, _, err := httpclient.Post(rpstir2RpUrl+"/sys/servicestate",
		`{"operate":"enter","state":"parsevalidate"}`, false)
	if err != nil {
		belogs.Error("ParseValidateStart(): servicestate enter parsevalidate fail, rpstir2RpUrl:", rpstir2RpUrl, err)
	}
	go func() {
		nextStep, err := parseValidateStart()
		belogs.Debug("ParseValidateStart():  parseValidateStart end,  nextStep is :", nextStep, err)
		// leave serviceState
		if err != nil {
			// will end this whole sync
			belogs.Error("ParseValidateStart():  parseValidateStart fail", err)
			_, _, err = httpclient.Post(rpstir2RpUrl+"/sys/servicestate",
				`{"operate":"leave","state":"end"}`, false)
			if err != nil {
				belogs.Error("ParseValidateStart(): after parseValidateStart fail, servicestate leave end fail, rpstir2RpUrl:", rpstir2RpUrl, err)
			}
		} else {
			_, _, err = httpclient.Post(rpstir2RpUrl+"/sys/servicestate",
				`{"operate":"leave","state":"parsevalidate"}`, false)
			if err != nil {
				belogs.Error("ParseValidateStart(): after parseValidateStart ok, servicestate leave parsevalidate fail, rpstir2RpUrl:", rpstir2RpUrl, err)
			}

			// will call chainValidate
			//	go httpclient.Post(rpstir2RpUrl+"/chainvalidate/start", "", false)
			belogs.Info("ParseValidateStart():  sync.Start end,  nextStep is :", nextStep)
			if conf.String("chain::adoptChainValidatePartial") == "true" {
				belogs.Info("ParseValidateStart(): will call partial chainvalidate")
				go httpclient.Post(rpstir2RpUrl+"/chainvalidatepartial/start", "", false)
			} else {
				belogs.Info("ParseValidateStart(): will call global chainvalidate")
				go httpclient.Post(rpstir2RpUrl+"/chainvalidate/start", "", false)
			}
		}

	}()

	ginserver.ResponseOk(c, nil)

}

// upload file to parse;
// only one file
func ParseValidateFile(c *gin.Context) {
	start := time.Now()
	tmpDir, err := os.MkdirTemp("", "ParseValidateFile") // temp dir
	if err != nil {
		belogs.Error("ParseValidateFile(): TempDir fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	defer os.RemoveAll(tmpDir)
	belogs.Debug("ParseValidateFile(): tmpDir:", tmpDir)

	receiveFile, err := ginserver.ReceiveFile(c, tmpDir)
	if err != nil {
		belogs.Error("ParseValidateFile(): ReceiveFile fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("ParseValidateFile(): ReceiveFile, receiveFile:", receiveFile)

	certType, certModel, stateModel, originModel, fileHash, err := parsevalidatecore.ParseValidateFile(receiveFile)
	if err != nil {
		belogs.Error("ParseValidateFile(): ParseValidateFile fail: receiveFile:", receiveFile, err, "  time(s):", time.Since(start))
		ginserver.ResponseFail(c, err, "")
		return
	}
	parseCertResponse := model.ParseCertResponse{
		CertType:    certType,
		CertModel:   certModel,
		OriginModel: originModel,
		StateModel:  stateModel,
		FileHash:    fileHash,
	}
	belogs.Debug("ParseValidateFile(): parseCertResponse:", jsonutil.MarshalJson(parseCertResponse), "  time(s):", time.Since(start))
	ginserver.ResponseOk(c, parseCertResponse)
}

// upload file to parse;
// only one file
func ParseFile(c *gin.Context) {
	start := time.Now()
	tmpDir, err := os.MkdirTemp("", "ParseFile") // temp dir
	if err != nil {
		belogs.Error("ParseFile(): TempDir fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	defer os.RemoveAll(tmpDir)
	belogs.Debug("ParseFile(): tmpDir:", tmpDir)

	receiveFile, err := ginserver.ReceiveFile(c, tmpDir)
	if err != nil {
		belogs.Error("ParseFile(): ReceiveFile fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("ParseFile(): ReceiveFile, receiveFile:", receiveFile)

	_, certModel, _, _, _, err := parsevalidatecore.ParseValidateFile(receiveFile)
	if err != nil {
		belogs.Error("ParseFile(): ParseFile: err:", receiveFile, err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("ParseFile(): ok, certModel:", jsonutil.MarshalJsonIndent(certModel),
		"  time(s):", time.Since(start))
	ginserver.ResponseOk(c, certModel)
}

// upload file to parse to get ca repo
func ParseFileSimple(c *gin.Context) {
	start := time.Now()
	tmpDir, err := os.MkdirTemp("", "ParseFileSimple") // temp dir
	if err != nil {
		belogs.Error("ParseFileSimple(): TempDir fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	defer os.RemoveAll(tmpDir)
	belogs.Debug("ParseFileSimple(): tmpDir:", tmpDir)

	receiveFile, err := ginserver.ReceiveFile(c, tmpDir)
	if err != nil {
		belogs.Error("ParseFileSimple(): ReceiveFile fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("ParseFileSimple(): ReceiveFile, receiveFile:", receiveFile)

	parseCerSimple, err := parsevalidatecore.ParseFileSimple(receiveFile)
	if err != nil {
		belogs.Error("ParseFileSimple(): ParseFileSimple: err:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("ParseFileSimple():ok, parseCerSimple:", jsonutil.MarshalJson(parseCerSimple))
	belogs.Info("ParseFileSimple():ok, parseCerSimple.RpkiNotify:", parseCerSimple.RpkiNotify,
		" parseCerSimple.CaRepository:", parseCerSimple.CaRepository, "   time(s):", time.Since(start))

	ginserver.ResponseOk(c, parseCerSimple)
}
