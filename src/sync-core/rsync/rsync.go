package rsync

import (
	"strings"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/bgpsecurity/rpstir2/sync-core/sync"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/rsyncutil"
)

func InsertSyncLogFiles(syncLogId uint64,
	addFiles, delFiles, updateFiles map[string]rsyncutil.RsyncFileHash, syncLogFilesCh chan []model.LabRpkiSyncLogFile) (err error) {
	belogs.Debug("InsertSyncLogFiles():rsync, syncLogId:", syncLogId,
		"   len(addFiles):", len(addFiles), "   len(delFiles):", len(delFiles), "   len(updateFiles):", len(updateFiles))

	rsyncTime := time.Now()
	syncLogFiles := make([]model.LabRpkiSyncLogFile, 0, 2*(len(addFiles)+len(delFiles)+len(updateFiles)))

	// add
	rsyncType := "add"
	for _, addFile := range addFiles {
		syncLogFile := ConvertToSyncLogFile(syncLogId, rsyncTime,
			rsyncType, &addFile)
		syncLogFiles = append(syncLogFiles, syncLogFile)
	}
	belogs.Debug("InsertSyncLogFiles():rsync, after len(addFiles):", syncLogId, len(addFiles))

	// update
	rsyncType = "update"
	for _, updateFile := range updateFiles {
		syncLogFile := ConvertToSyncLogFile(syncLogId, rsyncTime,
			rsyncType, &updateFile)
		syncLogFiles = append(syncLogFiles, syncLogFile)
	}
	belogs.Debug("InsertSyncLogFiles():rsync, after len(updateFiles):", syncLogId, len(updateFiles))

	// del
	rsyncType = "del"
	for _, delFile := range delFiles {
		syncLogFile := ConvertToSyncLogFile(syncLogId, rsyncTime,
			rsyncType, &delFile)
		syncLogFiles = append(syncLogFiles, syncLogFile)
	}

	belogs.Debug("InsertSyncLogFiles():rsync,  len(addFiles):", len(addFiles),
		"   len(delFiles):", len(delFiles), "   len(updateFiles):", len(updateFiles),
		"   len(syncLogFiles):", len(syncLogFiles))
	if syncLogFilesCh != nil {
		belogs.Debug("InsertSyncLogFiles():rsync len(syncLogFiles) -> syncLogFilesCh:", len(syncLogFiles))
		syncLogFilesCh <- syncLogFiles
	} else {
		belogs.Debug("InsertSyncLogFiles():rsync InsertSyncLogFilesDb():", len(syncLogFiles))
		sync.InsertSyncLogFilesDb(syncLogFiles)
	}
	return nil

}

func ConvertToSyncLogFile(syncLogId uint64, rsyncTime time.Time,
	syncType string, rsyncFileHash *rsyncutil.RsyncFileHash) (syncLogFile model.LabRpkiSyncLogFile) {

	rtr := "notNeed"
	if rsyncFileHash.FileType == "roa" || rsyncFileHash.FileType == "asa" || rsyncFileHash.FileType == "moa" {
		rtr = "notYet"
	}

	syncLogFileState := model.LabRpkiSyncLogFileState{
		Sync:            "finished",
		UpdateCertTable: "notYet",
		Rtr:             rtr,
	}
	state := jsonutil.MarshalJson(syncLogFileState)
	// /home/rpki/data/rrdprepo/rpki.ripe.net/repository/*** --> rsync://rpki.ripe.net/repository/***
	// so, when replace, keep "/" and add "rsync:/"
	sourceUrl := strings.Replace(rsyncFileHash.FilePath, conf.String("rsync::destPath"), "rsync:/", -1)
	// syncLogFile
	syncLogFile = model.LabRpkiSyncLogFile{
		SyncLogId: syncLogId,
		FileType:  rsyncFileHash.FileType,
		SyncTime:  rsyncTime,
		FilePath:  rsyncFileHash.FilePath,
		FileName:  rsyncFileHash.FileName,
		SourceUrl: sourceUrl,
		SyncType:  syncType,
		SyncStyle: "rsync",
		State:     state,
		FileHash:  rsyncFileHash.FileHash,
	}
	return syncLogFile

}
