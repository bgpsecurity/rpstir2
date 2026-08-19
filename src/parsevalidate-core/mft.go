package parsevalidatecore

import (
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	asn1 "github.com/bgpsecurity/rpstir2/parsevalidate-asn1"
	openssl "github.com/bgpsecurity/rpstir2/parsevalidate-openssl"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/regexputil"
)

func parseValidateMft(fileModel *model.FileModel) (mftModel model.MftModel, stateModel model.StateModel, err error) {
	start := time.Now()
	belogs.Debug("parseValidateMft(): fileModel:", jsonutil.MarshalJson(fileModel))
	stateModel = model.NewStateModel()
	err = parseMftModel(fileModel, &mftModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateMft():parseMftModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return mftModel, stateModel, err
	}
	belogs.Debug("parseValidateMft(): mftModel:", mftModel.String(), "  time(s):", time.Since(start))

	err = validateMftModel(&mftModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateMft():validateMftModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return mftModel, stateModel, err
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("parseValidateMft():stateModel have errors or warnings, fileModel:", jsonutil.MarshalJson(fileModel), "     stateModel:", jsonutil.MarshalJson(stateModel))
	}
	stateModel.JudgeState()
	belogs.Debug("parseValidateMft(): mftModel.FilePath, mftModel.FileName, mftModel.Ski, mftModel.Aki:",
		mftModel.FilePath, mftModel.FileName, mftModel.Ski, mftModel.Aki, "  time(s):", time.Since(start))
	return mftModel, stateModel, nil
}

// some parse may return err, will stop
func parseMftModel(fileModel *model.FileModel, mftModel *model.MftModel, stateModel *model.StateModel) error {
	belogs.Debug("parseMftModel(): fileModel:", jsonutil.MarshalJson(fileModel))

	fileLength, err := fileutil.GetFileLength(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseMftModel(): GetFileLength fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to open file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseMftModel(): fileLength:", fileLength, " tempFile:", fileModel.TempFilePathName)

	if fileLength == 0 {
		belogs.Error("parseMftModel(): GetFileLength fail, fileLenght is empty, fileModel:",
			jsonutil.MarshalJson(fileModel))
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "File is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
		return errors.New("File " + fileModel.FilePathName + " is empty")
	}

	mftModel.FilePath, mftModel.FileName = osutil.GetFilePathAndFileName(fileModel.FilePathName)
	//get file hash
	mftModel.FileHash, err = hashutil.Sha256File(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseMftModel(): Sha256File fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	err = asn1.ParseMftModelByAsn1(fileModel, mftModel, stateModel)
	if err != nil {
		if strings.Contains(err.Error(), "indefinite length found") {
			belogs.Debug("parseMftModel(): ParseMftModelByAsn1 fail, will parse again by openssl, fileModel:", jsonutil.MarshalJson(fileModel), err)
		} else {
			belogs.Error("parseMftModel(): ParseMftModelByAsn1 fail, will parse again by openssl, fileModel:", jsonutil.MarshalJson(fileModel), err)
		}
		// have set stateMsg in ParseMftModelByAsn1 .
		err = openssl.ParseMftModelByOpenssl(fileModel, mftModel, stateModel)
		if err != nil {
			belogs.Error("parseMftModel(): ParseMftModelByOpenssl fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			// have set stateMsg in ParseMftModelByOpenssl .
			return err
		}
	}

	/*
		//https://blog.csdn.net/Zhymax/article/details/7683925
		//openssl asn1parse -in -ard.mft -inform DER
		results, err := opensslutil.GetResultsByOpensslAns1(fileModel.TempFilePathName)
		if err != nil {
			belogs.Error("parseMftModel(): GetResultsByOpensslAns1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		belogs.Debug("parseMftModel(): len(results):", len(results))

		// get mft hex
		// first HEX DUMP

			   39:d=4  hl=2 l=  11 prim: OBJECT            :1.2.840.113549.1.9.16.1.26
			   52:d=4  hl=2 l=inf  cons: cont [ 0 ]
			   54:d=5  hl=2 l=inf  cons: OCTET STRING
			   56:d=6  hl=3 l= 137 prim: OCTET STRING      [HEX DUMP]:308186020200CA180F323031383036323831373030
			32345A180F32303138303632393138303032345A060960864801650304020130533051162C36353736393433633735383262
			3164656266666261303564363235343034323462633765626363352E63726C032100154269177B0346014642A367DA415F32
			C2BFE7C4EAD8AED59ACCF8F20220F89C


		err = openssl.ParseMftModelByOpensslResults(results, mftModel)
		if err != nil {
			belogs.Error("parseMftModel():ParseMftModelByOpensslResults fail, will try parseMftModelByPacket, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)

			err = parseMftModelByPacket(fileModel, mftModel)
			if err != nil {
				belogs.Error("parseMftModel():parseMftModelByPacket fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse file",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				return err
			}

		}

		mftModel.EContentType, err = openssl.ParseMftEContentTypeByOpensslResults(results)
		if err != nil {
			belogs.Error("parseMftModel():ParseEContentTypeByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		mftModel.SignerInfoModel, err = openssl.ParseSignerInfoModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseMftModel():ParseSignerInfoModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// get cer info in mft
		eeCerFile, fileByte, start, end, err := openssl.ParseByOpensslAns1ToX509(fileModel.TempFilePathName, results)
		if err != nil {
			belogs.Error("parseMftModel():ParseByOpensslAns1ToX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse ee certificate by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		defer osutil.CloseAndRemoveFile(eeCerFile)

		results, err = opensslutil.GetResultsByOpensslX509(eeCerFile.Name())
		if err != nil {
			belogs.Error("parseMftModel(): GetResultsByOpensslX509 fail, eeCerFile:", eeCerFile.Name(), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse ee certificate by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		belogs.Debug("parseMftModel(): len(results):", len(results))

		mftModel.Aki, mftModel.Ski, err = openssl.ParseAkiSkiByOpensslResults(results)
		if err != nil {
			belogs.Error("parseMftByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// AIA SIA
		mftModel.AiaModel, mftModel.SiaModel, err = openssl.ParseAiaModelSiaModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseMftModel(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// EE
		mftModel.EeCertModel, err = ParseEeCertModel(eeCerFile.Name(), fileByte, start, end)
		if err != nil {
			belogs.Error("parseMftModel(): ParseEeCertModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}
		// get IP address in EE: RFC9286, will check ipaddress(should be empty)
		mftModel.EeCertModel.CerIpAddressModel, _, err = openssl.ParseCerIpAddressModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseMftModel(): ParseCerIpAddressModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}
	*/
	belogs.Debug("parseMftModel(): mftModel:", mftModel.String())
	return nil
}

/*
func parseMftModelByPacket(fileModel *model.FileModel, mftModel *model.MftModel) error {

	fileByte, fileDecodeBase64Byte, err := fileutil.ReadFileAndDecodeCertBase64(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseMftModelByPacket():ReadFileAndDecodeCertBase64 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
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
	err = packet.ExtractMftOid(oidPacketss, fileModel.TempFilePathName, fileDecodeBase64Byte, mftModel)
	if err != nil {
		belogs.Error("parseMftModelByPacket():ExtractMftOid fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return err
	}

	return nil
}
*/
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

	//isHex, err := regexputil.IsHex(mftModel.MftNumber)
	//if !isHex || err != nil {
	_, ok := new(big.Int).SetString(mftModel.MftNumber, 16)
	if !ok {
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
		check := regexputil.CheckRpkiFileName(fileName)
		if !check {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "The haracters in file name is illegal",
				Detail: "The file is " + fileName}
			stateModel.AddError(&stateMsg)
		}

		hash := fileHash.Hash
		ext := osutil.Ext(fileName)
		// no .mft
		// https://www.iana.org/assignments/rpki/rpki.xhtml
		if ext != ".cer" && ext != ".roa" && ext != ".crl" && ext != ".gbr" &&
			ext != ".asa" && ext != ".sig" && ext != ".moa" {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "The file in fileList is not one of the types of cer/roa/crl/gbr/asa/sig/moa",
				Detail: "The file is " + fileName}
			stateModel.AddError(&stateMsg)
		}
		if len(hash) != 64 {
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "The length of the hash in fileList is not 64",
				Detail: "The illegal hash is " + hash}
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
			Detail: "The current time is " + convert.Time2StringZone(now) + ", thisUpdate is " + convert.Time2StringZone(mftModel.ThisUpdate)}
		stateModel.AddError(&stateMsg)
	}
	if mftModel.NextUpdate.Before(now) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", nextUpdate is " + convert.Time2StringZone(mftModel.NextUpdate)}
		stateModel.AddError(&stateMsg)
	}
	if mftModel.ThisUpdate.After(mftModel.NextUpdate) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is earlier than ThisUpdate",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}

	if mftModel.ThisUpdate.Before(mftModel.EeCertModel.NotBefore) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ThisUpdate of MFT is later than the NotBefore of EE",
			Detail: ""}
		stateModel.AddError(&stateMsg)
	}
	if !mftModel.ThisUpdate.Equal(mftModel.EeCertModel.NotBefore) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "ThisUpdate of MFT is not equal to the NotBefore of EE",
			Detail: ""}
		stateModel.AddWarning(&stateMsg)
	}
	if mftModel.NextUpdate.After(mftModel.EeCertModel.NotAfter) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate of MFT is later than the NotAfter of EE",
			Detail: ""}
		stateModel.AddWarning(&stateMsg)
	}
	if !mftModel.NextUpdate.Equal(mftModel.EeCertModel.NotAfter) {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate of MFT is not equal to the NotAfter of EE",
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

	// rfc9286
	if len(mftModel.EeCertModel.CerIpAddressModel.CerIpAddresses) != 0 {
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail: "IP address of MFT's EE should be empty",
			Detail: "INRs must use the 'inherit' attribute, the current length is " +
				convert.ToString(len(mftModel.EeCertModel.CerIpAddressModel.CerIpAddresses))}
		stateModel.AddError(&stateMsg)
	}

	//TODO, todo,Manifest's EE certificate has RFC3779 resources that are not marked inherit, in roa_vildate.c P1009
	ValidateEeCertModel(stateModel, &mftModel.EeCertModel)
	ValidateSignerInfoModel(stateModel, &mftModel.SignerInfoModel)

	belogs.Debug("validateMftModel():filePath, fileName,stateModel:",
		mftModel.FilePath, mftModel.FileName, jsonutil.MarshalJson(stateModel))
	return nil
}
