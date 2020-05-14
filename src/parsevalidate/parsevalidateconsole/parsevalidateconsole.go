package main

import (
	"container/list"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	belogs "github.com/astaxie/beego/logs"
	_ "github.com/cpusoft/goutil/conf"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	_ "github.com/cpusoft/goutil/logs"
	osutil "github.com/cpusoft/goutil/osutil"

	"model"
	"parsevalidate/parsevalidate"
)

func main() {
	belogs.Info("main(): start ")
	start := time.Now()

	fileList, err := GetAllFile()
	if err != nil {
		belogs.Error("main(): GetAllFile failed:", err)
		return
	}

	cerModelList, crlModelList, mftModelList, roaModelList, errList := ParseCert(fileList)
	if errList.Len() != 0 {
		belogs.Error("main(): ParseCert failed:", jsonutil.MarshalJson(errList))
		//if get err, just continue, not return
		//return
	}

	PrintResult(cerModelList, crlModelList, mftModelList, roaModelList)
	belogs.Info("main():  end,   duration time:", time.Now().Sub(start).String())

}

func GetAllFile() (*list.List, error) {

	var file string
	if len(os.Args) == 2 {
		file = os.Args[1]
	} else {
		//file = `638d3688-a849-3b8b-a63e-cb6b8f7dc77e.roa`
		//file = `9jsAaJV8kBhN8VyoJAOzTiZTxRU.roa`
		file = `./roa/`
	}
	belogs.Notice("GetAllFile():input read file or path :", file)
	file = strings.TrimSpace(file)

	// get all files, add to fileList
	isDir, err := osutil.IsDir(file)
	if err != nil {
		belogs.Error("GetAllFile():IsDir err:", file, err)
		return nil, err
	}
	fileList := list.New()
	if isDir {
		suffixs := make(map[string]string)
		suffixs[".cer"] = ".cer"
		suffixs[".mft"] = ".mft"
		suffixs[".roa"] = ".roa"
		suffixs[".crl"] = ".crl"
		//suffixs[".gbr"] = ".gbr"

		fileList = osutil.GetAllFilesInDirectoryBySuffixs(file, suffixs)
	} else {
		fileList.PushBack(file)
	}
	belogs.Notice("GetAllFile(): files count: ", fileList.Len())
	return fileList, nil
}

func ParseCert(fileList *list.List) (cerModelList *list.List, crlModelList *list.List,
	mftModelList *list.List, roaModelList *list.List, errList *list.List) {

	belogs.Notice("ParseCert():start, fileList.Len:", fileList.Len())

	// set param
	cerModelList = list.New()
	crlModelList = list.New()
	mftModelList = list.New()
	roaModelList = list.New()
	errList = list.New()
	allDuartionTime := float64(0)

	var wg sync.WaitGroup
	var cerlock sync.Mutex
	var crllock sync.Mutex
	var mftlock sync.Mutex
	var roalock sync.Mutex
	var errlock sync.Mutex
	cur := int32(0)

	maxProcs := runtime.NumCPU() // get cpu count
	runtime.GOMAXPROCS(maxProcs) //limit goroutines count

	ch := make(chan int, maxProcs*1)
	// get the ext of file, to process depend on ext

	for e := fileList.Front(); e != nil; e = e.Next() {
		fileName := e.Value.(string)
		ch <- 1
		wg.Add(1)
		atomic.AddInt32(&cur, 1)

		// start
		go func(certFile string, ch chan int) {
			defer wg.Done()
			belogs.Debug("ParseCert():cur:", cur, "    certFile:", certFile)

			startOneParse := time.Now()
			if strings.HasSuffix(certFile, ".cer") {
				cerModel, _, err := parsevalidate.ParseValidateCer(certFile)
				if err == nil {
					defer cerlock.Unlock()
					cerlock.Lock()
					cerModelList.PushBack(cerModel)

				} else {
					defer errlock.Unlock()
					errlock.Lock()
					errList.PushBack(certFile + "   " + err.Error())
				}

			} else if strings.HasSuffix(certFile, ".crl") {
				crlModel, _, err := parsevalidate.ParseValidateCrl(certFile)
				if err == nil {
					defer crllock.Unlock()
					crllock.Lock()
					crlModelList.PushBack(crlModel)
				} else {
					defer errlock.Unlock()
					errlock.Lock()
					errList.PushBack(certFile + "   " + err.Error())
				}

			} else if strings.HasSuffix(certFile, ".mft") {
				mftModel, _, err := parsevalidate.ParseValidateMft(certFile)
				if err == nil {
					defer mftlock.Unlock()
					mftlock.Lock()
					mftModelList.PushBack(mftModel)
				} else {
					defer errlock.Unlock()
					errlock.Lock()
					errList.PushBack(certFile + "   " + err.Error())
				}

			} else if strings.HasSuffix(certFile, ".roa") {
				roaModel, _, err := parsevalidate.ParseValidateRoa(certFile)
				if err == nil {
					defer roalock.Unlock()
					roalock.Lock()
					roaModelList.PushBack(roaModel)
				} else {
					defer errlock.Unlock()
					errlock.Lock()
					errList.PushBack(certFile + "   " + err.Error())
				}
			}
			endOneParse := time.Now()
			allDuartionTime += endOneParse.Sub(startOneParse).Seconds()
			atomic.AddInt32(&cur, int32(-1))
			<-ch
		}(fileName, ch)
	}

	wg.Wait()
	close(ch)

	belogs.Info("ParseCert(): parse files: cerModelList.Len:", cerModelList.Len(),
		"	crlModelList.Len:", crlModelList.Len(),
		"	mftModelList.Len:", mftModelList.Len(),
		"	roaModelList.Len:", roaModelList.Len(),
		"	errList.Len", errList.Len(),
		"  	average seconds is :", allDuartionTime/float64(fileList.Len()))
	belogs.Debug("ParseCert():cerModelList", jsonutil.MarshalJson(cerModelList))
	belogs.Debug("ParseCert():crlModelList", jsonutil.MarshalJson(crlModelList))
	belogs.Debug("ParseCert():mftModelList", jsonutil.MarshalJson(mftModelList))
	belogs.Debug("ParseCert():roaModelList", jsonutil.MarshalJson(roaModelList))
	belogs.Error("ParseCert():errList", jsonutil.MarshalJson(errList))

	return cerModelList, crlModelList, mftModelList, roaModelList, errList
}

func PrintResult(cerModelList *list.List, crlModelList *list.List,
	mftModelList *list.List, roaModelList *list.List) {

	if cerModelList.Len() > 0 {
		fmt.Println("cer:")
		cerModels := make([]model.CerModel, 0)
		for e := cerModelList.Front(); nil != e; e = e.Next() {
			cerModels = append(cerModels, e.Value.(model.CerModel))
		}
		fmt.Println(jsonutil.MarshallJsonIndent(cerModels))
	}
	if crlModelList.Len() > 0 {
		fmt.Println("crl:")
		crlModels := make([]model.CrlModel, 0)
		for e := crlModelList.Front(); nil != e; e = e.Next() {
			crlModels = append(crlModels, e.Value.(model.CrlModel))
		}
		fmt.Println(jsonutil.MarshallJsonIndent(crlModels))
	}
	if mftModelList.Len() > 0 {
		fmt.Println("mft:")
		mftModels := make([]model.MftModel, 0)
		for e := mftModelList.Front(); nil != e; e = e.Next() {
			mftModels = append(mftModels, e.Value.(model.MftModel))
		}
		fmt.Println(jsonutil.MarshallJsonIndent(mftModels))
	}
	if roaModelList.Len() > 0 {
		fmt.Println("roa:")
		roaModels := make([]model.RoaModel, 0)
		for e := roaModelList.Front(); nil != e; e = e.Next() {
			roaModels = append(roaModels, e.Value.(model.RoaModel))
		}
		fmt.Println(jsonutil.MarshallJsonIndent(roaModels))
	}
}
