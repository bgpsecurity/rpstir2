package parsevalidate

import (
	"strconv"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	convert "github.com/cpusoft/goutil/convert"
	hashutil "github.com/cpusoft/goutil/hashutil"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	opensslutil "github.com/cpusoft/goutil/opensslutil"
	osutil "github.com/cpusoft/goutil/osutil"
	regexputil "github.com/cpusoft/goutil/regexputil"

	"model"
	"parsevalidate/openssl"
	"parsevalidate/packet"
	"parsevalidate/util"
)

//Try to store the error in statemode instead of returning err
func ParseValidateMft(certFile string) (mftModel model.MftModel, stateModel model.StateModel, err error) {

	stateModel = model.NewStateModel()
	err = parseMftModel(certFile, &mftModel, &stateModel)
	if err != nil {
		belogs.Error("ParseValidateMft():parseMftByOpenssl err:", certFile, err)
		return mftModel, stateModel, nil
	}
	belogs.Debug("ParseValidateMft(): mftModel:", jsonutil.MarshalJson(mftModel))

	err = validateMftModel(&mftModel, &stateModel)
	if err != nil {
		belogs.Error("ParseValidateMft():validateMftModel err:", certFile, err)
		return mftModel, stateModel, nil
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("ParseValidateMft():stateModel have errors or warnings", certFile, "     stateModel:", jsonutil.MarshalJson(stateModel))
	}

	belogs.Debug("ParseValidateMft(): mftModel.FilePath, mftModel.FileName, mftModel.Ski, mftModel.Aki:",
		mftModel.FilePath, mftModel.FileName, mftModel.Ski, mftModel.Aki)
	return mftModel, stateModel, nil
}

// some parse may return err, will stop
func parseMftModel(certFile string, mftModel *model.MftModel, stateModel *model.StateModel) error {
	belogs.Debug("parseMftModel(): certFile: ", certFile)
	mftModel.FilePath, mftModel.FileName = osutil.GetFilePathAndFileName(certFile)

	//https://blog.csdn.net/Zhymax/article/details/7683925
	//openssl asn1parse -in -ard.mft -inform DER
	results, err := opensslutil.GetResultsByOpensslAns1(certFile)
	if err != nil {
		belogs.Error("parseMftModel(): GetResultsByOpensslAns1: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseMftModel(): len(results):", len(results))

	//get file hash
	mftModel.FileHash, err = hashutil.Sha256File(certFile)
	if err != nil {
		belogs.Error("parseMftModel(): Sha256File: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	// get mft hex
	// first HEX DUMP
	/*
		   39:d=4  hl=2 l=  11 prim: OBJECT            :1.2.840.113549.1.9.16.1.26
		   52:d=4  hl=2 l=inf  cons: cont [ 0 ]
		   54:d=5  hl=2 l=inf  cons: OCTET STRING
		   56:d=6  hl=3 l= 137 prim: OCTET STRING      [HEX DUMP]:308186020200CA180F323031383036323831373030
		32345A180F32303138303632393138303032345A060960864801650304020130533051162C36353736393433633735383262
		3164656266666261303564363235343034323462633765626363352E63726C032100154269177B0346014642A367DA415F32
		C2BFE7C4EAD8AED59ACCF8F20220F89C
	*/

	err = openssl.ParseMftModelByOpensslResults(results, mftModel)
	if err != nil {
		belogs.Error("parseMftModel():ParseMftModelByOpensslResults  certFile:", certFile, "  err:", err, " will try parseMftModelByPacket")
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)

		err = parseMftModelByPacket(certFile, mftModel)
		if err != nil {
			belogs.Error("parseMftModel():parseMftModelByPacket err:", certFile, err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}

	}

	mftModel.EContentType, err = openssl.ParseMftEContentTypeByOpensslResults(results)
	if err != nil {
		belogs.Error("parseMftModel():ParseEContentTypeByOpensslResults  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	mftModel.SignerInfoModel, err = openssl.ParseSignerInfoModelByOpensslResults(results)
	if err != nil {
		belogs.Error("parseMftModel():ParseSignerInfoModelByOpensslResults  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// get cer info in mft
	cerFile, fileByte, start, end, err := openssl.ParseByOpensslAns1ToX509(certFile, results)
	if err != nil {
		belogs.Error("parseMftModel():ParseByOpensslAns1ToX509  certFile:", certFile, "  err:", err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	defer osutil.CloseAndRemoveFile(cerFile)

	results, err = opensslutil.GetResultsByOpensslX509(cerFile.Name())
	if err != nil {
		belogs.Error("parseMftModel(): GetResultsByOpensslX509: err: ", err, ": "+cerFile.Name())
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse ee certificate by openssl",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseMftModel(): len(results):", len(results))

	mftModel.Aki, mftModel.Ski, err = openssl.ParseAkiSkiByOpensslResults(results)
	if err != nil {
		belogs.Error("parseMftByOpenssl(): ParseAiaModelSiaModelByOpensslResults: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// AIA SIA
	mftModel.AiaModel, mftModel.SiaModel, err = openssl.ParseAiaModelSiaModelByOpensslResults(results)
	if err != nil {
		belogs.Error("parseMftModel(): ParseAiaModelSiaModelByOpensslResults: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}

	// EE
	mftModel.EeCertModel, err = ParseEeCertModel(cerFile.Name(), fileByte, start, end)
	if err != nil {
		belogs.Error("parseMftModel(): ParseEeCertModel: err: ", err, ": "+certFile)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
	}
	belogs.Debug("parseMftModel(): mftModel:", jsonutil.MarshalJson(mftModel))
	return nil
}

func parseMftModelByPacket(certFile string, mftModel *model.MftModel) error {

	fileByte, fileDecodeBase64Byte, err := util.ReadFileAndDecodeBase64(certFile)
	if err != nil {
		belogs.Error("parseMftModelByPacket():ReadFile return err: ", certFile, err)
		return err
	}

	//get file hash
	mftModel.FileHash = hashutil.Sha256(fileByte)

	pack := packet.DecodePacket(fileDecodeBase64Byte)
	//packet.PrintPacketString("parseMftModelByPacket():DecodePacket: ", pack, true, true)

	var oidPacketss = &[]packet.OidPacket{}
	packet.TransformPacket(pack, oidPacketss)
	packet.PrintOidPacket(oidPacketss)

	// manifests,
	err = packet.ExtractMftOid(oidPacketss, certFile, fileDecodeBase64Byte, mftModel)
	if err != nil {
		belogs.Error("parseMftModelByPacket():ExtractMftOid err:", certFile, err)
		return err
	}

	return nil
}

// only validate mft self.  in chain check, will check file list;;;;
// https://datatracker.ietf.org/doc/rfc6486/?include_text=1   Manifests for the Resource Public Key Infrastructure (RPKI)  4.4.Manifest Validation;;;;;;;
// roa_validate.c  manifestValidate()
// TODO: sqhl.c P2036 updateManifestObjs(): check file and hash in mft to actually files
func validateMftModel(mftModel *model.MftModel, stateModel *model.StateModel) (err error) {

	// The version of the rpkiManifest is 0
	if mftModel.Version != 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Wrong Version number",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// check mft number ,should >0
	mftNumberByte := []byte(mftModel.MftNumber)
	//Manifest verifiers MUST be able to handle number values up to 20 octets. Conforming manifest issuers MUST NOT use number values longer than 20 octets.
	if len(mftNumberByte) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Manifest Number is zero",
			Detail: ""}
		stateModel.AddWarning(&stateMsg)
	}
	if len(mftNumberByte) > 20*2 {
		le := strconv.Itoa(len(mftNumberByte))
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Manifest Number is too long",
			Detail: "Manifest Number length is " + le}
		stateModel.AddWarning(&stateMsg)
	}
	isHex, err := regexputil.IsHex(mftModel.MftNumber)
	if !isHex || err != nil {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Manifest Number is not a hexadecimal number",
			Detail: mftModel.MftNumber}
		stateModel.AddError(&stateMsg)
	}
	if len(mftModel.MftNumber) > 1024 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Manifest Number is too long",
			Detail: mftModel.MftNumber}
		stateModel.AddError(&stateMsg)
	}

	// check the hash algorithm
	if mftModel.FileHashAlg != "2.16.840.1.101.3.4.2.1" {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Oid of fileHashAlg is not 2.16.840.1.101.3.4.2.1",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	// check check_mft_filenames will in chain check

	// check legal filename
	// on sync time ,file may not have sync, so only check filename is or not legal
	// not actually check file
	for _, fileHash := range mftModel.FileHashModels {
		fileName := fileHash.File
		hash := fileHash.Hash
		ext := osutil.Ext(fileName)
		if ext != ".cer" && ext != ".roa" && ext != ".crl" {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "The file in fileList is not one of the three types of cer/roa/crl",
				Detail: "The file is " + fileName}
			stateModel.AddError(&stateMsg)
		}
		if len(hash) != 64 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "The length of the hash in fileList is not 64",
				Detail: "illegal hash is " + hash}
			stateModel.AddError(&stateMsg)
		}
	}
	// check duplicate file name
	for i1 := 0; i1 < len(mftModel.FileHashModels); i1++ {
		duplicate := false
		fileHash1 := mftModel.FileHashModels[i1]
		for i2 := i1 + 1; i2 < len(mftModel.FileHashModels); i2++ {
			fileHash2 := mftModel.FileHashModels[i2]
			if fileHash1.File == fileHash2.File {
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "There are duplicate files in fileList",
					Detail: ""}
				stateModel.AddError(&stateMsg)
				duplicate = true
				break
			}
		}
		if duplicate {
			break
		}
	}

	//check time
	now := time.Now()
	if mftModel.ThisUpdate.IsZero() {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ThisUpdate is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if mftModel.NextUpdate.IsZero() {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	//thisUpdate precedes nextUpdate.
	if mftModel.ThisUpdate.After(now) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ThisUpdate is later than the current time",
			Detail: "now is " + convert.Time2StringZone(now) + ", thisUpdate is " + convert.Time2StringZone(mftModel.ThisUpdate)}
		if conf.Bool("policy::allowNotYetMft") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}
	}
	if mftModel.NextUpdate.Before(now) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is earlier than the current time",
			Detail: "now is " + convert.Time2StringZone(now) + ", nextUpdate is " + convert.Time2StringZone(mftModel.NextUpdate)}
		if conf.Bool("policy::allowStaleMft") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}
	}
	if mftModel.ThisUpdate.After(mftModel.NextUpdate) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is earlier than ThisUpdate",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	if mftModel.ThisUpdate.Before(mftModel.EeCertModel.NotBefore) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ThisUpdate is later than the NotBefore of EE",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if mftModel.NextUpdate.After(mftModel.EeCertModel.NotAfter) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is later than the NotAfter of EE",
			Detail: ""}
		stateModel.AddWarning(&stateMsg)
	}

	// ski aki
	if len(mftModel.Ski) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SKI is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	// hash is 160bit --> 20Byte --> 40Str
	if len(mftModel.Ski) != 40 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "SKI length is wrong",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if len(mftModel.Aki) == 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "AKI is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	// hash is 160bit --> 20Byte --> 40Str
	if len(mftModel.Aki) != 40 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "AKI length is wrong",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	//TODO, todo,Manifest's EE certificate has RFC3779 resources that are not marked inherit, in roa_vildate.c P1009
	err = ValidateEeCertModel(stateModel, &mftModel.EeCertModel)
	err = ValidateSignerInfoModel(stateModel, &mftModel.SignerInfoModel)

	belogs.Debug("validateMftModel():filePath, fileName,stateModel:",
		mftModel.FilePath, mftModel.FileName, jsonutil.MarshalJson(stateModel))
	return nil
}
