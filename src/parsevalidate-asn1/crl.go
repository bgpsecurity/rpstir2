package parsevalidateasn1

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/asn1util"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

func parseClrAkiByAsn1(data []byte) (string, error) {
	belogs.Debug("parseClrAkiByAsn1(): len(data):", len(data))
	raws := make([]asn1.RawValue, 0)
	_, err := asn1.Unmarshal(data, &raws)
	if err != nil {
		belogs.Error("parseClrAkiByAsn1(): Unmarshal raws fail,len(data):", len(data), err)
		return "", err
	}

	if len(raws) > 0 {
		return hex.EncodeToString(raws[0].Bytes), nil
	} else {
		return "", errors.New("it is no sequence of []byte")
	}
}

func ParseCrlModelByAsn1(fileModel *model.FileModel, crlModel *model.CrlModel, stateModel *model.StateModel) (err error) {
	start := time.Now()
	belogs.Debug("ParseCrlModelByAsn1(): fileModel:", jsonutil.MarshalJson(fileModel))

	fileByte, err := os.ReadFile(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("ParseCrlModelByAsn1(): ReadFile fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to read file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	crl, err := x509.ParseCRL(fileByte)
	if err != nil {
		belogs.Error("ParseCrlModelByAsn1(): ParseCRL fail, len(fileByte):", len(fileByte),
			"   fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to parse file by asn1",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}

	tbsCertList := crl.TBSCertList
	crlModel.Version = tbsCertList.Version
	crlModel.IssuerAll, _ = asn1util.GetDNFromRDNSeq(tbsCertList.Issuer, ",")
	crlModel.ThisUpdate = tbsCertList.ThisUpdate.Local()
	crlModel.NextUpdate = tbsCertList.NextUpdate.Local()
	crlModel.HasExpired = strconv.FormatBool(crl.HasExpired(time.Now()))
	if crl.SignatureAlgorithm.Algorithm.String() == "1.2.840.113549.1.1.11" {
		crlModel.CertAlgorithm = "sha256WithRSAEncryption"
	}
	if tbsCertList.Signature.Algorithm.String() == "1.2.840.113549.1.1.11" {
		crlModel.TbsAlgorithm = "sha256WithRSAEncryption"
	}
	//exts := tbsCertList.Extensions
	crlModel.RevokedCertModels = make([]model.RevokedCertModel, 0)
	revokedCerts := tbsCertList.RevokedCertificates
	belogs.Debug("ParseCrlModelByAsn1(): get crl, ThisUpdate:", crlModel.ThisUpdate, "  NextUpdate:", crlModel.NextUpdate,
		"   len(revokedCerts):", len(revokedCerts))
	for _, revokedCert := range revokedCerts {
		revokedCertModel := model.RevokedCertModel{}
		revokedCertModel.Sn = fmt.Sprintf("%x", revokedCert.SerialNumber)
		revokedCertModel.RevocationTime = revokedCert.RevocationTime.Local()
		crlModel.RevokedCertModels = append(crlModel.RevokedCertModels, revokedCertModel)
	}
	belogs.Debug("ParseCrlModelByAsn1():crlModel:", crlModel.String())

	belogs.Debug("ParseCrlModelByAsn1():len(Extensions):", len(crl.TBSCertList.Extensions))
	for i := range crl.TBSCertList.Extensions {
		extension := &crl.TBSCertList.Extensions[i]
		belogs.Debug("ParseCrlModelByAsn1(): extension.Id.String():", extension.Id.String())
		if extension.Id.String() == "2.5.29.35" {
			// authorityKeyIdentifier
			crlModel.Aki, err = parseClrAkiByAsn1(extension.Value)
			if err != nil {
				belogs.Error("ParseCrlModelByAsn1(): parseClrAkiByAsn1 fail, len(extension.Value):", len(extension.Value),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse aki file",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}

		} else if extension.Id.String() == "2.5.29.20" {
			// clrNumber
			var crlNumber *big.Int
			_, err = asn1.Unmarshal(extension.Value, &crlNumber)
			if err != nil {
				belogs.Error("ParseCrlModelByAsn1(): Unmarshal crlNumber fail, len(extension.Value):", len(extension.Value),
					"   fileModel:", jsonutil.MarshalJson(fileModel), err)
				stateMsg := model.StateMsg{Stage: "parsevalidate",
					Fail:   "Fail to parse crlNumber file",
					Detail: err.Error()}
				stateModel.AddError(&stateMsg)
				continue
			}
			belogs.Debug("ParseCrlModelByAsn1(): crlNumber:", crlNumber)
			crlModel.CrlNumber = crlNumber.Uint64()
		}
	}
	belogs.Info("ParseCrlModelByAsn1():crlModel:", crlModel.String(), "  time(s):", time.Since(start))
	return nil
}
