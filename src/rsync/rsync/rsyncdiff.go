package rsync

import (
	"strings"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	db "rsync/db"
	rsyncmodel "rsync/model"
)

func FoundDiffFiles(labRpkiSyncLogId uint64) (addFilesLen, delFilesLen, updateFilesLen, noChangeFilesLen uint64, err error) {
	start := time.Now()
	belogs.Info("FoundDiffFiles():start,  labRpkiSyncLogId:", labRpkiSyncLogId)

	filesFromDb, err := getFilesHashFromDb()
	if err != nil {
		belogs.Error("FoundDiffFiles():GetFilesHashFromDb fail:", err)
		return 0, 0, 0, 0, err
	}
	filesFromDisk, err := getFilesHashFromDisk()
	if err != nil {
		belogs.Error("FoundDiffFiles():GetFilesHashFromDisk fail:", err)
		return 0, 0, 0, 0, err
	}
	addFiles, delFiles, updateFiles, noChangeFiles, err := diffFiles(filesFromDb, filesFromDisk)
	if err != nil {
		belogs.Error("FoundDiffFiles():diffFiles:", err)
		return 0, 0, 0, 0, err
	}

	err = db.InsertSyncLogFiles(labRpkiSyncLogId, addFiles, delFiles, updateFiles)
	if err != nil {
		belogs.Error("FoundDiffFiles():InsertRsyncLogFiles:", err)
		return 0, 0, 0, 0, err
	}

	addFilesLen = uint64(len(addFiles))
	delFilesLen = uint64(len(delFiles))
	updateFilesLen = uint64(len(updateFiles))
	noChangeFilesLen = uint64(len(noChangeFiles))

	belogs.Info("FoundDiffFiles():end, addFilesLen, delFilesLen, updateFilesLen, noChangeFilesLen: ",
		addFilesLen, delFilesLen, updateFilesLen, noChangeFilesLen, "  time(s):", time.Now().Sub(start).Seconds())
	return addFilesLen, delFilesLen, updateFilesLen, noChangeFilesLen, nil
}

// db is old, disk is new
func diffFiles(filesFromDb, filesFromDisk map[string]rsyncmodel.RsyncFileHash) (addFiles,
	delFiles, updateFiles, noChangeFiles map[string]rsyncmodel.RsyncFileHash, err error) {

	start := time.Now()
	// if db is empty, so all filesFromDisk is add
	if len(filesFromDb) == 0 {
		return filesFromDisk, nil, nil, nil, nil
	}

	// if disk is empty, so all filesFromDb is del
	if len(filesFromDisk) == 0 {
		return nil, filesFromDb, nil, nil, nil
	}

	// for db, check. add/update/nochange from disk, del from db
	addFiles = make(map[string]rsyncmodel.RsyncFileHash, len(filesFromDb))
	delFiles = make(map[string]rsyncmodel.RsyncFileHash, len(filesFromDb))
	updateFiles = make(map[string]rsyncmodel.RsyncFileHash, len(filesFromDb))
	noChangeFiles = make(map[string]rsyncmodel.RsyncFileHash, len(filesFromDb))

	// for db, check disk
	for keyDb, valueDb := range filesFromDb {
		// if found in disk,
		if valueDisk, ok := filesFromDisk[keyDb]; ok {
			// if hash is equal, then save to noChangeFiles, else save to updateFiles
			// and db.jsonall should save as lasjsonall
			if valueDb.FileHash == valueDisk.FileHash {
				valueDisk.LastJsonAll = valueDb.LastJsonAll
				noChangeFiles[keyDb] = valueDisk

			} else {
				valueDisk.LastJsonAll = valueDb.LastJsonAll
				updateFiles[keyDb] = valueDisk
			}
			//have found in disk, then del it in disk map, so remain in disk will be add
			delete(filesFromDisk, keyDb)
		} else {

			// if not found in disk ,then is del, so save to delFiles, and value is db
			delFiles[keyDb] = valueDb
		}
	}
	addFiles = filesFromDisk
	belogs.Debug("diffFiles(): len(addFiles):", len(addFiles), jsonutil.MarshalJson(addFiles))
	belogs.Debug("diffFiles(): len(delFiles):", len(delFiles), jsonutil.MarshalJson(delFiles))
	belogs.Debug("diffFiles(): len(updateFiles):", len(updateFiles), jsonutil.MarshalJson(updateFiles))
	belogs.Debug("diffFiles(): len(noChangeFiles):", len(noChangeFiles), jsonutil.MarshalJson(noChangeFiles))
	belogs.Debug("diffFiles(): time(s):", time.Now().Sub(start).Seconds())
	belogs.Info("diffFiles(): len(addFiles):", len(addFiles), "  len(delFiles):", len(delFiles),
		"  len(updateFiles):", len(updateFiles), "  len(noChangeFiles):", len(noChangeFiles), "  time(s):", time.Now().Sub(start).Seconds())

	return addFiles, delFiles, updateFiles, noChangeFiles, nil

}

func getFilesHashFromDb() (files map[string]rsyncmodel.RsyncFileHash, err error) {
	start := time.Now()

	// init cap
	cerFileHashs := make([]rsyncmodel.RsyncFileHash, 0, 25000)
	sql := `select c.filePath , c.fileName, c.fileHash, c.jsonAll as lastJsonAll,  'cer' as fileType  
			from lab_rpki_cer c , lab_rpki_sync_log_file f  
			where c.syncLogFileId = f.id   and f.syncStyle='rsync' order by c.id `
	err = xormdb.XormEngine.Sql(sql).Find(&cerFileHashs)
	if err != nil {
		belogs.Error("GetFilesHashFromDb(): get lab_rpki_cer fail:", err)
		return nil, nil
	}
	belogs.Debug("getFilesHashFromDb(): len(cerrsyncmodel.RsyncFileHashs):", len(cerFileHashs))

	// init cap
	roaFileHashs := make([]rsyncmodel.RsyncFileHash, 0, 25000)
	sql = `select c.filePath , c.fileName, c.fileHash, c.jsonAll as lastJsonAll,  'roa' as fileType  
			from lab_rpki_roa c , lab_rpki_sync_log_file f  
			where c.syncLogFileId = f.id   and f.syncStyle='rsync' order by c.id `
	err = xormdb.XormEngine.Sql(sql).Find(&roaFileHashs)
	if err != nil {
		belogs.Error("GetFilesHashFromDb(): get lab_rpki_roa fail:", err)
		return nil, nil
	}
	belogs.Debug("getFilesHashFromDb(): len(roarsyncmodel.RsyncFileHashs):", len(roaFileHashs))

	// init cap
	crlFileHashs := make([]rsyncmodel.RsyncFileHash, 0, 25000)
	sql = `select c.filePath , c.fileName, c.fileHash, c.jsonAll as lastJsonAll,  'roa' as fileType  
			from lab_rpki_crl c , lab_rpki_sync_log_file f  
			where c.syncLogFileId = f.id   and f.syncStyle='rsync' order by c.id `
	err = xormdb.XormEngine.Sql(sql).Find(&crlFileHashs)
	if err != nil {
		belogs.Error("GetFilesHashFromDb(): get lab_rpki_crl fail:", err)
		return nil, nil
	}
	belogs.Debug("getFilesHashFromDb(): len(crlrsyncmodel.RsyncFileHashs):", len(crlFileHashs))

	// init cap
	mftFileHashs := make([]rsyncmodel.RsyncFileHash, 0, 25000)
	sql = `select c.filePath , c.fileName, c.fileHash, c.jsonAll as lastJsonAll,  'roa' as fileType  
			from lab_rpki_mft c , lab_rpki_sync_log_file f  
			where c.syncLogFileId = f.id   and f.syncStyle='rsync' order by c.id `
	err = xormdb.XormEngine.Sql(sql).Find(&mftFileHashs)
	if err != nil {
		belogs.Error("GetFilesHashFromDb(): get lab_rpki_mft fail:", err)
		return nil, nil
	}
	belogs.Debug("getFilesHashFromDb(): len(mftrsyncmodel.RsyncFileHashs):", len(mftFileHashs))

	files = make(map[string]rsyncmodel.RsyncFileHash, len(cerFileHashs)+len(roaFileHashs)+
		len(crlFileHashs)+len(mftFileHashs)+1000)
	for i := range cerFileHashs {
		files[osutil.JoinPathFile(cerFileHashs[i].FilePath, cerFileHashs[i].FileName)] = cerFileHashs[i]
	}
	for i := range roaFileHashs {
		files[osutil.JoinPathFile(roaFileHashs[i].FilePath, roaFileHashs[i].FileName)] = roaFileHashs[i]
	}
	for i := range crlFileHashs {
		files[osutil.JoinPathFile(crlFileHashs[i].FilePath, crlFileHashs[i].FileName)] = crlFileHashs[i]
	}
	for i := range mftFileHashs {
		files[osutil.JoinPathFile(mftFileHashs[i].FilePath, mftFileHashs[i].FileName)] = mftFileHashs[i]
	}
	belogs.Debug("getFilesHashFromDb(): len(files):", len(files),
		"     files:", jsonutil.MarshalJson(files), "  time(s):", time.Now().Sub(start).Seconds())
	return files, nil
}

func getFilesHashFromDisk() (files map[string]rsyncmodel.RsyncFileHash, err error) {
	start := time.Now()

	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	m[".crl"] = ".crl"
	m[".roa"] = ".roa"
	m[".mft"] = ".mft"

	fileStats, err := osutil.GetAllFileStatsBySuffixs(conf.VariableString("rsync::destPath")+"/", m)
	if err != nil {
		belogs.Error("GetFilesHashFromDisk(): GetAllFileStatsBySuffixs fail:", conf.VariableString("rsync::destPath")+"/", err)
		return nil, err
	}
	belogs.Debug("getFilesHashFromDisk(): len(fileStats):", len(fileStats), "    fileStats:", jsonutil.MarshalJson(fileStats))
	files = make(map[string]rsyncmodel.RsyncFileHash, len(fileStats))
	for i := range fileStats {
		fileHash := rsyncmodel.RsyncFileHash{}
		fileHash.FileHash = fileStats[i].Hash256
		fileHash.FileName = fileStats[i].FileName
		fileHash.FilePath = fileStats[i].FilePath
		fileHash.FileType = strings.Replace(osutil.Ext(fileStats[i].FileName), ".", "", -1) //remove dot, should be cer/crl/roa/mft
		files[osutil.JoinPathFile(fileStats[i].FilePath, fileStats[i].FileName)] = fileHash
	}

	belogs.Info("getFilesHashFromDisk(): len(files):", len(files), "  time(s):", time.Now().Sub(start).Seconds())
	return files, nil

}
