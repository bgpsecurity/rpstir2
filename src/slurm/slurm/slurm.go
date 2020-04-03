package slurm

import (
	"bytes"
	"errors"
	"io/ioutil"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	db "slurm/db"
)

func UploadFiles(receiveFiles map[string]string) error {

	//defer httpserver.RemoveReceiveFiles(receiveFiles)
	belogs.Info("UploadFiles(): ReceiveFiles: ", receiveFiles)
	for _, item := range receiveFiles {
		return uploadFile(item)
	}
	return errors.New("UploadFiles(): receiveFiles is empty")

}

// upload file to parse
func uploadFile(slurmFile string) (err error) {
	belogs.Debug("uploadFile(): slurmFile:", slurmFile)

	slurm, err := ParseSlurm(slurmFile)
	if err != nil {
		return err
	}

	err = CheckSlurm(&slurm)
	if err != nil {
		return err
	}

	err = db.SaveSlurm(&slurm, slurmFile)
	if err != nil {
		return err
	}

	return nil
}

func ParseSlurm(slurmFile string) (slurm model.Slurm, err error) {

	// load slurm json file
	f, err := ioutil.ReadFile(slurmFile)
	if err != nil {
		belogs.Error("ParseSlurm():ReadFile fail: ", slurmFile, err)
		return slurm, err
	}

	err = jsonutil.UnmarshalJson(string(f), &slurm)
	if err != nil {
		belogs.Error("ParseSlurm():Unmarshal json fail: ", slurmFile, err)
		return slurm, err
	}

	return slurm, nil
}

func CheckSlurm(slurm *model.Slurm) (err error) {
	var errMsg bytes.Buffer
	// check slurm file is right ?
	if slurm.SlurmVersion != 1 {
		belogs.Debug("CheckSlurm():slurm version is not 1 ")
		errMsg.WriteString("slurm version is not 1; ")
	}

	prefixAndAsn := PrefixAndAsn{}
	//check filter list
	if len(slurm.ValidationOutputFilters.PrefixFilters) > 0 ||
		len(slurm.ValidationOutputFilters.BgpsecFilters) > 0 {
		var prefixFilters = slurm.ValidationOutputFilters.PrefixFilters
		var bgpsecFilters = slurm.ValidationOutputFilters.BgpsecFilters

		// check prefixfilters
		if len(prefixFilters) > 0 {
			for i := 0; i < len(prefixFilters); i++ {
				if prefixAndAsn, err = CheckPrefixAsnAndGetPrefixLength(prefixFilters[i].Prefix, prefixFilters[i].Asn,
					0); err != nil {
					errMsg.WriteString(err.Error() + ";")
					belogs.Debug("CheckSlurm():check prefixFilter", prefixFilters[i], ", found err ", errMsg)
				}
				belogs.Debug("CheckSlurm():from prefixFilter: get prefixAndAsn", prefixAndAsn)

				prefixFilters[i].FormatPrefix = prefixAndAsn.FormatPrefix
				prefixFilters[i].PrefixLength = prefixAndAsn.PrefixLength
				prefixFilters[i].MaxPrefixLength = prefixAndAsn.MaxPrefixLength
				belogs.Debug("CheckSlurm():prefixFilter:", prefixFilters[i])

				// check in lab_rpki_slurm,
				// should no conflict: same asn/prefix, show as filter in slurm file, but as assertion in db
				// if same asn/prefix, show as filter in slurm file and alsow as filter in db , then no conflict, just ignore
				if err = db.CheckConflictInDb("prefixFilters", prefixFilters[i].Asn,
					prefixFilters[i].Prefix, prefixFilters[i].MaxPrefixLength); err != nil {
					errMsg.WriteString(err.Error() + ";")
					belogs.Debug("CheckSlurm(): CheckConflictInDb prefixFilter in db, ", prefixFilters[i], ", found err ", errMsg)
				}

			}
			belogs.Debug("CheckSlurm():prefixFilters: ", prefixFilters)
			slurm.ValidationOutputFilters.PrefixFilters = prefixFilters
		}
		// bgpsec is not realized in rpstir,so bgpsecFilter will be ignored in slurm now
		if len(bgpsecFilters) > 0 {

		}

	}
	//check assertions
	if len(slurm.LocallyAddedAssertions.PrefixAssertions) > 0 ||
		len(slurm.LocallyAddedAssertions.BgpsecAssertions) > 0 {
		var prefixAssertions = slurm.LocallyAddedAssertions.PrefixAssertions
		var bgpsecAssertions = slurm.LocallyAddedAssertions.BgpsecAssertions

		if len(prefixAssertions) > 0 {
			for i := 0; i < len(prefixAssertions); i++ {
				if prefixAndAsn, err = CheckPrefixAsnAndGetPrefixLength(prefixAssertions[i].Prefix, prefixAssertions[i].Asn,
					prefixAssertions[i].MaxPrefixLength); err != nil {
					errMsg.WriteString(err.Error() + ";")
					belogs.Debug("CheckSlurm():check prefixAssertion", prefixAssertions[i], ", found err ", errMsg)
				}
				belogs.Debug("CheckSlurm():from prefixAssertions, get prefixAndAsn:", prefixAndAsn)

				prefixAssertions[i].MaxPrefixLength = prefixAndAsn.MaxPrefixLength
				belogs.Debug("CheckSlurm():prefixAssertion:", prefixAssertions[i])

				if err = db.CheckConflictInDb("prefixAssertions", prefixAssertions[i].Asn,
					prefixAssertions[i].Prefix, prefixAssertions[i].MaxPrefixLength); err != nil {
					errMsg.WriteString(err.Error() + ";")
					belogs.Debug("CheckSlurm(): CheckConflictInDb prefixAssertion in db, ", prefixAssertions[i], ", found err ", errMsg)
				}

			}
			belogs.Debug("CheckSlurm():prefixAssertions: ", prefixAssertions)
			slurm.LocallyAddedAssertions.PrefixAssertions = prefixAssertions
		}
		// bgpsec is not realized in rpstir,so bgpsecAssertions will be ignored in slurm now
		if len(bgpsecAssertions) > 0 {

		}
	}

	// return error
	if errMsg.Len() > 0 {
		belogs.Info("CheckSlurm(): fail:", errMsg.String())
		return errors.New(errMsg.String())
	} else {
		belogs.Info("CheckSlurm():ok :")
		return nil
	}
}
