package model

import (
	"container/list"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	belogs "github.com/astaxie/beego/logs"

	"model"
)

// queue for rsync url
type RsyncParseQueue struct {

	//rsync channel, store will rsync url and destPath
	RsyncModelChan chan RsyncModelChan

	// parse cer channel, store will parse filepathname
	ParseModelChan chan ParseModelChan

	// rsyncing and parsing count, all are zero, will end rsync
	RsyncingParsingCount int64
	// rsyncing count , will decide rsync wait time
	CurRsyncingCount int64

	// rsync and parse end channel, to call check whether rsync is real end ?
	RsyncParseEndChan chan RsyncParseEndChan

	// have added syncurls List
	rsyncAddedUrlsMutex *sync.RWMutex
	rsyncAddedUrls      *list.List

	// other save to synclog,
	LabRpkiSyncLogId uint64
	RsyncResult      model.SyncResult
}

func NewQueue() *RsyncParseQueue {
	rq := &RsyncParseQueue{}

	rq.RsyncModelChan = make(chan RsyncModelChan, 100)
	rq.ParseModelChan = make(chan ParseModelChan, 100)
	rq.RsyncParseEndChan = make(chan RsyncParseEndChan, 100)
	rq.RsyncingParsingCount = 0
	rq.CurRsyncingCount = 0

	rq.rsyncAddedUrlsMutex = new(sync.RWMutex)
	rq.rsyncAddedUrls = list.New()

	rq.RsyncResult.StartTime = time.Now()
	rq.RsyncResult.OkUrls = make([]string, 0, 100000)
	rq.RsyncResult.FailUrls = make(map[string]string, 100)
	rq.RsyncResult.FailParseValidateCerts = make(map[string]string, 100)
	return rq
}

func (r *RsyncParseQueue) Close() {
	close(r.RsyncModelChan)
	close(r.ParseModelChan)
	close(r.RsyncParseEndChan)
	r.rsyncAddedUrlsMutex = nil
	r.rsyncAddedUrls = nil
	r.RsyncResult.OkUrls = nil
	r.RsyncResult.FailUrls = nil
	r.RsyncResult.FailParseValidateCerts = nil
	r = nil

}

func (r *RsyncParseQueue) DelRsyncAddedUrl(url string) {
	r.rsyncAddedUrlsMutex.Lock()
	defer r.rsyncAddedUrlsMutex.Unlock()
	if len(url) == 0 {
		belogs.Debug("DelRsyncAddedUrl():url is len:", url)
		return
	}

	e := r.rsyncAddedUrls.Front()
	for e != nil {
		if url == e.Value.(RsyncModelChan).Url {
			belogs.Debug("DelRsyncAddedUrl():have existed, will remove:", url, " in ", e.Value.(RsyncModelChan).Url)
			r.rsyncAddedUrls.Remove(e)
			break
		} else {
			e = e.Next()
		}
	}
}

func (r *RsyncParseQueue) PreCheckRsyncUrl(url string) (ok bool) {
	r.rsyncAddedUrlsMutex.RLock()
	defer r.rsyncAddedUrlsMutex.RUnlock()
	if len(url) == 0 {
		belogs.Error("PreCheckRsyncUrl():url  is 0")
		return false
	}
	if strings.HasPrefix(url, "rsync://localhost") || strings.HasPrefix(url, "rsync://127.0.0.1") {
		belogs.Error("PreCheckRsyncUrl():url is localhost:", url)
		return false
	}
	e := r.rsyncAddedUrls.Front()
	for e != nil {
		if strings.Contains(url, e.Value.(RsyncModelChan).Url) {
			belogs.Debug("PreCheckRsyncUrl():have existed:", url, " in ", e.Value.(RsyncModelChan).Url)
			return false
		}
		e = e.Next()
	}
	return true
}

// add resync url
// if have error, should set RsyncingParsingCount-1
func (r *RsyncParseQueue) AddRsyncUrl(url string, dest string) {

	r.rsyncAddedUrlsMutex.Lock()
	defer r.rsyncAddedUrlsMutex.Unlock()
	defer func() {
		belogs.Debug("AddRsyncUrl():defer rpQueue.RsyncingParsingCount:", atomic.LoadInt64(&r.RsyncingParsingCount))
		if atomic.LoadInt64(&r.RsyncingParsingCount) == 0 {
			r.RsyncParseEndChan <- RsyncParseEndChan{}
		}
	}()
	belogs.Debug("AddRsyncUrl():url:", url, "    dest:", dest)
	if len(url) == 0 || len(dest) == 0 {
		belogs.Error("AddRsyncUrl():len(url) == 0 || len(dest) == 0, before RsyncingParsingCount-1:", atomic.LoadInt64(&r.RsyncingParsingCount))
		atomic.AddInt64(&r.RsyncingParsingCount, -1)
		belogs.Debug("AddRsyncUrl():len(url) == 0 || len(dest) == 0, after RsyncingParsingCount-1:", atomic.LoadInt64(&r.RsyncingParsingCount))
		return
	}
	if strings.HasPrefix(url, "rsync://localhost") || strings.HasPrefix(url, "rsync://127.0.0.1") {
		belogs.Error("AddRsyncUrl():url is localhost:", url)
		belogs.Debug("AddRsyncUrl():url is localhost, before RsyncingParsingCount-1:", atomic.LoadInt64(&r.RsyncingParsingCount))
		atomic.AddInt64(&r.RsyncingParsingCount, -1)
		belogs.Debug("AddRsyncUrl()::url is localhost, after RsyncingParsingCount-1:", atomic.LoadInt64(&r.RsyncingParsingCount))
		return
	}
	e := r.rsyncAddedUrls.Front()
	for e != nil {
		if strings.Contains(url, e.Value.(RsyncModelChan).Url) {
			belogs.Debug("AddRsyncUrl():have existed:", url, " in ", e.Value.(RsyncModelChan).Url,
				"   len(r.RsyncModelChan):", len(r.RsyncModelChan))
			belogs.Debug("AddRsyncUrl():have existed, before RsyncingParsingCount-1:", atomic.LoadInt64(&r.RsyncingParsingCount))
			atomic.AddInt64(&r.RsyncingParsingCount, -1)
			belogs.Debug("AddRsyncUrl():have existed, after RsyncingParsingCount-1:", atomic.LoadInt64(&r.RsyncingParsingCount))
			return
		}
		e = e.Next()
	}

	rsyncModelChan := RsyncModelChan{Url: url, Dest: dest}
	e = r.rsyncAddedUrls.PushBack(rsyncModelChan)
	belogs.Debug("AddRsyncUrl():will send to rsyncModelChan:", rsyncModelChan,
		"   len(r.RsyncModelChan):", len(r.RsyncModelChan), "   len(rsyncAddedUrls):", r.rsyncAddedUrls.Len())
	r.RsyncModelChan <- rsyncModelChan
	belogs.Debug("AddRsyncUrl():after send to rsyncModelChan:", rsyncModelChan,
		"   len(r.RsyncModelChan):", len(r.RsyncModelChan), "   len(rsyncAddedUrls):", r.rsyncAddedUrls.Len())
	return
}

func (r *RsyncParseQueue) GetRsyncUrls() (urls []string) {
	r.rsyncAddedUrlsMutex.Lock()
	defer r.rsyncAddedUrlsMutex.Unlock()
	urls = make([]string, 0, r.rsyncAddedUrls.Len())
	belogs.Debug("GetRsyncUrls():r.rsyncAddedUrls.Len():", r.rsyncAddedUrls.Len())
	for e := r.rsyncAddedUrls.Front(); e != nil; e = e.Next() {
		urls = append(urls, e.Value.(RsyncModelChan).Url)
	}
	belogs.Debug("GetRsyncUrls():urls:", urls)
	return urls
}
