package sys

import (
	"errors"
	"math/rand"
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

var initSqls []string = []string{
	`drop table if exists lab_rpki_conf`,
	`drop table if exists lab_rpki_cer_aia`,
	`drop table if exists lab_rpki_cer_asn`,
	`drop table if exists lab_rpki_cer_crldp`,
	`drop table if exists lab_rpki_cer_ipaddress`,
	`drop table if exists lab_rpki_cer_sia`,
	`drop table if exists lab_rpki_cer`,
	`drop table if exists lab_rpki_crl_revoked_cert`,
	`drop table if exists lab_rpki_crl`,
	`drop table if exists lab_rpki_mft_aia`,
	`drop table if exists lab_rpki_mft_file_hash`,
	`drop table if exists lab_rpki_mft_sia`,
	`drop table if exists lab_rpki_mft`,
	`drop table if exists lab_rpki_roa_aia`,
	`drop table if exists lab_rpki_roa_ee_ipaddress`,
	`drop table if exists lab_rpki_roa_ipaddress`,
	`drop table if exists lab_rpki_roa_sia`,
	`drop table if exists lab_rpki_roa`,
	`drop table if exists lab_rpki_asa_provider_asn`,
	`drop table if exists lab_rpki_asa_customer_asn`,
	`drop table if exists lab_rpki_asa_aia`,
	`drop table if exists lab_rpki_asa_sia`,
	`drop table if exists lab_rpki_asa`,
	`drop table if exists lab_rpki_rtr_full_log`,
	`drop table if exists lab_rpki_rtr_full`,
	`drop table if exists lab_rpki_rtr_incremental`,
	`drop table if exists lab_rpki_rtr_asa_full_log`,
	`drop table if exists lab_rpki_rtr_asa_full`,
	`drop table if exists lab_rpki_rtr_asa_incremental`,
	`drop table if exists lab_rpki_rtr_serial_number`,
	`drop table if exists lab_rpki_rtr_session`,
	`drop table if exists lab_rpki_slurm`,
	`drop table if exists lab_rpki_sync_log_file`,
	`drop table if exists lab_rpki_sync_log`,
	`drop table if exists lab_rpki_sync_rrdp_log`,
	`drop table if exists lab_rpki_sync_url`,
	`drop view if exists lab_rpki_crl_revoked_cert_view`,
	`drop view if exists lab_rpki_mft_file_hash_view`,
	`drop view if exists lab_rpki_roa_ipaddress_count_view`,
	`drop view if exists lab_rpki_roa_ipaddress_view`,
	`drop view if exists lab_rpki_sync_rrdp_log_maxid_view`,

	`
#################################
## main table for cer/crl/roa/mft
#################################	
CREATE TABLE lab_rpki_cer (
	id int(10) unsigned not null primary key auto_increment,
	sn varchar(1024) NOT NULL,
	notBefore datetime NOT NULL,
	notAfter datetime NOT NULL,
	subject varchar(1024) ,
	issuer varchar(1024) ,
	ski varchar(128) ,
	aki varchar(128) ,
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL ,
	state json comment 'state info in json',
	jsonAll json not null comment 'all cer info in json',
	chainCerts json comment 'chain certs(cer/crl/mft/roa) in json',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncLogFileId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log_file(id)',
	updateTime datetime NOT NULL,
	fileHash varchar(512) NOT NULL ,
	origin json comment 'origin(rir->repo) in json',
	key ski (ski),
	key aki (aki),
	key filePath (filePath(256)),
	key fileName (fileName),
	key syncLogId (syncLogId),
	key syncLogFileId (syncLogFileId),
	unique cerFilePathFileName (filePath(256),fileName),
	unique cerSkiFilePath (ski,filePath(256))
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='main cer table'
`,

	`
CREATE TABLE lab_rpki_cer_sia (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	rpkiManifest varchar(512) comment 'mft sync url',
	rpkiNotify varchar(512),
	caRepository varchar(512) comment 'ca repository url(directory)',
	signedObject varchar(512) ,
	FOREIGN key (cerid) REFERENCES lab_rpki_cer(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='cer sia'
`,

	`
CREATE TABLE lab_rpki_cer_aia (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	caIssuers varchar(512) comment 'father ca url (cer file)',
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='cer aia'
`,

	`
CREATE TABLE lab_rpki_cer_crldp (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	crldp varchar(512) comment 'crl sync url(file)',
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='cer crl'
`,

	`
CREATE TABLE lab_rpki_cer_ipaddress (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	addressFamily int(10) unsigned not null,
	addressPrefix varchar(512) comment 'address prefix: 147.28.83.0/24 ',
	min varchar(512) comment 'min address: 99.96.0.0',
	max varchar(512) comment 'max address: 99.105.127.255',
	rangeStart varchar(512) comment 'min address range from addressPrefix or min/max, in hex:  63.60.00.00',
	rangeEnd varchar(512) comment 'max address range from addressPrefix or min/max, in hex:  63.69.7f.ff',
	addressPrefixRange json comment 'min--max, such as 192.0.2.0--192.0.2.130, will convert to addressprefix range in json:{192.0.2.0/25, 192.0.2.128/31, 192.0.2.130/32}',
	key addressPrefix (addressPrefix),
	key rangeStart (rangeStart),
	key rangeEnd (rangeEnd),
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='cer ip address range'
`,
	`
## because 0 of asn has special meaning, so the default of asn is -1, and is "bigint signed" in mysql
CREATE TABLE lab_rpki_cer_asn (
	id int(10) unsigned not null primary key auto_increment,
	cerId int(10) unsigned not null,
	asn bigint(20) signed,
	min bigint(20) signed,
	max bigint(20) signed,
	key asn (asn),
	foreign key (cerId) references lab_rpki_cer(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='cer asn range'
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
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL ,
	state json comment 'state info in json',
	jsonAll json NOT NULL,
	chainCerts json comment 'chain certs(cer/crl/mft/roa) in json',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncLogFileId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log_file(id)',
	updateTime datetime NOT NULL,
	fileHash varchar(512) NOT NULL ,
	origin json comment 'origin(rir->repo) in json',
	key aki (aki),
	key filePath (filePath(256)), 
	key fileName (fileName),
	key syncLogId (syncLogId),
	key syncLogFileId (syncLogFileId), 
	unique crlFilePathFileName (filePath(256),fileName)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='crl '
`,

	`
CREATE TABLE lab_rpki_crl_revoked_cert (
	id int(10) unsigned not null primary key auto_increment,
	crlId int(10) unsigned not null,
	sn varchar(512) NOT NULL,
	revocationTime datetime NOT NULL,
	key sn (sn),
	foreign key (crlId) references lab_rpki_crl(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='all sn and revocationTime in crl'
`,

	`
###### manifest
CREATE TABLE lab_rpki_mft (
	id int(10) unsigned not null primary key auto_increment,
	mftNumber varchar(1024) NOT NULL,
	thisUpdate datetime NOT NULL,
	nextUpdate datetime NOT NULL,
	ski varchar(128) ,
	aki varchar(128) ,
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL ,
	state json comment 'state info in json',
	jsonAll json NOT NULL,
	chainCerts json comment 'chain certs(cer/crl/mft/roa) in json',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncLogFileId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log_file(id)',
	updateTime datetime NOT NULL,
	fileHash varchar(512) NOT NULL ,
	origin json comment 'origin(rir->repo) in json',
	key ski (ski),
	key aki (aki),
	key filePath (filePath(256)), 
	key fileName (fileName),
	key syncLogId (syncLogId),
	key syncLogFileId (syncLogFileId), 
	unique mftFilePathFileName (filePath(256),fileName),
	unique mftSkiFilePath (ski,filePath(256)) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='manifest'
`,

	`
CREATE TABLE lab_rpki_mft_sia (
	id int(10) unsigned not null primary key auto_increment,
	mftId int(10) unsigned not null,
	rpkiManifest varchar(512) ,
	rpkiNotify varchar(512) ,
	caRepository varchar(512) ,
	signedObject varchar(512) ,
	FOREIGN key (mftId) REFERENCES lab_rpki_mft(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='mft sia'
`,

	`
CREATE TABLE lab_rpki_mft_aia (
	id int(10) unsigned not null primary key auto_increment,
	mftId int(10) unsigned not null,
	caIssuers varchar(512) ,
	foreign key (mftId) references lab_rpki_mft(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='mft aia'
`,

	`
CREATE TABLE lab_rpki_mft_file_hash (
	id int(10) unsigned not null primary key auto_increment,
	mftId int(10) unsigned not null,
	file varchar(1024),
	hash varchar(1024),
	foreign key (mftId) references lab_rpki_mft(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='files in manifest'
`,

	`
###### roa
## because 0 of asn has special meaning, so the default of asn is -1, and is "bigint signed" in mysql
CREATE TABLE lab_rpki_roa (
	id int(10) unsigned not null primary key auto_increment,
	asn bigint(20) signed not null,
	ski varchar(128) ,
	aki varchar(128) ,
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL ,
	state json comment 'state info in json',
	jsonAll json NOT NULL,
	chainCerts json comment 'chain certs(cer/crl/mft/roa) in json',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncLogFileId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log_file(id)',
	updateTime datetime NOT NULL,
	fileHash varchar(512) NOT NULL ,
	origin json comment 'origin(rir->repo) in json',
	key ski (ski),
	key aki (aki),
	key filePath (filePath(256)),
	key fileName (fileName),
	key syncLogId (syncLogId),
	key syncLogFileId (syncLogFileId), 
	unique roaFilePathFileName (filePath(256),fileName),
	unique roaSkiFilePath (ski,filePath(256)) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa info'
`,

	`	
CREATE TABLE lab_rpki_roa_sia (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	rpkiManifest varchar(512) ,
	rpkiNotify varchar(512) ,
	caRepository varchar(512) ,
	signedObject varchar(512) ,
	FOREIGN key (roaId) REFERENCES lab_rpki_roa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa sia'
`,

	`
CREATE TABLE lab_rpki_roa_aia (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	caIssuers varchar(512) ,
	foreign key (roaId) references lab_rpki_roa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa aia'
`,

	`
CREATE TABLE lab_rpki_roa_ipaddress (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	addressFamily int(10) unsigned not null,
	addressPrefix varchar(512),
	maxLength int(10) unsigned,
	rangeStart varchar(512),
	rangeEnd varchar(512),
	addressPrefixRange json comment 'min--max, such as 192.0.2.0--192.0.2.130, will convert to addressprefix range in json:{192.0.2.0/25, 192.0.2.128/31, 192.0.2.130/32}',
	key addressPrefix (addressPrefix),
	key rangeStart (rangeStart),
	key rangeEnd (rangeEnd),
	foreign key (roaId) references lab_rpki_roa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa ip prefix'
`,

	`
CREATE TABLE lab_rpki_roa_ee_ipaddress (
	id int(10) unsigned not null primary key auto_increment,
	roaId int(10) unsigned not null,
	addressFamily int(10) unsigned not null,
	addressPrefix varchar(512) comment 'address prefix: 147.28.83.0/24 ',
	min varchar(512) comment 'min address: 99.96.0.0',
	max varchar(512) comment 'max address: 99.105.127.255',
	rangeStart varchar(512) comment 'min address range from addressPrefix or min/max, in hex:  63.60.00.00',
	rangeEnd varchar(512) comment 'max address range from addressPrefix or min/max, in hex:  63.69.7f.ff',
	addressPrefixRange json comment 'min--max, such as 192.0.2.0--192.0.2.130, will convert to addressprefix range in json:{192.0.2.0/25, 192.0.2.128/31, 192.0.2.130/32}',
	key addressPrefix (addressPrefix),
	key rangeStart (rangeStart),
	key rangeEnd (rangeEnd),
	foreign key (roaId) references lab_rpki_roa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='roa ee ip prefix'
`,

	`
###### asa
CREATE TABLE lab_rpki_asa (
	id int(10) unsigned not null primary key auto_increment,
	ski varchar(128) ,
	aki varchar(128) ,
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL ,
	state json comment 'state info in json',
	jsonAll json NOT NULL,
	chainCerts json comment 'chain certs(cer/crl/mft/roa/asa) in json',
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncLogFileId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log_file(id)',
	updateTime datetime NOT NULL,
	fileHash varchar(512) NOT NULL ,
	origin json comment 'origin(rir->repo) in json',
	key ski (ski),
	key aki (aki),
	key filePath (filePath(256)),
	key fileName (fileName),
	key syncLogId (syncLogId),
	key syncLogFileId (syncLogFileId), 
	unique asaFilePathFileName (filePath(256),fileName),
	unique asaSkiFilePath (ski,filePath(256)) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='asa info'
`,

	`	
CREATE TABLE lab_rpki_asa_sia (
	id int(10) unsigned not null primary key auto_increment,
	asaId int(10) unsigned not null,
	rpkiManifest varchar(512) ,
	rpkiNotify varchar(512) ,
	caRepository varchar(512) ,
	signedObject varchar(512) ,
	FOREIGN key (asaId) REFERENCES lab_rpki_asa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='asa sia'
`,

	`
CREATE TABLE lab_rpki_asa_aia (
	id int(10) unsigned not null primary key auto_increment,
	asaId int(10) unsigned not null,
	caIssuers varchar(512) ,
	foreign key (asaId) references lab_rpki_asa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='asa aia'
`,

	`
##### one asa may have many customerAsn
CREATE TABLE lab_rpki_asa_customer_asn (
	id int(10) unsigned not null primary key auto_increment,
	asaId int(10) unsigned not null,
	customerAsn int(10) unsigned not null,
	addressFamily int(10) unsigned,
	key customerAsn (customerAsn),
	foreign key (asaId) references lab_rpki_asa(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='asa customerAsn'
`,

	`
##### one customerAsn may have many providerAsn
CREATE TABLE lab_rpki_asa_provider_asn (
	id int(10) unsigned not null primary key auto_increment,
	asaId int(10) unsigned not null,
	customerAsnId int(10) unsigned not null,
	providerAsn int(10) unsigned not null,
	addressFamily int(10) unsigned,
	providerOrder int(10) unsigned not null,
	key providerAsn (providerAsn),
	foreign key (asaId) references lab_rpki_asa(id),
	foreign key (customerAsnId) references lab_rpki_asa_customer_asn(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='asa providerAsn'
`,

	`
################################################
## recored every sync log for cer/crl/roa/mft
################################################
CREATE TABLE lab_rpki_sync_log (
	id int(10) unsigned not null primary key auto_increment,
	syncState MEDIUMTEXT,
	parseValidateState MEDIUMTEXT,
	chainValidateState MEDIUMTEXT,
	rtrState MEDIUMTEXT,
	state varchar(16) not null comment 'rsyncing/rsynced ddrping/ddrped  diffing/diffed   parsevalidating/parsevalidated   rtring/rtred idle',
	syncStyle varchar(16) not null comment 'rsync/rrdp' 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='recored every sync log'
`,

	`
CREATE TABLE lab_rpki_sync_log_file (
	id int(10) unsigned not null primary key auto_increment,
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	syncTime datetime not null comment 'sync time for every file',
	syncStyle varchar(16) not null comment 'rrdp/rsync' , 
	syncType varchar(16) not null comment 'add/del/update' ,
	fileType varchar(16) not null comment 'cer/roa/mft/crl/',
	filePath varchar(1024) NOT NULL ,
	fileName varchar(128) NOT NULL ,
	sourceUrl varchar(512) , 
	jsonAll json comment 'cert json info from cer/crl/mft/roa.jsonAll' ,
	fileHash varchar(512) ,
	state json comment '{"sync":"finished","updateCertTable":"notYet/finished"}: have synced ,have published to main table',
	key fileType (fileType),
	key syncType (syncType),
	key filePath (filePath(256)),
	key fileName (fileName),
	unique synclogfileFilePathFileName (filePath(256),fileName,syncLogId),
	foreign key (syncLogId) references lab_rpki_sync_log(id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='recored sync log for cer/roa/mft/crl'
`,

	`
CREATE TABLE lab_rpki_sync_rrdp_log (
	id int(10) unsigned not null primary key auto_increment,
	syncLogId int(10) unsigned not null comment 'foreign key references lab_rpki_sync_log(id)',
	notifyUrl varchar(512) not null comment 'notification.xml url',
	sessionId varchar(512) not null comment 'session_id',
	lastSerial int(10) unsigned comment 'last serial',
	curSerial int(10) unsigned not null comment 'current serial',
	rrdpTime datetime not null comment 'rrdp time',
	rrdpType varchar(16) not null comment 'snapshot/delta' ,
	snapshotOrDeltaUrl varchar(256) not null comment 'snapshot/delta url' ,
	foreign key (syncLogId) references lab_rpki_sync_log(id),
	index notifyUrl (notifyUrl) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='recored notification.xml update log'
`,

	`
CREATE TABLE lab_rpki_sync_url (
	id int(10) unsigned not null primary key auto_increment,
	rrdpUrl varchar(256) comment 'rrdp url',
	rrdpUrlState json comment '{state:valid/invalid}', 
	rrdpUpdateTime datetime comment 'rrdp update time', 
	rsyncUrls varchar(512) not null comment 'rsync url',
	rsyncUrlState json not null comment '{state:valid/invalid}', 
	rsyncUpdateTime datetime not null comment 'rsync update time',
	addTime datetime not null comment 'add time',
	index rrdpUrl (rrdpUrl),
	unique syncUrlRrdpUrlRsyncUrls (rrdpUrl , rsyncUrls) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='all rsync/rrdp url'
`,

	`
##################
## RTR
##################
CREATE TABLE lab_rpki_rtr_session (
	sessionId int(10) unsigned not null primary key comment 'sessionId, after init will not change',
	createTime datetime NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rtr session'
`,

	`
# insert random sessionId 
INSERT INTO lab_rpki_rtr_session(sessionId, createTime) VALUES(ROUND(RAND() * 999 + 99), NOW())
`,

	`
## serialNumber should not be auto_increment, because it will be wraped
CREATE TABLE lab_rpki_rtr_serial_number (
	id bigint(20) unsigned not null primary key auto_increment comment 'id',
	serialNumber bigint(20) unsigned not null comment 'serialNumber for rtr_full, rtr_incremental',
	globalSerialNumber bigint(20) unsigned not null comment 'serialNumber for center vc update by sync and slurm',
	subpartSerialNumber bigint(20) unsigned not null comment 'serialNumber for sub vc update by slurm',
	createTime datetime NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='after every sync repo, serial num  will generate new serialnumber'
`,

	`

CREATE TABLE lab_rpki_rtr_full (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	asn bigint(20) signed not null,
	address varchar(512) not null comment 'address : 147.28.83 ',
	prefixLength int(10) unsigned not null,
	maxLength int(10) unsigned not null,
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key asn(asn),
	key address(address),
	key prefixLength(prefixLength),
	key maxLength(maxLength),
	unique rtrFullSerialNumberAsnAddressPrefixLengthMaxLength (serialNumber , asn,address,prefixLength,maxLength)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='after every sync repo, will insert all full'
`,

	`
CREATE TABLE lab_rpki_rtr_full_log (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	asn bigint(20) signed not null,
	address varchar(512) not null comment 'address : 147.28.83 ',
	prefixLength int(10) unsigned not null,
	maxLength int(10) unsigned not null,
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key asn(asn),
	key address(address),
	key prefixLength(prefixLength),
	key maxLength(maxLength) 
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='full rtr log history'
`,

	`
CREATE TABLE lab_rpki_rtr_incremental (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	style varchar(16) not null comment 'announce/withdraw, is 1/0 in protocol',
	asn bigint(20) signed not null,
	address varchar(512) not null comment 'address : 147.28.83 ',
	prefixLength int(10) unsigned not null,
	maxLength int(10) unsigned not null,
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key asn(asn),
	key address(address),
	key prefixLength(prefixLength),
	key maxLength(maxLength),
	unique rtrIncrementalSerialNumberAsnAddrPrefixMaxStyle (serialNumber , asn,address,prefixLength,maxLength,style)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='incremental rtr'
`,

	`
CREATE TABLE lab_rpki_rtr_asa_full (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	addressFamily int(10) unsigned,
	customerAsn int(10) unsigned not null comment 'customer asn',
	providerAsns varchar(255) comment '[{"providerAsn":65000},{"providerAsn":65001},{"providerAsn":65002}]',
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key customerAsn(customerAsn),
	unique rtrAsaFullSerialNumberCustomerAsnProviderAsns(serialNumber,customerAsn,providerAsns)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='full rtr asa'
`,

	`
CREATE TABLE lab_rpki_rtr_asa_full_log (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	addressFamily int(10) unsigned,
	customerAsn int(10) unsigned not null comment 'customer asn',
	providerAsns varchar(255) comment '[{"providerAsn":65000},{"providerAsn":65001},{"providerAsn":65002}]',
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key customerAsn(customerAsn)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='full rtr asa log history'
`,

	`
CREATE TABLE lab_rpki_rtr_asa_incremental (
	id int(10) unsigned not null primary key auto_increment,
	serialNumber bigint(20) unsigned not null,
	style varchar(16) not null comment 'announce/withdraw, is 1/0 in protocol',
	addressFamily int(10) unsigned,
	customerAsn int(10) unsigned not null comment 'customer asn',
	providerAsns varchar(255) comment '[{"providerAsn":65000},{"providerAsn":65001},{"providerAsn":65002}]',
	sourceFrom json not null comment 'come from : {souce:sync/slurm/rush,syncLogId/syncLogFileId/slurmId/slurmFileId/rushDataLogId}',
	key serialNumber(serialNumber),
	key customerAsn(customerAsn),
	unique rtrIncrementalSerialNumberCustomerAsnProviderAsns(serialNumber,customerAsn,providerAsns)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='incremental rtr asa'
`,

	`
################################
## SLURM
################################
CREATE TABLE lab_rpki_slurm (
	id int(10) unsigned not null primary key auto_increment,
	version int(10) unsigned default 1,
	style varchar(128) not null comment 'prefixFilter/bgpsecFilter/prefixAssertion/bgpsecAssertion',
	asn bigint(20) signed ,
	addressPrefix varchar(512) comment '198.51.100.0/24 or 2001:DB8::/32',
	maxLength int(10) unsigned ,
	ski varchar(256) comment 'some base64 ski',
	routerPublicKey varchar(256) comment 'some base64 ski', 
	comment varchar(256),
	treatLevel varchar(64) comment 'critical/major/normal',
	slurmLogId int(10) unsigned not null comment 'lab_rpki_slurm_log.id',
	slurmLogFileId int(10) unsigned not null comment 'lab_rpki_slurm_log_file.id',
	state json not null comment '[rtr:notYet/finished]',
	unique slurmAsnAddressPrefix_maxLength (asn,addressPrefix,maxLength)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='valid slurms'
`,

	` 
#####################
#### conf
#####################
CREATE TABLE lab_rpki_conf (
	id int(10) unsigned NOT NULL primary key auto_increment,
	section varchar(128) not null comment 'section',
	myKey varchar(128) not null comment 'key',
	myValue varchar(1024) not null comment 'value',
	defaultMyValue varchar(1024) not null comment 'default value',
	updateTime datetime not null comment 'update time',
	unique sectionMyKey (section,myKey)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='rpstir2 configuration'
 

`,

	`
#########################
## create view roa ipaddress 
#########################
CREATE VIEW lab_rpki_roa_ipaddress_view AS 
select r.id AS id, 
	r.asn AS asn, 
	i.addressPrefix AS addressPrefix, 
	i.maxLength AS maxLength, 
	r.syncLogId AS syncLogId, 
	r.syncLogFileId AS syncLogFileId, 
	r.origin->>'$.rir' as rir, 
	r.origin->>'$.repo' as repo  
from lab_rpki_roa r join lab_rpki_roa_ipaddress i 
where i.roaId = r.id and 
	r.state->>'$.state' in ('valid','warning') 
order by 
	r.origin->>'$.rir', 
	r.origin->>'$.repo', 
	i.addressPrefix, 
	i.maxLength, 
	r.asn, 
	r.id 
`,

	`
#########################
## create view crl revoked sn
#########################
CREATE VIEW lab_rpki_crl_revoked_cert_view AS  
select l.id, l.fileName, l.aki , r.revocationTime, r.sn
from lab_rpki_crl l, lab_rpki_crl_revoked_cert r 
where l.id = r.crlId order by l.id
`,

	`
#########################
## create view mft file hash
#########################
CREATE VIEW lab_rpki_mft_file_hash_view AS 
SELECT	m.id as mftId,	m.aki as aki,	fh.id as mftFileHashId,	fh.file as file, fh.hash as hash 
FROM lab_rpki_mft m , lab_rpki_mft_file_hash fh 
WHERE m.id = fh.mftId 
ORDER BY m.id, fh.id
`,

	`
#########################
## create view roaIpAddressCount
#########################
CREATE VIEW lab_rpki_roa_ipaddress_count_view AS 
select roaId, count(*) as roaIpAddressCount 
from lab_rpki_roa_ipaddress 
group by roaId order by roaIpAddressCount 
`,
	`
#########################
## create view roaIpAddressCount
#########################
CREATE VIEW lab_rpki_sync_rrdp_log_maxid_view AS 
select max(cc.id) AS maxId from lab_rpki_sync_rrdp_log cc group by cc.notifyUrl order by cc.notifyUrl 
`,
}

var fullSyncSqls []string = []string{
	`truncate  table  lab_rpki_cer`,
	`truncate  table  lab_rpki_cer_sia`,
	`truncate  table  lab_rpki_cer_aia`,
	`truncate  table  lab_rpki_cer_crldp`,
	`truncate  table  lab_rpki_cer_ipaddress`,
	`truncate  table  lab_rpki_cer_asn`,
	`truncate  table  lab_rpki_crl`,
	`truncate  table  lab_rpki_crl_revoked_cert`,
	`truncate  table  lab_rpki_mft`,
	`truncate  table  lab_rpki_mft_sia`,
	`truncate  table  lab_rpki_mft_aia`,
	`truncate  table  lab_rpki_mft_file_hash`,
	`truncate  table  lab_rpki_roa`,
	`truncate  table  lab_rpki_roa_sia`,
	`truncate  table  lab_rpki_roa_aia`,
	`truncate  table  lab_rpki_roa_ipaddress`,
	`truncate  table  lab_rpki_roa_ee_ipaddress`,
	`truncate  table  lab_rpki_asa`,
	`truncate  table  lab_rpki_asa_sia`,
	`truncate  table  lab_rpki_asa_aia`,
	`truncate  table  lab_rpki_asa_customer_asn`,
	`truncate  table  lab_rpki_asa_provider_asn`,
	`truncate  table  lab_rpki_sync_rrdp_log`,
	`truncate  table  lab_rpki_sync_log_file`,
	`truncate  table  lab_rpki_sync_log`,
	`truncate  table  lab_rpki_sync_url`,
}
var resetAllOtherSqls []string = []string{
	`truncate  table  lab_rpki_conf`,
	`truncate  table  lab_rpki_rtr_session`,
	`truncate  table  lab_rpki_rtr_serial_number`,
	`truncate  table  lab_rpki_rtr_full`,
	`truncate  table  lab_rpki_rtr_full_log`,
	`truncate  table  lab_rpki_rtr_incremental`,
	`truncate  table  lab_rpki_rtr_asa_full`,
	`truncate  table  lab_rpki_rtr_asa_full_log`,
	`truncate  table  lab_rpki_rtr_asa_incremental`,
	`truncate  table  lab_rpki_slurm`,
}

var optimizeSqls []string = []string{
	`optimize  table  lab_rpki_cer`,
	`optimize  table  lab_rpki_cer_sia`,
	`optimize  table  lab_rpki_cer_aia`,
	`optimize  table  lab_rpki_cer_crldp`,
	`optimize  table  lab_rpki_cer_ipaddress`,
	`optimize  table  lab_rpki_cer_asn`,
	`optimize  table  lab_rpki_crl`,
	`optimize  table  lab_rpki_crl_revoked_cert`,
	`optimize  table  lab_rpki_mft`,
	`optimize  table  lab_rpki_mft_sia`,
	`optimize  table  lab_rpki_mft_aia`,
	`optimize  table  lab_rpki_mft_file_hash`,
	`optimize  table  lab_rpki_roa`,
	`optimize  table  lab_rpki_roa_sia`,
	`optimize  table  lab_rpki_roa_aia`,
	`optimize  table  lab_rpki_roa_ipaddress`,
	`optimize  table  lab_rpki_roa_ee_ipaddress`,
	`optimize  table  lab_rpki_asa`,
	`optimize  table  lab_rpki_asa_sia`,
	`optimize  table  lab_rpki_asa_aia`,
	`optimize  table  lab_rpki_asa_customer_asn`,
	`optimize  table  lab_rpki_asa_provider_asn`,
	`optimize  table  lab_rpki_sync_log_file`,
	`optimize  table  lab_rpki_sync_rrdp_log`,
	`optimize  table  lab_rpki_sync_log`,
	`optimize  table  lab_rpki_sync_url`,
	`optimize  table  lab_rpki_rtr_session`,
	`optimize  table  lab_rpki_rtr_serial_number`,
	`optimize  table  lab_rpki_rtr_full`,
	`optimize  table  lab_rpki_rtr_full_log`,
	`optimize  table  lab_rpki_rtr_incremental`,
	`optimize  table  lab_rpki_rtr_asa_full`,
	`optimize  table  lab_rpki_rtr_asa_full_log`,
	`optimize  table  lab_rpki_rtr_asa_incremental`,
	`optimize  table  lab_rpki_slurm`,
}

// when isInit is true, then init all db. otherwise will reset all db
func initResetDb(sysStyle SysStyle) error {
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	//truncate all table
	err = initResetImplDb(session, sysStyle)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "initResetDb(): initResetImplDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "initResetDb(): CommitSession fail", err)
	}
	return nil
}

// need to init sessionId when it is empty
func initResetImplDb(session *xorm.Session, sysStyle SysStyle) error {
	defer func(session1 *xorm.Session) {
		sql := `set foreign_key_checks=1;`
		if _, err := session1.Exec(sql); err != nil {
			belogs.Error("initResetImplDb(): SET foreign_key_checks=1 fail", err)
			xormdb.RollbackAndLogError(session, "initResetImplDb():SET foreign_key_checks=1 fail", err)
		}
	}(session)

	start := time.Now()
	sql := `set foreign_key_checks=0;`
	if _, err := session.Exec(sql); err != nil {
		belogs.Error("initResetImplDb(): SET foreign_key_checks=0 fail", err)
		return xormdb.RollbackAndLogError(session, "initResetImplDb():SET foreign_key_checks=0 fail: ", err)
	}
	belogs.Debug("initResetImplDb():foreign_key_checks=0;   time(s):", time.Since(start))

	// delete rtr_session
	var sqls []string
	if sysStyle.SysStyle == "init" {
		sqls = initSqls
	} else if sysStyle.SysStyle == "fullsync" || sysStyle.SysStyle == "resetall" {
		sqls = fullSyncSqls
		if sysStyle.SysStyle == "resetall" {
			sqls = append(sqls, resetAllOtherSqls...)
		}
		sqls = append(sqls, optimizeSqls...)
	}
	belogs.Debug("initResetImplDb():will Exec sqls:", jsonutil.MarshalJson(sqls))
	belogs.Info("initResetImplDb():will Exec len(sqls):", len(sqls))
	for _, sq := range sqls {
		now := time.Now()
		if _, err := session.Exec(sq); err != nil {
			belogs.Error("initResetImplDb():  "+sq+" fail", err)
			return xormdb.RollbackAndLogError(session, "initResetImplDb():sql fail: "+sq, err)
		}
		belogs.Info("initResetImplDb(): sq:", sq, ", sql time(s):", time.Since(now))
	}
	belogs.Info("initResetImplDb(): len(sqls):", len(sqls), ",  time(s):", time.Since(start))

	// when resetall,
	if sysStyle.SysStyle == "resetall" {
		// generate new session random, insert lab_rpki_rtr_session
		rand.Seed(time.Now().UnixNano())
		rtrSession := model.LabRpkiRtrSession{}
		rtrSession.SessionId = uint64(rand.Intn(999) + 99)
		rtrSession.CreateTime = time.Now()
		belogs.Info("initResetImplDb():insert lab_rpki_rtr_session:  ", rtrSession)
		if _, err := session.Insert(&rtrSession); err != nil {
			belogs.Error("initResetImplDb():insert rtr_session fail", err)
			return xormdb.RollbackAndLogError(session, "initResetImplDb():insert rtr_session fail", err)
		}
	}
	if sysStyle.SysStyle == "init" || sysStyle.SysStyle == "resetall" {
		// insert lab_rpki_conf
		sql = `insert lab_rpki_conf ( section, myKey, myValue, defaultMyValue, updateTime) 
			values(?,?,?,?,?) `
		_, err := session.Exec(sql, "rpOperate", "cacheUpdateType", "manual", "manual", time.Now())
		if err != nil {
			belogs.Error("initResetImplDb(): insert lab_rpki_conf fail", err)
			return xormdb.RollbackAndLogError(session, "initResetImplDb():insert lab_rpki_conf fail", err)
		}
	}
	if err := session.Commit(); err != nil {
		return xormdb.RollbackAndLogError(session, "initResetImplDb():commit fail", err)
	}
	return nil
}

func getResultsDb() (results CertResults, err error) {
	results.CerResult, err = getResultDb("lab_rpki_cer", "cer")
	if err != nil {
		belogs.Error("getResultsDb():select lab_rpki_cer, fail:", err)
		return results, err
	}
	results.CrlResult, err = getResultDb("lab_rpki_crl", "crl")
	if err != nil {
		belogs.Error("getResultsDb():select lab_rpki_crl , fail:", err)
		return results, err
	}
	results.MftResult, err = getResultDb("lab_rpki_mft", "mft")
	if err != nil {
		belogs.Error("getResultsDb():select lab_rpki_mft, fail:", err)
		return results, err
	}
	results.RoaResult, err = getResultDb("lab_rpki_roa", "roa")
	if err != nil {
		belogs.Error("getResultsDb():select lab_rpki_roa, fail:", err)
		return results, err
	}
	return results, nil
}

func getResultDb(table, fileType string) (result CertResult, err error) {
	sql :=
		`select al.count as allCount, va.count as validCount, wa.count as warnigCount, ia.count as invalidCount , '` + fileType + `' as fileType  from 
		(select count(*) as count from ` + table + ` c) al,
		(select count(*) as count from ` + table + ` c where c.state->>"$.state" ='valid' ) va,
		(select count(*) as count from ` + table + ` c where c.state->>"$.state" ='warning') wa,
		(select count(*) as count from ` + table + ` c where c.state->>"$.state" ='invalid') ia`
	has, err := xormdb.XormEngine.SQL(sql).Get(&result)
	if err != nil {
		belogs.Error("getResultDb():select count, fail:", table, err)
		return result, err
	}
	if !has {
		belogs.Error("getResultDb(): not get count, fail:", table)
		return result, errors.New("not get count")
	}
	belogs.Debug("getResultDb():result :", jsonutil.MarshalJson(result))
	return result, nil
}

func exportRoasDb() (exportRoas []ExportRoa, err error) {
	sql :=
		`select asn, addressPrefix, maxLength, rir, repo 
		from lab_rpki_roa_ipaddress_view v
		order by rir, repo,addressPrefix,maxLength,asn`
	err = xormdb.XormEngine.SQL(sql).Find(&exportRoas)
	if err != nil {
		belogs.Error("exportRoasDb():Find, fail:", err)
		return nil, err
	}

	belogs.Debug("exportRoasDb():len(exportRoas):", len(exportRoas))
	return exportRoas, nil
}

func exportRtrForManrsDb() (rtrForManrss []RtrForManrs, err error) {
	rtrForManrss = make([]RtrForManrs, 0)
	sql :=
		`select asn, address, prefixLength,maxLength as max_length 
		from lab_rpki_rtr_full order by id `
	err = xormdb.XormEngine.SQL(sql).Find(&rtrForManrss)
	if err != nil {
		belogs.Error("exportRtrForManrsDb():Find, fail:", err)
		return nil, err
	}

	belogs.Debug("exportRtrForManrsDb():len(rtrForManrss):", len(rtrForManrss))
	return rtrForManrss, nil
}
