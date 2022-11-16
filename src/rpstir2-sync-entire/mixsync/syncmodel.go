package mixsync

import (
	"container/list"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
)

// queue for rrdp url of notify.xml
type SyncParseQueue struct {

	//rrdp channel, store will rrdp url and destPath
	SyncChan chan SyncChan

	// parse cer channel, store will parse filepathname
	ParseChan chan ParseChan

	// syncing and parsing count, all are zero, will end rrdp
	SyncingAndParsingCount int64
	// syncing count , will decide rrdp wait time
	SyncingCount int64

	// rrdp and parse end channel, to call check whether rrdp is real end ?
	SyncAndParseEndChan chan SyncAndParseEndChan

	// have added syncurls List
	syncUrlsMutex *sync.RWMutex
	syncUrls      *list.List

	// other save to synclog,
	LabRpkiSyncLogId uint64

	// cannot use map, will cause panic:"fatal error: concurrent map iteration and map write"
	// jsonutil.MarshalJson(rrQueue.SyncResult)
	SyncResult model.SyncResult

	// last saved syncRrdpLogs
	LastSyncRrdpLogs map[string]model.LabRpkiSyncRrdpLog
}

func NewSyncParseQueue() *SyncParseQueue {
	spq := &SyncParseQueue{}

	spq.SyncChan = make(chan SyncChan, 100)
	spq.ParseChan = make(chan ParseChan, 100)
	spq.SyncAndParseEndChan = make(chan SyncAndParseEndChan, 100)
	spq.SyncingAndParsingCount = 0
	spq.SyncingCount = 0

	spq.syncUrlsMutex = new(sync.RWMutex)
	spq.syncUrls = list.New()

	spq.SyncResult.StartTime = time.Now()
	spq.SyncResult.OkUrls = make([]string, 0, 100000)
	spq.SyncResult.FailUrls = sync.Map{}
	spq.SyncResult.FailParseValidateCerts = sync.Map{}
	belogs.Debug("NewQueue():spq:", jsonutil.MarshalJson(spq))
	return spq
}

func (r *SyncParseQueue) Close() {
	close(r.SyncChan)
	close(r.ParseChan)
	close(r.SyncAndParseEndChan)
	r.syncUrlsMutex = nil
	r.syncUrls = nil
	r.SyncResult.OkUrls = nil
	r.SyncResult.FailUrls = sync.Map{}
	r.SyncResult.FailParseValidateCerts = sync.Map{}
	r = nil

}
func (r *SyncParseQueue) IsClose() bool {
	return r == nil || r.syncUrlsMutex == nil
}

func (r *SyncParseQueue) PreCheckSyncUrl(url string) (ok bool) {
	r.syncUrlsMutex.RLock()
	defer r.syncUrlsMutex.RUnlock()
	if len(url) == 0 {
		belogs.Error("PreCheckSyncUrl():url  is 0")
		return false
	}
	if strings.HasPrefix(url, "rrdp://localhost") || strings.HasPrefix(url, "rrdp://127.0.0.1") ||
		strings.HasPrefix(url, "https://localhost") || strings.HasPrefix(url, "https://127.0.0.1") ||
		strings.HasPrefix(url, "http://localhost") || strings.HasPrefix(url, "http://127.0.0.1") ||
		strings.HasPrefix(url, "rsync://localhost") || strings.HasPrefix(url, "rsync://127.0.0.1") {
		belogs.Error("PreCheckSyncUrl():url is localhost:", url)
		return false
	}

	if strings.Index(url, "ca.rg.net/rrdp/notify.xml") > 0 {
		belogs.Error("PreCheckSyncUrl():this url is not rrdp url, so just ignore:", url)
		return false
	}

	repoNum := uint64(0)
	e := r.syncUrls.Front()
	for e != nil {
		if strings.Contains(url, e.Value.(SyncChan).Url) {
			belogs.Debug("PreCheckSyncUrl():have existed:", url, " in ", e.Value.(SyncChan).Url)
			return false
		}
		e = e.Next()
		repoNum++
	}
	limitOfRepoNum := uint64(conf.Int("sync::limitOfRepoNum"))
	if limitOfRepoNum > 0 && repoNum > limitOfRepoNum {
		belogs.Error("PreCheckSyncUrl():repoNum is more than limit, repoNum is ", repoNum, "   limitOfRepoNum is ", limitOfRepoNum)
		return false
	}
	return true
}

// add resync url
// if have error, should set SyncingAndParsingCount-1
func (r *SyncParseQueue) AddSyncUrl(url string, dest string) {

	r.syncUrlsMutex.Lock()
	defer r.syncUrlsMutex.Unlock()
	defer func() {
		belogs.Debug("AddSyncUrl():defer rpQueue.SyncingAndParsingCount:", atomic.LoadInt64(&r.SyncingAndParsingCount))
		if atomic.LoadInt64(&r.SyncingAndParsingCount) == 0 {
			r.SyncAndParseEndChan <- SyncAndParseEndChan{}
		}
	}()
	belogs.Debug("AddSyncUrl():url:", url, "    dest:", dest)
	if len(url) == 0 || len(dest) == 0 {
		belogs.Error("AddSyncUrl():len(url) == 0 || len(dest) == 0, before SyncingAndParsingCount-1:", atomic.LoadInt64(&r.SyncingAndParsingCount))
		atomic.AddInt64(&r.SyncingAndParsingCount, -1)
		belogs.Debug("AddSyncUrl():len(url) == 0 || len(dest) == 0, after SyncingAndParsingCount-1:", atomic.LoadInt64(&r.SyncingAndParsingCount))
		return
	}

	e := r.syncUrls.Front()
	for e != nil {
		if strings.Contains(url, e.Value.(SyncChan).Url) {
			belogs.Debug("AddSyncUrl():have existed:", url, " in ", e.Value.(SyncChan).Url,
				"   len(r.SyncChan):", len(r.SyncChan))
			belogs.Debug("AddSyncUrl():have existed, before SyncingAndParsingCount-1:", atomic.LoadInt64(&r.SyncingAndParsingCount))
			atomic.AddInt64(&r.SyncingAndParsingCount, -1)
			belogs.Debug("AddSyncUrl():have existed, after SyncingAndParsingCount-1:", atomic.LoadInt64(&r.SyncingAndParsingCount))
			return
		}
		e = e.Next()
	}

	syncChan := SyncChan{Url: url, Dest: dest}
	e = r.syncUrls.PushBack(syncChan)
	belogs.Info("AddSyncUrl():will send to syncChan:", syncChan,
		"   len(syncUrls):", r.syncUrls.Len())
	r.SyncChan <- syncChan
	belogs.Debug("AddSyncUrl():after send to syncChan:", syncChan,
		"    syncUrls:", r.syncUrls)
	return
}

func (r *SyncParseQueue) GetSyncUrls() (urls []string) {
	r.syncUrlsMutex.Lock()
	defer r.syncUrlsMutex.Unlock()
	urls = make([]string, 0, r.syncUrls.Len())
	belogs.Debug("GetSyncUrls():r.syncUrls.Len():", r.syncUrls.Len())
	for e := r.syncUrls.Front(); e != nil; e = e.Next() {
		urls = append(urls, e.Value.(SyncChan).Url)
	}
	belogs.Debug("GetSyncUrls():urls:", urls)
	return urls
}

// rrdp channel
type SyncChan struct {
	Url  string `json:"url"`
	Dest string `jsong:"dest"`
}

// parse channel
type ParseChan struct {
	Url           string   `json:"url"`
	FilePathNames []string `json:"filePathNames"`
}

// rrdp and parse end channel, may be end
type SyncAndParseEndChan struct {
}

type SyncState struct {
	SyncStyle string `json:"syncStyle"`

	StartTime time.Time `json:"startTime,omitempty"`
	EndTime   time.Time `json:"endTime,omitempty"`

	SyncUrls   []string         `json:"syncUrls"`
	SyncResult model.SyncResult `json:"syncResult"`
}
