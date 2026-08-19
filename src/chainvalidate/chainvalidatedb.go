package chainvalidate

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
)

/*
// SQLDataSource 实现了 DataSource 接口，用于 MySQL 数据源
type SQLDataSource struct{}

func getDataSource() *SQLDataSource {
	return &SQLDataSource{}
}
//func getChainCertSqlDb(oneSize, myCount int64, sql string,
*/
//	wg *sync.WaitGroup, chainCertSqlsCh chan []*ChainCertSql) (selectCount int64) {
//	start := time.Now()
//	var offset int64
//	belogs.Debug("getChainCertSqlDb(): will select certSql limit, sql:", sql,
//		"  oneSize:", oneSize, "  myCount:", myCount)
//	for offset = 0; offset < myCount; offset += oneSize {
//		startOne := time.Now()
//		chainCertSqls := make([]*ChainCertSql, 0)
//		err := xormdb.XormEngine.SQL(sql, offset, oneSize).Find(&chainCertSqls)
//		if err != nil {
//			belogs.Error("getChainCertSqlDb():select rpki_roa limit, fail:", err, " one sql time(s):", time.Since(startOne))
//			return
//		}
//		atomic.AddInt64(&selectCount, int64(len(chainCertSqls)))
//		wg.Add(1)
//		chainCertSqlsCh <- chainCertSqls
//		belogs.Debug("getChainCertSqlDb(): select certSql limit, sql:", sql,
//			"  len(chainCertSqls):", len(chainCertSqls),
//			"  offset:", offset, " oneSize:", oneSize,
//			"  one sql time(s):", time.Since(startOne))
//	}
//	belogs.Info("getChainCertSqlDb(): done, sql:", sql, "  oneSize:", oneSize,
//		"  myCount:", myCount, "  selectCount:", selectCount, "  time(s):", time.Since(start))
//	return selectCount
//}

func getChainCertSqlDb(oneSize, myCount int64, sql string,
	wg *sync.WaitGroup, chainCertSqlsCh chan []*ChainCertData) (selectCount int64) {
	start := time.Now()
	var offset int64
	belogs.Debug("getChainCertSqlDb(): will select certSql limit, sql:", sql,
		"  oneSize:", oneSize, "  myCount:", myCount)
	for offset = 0; offset < myCount; offset += oneSize {
		startOne := time.Now()
		chainCertDatas := make([]*ChainCertData, 0)
		err := xormdb.XormEngine.SQL(sql, offset, oneSize).Find(&chainCertDatas)
		if err != nil {
			belogs.Error("getChainCertSqlDb():select rpki_roa limit, fail:", err, " one sql time(s):", time.Since(startOne))
			return
		}
		atomic.AddInt64(&selectCount, int64(len(chainCertDatas)))
		wg.Add(1)
		chainCertSqlsCh <- chainCertDatas
		belogs.Debug("getChainCertSqlDb(): select certSql limit, sql:", sql,
			"  len(chainCertSqls):", len(chainCertDatas),
			"  offset:", offset, " oneSize:", oneSize,
			"  one sql time(s):", time.Since(startOne))
	}
	belogs.Info("getChainCertSqlDb(): done, sql:", sql, "  oneSize:", oneSize,
		"  myCount:", myCount, "  selectCount:", selectCount, "  time(s):", time.Since(start))
	return selectCount
}
