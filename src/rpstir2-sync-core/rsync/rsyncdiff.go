package rsync

import (
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/rsyncutil"
)

// destPath: when it is "", will diff all. if not, just diff destPath
func FoundDiffFiles(syncLogId uint64, destPath string, syncLogFilesCh chan []model.SyncLogFile) (addFilesLen, delFilesLen, updateFilesLen, noChangeFilesLen uint64, err error) {
	start := time.Now()
	belogs.Info("FoundDiffFiles():start,  syncLogId:", syncLogId, " destPath:", destPath)

	filesFromDb, err := getFilesHashFromDb(destPath)
	if err != nil {
		belogs.Error("FoundDiffFiles():GetFilesHashFromDb fail:", err)
		return 0, 0, 0, 0, err
	}
	filesFromDisk, err := rsyncutil.GetFilesHashFromDisk(destPath)
	if err != nil {
		belogs.Error("FoundDiffFiles():GetFilesHashFromDisk fail:", err)
		return 0, 0, 0, 0, err
	}
	addFiles, delFiles, updateFiles, noChangeFiles, err := rsyncutil.DiffFiles(filesFromDb, filesFromDisk)
	if err != nil {
		belogs.Error("FoundDiffFiles():diffFiles:", err)
		return 0, 0, 0, 0, err
	}

	err = InsertSyncLogFiles(syncLogId, addFiles, delFiles, updateFiles, syncLogFilesCh)
	if err != nil {
		belogs.Error("FoundDiffFiles():InsertRsyncLogFiles:", err)
		return 0, 0, 0, 0, err
	}

	addFilesLen = uint64(len(addFiles))
	delFilesLen = uint64(len(delFiles))
	updateFilesLen = uint64(len(updateFiles))
	noChangeFilesLen = uint64(len(noChangeFiles))

	belogs.Info("FoundDiffFiles():end, addFilesLen, delFilesLen, updateFilesLen, noChangeFilesLen: ",
		addFilesLen, delFilesLen, updateFilesLen, noChangeFilesLen, "  time(s):", time.Since(start))
	return addFilesLen, delFilesLen, updateFilesLen, noChangeFilesLen, nil
}
