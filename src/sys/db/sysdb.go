package sys

import (
	"errors"
	"math/rand"
	"time"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"

	"model"
	sysmodel "sys/model"
)

var intiSqls []string = []string{
	`SET FOREIGN_KEY_CHECKS = 0`,
	`DROP TABLE IF EXISTS	lab_rpki_cer`,
	`DROP TABLE IF EXISTS	lab_rpki_cer_sia`,
	`DROP TABLE IF EXISTS	lab_rpki_cer_aia`,
	`DROP TABLE IF EXISTS	lab_rpki_cer_crldp`,
	`DROP TABLE IF EXISTS	lab_rpki_cer_ipaddress`,
	`DROP TABLE IF EXISTS	lab_rpki_cer_asn`,
	`DROP TABLE IF EXISTS	lab_rpki_crl`,
	`DROP TABLE IF EXISTS	lab_rpki_crl_revoked_cert`,
	`DROP TABLE IF EXISTS	lab_rpki_mft`,
	`DROP TABLE IF EXISTS	lab_rpki_mft_sia`,
	`DROP TABLE IF EXISTS	lab_rpki_mft_aia`,
	`DROP TABLE IF EXISTS	lab_rpki_mft_file_hash`,
	`DROP TABLE IF EXISTS	lab_rpki_roa`,
	`DROP TABLE IF EXISTS	lab_rpki_roa_sia`,
	`DROP TABLE IF EXISTS	lab_rpki_roa_aia`,
	`DROP TABLE IF EXISTS	lab_rpki_roa_ipaddress`,
	`DROP TABLE IF EXISTS	lab_rpki_roa_ee_ipaddress`,
	`DROP TABLE IF EXISTS	lab_rpki_sync_log_file`,
	`DROP TABLE IF EXISTS	lab_rpki_sync_rrdp_log`,
	`DROP TABLE IF EXISTS	lab_rpki_sync_log`,
	`DROP TABLE IF EXISTS	lab_rpki_rtr_session`,
	`DROP TABLE IF EXISTS	lab_rpki_rtr_serial_number`,
	`DROP TABLE IF EXISTS	lab_rpki_rtr_full`,
	`DROP TABLE IF EXISTS	lab_rpki_rtr_full_log`,
	`DROP TABLE IF EXISTS	lab_rpki_rtr_incremental`,
	`DROP TABLE IF EXISTS	lab_rpki_slurm`,
	`DROP TABLE IF EXISTS	lab_rpki_slurm_file`,
	`DROP TABLE IF EXISTS	lab_rpki_statistic`,
	`DROP TABLE IF EXISTS	lab_rpki_stat_roa_competation`,
	`DROP TABLE IF EXISTS	lab_rpki_transfer_target`,
	`DROP TABLE IF EXISTS	lab_rpki_transfer_log`,
	`DROP TABLE IF EXISTS	lab_rpki_statistic`,
	`DROP TABLE IF EXISTS	lab_rpki_conf`,
	`DROP VIEW  IF EXISTS	lab_rpki_roa_ipaddress_view`,
	`SET FOREIGN_KEY_CHECKS = 1`,
	`
#################################
## main table for cer/crl/roa/mft
#################################	
CREATE TABLE lab_rpki_cer (
  id int(10) unsigned not null primary key auto_increment,
  sn varchar(128) NOT NULL,
  notBefore datetime NOT NULL,
  notAfter datetime NOT NULL,
  subject varchar(512) ,
  issuer varchar(512) ,
  ski varchar(128) ,
  aki varchar(128) ,
  filePath varchar(512) NOT NULL ,
  fileName varchar(128) NOT NULL ,
  state json COMMENT 'state info in json',
  jsonAll json NOT NULL  COMMENT 'all cer info in json',
  chainCerts json   COMMENT 'chain certs(cer/crl/mft/roa) in json',
  syncLogId int(10) unsigned not null  COMMENT 'foreign key  references lab_rpki_sync_log(id)',
  syncLogFileId int(10) unsigned not null  COMMENT 'foreign key  references lab_rpki_sync_log_file(id)',
  updateTime datetime NOT NULL,
  fileHash varchar(512) NOT NULL ,
  origin json COMMENT 'origin(rir->repo) in json',
  key  ski (ski),
  key  aki (aki),
  key  filePath (filePath),
  key  fileName (fileName),
  unique  cerFilePathFileName (filePath,fileName),
  unique  cerSkiFilePath (ski,filePath)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4   COLLATE=utf8mb4_bin comment='main cer table'
`,

	`
CREATE TABLE lab_rpki_cer_sia (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	rpkiManifest  varchar(512)  COMMENT 'mft sync url',
	rpkiNotify  varchar(512) ,
	caRepository  varchar(512)  COMMENT 'ca repository url(direcotry)',
	signedObject  varchar(512) ,
	FOREIGN key (cerid) REFERENCES lab_rpki_cer(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='cer sia'
`,

	`
CREATE TABLE lab_rpki_cer_aia (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	caIssuers  varchar(512) COMMENT 'father ca url (cer file)',
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='cer aia'
`,

	`
CREATE TABLE lab_rpki_cer_crldp (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	crldp varchar(512) COMMENT 'crl sync url(file)',
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='cer crl'
`,

	`
CREATE TABLE lab_rpki_cer_ipaddress (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	addressFamily  int(10) unsigned not null,
	addressPrefix  varchar(512) COMMENT 'address prefix: 147.28.83.0/24 ',
	min  varchar(512) COMMENT 'min address:  99.96.0.0',
	max  varchar(512) COMMENT 'max address:  99.105.127.255',
	rangeStart  varchar(512) COMMENT 'min address range from addressPrefix or min/max, in hex:  63.60.00.00',
	rangeEnd  varchar(512) COMMENT 'max address range from addressPrefix or min/max, in hex:  63.69.7f.ff',
	addressPrefixRange json COMMENT 'min--max, such as 192.0.2.0--192.0.2.130, will convert to addressprefix range in json:{192.0.2.0/25, 192.0.2.128/31, 192.0.2.130/32}',
	key  addressPrefix (addressPrefix),
	key  rangeStart (rangeStart),
	key  rangeEnd (rangeEnd),
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='cer ip address range'
`,
	`
## because 0 of asn has special meaning, so the default of asn is -1, and is "bigint signed" in mysql
CREATE TABLE lab_rpki_cer_asn (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	asn bigint(20) signed,
	min bigint(20) signed,
	max bigint(20) signed,
	 key  asn (asn),
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='cer asn range'
`,

	`
###### crl	  		    
CREATE TABLE lab_rpki_crl (
  id int(10) unsigned not null primary key auto_increment,
  thisUpdate datetime NOT NULL,
  nextUpdate datetime NOT NULL,
  hasExpired varchar(8) ,
  aki varchar(128) ,
  crlNumber bigint(20) unsigned not null ,
  filePath varchar(512) NOT NULL ,
  fileName varchar(128) NOT NULL ,
  state json COMMENT 'state info in json',
  jsonAll json NOT NULL,
  chainCerts json   COMMENT 'chain certs(cer/crl/mft/roa) in json',
  syncLogId int(10) unsigned not null  COMMENT 'foreign key  references lab_rpki_sync_log(id)',
  syncLogFileId int(10) unsigned not null  COMMENT 'foreign key  references lab_rpki_sync_log_file(id)',
  updateTime datetime NOT NULL,
  fileHash varchar(512) NOT NULL ,
  origin json COMMENT 'origin(rir->repo) in json',
  key  aki (aki),
  key  filePath (filePath),  
  key  fileName (fileName),
  unique  crlFilePathFileName (filePath,fileName)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='crl '
`,

	`
CREATE TABLE lab_rpki_crl_revoked_cert (
	id int(10) unsigned not null primary key auto_increment,
	crlId int(10) unsigned not null,
	sn varchar(512) NOT NULL,
	revocationTime datetime NOT NULL,
	key  sn (sn),
	foreign key (crlId) references lab_rpki_crl(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='all sn and revocationTime in crl'
`,

	`
###### manifest
CREATE TABLE lab_rpki_mft (
  id int(10) unsigned not null primary key auto_increment,
  mftNumber  varchar(1024) NOT NULL,
  thisUpdate datetime NOT NULL,
  nextUpdate datetime NOT NULL,
  ski varchar(128) ,
  aki varchar(128) ,
  filePath varchar(512) NOT NULL ,
  fileName varchar(128) NOT NULL ,
  state json COMMENT 'state info in json',
  jsonAll json NOT NULL,
  chainCerts json   COMMENT 'chain certs(cer/crl/mft/roa) in json',
  syncLogId int(10) unsigned not null  COMMENT 'foreign key  references lab_rpki_sync_log(id)',
  syncLogFileId int(10) unsigned not null  COMMENT 'foreign key  references lab_rpki_sync_log_file(id)',
  updateTime datetime NOT NULL,
  fileHash varchar(512) NOT NULL ,
  origin json COMMENT 'origin(rir->repo) in json',
  key  ski (ski),
  key  aki (aki),
  key  filePath (filePath),  
  key  fileName (fileName),
  unique  mftFilePathFileName (filePath,fileName),
  unique  mftSkiFilePath (ski,filePath) 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='manifest'
`,

	`
CREATE TABLE lab_rpki_mft_sia (
	id int(10) unsigned not null primary key auto_increment,
	mftId int(10) unsigned not null,
	rpkiManifest  varchar(512) ,
	rpkiNotify  varchar(512) ,
	caRepository  varchar(512) ,
	signedObject  varchar(512) ,
	FOREIGN key (mftId) REFERENCES lab_rpki_mft(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='mft sia'
`,

	`
CREATE TABLE lab_rpki_mft_aia (
	id int(10) unsigned not null primary key auto_increment,
	mftId int(10) unsigned not null,
	caIssuers  varchar(512) ,
	foreign key (mftId) references lab_rpki_mft(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='mft aia'
`,

	`
CREATE TABLE lab_rpki_mft_file_hash (
	id int(10) unsigned not null primary key auto_increment,
	mftId int(10) unsigned not null,
	file varchar(1024),
  hash varchar(1024),
	foreign key (mftId) references lab_rpki_mft(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='files in manifest'
`,

	`
###### roa
## because 0 of asn has special meaning, so the default of asn is -1, and is "bigint signed" in mysql
CREATE TABLE lab_rpki_roa (
  id int(10) unsigned not null primary key auto_increment,
  asn bigint(20) signed not null,
  ski varchar(128) ,
  aki varchar(128) ,
  filePath varchar(512) NOT NULL ,
  fileName varchar(128) NOT NULL ,
  state json COMMENT 'state info in json',
  jsonAll json NOT NULL,
  chainCerts json   COMMENT 'chain certs(cer/crl/mft/roa) in json',
  syncLogId int(10) unsigned not null  COMMENT 'foreign key  references lab_rpki_sync_log(id)',
  syncLogFileId int(10) unsigned not null  COMMENT 'foreign key  references lab_rpki_sync_log_file(id)',
  updateTime datetime NOT NULL,
  fileHash varchar(512) NOT NULL ,
  origin json COMMENT 'origin(rir->repo) in json',
  key  ski (ski),
  key  aki (aki),
  key  filePath (filePath),
  key  fileName (fileName),
  unique  roaFilePathFileName (filePath,fileName),
  unique  roaSkiFilePath (ski,filePath) 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4   COLLATE=utf8mb4_bin comment='roa info'
`,

	`	
CREATE TABLE lab_rpki_roa_sia (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	rpkiManifest  varchar(512) ,
	rpkiNotify  varchar(512) ,
	caRepository  varchar(512) ,
	signedObject  varchar(512) ,
	FOREIGN key (roaId) REFERENCES lab_rpki_roa(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='roa sia'
`,

	`
CREATE TABLE lab_rpki_roa_aia (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	caIssuers  varchar(512) ,
	foreign key (roaId) references lab_rpki_roa(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='roa aia'
`,

	`
CREATE TABLE lab_rpki_roa_ipaddress (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	addressFamily  int(10) unsigned not null,
	addressPrefix  varchar(512),
	maxLength int(10) unsigned,
	rangeStart  varchar(512),
	rangeEnd  varchar(512),
	addressPrefixRange json COMMENT 'min--max, such as 192.0.2.0--192.0.2.130, will convert to addressprefix range in json:{192.0.2.0/25, 192.0.2.128/31, 192.0.2.130/32}',
	key  addressPrefix (addressPrefix),
	key  rangeStart (rangeStart),
	key  rangeEnd (rangeEnd),
	foreign key (roaId) references lab_rpki_roa(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='roa ip prefix'
`,

	`
CREATE TABLE lab_rpki_roa_ee_ipaddress (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	addressFamily  int(10) unsigned not null,
	addressPrefix  varchar(512) COMMENT 'address prefix: 147.28.83.0/24 ',
	min  varchar(512) COMMENT 'min address:  99.96.0.0',
	max  varchar(512) COMMENT 'max address:  99.105.127.255',
	rangeStart  varchar(512) COMMENT 'min address range from addressPrefix or min/max, in hex:  63.60.00.00',
	rangeEnd  varchar(512) COMMENT 'max address range from addressPrefix or min/max, in hex:  63.69.7f.ff',
	addressPrefixRange json COMMENT 'min--max, such as 192.0.2.0--192.0.2.130, will convert to addressprefix range in json:{192.0.2.0/25, 192.0.2.128/31, 192.0.2.130/32}',
	key  addressPrefix (addressPrefix),
	key  rangeStart (rangeStart),
	key  rangeEnd (rangeEnd),
	foreign key (roaId) references lab_rpki_roa(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='roa ee ip prefix'
`,

	`
################################################
## recored every sync log for cer/crl/roa/mft
################################################
CREATE TABLE lab_rpki_sync_log (
  id int(10) unsigned not null primary key auto_increment,
  rsyncState json,
  rrdpState json,
  diffState json,
  parseValidateState json,
  chainValidateState json,
  rtrState json,
  state varchar(16) not null COMMENT 'rsyncing/rsynced ddrping/ddrped  diffing/diffed   parsevalidating/parsevalidated   rtring/rtred idle',
  syncStyle varchar(16) not null COMMENT 'rsync/rrdp' 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='recored every sync log'
`,

	`
CREATE TABLE lab_rpki_sync_log_file (
  id int(10) unsigned not null primary key auto_increment,
  syncLogId int(10) unsigned not null  COMMENT 'foreign key  references lab_rpki_sync_log(id)',
  syncTime datetime NOT NULL   COMMENT 'sync time for every file',
  syncType varchar(16) NOT NULL COMMENT 'add/del/update' ,
  fileType varchar(16) NOT NULL  COMMENT 'cer/roa/mft/crl/',
  ski varchar(128) ,
  aki varchar(128) ,
  filePath varchar(512) NOT NULL ,
  fileName varchar(128) NOT NULL ,
  parseValidateResultJson json COMMENT 'parse json info' ,
  jsonAll json COMMENT 'cert json info from cer/crl/mft/roa.jsonAll' ,
  lastJsonAll json COMMENT 'when update, last cert json info from cer/crl/mft/roa.jsonAll' ,
  fileHash varchar(512) ,
  lastFileHash varchar(512) ,
  state json  COMMENT '{"sync":"finished","updateCertTable":"notYet/finished"}: have synced ,have published to main table',
  key  fileType (fileType),
  key  syncType (syncType),
  key  ski (ski),
  key  aki (aki),
  key  filePath (filePath),
  key  fileName (fileName),
  unique  synclogfileFilePathFileName (filePath,fileName,syncLogId),
  foreign key (syncLogId) references lab_rpki_sync_log(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='recored sync log for cer/roa/mft/crl'
`,

	`
CREATE TABLE lab_rpki_sync_rrdp_log (
  id int(10) unsigned not null primary key auto_increment,
  syncLogId int(10) unsigned not null  COMMENT 'foreign key  references lab_rpki_sync_log(id)',
  notifyUrl varchar(512) NOT NULL  COMMENT 'notification.xml url',
  sessionId varchar(512) not null  COMMENT 'session_id',
  lastSerial int(10) unsigned  COMMENT 'last serial',
  curSerial int(10) unsigned not null  COMMENT 'current serial',
  rrdpTime datetime NOT NULL   COMMENT 'rrdp time',
  rrdpType varchar(16) NOT NULL COMMENT 'snapshot/delta' ,
  foreign key (syncLogId) references lab_rpki_sync_log(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='recored notification.xml update log'
`,

	`
##################
## RTR
##################
CREATE TABLE lab_rpki_rtr_session (
  sessionId int(10) unsigned not null primary key   COMMENT 'sessionId, after init will not change',
  createTime datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='rtr session'
`,

	`
# insert random sessionId 
INSERT INTO lab_rpki_rtr_session(sessionId, createTime) VALUES(ROUND(RAND() * 999 + 99), NOW())
`,

	`
## serialNumber should not be auto_increment, because it will be wraped
CREATE TABLE lab_rpki_rtr_serial_number (
  id bigint(20) unsigned not null primary key auto_increment,
  serialNumber bigint(20) unsigned not null ,
  createTime datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='after every sync repo, serial num  will generate new serialnumber'
`,

	`

CREATE TABLE lab_rpki_rtr_full (
  id int(10) unsigned not null primary key auto_increment,
  serialNumber bigint(20) unsigned not null,
  asn bigint(20) signed not null,
  address  varchar(512) not null COMMENT 'address : 147.28.83.0 ',
  prefixLength  int(10) unsigned not null,
  maxLength int(10) unsigned not null,
  sourceFrom  json not null comment 'come from : {souce:sync/slurm/transfer,syncLogId/syncLogFileId/slurmId/slurmFileId/transferLogId}',
  unique  rtrFullSerialNumberAsnAddressPrefixLengthMaxLength (serialNumber , asn,address,prefixLength,maxLength)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='after every sync repo, will insert all full'
`,

	`
CREATE TABLE lab_rpki_rtr_full_log (
  id int(10) unsigned not null primary key auto_increment,
  serialNumber bigint(20) unsigned not null,
  asn bigint(20) signed not null,
  address  varchar(512) not null COMMENT 'address : 147.28.83.0 ',
  prefixLength  int(10) unsigned not null,
  maxLength int(10) unsigned not null,
  sourceFrom  json not null comment 'come from : {souce:sync/slurm/transfer,syncLogId/syncLogFileId/slurmId/slurmFileId/transferLogId}'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='full rtr log history'
`,

	`
CREATE TABLE lab_rpki_rtr_incremental (
  id int(10) unsigned not null primary key auto_increment,
  serialNumber bigint(20) unsigned not null,
  style  varchar(16) not null  comment 'announce/withdraw, is 1/0 in protocol',
  asn bigint(20) signed not null,
  address  varchar(512) not null COMMENT 'address : 147.28.83.0 ',
  prefixLength  int(10) unsigned not null,
  maxLength int(10) unsigned not null,
  sourceFrom  json not null comment 'come from : {souce:sync/slurm/transfer,syncLogId/syncLogFileId/slurmId/slurmFileId/transferLogId}',
  unique  rtrIncrementalSerialNumberAsnAddrPrefixMaxStyle (serialNumber , asn,address,prefixLength,maxLength,style)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COLLATE=utf8mb4_bin comment='after every sync repo, will insert all full'
`,

	`
################################
##  SLURM
################################
CREATE TABLE lab_rpki_slurm (
  id int(10) unsigned not null primary key auto_increment,
  version  int(10) unsigned default 1,
  style varchar(128) NOT NULL COMMENT 'prefixFilter/bgpsecFilter/prefixAssertion/bgpsecAssertion',
  asn bigint(20) signed ,
  addressPrefix  varchar(512) COMMENT '198.51.100.0/24 or 2001:DB8::/32',
  maxLength  int(10) unsigned ,
  ski  varchar(256) COMMENT 'some base64 ski',
  routerPublicKey  varchar(256) COMMENT 'some base64 ski',  
  comment  varchar(256),
  slurmFileId int(10) unsigned not null  COMMENT 'lab_rpki_slurm_file.id',
  priority int(10) unsigned not null default 5  COMMENT '0-10, 0 is highest level, 10 is  lowest. default 5. the higher level users slurm will conver lower ',
  state json not null COMMENT '[rtr:notYet/finished]',
  unique  slurmAsnAddressPrefix_maxLength (asn,addressPrefix,maxLength)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4   COLLATE=utf8mb4_bin comment='support different zone slurm file'
`,

	`
CREATE TABLE lab_rpki_slurm_file (
  id int(10) unsigned not null primary key auto_increment,
  jsonAll  json not null  COMMENT 'slurm content',
  uploadTime  datetime NOT NULL,
  fileName varchar(128) NOT NULL ,
  priority int(10) unsigned not null default 5  COMMENT '0-10, 0 is highest level, 10 is  lowest. default 5. the higher level users slurm will conver lower '
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4   COLLATE=utf8mb4_bin comment='support different user upload different zone slurm file'
`,

	`
########################################
## stat: 
## 1. competetation result
########################################

##after every sync, will delete and re-caculate all competation result
##CREATE TABLE lab_rpki_stat_roa_competation (
##	id int(10) unsigned not null primary key auto_increment,
##	roaId int(10) unsigned not null  COMMENT 'REFERENCES lab_rpki_roa(id)',
##	htmlResult  mediumtext ,
##	jsonResult  json 
##) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4   COLLATE=utf8mb4_bin comment='roa competation result'
`,

	`
#####################
####  transfer
#####################
CREATE TABLE lab_rpki_transfer_target (
  id int(10) unsigned NOT NULL primary key auto_increment,
  protocol varchar(64) NOT NULL COMMENT 'http or https',
  address varchar(64) NOT NULL COMMENT 'IP or domain ',
  port    int(10) unsigned NOT NULL COMMENT 'port',
  targetType   varchar(64) NOT NULL COMMENT 'vc/rp',
  createTime datetime NOT NULL COMMENT 'create time',
  state varchar(16) NOT NULL DEFAULT 'valid'  COMMENT 'valid/invalid'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='linked server, as target. so every server(rp/vc) may have different targets'
`,

	`
CREATE TABLE lab_rpki_transfer_log (
  id int(10) unsigned NOT NULL primary key auto_increment,
  transferTargetId int(10) unsigned NOT NULL  COMMENT 'lab_rpki_transfer_target.id',
  operate   varchar(64) NOT NULL COMMENT 'all/update',
  transferTime datetime NOT NULL COMMENT 'update time',
  uuid varchar(64) NOT NULL COMMENT 'used to uniquely identify transfer data',
  content longtext   COMMENT 'transfer content',
  transferType varchar(64) NOT NULL COMMENT 'send/receive',
  result   varchar(64) COMMENT 'ok/fail',
  errMsg   varchar(256) COMMENT 'fail message'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='log distribute'
`,

	`
#####################
####  statistic
#####################
CREATE TABLE lab_rpki_statistic (
  id int(10) unsigned NOT NULL primary key auto_increment,
  rir varchar(64)  NOT NULL COMMENT 'which nic',
  cerFileCount json NOT NULL COMMENT 'cer Count',
  crlFileCount json NOT NULL COMMENT 'crl Count',
  mftFileCount json NOT NULL COMMENT 'mft Count',
  roaFileCount json NOT NULL COMMENT 'roa Count',
  repos json NOT NULL COMMENT 'repos, big json',
  sync json  NOT NULL COMMENT 'sync info'  
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='statis, update after every sync'
`,

	`
#####################
####  configure
#####################
CREATE TABLE lab_rpki_conf (
  id int(10) unsigned NOT NULL primary key auto_increment,
  systemName varchar(256) NOT NULL  COMMENT 'system name',
  parameterName varchar(256) NOT NULL  COMMENT 'parameter name',
  parameterValue varchar(256) NOT NULL COMMENT 'parameter value',
  parameterDefaultValue varchar(256) NOT NULL COMMENT 'parameter default value',
  updateTime datetime NOT NULL COMMENT 'update time'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='configuration of rpstir2'
`,

	`
#########################
## create view
#########################
CREATE VIEW lab_rpki_roa_ipaddress_view AS 
select r.id AS id,r.asn AS asn,i.addressPrefix AS addressPrefix,i.maxLength AS maxLength, 
    r.syncLogId AS syncLogId,r.syncLogFileId AS syncLogFileId  
from (lab_rpki_roa r join lab_rpki_roa_ipaddress i)  
where ((i.roaId = r.id) and (json_extract(r.state,'$.state') = 'valid')) order by i.id
`}

var resetSqls []string = []string{
	`truncate  table  lab_rpki_cer  `,
	`truncate  table  lab_rpki_cer_sia `,
	`truncate  table  lab_rpki_cer_aia  `,
	`truncate  table  lab_rpki_cer_crldp `,
	`truncate  table  lab_rpki_cer_ipaddress `,
	`truncate  table  lab_rpki_cer_asn  `,
	`truncate  table  lab_rpki_crl  `,
	`truncate  table  lab_rpki_crl_revoked_cert  `,
	`truncate  table  lab_rpki_mft  `,
	`truncate  table  lab_rpki_mft_sia  `,
	`truncate  table  lab_rpki_mft_aia  `,
	`truncate  table  lab_rpki_mft_file_hash  `,
	`truncate  table  lab_rpki_roa  `,
	`truncate  table  lab_rpki_roa_sia  `,
	`truncate  table  lab_rpki_roa_aia  `,
	`truncate  table  lab_rpki_roa_ipaddress  `,
	`truncate  table  lab_rpki_roa_ee_ipaddress  `,
	`truncate  table  lab_rpki_sync_log_file  `,
	`truncate  table  lab_rpki_sync_rrdp_log  `,
	`truncate  table  lab_rpki_sync_log  `,
	`truncate  table  lab_rpki_rtr_session  `,
	`truncate  table  lab_rpki_rtr_serial_number  `,
	`truncate  table  lab_rpki_rtr_full  `,
	`truncate  table  lab_rpki_rtr_full_log  `,
	`truncate  table  lab_rpki_rtr_incremental  `,
	`truncate  table  lab_rpki_slurm  `,
	`truncate  table  lab_rpki_slurm_file  `,
	`truncate  table  lab_rpki_statistic  `,
	//	`truncate  table  lab_rpki_stat_roa_competation  `,
	`truncate  table  lab_rpki_transfer_target  `,
	`truncate  table  lab_rpki_transfer_log  `,

	`optimize  table  lab_rpki_cer  `,
	`optimize  table  lab_rpki_cer_sia `,
	`optimize  table  lab_rpki_cer_aia  `,
	`optimize  table  lab_rpki_cer_crldp `,
	`optimize  table  lab_rpki_cer_ipaddress `,
	`optimize  table  lab_rpki_cer_asn  `,
	`optimize  table  lab_rpki_crl  `,
	`optimize  table  lab_rpki_crl_revoked_cert  `,
	`optimize  table  lab_rpki_mft  `,
	`optimize  table  lab_rpki_mft_sia  `,
	`optimize  table  lab_rpki_mft_aia  `,
	`optimize  table  lab_rpki_mft_file_hash  `,
	`optimize  table  lab_rpki_roa  `,
	`optimize  table  lab_rpki_roa_sia  `,
	`optimize  table  lab_rpki_roa_aia  `,
	`optimize  table  lab_rpki_roa_ipaddress  `,
	`optimize  table  lab_rpki_roa_ee_ipaddress  `,
	`optimize  table  lab_rpki_sync_log_file  `,
	`optimize  table  lab_rpki_sync_rrdp_log  `,
	`optimize  table  lab_rpki_sync_log  `,
	`optimize  table  lab_rpki_rtr_session  `,
	`optimize  table  lab_rpki_rtr_serial_number  `,
	`optimize  table  lab_rpki_rtr_full  `,
	`optimize  table  lab_rpki_rtr_full_log  `,
	`optimize  table  lab_rpki_rtr_incremental  `,
	`optimize  table  lab_rpki_slurm  `,
	`optimize  table  lab_rpki_slurm_file  `,
	`optimize  table  lab_rpki_statistic  `,
	//	`optimize  table  lab_rpki_stat_roa_competation  `,
	`optimize  table  lab_rpki_transfer_target  `,
	`optimize  table  lab_rpki_transfer_log  `}

// when isInit is true, then init all db. otherwise will reset all db
func InitResetDb(isInit bool) error {
	session, err := xormdb.NewSession()
	defer session.Close()

	//truncate all table
	err = initResetDb(session, isInit)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "truncateDb(): truncateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "truncateDb(): CommitSession fail", err)

	}
	return nil
}

// need to init sessionId when it is emtpy
func initResetDb(session *xorm.Session, isInit bool) error {
	defer func(session1 *xorm.Session) {
		sql := `set foreign_key_checks=1;`
		if _, err := session1.Exec(sql); err != nil {
			belogs.Error("initResetDb(): SET FOREING_KEY_CHECKS=1 fail", err)

		}
	}(session)

	sql := `set foreign_key_checks=0;`
	if _, err := session.Exec(sql); err != nil {
		belogs.Error("initResetDb(): SET FOREING_KEY_CHECKS=0 fail", err)
		return err
	}

	// delete rtr_session
	var sqls []string
	if isInit {
		sqls = intiSqls
	} else {
		sqls = resetSqls
	}
	for _, sq := range sqls {
		if _, err := session.Exec(sq); err != nil {
			belogs.Error("initResetDb():  "+sq+" fail", err)
			return err
		}
	}

	// generate new session random, insert lab_rpki_rtr_session
	rand.Seed(time.Now().UnixNano())
	rtrSession := model.LabRpkiRtrSession{}
	rtrSession.SessionId = uint64(rand.Intn(999) + 99)
	rtrSession.CreateTime = time.Now()
	belogs.Info("initResetDb():insert lab_rpki_rtr_session:  ", rtrSession)
	if _, err := session.Insert(&rtrSession); err != nil {
		belogs.Error("initResetDb():insert rtr_session fail", err)
		return err
	}

	return nil
}

func GetMaxSyncLog() (syncLog model.LabRpkiSyncLog, err error) {
	sql := `select id,rsyncState,diffState,parseValidateState ,chainValidateState,
	 rtrState,rrdpState,state,syncStyle from lab_rpki_sync_log order by id desc limit 1`
	has, err := xormdb.XormEngine.Sql(sql).Get(&syncLog)
	if err != nil {
		belogs.Error("GetMaxSyncLog():select from lab_rpki_sync_log, fail:", err)
		return syncLog, err
	}
	if !has {
		belogs.Error("GetMaxSyncLog(): syncLog no exist :")
		return syncLog, errors.New("syncLog is no exist")
	}
	belogs.Debug("GetMaxSyncLog():syncLog :", jsonutil.MarshalJson(syncLog))
	return syncLog, nil
}

func Results() (results sysmodel.Results, err error) {
	results.CerResult, err = result("lab_rpki_cer", "cer")
	if err != nil {
		belogs.Error("result():select lab_rpki_cer, fail:", err)
		return results, err
	}
	results.CrlResult, err = result("lab_rpki_crl", "crl")
	if err != nil {
		belogs.Error("result():select lab_rpki_crl , fail:", err)
		return results, err
	}
	results.MftResult, err = result("lab_rpki_mft", "mft")
	if err != nil {
		belogs.Error("result():select lab_rpki_mft, fail:", err)
		return results, err
	}
	results.RoaResult, err = result("lab_rpki_roa", "roa")
	if err != nil {
		belogs.Error("result():select lab_rpki_roa, fail:", err)
		return results, err
	}
	return results, nil
}

func result(table, fileType string) (result sysmodel.Result, err error) {
	sql :=
		`select al.count as allCount, va.count as validCount, wa.count as warnigCount, ia.count as invalidCount , '` + fileType + `' as fileType  from 
		(select count(*) as count from ` + table + ` c) al,
		(select count(*) as count from ` + table + ` c where c.state->>"$.state" ='valid' ) va,
		(select count(*) as count from ` + table + ` c where c.state->>"$.state" ='warning') wa,
		(select count(*) as count from ` + table + ` c where c.state->>"$.state" ='invalid') ia`
	has, err := xormdb.XormEngine.Sql(sql).Get(&result)
	if err != nil {
		belogs.Error("result():select count, fail:", table, err)
		return result, err
	}
	if !has {
		belogs.Error("result(): not get count, fail:", table)
		return result, errors.New("not get count")
	}
	belogs.Debug("result():result :", jsonutil.MarshalJson(result))
	return result, nil
}
