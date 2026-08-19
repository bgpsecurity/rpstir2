package rsync

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/rsyncutil"
	"github.com/cpusoft/goutil/xormdb"
)

func getFilesHashFromDb(destPath string) (files map[string]rsyncutil.RsyncFileHash, err error) {
	start := time.Now()
	condition := ` and c.filePath like ? `
	param := destPath + "%"
	if len(destPath) == 0 {
		condition = ` and 'a' = ? `
		param = `a`
	}

	// init cap
	cerFileHashs := make([]rsyncutil.RsyncFileHash, 0, 25000)
	sql := `select c.filePath , c.fileName, c.fileHash, c.jsonAll as lastJsonAll,  'cer' as fileType  
			from lab_rpki_cer c , lab_rpki_sync_log_file f  
			where c.syncLogFileId = f.id   and f.syncStyle='rsync' 
			` + condition + `			
			order by c.filePath `
	err = xormdb.XormEngine.SQL(sql, param).Find(&cerFileHashs)
	if err != nil {
		belogs.Error("getFilesHashFromDb(): get lab_rpki_cer fail:", sql, param, err)
		return nil, err
	}
	belogs.Debug("getFilesHashFromDb(): len(cerrsyncmodel.RsyncFileHashs):", sql, param, len(cerFileHashs))

	// init cap
	crlFileHashs := make([]rsyncutil.RsyncFileHash, 0, 25000)
	sql = `select c.filePath , c.fileName, c.fileHash, c.jsonAll as lastJsonAll,  'crl' as fileType  
			from lab_rpki_crl c , lab_rpki_sync_log_file f  
			where c.syncLogFileId = f.id   and f.syncStyle='rsync' 
			` + condition + `			
			order by c.filePath `
	err = xormdb.XormEngine.SQL(sql, param).Find(&crlFileHashs)
	if err != nil {
		belogs.Error("getFilesHashFromDb(): get lab_rpki_crl fail:", sql, param, err)
		return nil, err
	}
	belogs.Debug("getFilesHashFromDb(): len(crlrsyncmodel.RsyncFileHashs):", sql, param, len(crlFileHashs))

	// init cap
	mftFileHashs := make([]rsyncutil.RsyncFileHash, 0, 25000)
	sql = `select c.filePath , c.fileName, c.fileHash, c.jsonAll as lastJsonAll,  'mft' as fileType  
			from lab_rpki_mft c , lab_rpki_sync_log_file f  
			where c.syncLogFileId = f.id   and f.syncStyle='rsync' 
			` + condition + `			
			order by c.filePath `
	err = xormdb.XormEngine.SQL(sql, param).Find(&mftFileHashs)
	if err != nil {
		belogs.Error("getFilesHashFromDb(): get lab_rpki_mft fail:", sql, param, err)
		return nil, err
	}
	belogs.Debug("getFilesHashFromDb(): len(mftrsyncmodel.RsyncFileHashs):", sql, param, len(mftFileHashs))

	// init cap
	roaFileHashs := make([]rsyncutil.RsyncFileHash, 0, 25000)
	sql = `select c.filePath , c.fileName, c.fileHash, c.jsonAll as lastJsonAll,  'roa' as fileType  
			from lab_rpki_roa c , lab_rpki_sync_log_file f  
			where c.syncLogFileId = f.id   and f.syncStyle='rsync' 
			` + condition + `			
			order by c.filePath `
	err = xormdb.XormEngine.SQL(sql, param).Find(&roaFileHashs)
	if err != nil {
		belogs.Error("getFilesHashFromDb(): get lab_rpki_roa fail:", sql, param, err)
		return nil, err
	}
	belogs.Debug("getFilesHashFromDb(): len(roarsyncmodel.RsyncFileHashs):", sql, param, len(roaFileHashs))

	// init cap
	asaFileHashs := make([]rsyncutil.RsyncFileHash, 0, 100)
	sql = `select c.filePath , c.fileName, c.fileHash, c.jsonAll as lastJsonAll,  'asa' as fileType  
			from lab_rpki_asa c , lab_rpki_sync_log_file f  
			where c.syncLogFileId = f.id   and f.syncStyle='rsync' 
			` + condition + `			
			order by c.filePath `
	err = xormdb.XormEngine.SQL(sql, param).Find(&asaFileHashs)
	if err != nil {
		belogs.Error("getFilesHashFromDb(): get lab_rpki_asa fail:", sql, param, err)
		return nil, err
	}
	belogs.Debug("getFilesHashFromDb(): len(asarsyncmodel.RsyncFileHashs):", sql, param, len(asaFileHashs))

	// init cap
	moaFileHashs := make([]rsyncutil.RsyncFileHash, 0, 10)
	sql = `select c.filePath , c.fileName, c.fileHash, c.jsonAll as lastJsonAll,  'moa' as fileType  
			from lab_rpki_moa c , lab_rpki_sync_log_file f  
			where c.syncLogFileId = f.id   and f.syncStyle='rsync' 
			` + condition + `			
			order by c.filePath `
	err = xormdb.XormEngine.SQL(sql, param).Find(&moaFileHashs)
	if err != nil {
		belogs.Error("getFilesHashFromDb(): get lab_rpki_moa fail:", sql, param, err)
		return nil, err
	}
	belogs.Debug("getFilesHashFromDb(): len(moarsyncmodel.RsyncFileHashs):", sql, param, len(moaFileHashs))

	files = make(map[string]rsyncutil.RsyncFileHash, len(cerFileHashs)+
		len(crlFileHashs)+len(mftFileHashs)+len(roaFileHashs)+len(asaFileHashs)+len(moaFileHashs)+10)
	for i := range cerFileHashs {
		files[osutil.JoinPathFile(cerFileHashs[i].FilePath, cerFileHashs[i].FileName)] = cerFileHashs[i]
	}
	for i := range crlFileHashs {
		files[osutil.JoinPathFile(crlFileHashs[i].FilePath, crlFileHashs[i].FileName)] = crlFileHashs[i]
	}
	for i := range mftFileHashs {
		files[osutil.JoinPathFile(mftFileHashs[i].FilePath, mftFileHashs[i].FileName)] = mftFileHashs[i]
	}
	for i := range roaFileHashs {
		files[osutil.JoinPathFile(roaFileHashs[i].FilePath, roaFileHashs[i].FileName)] = roaFileHashs[i]
	}
	for i := range asaFileHashs {
		files[osutil.JoinPathFile(asaFileHashs[i].FilePath, asaFileHashs[i].FileName)] = asaFileHashs[i]
	}
	for i := range moaFileHashs {
		files[osutil.JoinPathFile(moaFileHashs[i].FilePath, moaFileHashs[i].FileName)] = moaFileHashs[i]
	}
	belogs.Info("getFilesHashFromDb(): len(files):", len(files),
		"     files:", jsonutil.MarshalJson(files), "  time(s):", time.Since(start))
	return files, nil
}
