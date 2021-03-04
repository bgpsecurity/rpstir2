package model

import (
	"container/list"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
)

// queue for rrdp url of notify.xml
type RrdpParseQueue struct {

	//rrdp channel, store will rrdp url and destPath
	RrdpModelChan chan RrdpModelChan

	// parse cer channel, store will parse filepathname
	ParseModelChan chan ParseModelChan

	// rrdping and parsing count, all are zero, will end rrdp
	RrdpingParsingCount int64
	// rrdping count , will decide rrdp wait time
	CurRrdpingCount int64

	// rrdp and parse end channel, to call check whether rrdp is real end ?
	RrdpParseEndChan chan RrdpParseEndChan

	// have added syncurls List
	rrdpAddedUrlsMutex *sync.RWMutex
	rrdpAddedUrls      *list.List

	// other save to synclog,
	LabRpkiSyncLogId uint64
	RrdpResult       model.SyncResult

	// last saved syncRrdpLogs
	LastSyncRrdpLogs map[string]model.LabRpkiSyncRrdpLog
}

func NewQueue() *RrdpParseQueue {
	rq := &RrdpParseQueue{}

	rq.RrdpModelChan = make(chan RrdpModelChan, 100)
	rq.ParseModelChan = make(chan ParseModelChan, 100)
	rq.RrdpParseEndChan = make(chan RrdpParseEndChan, 100)
	rq.RrdpingParsingCount = 0
	rq.CurRrdpingCount = 0

	rq.rrdpAddedUrlsMutex = new(sync.RWMutex)
	rq.rrdpAddedUrls = list.New()

	rq.RrdpResult.StartTime = time.Now()
	rq.RrdpResult.OkUrls = make([]string, 0, 100000)
	rq.RrdpResult.FailUrls = make(map[string]string, 100)
	rq.RrdpResult.FailParseValidateCerts = make(map[string]string, 100)
	belogs.Debug("NewQueue():rq:", jsonutil.MarshalJson(rq))
	return rq
}

func (r *RrdpParseQueue) Close() {
	close(r.RrdpModelChan)
	close(r.ParseModelChan)
	close(r.RrdpParseEndChan)
	r.rrdpAddedUrlsMutex = nil
	r.rrdpAddedUrls = nil
	r.RrdpResult.OkUrls = nil
	r.RrdpResult.FailUrls = nil
	r.RrdpResult.FailParseValidateCerts = nil
	r = nil

}

func (r *RrdpParseQueue) DelRrdpAddedUrl(url string) {
	r.rrdpAddedUrlsMutex.Lock()
	defer r.rrdpAddedUrlsMutex.Unlock()
	if len(url) == 0 {
		belogs.Debug("DelRrdpAddedUrl():url is len:", url)
		return
	}

	e := r.rrdpAddedUrls.Front()
	for e != nil {
		if url == e.Value.(RrdpModelChan).Url {
			belogs.Debug("DelRrdpAddedUrl():have existed, will remove:", url, " in ", e.Value.(RrdpModelChan).Url)
			r.rrdpAddedUrls.Remove(e)
			break
		} else {
			e = e.Next()
		}
	}
}

func (r *RrdpParseQueue) PreCheckRrdpUrl(url string) (ok bool) {
	r.rrdpAddedUrlsMutex.RLock()
	defer r.rrdpAddedUrlsMutex.RUnlock()
	if len(url) == 0 {
		belogs.Error("PreCheckRrdpUrl():url  is 0")
		return false
	}
	if strings.HasPrefix(url, "rrdp://localhost") || strings.HasPrefix(url, "rrdp://127.0.0.1") {
		belogs.Error("PreCheckRrdpUrl():url is localhost:", url)
		return false
	}
	if strings.Index(url, "chloe.sobornost.net/rpki/news-public.xml") > 0 {
		belogs.Error("PreCheckRrdpUrl():this url is not rrdp url, so just ignore:", url)
		return false
	}

	e := r.rrdpAddedUrls.Front()
	for e != nil {
		if strings.Contains(url, e.Value.(RrdpModelChan).Url) {
			belogs.Debug("PreCheckRrdpUrl():have existed:", url, " in ", e.Value.(RrdpModelChan).Url)
			return false
		}
		e = e.Next()
	}
	return true
}

// add resync url
// if have error, should set RrdpingParsingCount-1
func (r *RrdpParseQueue) AddRrdpUrl(url string, dest string) {

	r.rrdpAddedUrlsMutex.Lock()
	defer r.rrdpAddedUrlsMutex.Unlock()
	defer func() {
		belogs.Debug("AddRrdpUrl():defer rpQueue.RrdpingParsingCount:", atomic.LoadInt64(&r.RrdpingParsingCount))
		if atomic.LoadInt64(&r.RrdpingParsingCount) == 0 {
			r.RrdpParseEndChan <- RrdpParseEndChan{}
		}
	}()
	belogs.Debug("AddRrdpUrl():url:", url, "    dest:", dest)
	if len(url) == 0 || len(dest) == 0 {
		belogs.Error("AddRrdpUrl():len(url) == 0 || len(dest) == 0, before RrdpingParsingCount-1:", atomic.LoadInt64(&r.RrdpingParsingCount))
		atomic.AddInt64(&r.RrdpingParsingCount, -1)
		belogs.Debug("AddRrdpUrl():len(url) == 0 || len(dest) == 0, after RrdpingParsingCount-1:", atomic.LoadInt64(&r.RrdpingParsingCount))
		return
	}
	if strings.HasPrefix(url, "https://localhost") || strings.HasPrefix(url, "https://127.0.0.1") {
		belogs.Error("AddRrdpUrl():url is localhost:", url)
		belogs.Debug("AddRrdpUrl():url is localhost, before RrdpingParsingCount-1:", atomic.LoadInt64(&r.RrdpingParsingCount))
		atomic.AddInt64(&r.RrdpingParsingCount, -1)
		belogs.Debug("AddRrdpUrl()::url is localhost, after RrdpingParsingCount-1:", atomic.LoadInt64(&r.RrdpingParsingCount))
		return
	}
	e := r.rrdpAddedUrls.Front()
	for e != nil {
		if strings.Contains(url, e.Value.(RrdpModelChan).Url) {
			belogs.Debug("AddRrdpUrl():have existed:", url, " in ", e.Value.(RrdpModelChan).Url,
				"   len(r.RrdpModelChan):", len(r.RrdpModelChan))
			belogs.Debug("AddRrdpUrl():have existed, before RrdpingParsingCount-1:", atomic.LoadInt64(&r.RrdpingParsingCount))
			atomic.AddInt64(&r.RrdpingParsingCount, -1)
			belogs.Debug("AddRrdpUrl():have existed, after RrdpingParsingCount-1:", atomic.LoadInt64(&r.RrdpingParsingCount))
			return
		}
		e = e.Next()
	}

	rrdpModelChan := RrdpModelChan{Url: url, Dest: dest}
	e = r.rrdpAddedUrls.PushBack(rrdpModelChan)
	belogs.Debug("AddRrdpUrl():will send to rrdpModelChan:", rrdpModelChan,
		"   len(r.RrdpModelChan):", len(r.RrdpModelChan), "   len(rrdpAddedUrls):", r.rrdpAddedUrls.Len())
	r.RrdpModelChan <- rrdpModelChan
	belogs.Debug("AddRrdpUrl():after send to rrdpModelChan:", rrdpModelChan,
		"   len(r.RrdpModelChan):", len(r.RrdpModelChan), "   len(rrdpAddedUrls):", r.rrdpAddedUrls.Len())
	return
}

func (r *RrdpParseQueue) GetRrdpUrls() (urls []string) {
	r.rrdpAddedUrlsMutex.Lock()
	defer r.rrdpAddedUrlsMutex.Unlock()
	urls = make([]string, 0, r.rrdpAddedUrls.Len())
	belogs.Debug("GetRrdpUrls():r.rrdpAddedUrls.Len():", r.rrdpAddedUrls.Len())
	for e := r.rrdpAddedUrls.Front(); e != nil; e = e.Next() {
		urls = append(urls, e.Value.(RrdpModelChan).Url)
	}
	belogs.Debug("GetRrdpUrls():urls:", urls)
	return urls
}
