package sys

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

// exportChainRepos 导出资料库列表
//
//	@return chainReposs 返回资料库列表
//	@return err 返回错误
func exportChainRepos() (chainReposs []ChainRepos, err error) {
	start := time.Now()
	chainCertIds, certIdRepos, err := getChainCertsDb()
	if err != nil {
		belogs.Error("exportChainRepos(): getChainCertsDb fail:", err)
		return nil, err
	}
	belogs.Info("exportChainRepos(): len(chainCertIds):", len(chainCertIds), "  len(certIdRepos):", len(certIdRepos), "   time(s):", time.Since(start))

	chainReposs, err = mergeCertIdAndRepo(chainCertIds, certIdRepos)
	if err != nil {
		belogs.Error("exportChainRepos(): mergeCertIdAndRepo fail:", err)
		return nil, err
	}
	return chainReposs, nil

}

func mergeCertIdAndRepo(chainCertIds []ChainCertId, certIdRepos []CertIdRepo) (chainReposs []ChainRepos, err error) {
	belogs.Debug("mergeCertIdAndRepo(): len(chainCertIds):", len(chainCertIds), "   len(certIdRepos):", len(certIdRepos))
	certIdReposMap := make(map[string]string)
	for i := range certIdRepos {
		if len(certIdRepos[i].Repo) == 0 {
			continue
		}
		key := convert.ToString(certIdRepos[i].Id) + "_" + certIdRepos[i].FileType
		certIdReposMap[key] = certIdRepos[i].Repo
	}
	belogs.Debug("mergeCertIdAndRepo(): len(certIdReposMap):", len(certIdReposMap))

	chainReposs = make([]ChainRepos, 0)
	reposMaps := make(map[string]string)
	for i := range chainCertIds {
		if len(chainCertIds[i].ParentCers) == 0 {
			continue
		}
		repos := make([]string, 0)

		certKey := convert.ToString(chainCertIds[i].Id) + "_" + chainCertIds[i].FileType
		certRepo, ok := certIdReposMap[certKey]
		if !ok {
			belogs.Error("mergeCertIdAndRepo(): not found in certIdReposMap, certKey:", certKey)
			continue
		}
		belogs.Debug("mergeCertIdAndRepo(): certKey:", certKey, "  certRepo:", certRepo)
		repos = append(repos, certRepo)
		cetReposMap := make(map[string]string, 0)
		cetReposMap[certRepo] = certRepo

		parentCerIds := make([]ParentCerId, 0)
		err = jsonutil.UnmarshalJson(chainCertIds[i].ParentCers, &parentCerIds)
		if err != nil {
			belogs.Error("mergeCertIdAndRepo(): UnMarshalJson fail, ParentCers:", chainCertIds[i].ParentCers)
			continue
		}

		for j := range parentCerIds {
			cerKey := convert.ToString(parentCerIds[j].Id) + "_cer"
			cerRepo, ok := certIdReposMap[cerKey]
			if !ok {
				belogs.Error("mergeCertIdAndRepo(): not found in parentCerIds, cerKey:", cerKey)
				continue
			}
			if _, ok := cetReposMap[cerRepo]; ok {
				continue
			}
			cetReposMap[cerRepo] = cerRepo
			belogs.Debug("mergeCertIdAndRepo(): add cerRepo:", cerRepo)
			repos = append(repos, cerRepo)
		}
		belogs.Debug("mergeCertIdAndRepo(): repos:", jsonutil.MarshalJson(repos),
			"  certKey:", certKey, "   certRepo:", certRepo,
			"  parentCerIds:", parentCerIds, " repos(include certsRepo):", repos)

		reposJson := jsonutil.MarshalJson(repos)
		if _, ok := reposMaps[reposJson]; ok {
			continue
		}
		reposMaps[reposJson] = reposJson

		//sort.Reverse(sort.StringSlice(repos))
		for m, n := 0, len(repos)-1; m < n; m, n = m+1, n-1 {
			repos[m], repos[n] = repos[n], repos[m]
		}
		chainRepos := ChainRepos{Repos: repos}
		belogs.Debug("mergeCertIdAndRepo(): chainRepos:", jsonutil.MarshalJson(chainRepos))
		chainReposs = append(chainReposs, chainRepos)
	}
	belogs.Debug("mergeCertIdAndRepo(): chainReposs:", jsonutil.MarshalJson(chainReposs))
	return chainReposs, nil
}
