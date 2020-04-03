package statistic

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	db "statistic/db"
	statisticmodel "statistic/model"
	util "statistic/util"
)

func Start() {
	start := time.Now()
	belogs.Info("Start():statistic start")

	// get rir
	rirs, err := GetRirStatisticModels()
	if err != nil {
		belogs.Error("Start(): GetRirStatisticModels fail:", len(rirs), err, "  time(s):", time.Now().Sub(start).Seconds())
		return
	}

	// save to db
	err = db.UpdateStatistic(rirs)
	if err != nil {
		belogs.Error("Start(): UpdateStatistic fail:", len(rirs), err, "  time(s):", time.Now().Sub(start).Seconds())
		return
	}
	belogs.Debug("Start(): end  time(s):", time.Now().Sub(start).Seconds())

}

func GetRirStatisticModels() (rirs []statisticmodel.RirStatisticModel, err error) {
	start := time.Now()
	// init 5 rir
	rirs = make([]statisticmodel.RirStatisticModel, 5)
	rirs[0] = statisticmodel.RirStatisticModel{Rir: model.ORIGIN_RIR_AFRINIC}
	rirs[1] = statisticmodel.RirStatisticModel{Rir: model.ORIGIN_RIR_APNIC}
	rirs[2] = statisticmodel.RirStatisticModel{Rir: model.ORIGIN_RIR_ARIN}
	rirs[3] = statisticmodel.RirStatisticModel{Rir: model.ORIGIN_RIR_LACNIC}
	rirs[4] = statisticmodel.RirStatisticModel{Rir: model.ORIGIN_RIR_RIPE_NCC}

	// get all count
	rirFileCountDbs, err := db.GetRirFileCountDb()
	if err != nil {
		belogs.Error("GetRirStatisticModels(): GetFileCountDb fail:", err)
		return nil, err
	}

	// set count to each rir
	for rirFileIndex, _ := range rirFileCountDbs {
		for i, _ := range rirs {
			if rirFileCountDbs[rirFileIndex].Rir == rirs[i].Rir {
				util.RirFileCountDbToRirStatisticModel(&rirFileCountDbs[rirFileIndex], &rirs[i])
				belogs.Debug("GetRirStatisticModels(): RirFileCountDbToRirStatisticModel  rirs[i]:", rirs[i].Rir, jsonutil.MarshalJson(rirs[i]))
				break
			}
		}
	}
	// get repo count(in each rir)
	for i, _ := range rirs {
		err = getRepoByRir(rirs[i].Rir, &rirs[i])
		if err != nil {
			belogs.Error("GetRirStatisticModels(): GetFileCountDb fail:", err)
			return nil, err
		}
		belogs.Debug("GetRirStatisticModels(): GetRepoByRir  rirs[i]:", rirs[i].Rir, jsonutil.MarshalJson(rirs[i]))
	}

	// get sync info
	syncModel, err := db.GetLatestSync()
	if err != nil {
		belogs.Error("GetRirStatisticModels(): GetLatestSync fail:", err)
		return nil, err
	}
	belogs.Debug("GetRirStatisticModels(): syncModel:", jsonutil.MarshalJson(syncModel))

	// set sync to each rir
	for i, _ := range rirs {
		rirs[i].SyncModel = syncModel
	}
	belogs.Debug("GetRirStatisticModels(): end time(s):", time.Now().Sub(start).Seconds())
	return rirs, nil
}

// get repo in each rir
func getRepoByRir(rir string, rirStatisticModel *statisticmodel.RirStatisticModel) (err error) {
	start := time.Now()
	belogs.Debug("getRepoByRir(): rir:", rir)

	// get repo by rir
	repoFileCountDbs, err := db.GetRepoFileCountDb(rir)
	if err != nil {
		belogs.Error("getRepoByRir(): GetRepoFileCountDb fail:", err)
		return err
	}
	belogs.Debug("getRepoByRir(): rir,len(repoFileCountDbs):", rir, len(repoFileCountDbs))

	// map to get each repo file count . must be map[string]*RepoStatisticModel
	repoMap := make(map[string]*statisticmodel.RepoStatisticModel, 0)
	for repoIndex, _ := range repoFileCountDbs {
		repo := repoFileCountDbs[repoIndex].Repo
		repoStatisModel, ok := repoMap[repo]
		if ok {

		} else {
			repoStatisModel = new(statisticmodel.RepoStatisticModel)
			repoStatisModel.Rir = rir
			repoStatisModel.Repo = repo
			repoMap[repo] = repoStatisModel
		}
		util.RepoFileCountDbToRepoStatisticModel(&repoFileCountDbs[repoIndex], repoStatisModel)

	}
	belogs.Debug("getRepoByRir(): repoMap:", jsonutil.MarshalJson(repoMap))

	// map to get each repo state
	for _, repoStatisticModel := range repoMap {
		fileStates, err := db.GetFileState(repoStatisticModel.Rir, repoStatisticModel.Repo)
		if err != nil {
			belogs.Error("GetRepoByRir(): GetFileState fail:", err)
			return err
		}
		repoStatisticModel.FileStates = fileStates
		rirStatisticModel.RepoStatisticModels = append(rirStatisticModel.RepoStatisticModels, *repoStatisticModel)
	}
	belogs.Debug("getRepoByRir(): end  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
