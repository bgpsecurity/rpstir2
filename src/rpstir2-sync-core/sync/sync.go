package sync

import (
	"os"
	"time"

	"github.com/cpusoft/goutil/belogs"
	model "rpstir2-model"
)

func CallRemoveAndDelDbByFilePaths(filePaths []string, connectCh chan bool) {
	select {
	case connect := <-connectCh:
		belogs.Debug("CallRemoveAndDelDbByFilePaths():  filePaths:", filePaths, "connect:", connect)
		if connect {
			RemoveAndDelDbByFilePaths(filePaths)
		}
		return
	case <-time.After(15 * time.Minute):
		belogs.Debug("CallRemoveAndDelDbByFilePaths(): time out, 15 minute:", filePaths)
		return
	}
}

//
func RemoveAndDelDbByFilePaths(filePaths []string) (err error) {
	start := time.Now()
	belogs.Debug("RemoveAndDelDbByFilePaths(): filePaths:", filePaths)
	// Do not delete all paths together, which will lead to too long transaction time
	for i := range filePaths {
		err = DelByFilePathDb(filePaths[i])
		if err != nil {
			belogs.Error("RemoveAndDelDbByFilePaths(): DelByFilePathDb, filePaths: ", filePaths, err)
			return err
		}

		err = os.RemoveAll(filePaths[i])
		if err != nil {
			// just log error
			belogs.Error("RemoveAndDelDbByFilePaths(): RemoveAll, filePaths[i]: ", filePaths[i], err)
		}
	}

	belogs.Debug("RemoveAndDelDbByFilePaths(): filePaths:", filePaths, "  time(s):", time.Since(start))
	return nil
}

func CallInsertSyncLogFiles(syncLogFilesCh chan []model.SyncLogFile, endCh chan bool) {
	for {
		select {
		case syncLogFiles, ok := <-syncLogFilesCh:
			belogs.Debug("CallInsertSyncLogFile(): len(syncLogFiles):", len(syncLogFiles), ok)
			if ok {
				InsertSyncLogFilesDb(syncLogFiles)
			}
		case end := <-endCh:
			if end {
				belogs.Debug("CallInsertSyncLogFile(): end:")
				return
			}
		}
	}
}
