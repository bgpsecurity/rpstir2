package util

import (
	"strings"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	statisticmodel "statistic/model"
)

func RirFileCountDbToRirStatisticModel(rirFileCountDb *statisticmodel.RirFileCountDb, rirStatisticModel *statisticmodel.RirStatisticModel) {
	belogs.Debug("RirFileCountDbToRirStatisticModel(): rirFileCountDb.FileType:", rirFileCountDb.FileType)
	if rirFileCountDb.FileType == "cer" {
		rirFileCountDbToFileCount(rirFileCountDb, &rirStatisticModel.CerFileCount)
	} else if rirFileCountDb.FileType == "crl" {
		rirFileCountDbToFileCount(rirFileCountDb, &rirStatisticModel.CrlFileCount)
	} else if rirFileCountDb.FileType == "mft" {
		rirFileCountDbToFileCount(rirFileCountDb, &rirStatisticModel.MftFileCount)
	} else if rirFileCountDb.FileType == "roa" {
		rirFileCountDbToFileCount(rirFileCountDb, &rirStatisticModel.RoaFileCount)
	}

}

// distribute to each fileType
func RepoFileCountDbToRepoStatisticModel(repoFileCountDb *statisticmodel.RepoFileCountDb, repoStatisticModel *statisticmodel.RepoStatisticModel) {
	if repoFileCountDb.FileType == "cer" {
		repoFileCountDbToFileCount(repoFileCountDb, &repoStatisticModel.CerFileCount)
	} else if repoFileCountDb.FileType == "crl" {
		repoFileCountDbToFileCount(repoFileCountDb, &repoStatisticModel.CrlFileCount)
	} else if repoFileCountDb.FileType == "mft" {
		repoFileCountDbToFileCount(repoFileCountDb, &repoStatisticModel.MftFileCount)
	} else if repoFileCountDb.FileType == "roa" {
		repoFileCountDbToFileCount(repoFileCountDb, &repoStatisticModel.RoaFileCount)
	}
}

func rirFileCountDbToFileCount(rirFileCountDb *statisticmodel.RirFileCountDb, fileCount *statisticmodel.FileCount) {
	belogs.Debug("rirFileCountDbToFileCount(): rirFileCountDb.State:", rirFileCountDb.State)
	if rirFileCountDb.State == "valid" {
		fileCount.ValidCount = rirFileCountDb.Count
	} else if rirFileCountDb.State == "invalid" {
		fileCount.InvalidCount = rirFileCountDb.Count
	} else if rirFileCountDb.State == "warning" {
		fileCount.WarningCount = rirFileCountDb.Count
	}
	return

}
func repoFileCountDbToFileCount(repoFileCountDb *statisticmodel.RepoFileCountDb, fileCount *statisticmodel.FileCount) {
	belogs.Debug("repoFileCountDbToFileCount(): repoFileCountDb.State:", repoFileCountDb.State)
	if repoFileCountDb.State == "valid" {
		fileCount.ValidCount = repoFileCountDb.Count
	} else if repoFileCountDb.State == "invalid" {
		fileCount.InvalidCount = repoFileCountDb.Count
	} else if repoFileCountDb.State == "warning" {
		fileCount.WarningCount = repoFileCountDb.Count
	}
	return

}

func GetUrlByFilePathName(filePath, fileName string) (url string) {
	file := filePath + fileName
	url = strings.Replace(file, conf.VariableString("rsync::destpath")+"/", "rsync://", -1)
	url = strings.Replace(url, conf.VariableString("rrdp::destpath")+"/", "rsync://", -1)
	return url
}

func GetFailDetailsByState(state string) (failDetails []string) {
	stateModel := model.StateModel{}
	jsonutil.UnmarshalJson(state, &stateModel)
	for i, _ := range stateModel.Errors {
		failDetails = append(failDetails, stateModel.Errors[i].Fail)
	}
	for i, _ := range stateModel.Warnings {
		failDetails = append(failDetails, stateModel.Warnings[i].Fail)
	}
	return
}
