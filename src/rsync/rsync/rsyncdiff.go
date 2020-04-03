package rsync

import (
	"strings"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	httpclient "github.com/cpusoft/goutil/httpclient"
	osutil "github.com/cpusoft/goutil/osutil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"rsync/db"
	rsyncmodel "rsync/model"
)

func FoundDiffFiles(labRpkiSyncLogId uint64) {
	belogs.Info("FoundDiffFiles():start:")
	// save starttime to lab_rpki_sync_log
	err := db.UpdateRsyncLogDiffStateStart(labRpkiSyncLogId, "diffing")
	if err != nil {
		belogs.Error("Start():InsertRsyncLogRsyncStat fail:", err)
		return
	}

	filesFromDb, err := getFilesHashFromDb()
	if err != nil {
		belogs.Error("FoundDiffFiles():GetFilesHashFromDb fail:", err)
		return
	}
	filesFromDisk, err := getFilesHashFromDisk()
	if err != nil {
		belogs.Error("FoundDiffFiles():GetFilesHashFromDiskfail:", err)
		return
	}
	addFiles, delFiles, updateFiles, noChangeFiles, err := diffFiles(filesFromDb, filesFromDisk)
	if err != nil {
		belogs.Error("FoundDiffFiles():diffFiles:", err)
		return
	}

	err = db.UpdateRsyncLogDiffStateEnd(labRpkiSyncLogId, "diffed", filesFromDb,
		filesFromDisk, addFiles, delFiles, updateFiles, noChangeFiles)
	if err != nil {
		belogs.Error("FoundDiffFiles():UpdateRsyncLogDiffState fail:", err)
		return
	}
	belogs.Info("FoundDiffFiles():end , will call parsevalidate")

	// call parse validate
	go func() {
		httpclient.Post("http", conf.String("rpstir2::parsevalidateserver"), conf.Int("rpstir2::httpport"),
			"/parsevalidate/start", "")
	}()
}

// db is old, disk is new
func diffFiles(filesFromDb, filesFromDisk map[string]rsyncmodel.RsyncFileHash) (addFiles,
	delFiles, updateFiles, noChangeFiles map[string]rsyncmodel.RsyncFileHash, err error) {

	// if db is empty, so all filesFromDisk is add
	if len(filesFromDb) == 0 {
		return filesFromDisk, nil, nil, nil, nil
	}

	// if disk is emtpy, so all filesFromDb is del
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
	belogs.Debug("diffFiles(): len(addFiles):", len(addFiles), "  len(delFiles):", len(delFiles),
		"  len(updateFiles):", len(updateFiles), "  len(noChangeFiles):", len(noChangeFiles))
	return addFiles, delFiles, updateFiles, noChangeFiles, nil

}

func getFilesHashFromDb() (files map[string]rsyncmodel.RsyncFileHash, err error) {

	// init cap
	cerFileHashs := make([]rsyncmodel.RsyncFileHash, 15000)
	err = xormdb.XormEngine.Table("lab_rpki_cer").
		Select("filePath , fileName, fileHash, jsonAll as lastJsonAll, 'cer' as fileType").
		Asc("id").Find(&cerFileHashs)
	if err != nil {
		belogs.Error("GetFilesHashFromDb(): get lab_rpki_cer fail:", err)
		return nil, nil
	}
	belogs.Debug("getFilesHashFromDb(): len(cerrsyncmodel.RsyncFileHashs):", len(cerFileHashs))

	// init cap
	roaFileHashs := make([]rsyncmodel.RsyncFileHash, 25000)
	err = xormdb.XormEngine.Table("lab_rpki_roa").
		Select("filePath , fileName, fileHash, jsonAll as lastJsonAll, 'roa' as fileType").
		Asc("id").Find(&roaFileHashs)
	if err != nil {
		belogs.Error("GetFilesHashFromDb(): get lab_rpki_roa fail:", err)
		return nil, nil
	}
	belogs.Debug("getFilesHashFromDb(): len(roarsyncmodel.RsyncFileHashs):", len(roaFileHashs))

	// init cap
	crlFileHashs := make([]rsyncmodel.RsyncFileHash, 15000)
	err = xormdb.XormEngine.Table("lab_rpki_crl").
		Select("filePath,fileName,fileHash,jsonAll as 'lastJsonAll','crl' as fileType").
		Asc("id").Find(&crlFileHashs)
	if err != nil {
		belogs.Error("GetFilesHashFromDb(): get lab_rpki_crl fail:", err)
		return nil, nil
	}
	belogs.Debug("getFilesHashFromDb(): len(crlrsyncmodel.RsyncFileHashs):", len(crlFileHashs))

	// init cap
	mftFileHashs := make([]rsyncmodel.RsyncFileHash, 15000)
	err = xormdb.XormEngine.Table("lab_rpki_mft").
		Select("filePath,fileName,fileHash,jsonAll as 'lastJsonAll','mft' as fileType").
		Asc("id").Find(&mftFileHashs)
	if err != nil {
		belogs.Error("GetFilesHashFromDb(): get lab_rpki_mft fail:", err)
		return nil, nil
	}
	belogs.Debug("getFilesHashFromDb(): len(mftrsyncmodel.RsyncFileHashs):", len(mftFileHashs))

	files = make(map[string]rsyncmodel.RsyncFileHash, len(cerFileHashs)+len(roaFileHashs)+
		len(crlFileHashs)+len(mftFileHashs)+1000)
	for i, _ := range cerFileHashs {
		files[osutil.JoinPathFile(cerFileHashs[i].FilePath, cerFileHashs[i].FileName)] = cerFileHashs[i]
	}
	for i, _ := range roaFileHashs {
		files[osutil.JoinPathFile(roaFileHashs[i].FilePath, roaFileHashs[i].FileName)] = roaFileHashs[i]
	}
	for i, _ := range crlFileHashs {
		files[osutil.JoinPathFile(crlFileHashs[i].FilePath, crlFileHashs[i].FileName)] = crlFileHashs[i]
	}
	for i, _ := range mftFileHashs {
		files[osutil.JoinPathFile(mftFileHashs[i].FilePath, mftFileHashs[i].FileName)] = mftFileHashs[i]
	}
	belogs.Debug("getFilesHashFromDb(): len(files):", len(files))
	return files, nil
}

func getFilesHashFromDisk() (files map[string]rsyncmodel.RsyncFileHash, err error) {

	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	m[".crl"] = ".crl"
	m[".roa"] = ".roa"
	m[".mft"] = ".mft"

	fileStats, err := osutil.GetAllFileStatsBySuffixs(conf.VariableString("rsync::destpath")+"/", m)
	if err != nil {
		belogs.Error("GetFilesHashFromDisk(): GetAllFileStatsBySuffixs fail:", conf.VariableString("rsync::destpath")+"/", err)
		return nil, err
	}
	belogs.Debug("getFilesHashFromDisk(): len(fileStats):", len(fileStats))
	files = make(map[string]rsyncmodel.RsyncFileHash, len(fileStats))
	for i, _ := range fileStats {
		fileHash := rsyncmodel.RsyncFileHash{}
		fileHash.FileHash = fileStats[i].Hash256
		fileHash.FileName = fileStats[i].FileName
		fileHash.FilePath = fileStats[i].FilePath
		fileHash.FileType = strings.Replace(osutil.Ext(fileStats[i].FileName), ".", "", -1) //remove dot, should be cer/crl/roa/mft
		belogs.Debug("getFilesHashFromDisk(): fileHash:", fileHash)
		files[osutil.JoinPathFile(fileStats[i].FilePath, fileStats[i].FileName)] = fileHash
	}
	return files, nil

}
