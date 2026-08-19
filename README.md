

# RPSTIR2
## 1. Introduction
RPKI is a hierarchical Public Key Infrastructure(PKI) that binds Internet Number Resources(INRs) such as Autonomous System Numbers(ASNs) and IP addresses to public keys via certificates. RPKI allows INR holder(certificate holder) to allocate certain IP prefix to their customers via issuing resource certificates(RCs) and authorizing an ASN to announce certain IP prefixes via issuing ROAs, and all of these RPKI objects are published in RPKI repository.

As the bridge between inter-domain routing system and RPKI repository, RPKI Relying Party(RP) is designed to assist BGP Speakers in synchronization of RPKI objects, validation of certificate chain, cache management and transmission of Validated ROA Payloads(VRPs).

RPSTIR2 is a kind of RP software written in GO, which based on design idea of RPSTIR, provides all the standard functions mentioned above. RPSTIR2 also supports more RPKI-related protocols and optimizes performance.

RPSTIR2 is capable of running on CentOS7(64bit)/Ubuntu18(64bit) or higher.
&nbsp;

## 2. Install RPSTIR2

### 2.1 Install OpenSSL
OpenSSL version must be 1.1.1b or higher, and  "enable-rfc3779" needs to be set when compiling OpenSSL.

```shell
$ mkdir -p /home/openssl 
$ cd /home/openssl
$ wget  https://github.com/openssl/openssl/releases/download/openssl-3.5.7/openssl-3.5.7.tar.gz
$ tar xzvf openssl-3.5.7.tar.gz
$ cd openssl-3.5.7
$ ./config  shared enable-rfc3779 --prefix=/home/openssl/openssl -Wl,-rpath,'/home/openssl/openssl/lib64'
$ make
$ make install
$ /home/openssl/openssl/bin/openssl version
$ OpenSSL 3.5.7
```

### 2.2 Install MySQL
You can download and install MySQL from https://dev.mysql.com/downloads/ according to your platform. MySQL version must be 8 or higher and should support JSON. You should login in MySQL as root, and create user accounts and database of RPSTIR2. 

```mysql
CREATE USER 'rpstir2'@'localhost' IDENTIFIED WITH mysql_native_password BY 'Rpstir-123';
CREATE USER 'rpstir2'@'%' IDENTIFIED WITH mysql_native_password BY 'Rpstir-123';
flush privileges;

CREATE DATABASE rpstir2;
GRANT ALL PRIVILEGES ON rpstir2.* TO 'rpstir2'@'localhost'  with grant option;
GRANT ALL PRIVILEGES ON rpstir2.* TO 'rpstir2'@'%'  with grant option;
flush privileges;
```

Note: You also can use docker to run MySQL, and make sure that the time zone of docker is the same as that of the host. 

### 2.3 Install GoLang(Optional)
If you plan to compile the program by yourself, you need to install a version of Golang higher than 1.17. Otherwise you don't need to install it.


### 2.4 Create RPSTIR2 directories
Before installing RPSTIR2, you should create directories in advance, one of which is for program and the other is for the cache data. you can modify the shell, and change "conf/project.conf". 

| Directory  | Path                      |
| :--------: | ------------------------- |
| programDir | /home/rpki/rpstir2        |
| dataDir    | /home/rpki/data           |


```shell
$ mkdir -p /home/rpki/ /home/rpki/rpstir2  /home/rpki/data  /home/rpki/data/rrdprepo  /home/rpki/data/rsyncrepo /home/rpki/data/tal
```

### 2.5 Download RPSTIR2 

```shell
$ cd /home/rpki/
$ git clone https://github.com/bgpsecurity/rpstir2.git 
$ cd /home/rpki/rpstir2/bin
$ chmod +x *
$ cp /home/rpki/rpstir2/build/tal/*  /home/rpki/data/tal/
```

### 2.6 Configure RPSTIR2
You can modify configuration parameters of programDir, dataDir, mysql, and  port in configuration file(/home/rpki/rpstir2/conf/project.conf). 

## 3 Running RPSTIR2
The RPSTIR2 must be started first, you can check for errors by looking at the log files in ./log/ directory.

### 3.1 Start and stop the RPSTIR2
You can check the log files in ./log/ to see whether the program is started successfully.

```shell
$ cd /home/rpki/rpstir2/bin
$./rpstir2.sh start 
```

```shell
$ cd /home/rpki/rpstir2/bin
$./rpstir2.sh stop 
```

### 3.2 Initialize the RPSTIR2
This command is used to initialize or reset the database. Please check the log files in ./log/ to see if the execution is successful.

```shell
$ cd /home/rpki/rpstir2/bin
$./rpstir2.sh start 
$./rpstir2.sh init 
```

### 3.3 Configure scheduled task
You can use crontab to perform scheduled synchronization tasks. Then RPSTIR2 will download RPKI objects, and complete the subsequent validation procedure according to the schedule you set. 

```shell
$ crontab -e
10 */4 * * *  /home/rpki/rpstir2/bin/rpstir2.sh sync
```
Note: The RPSTIR2 service must start first. 

### 3.4 Sync and validate RPKI objects
You can download RPKI objects with rsync or RRDP protocol, and complete the subsequent validation procedure. 

```shell
$ cd /home/rpki/rpstir2/bin
$ ./rpstir2.sh sync  
```

### 3.5 Get sync and validation status
Because rsync and RRDP take long time to run, they are executed in the background. So you need a command to determine if the synchronization and validation process is complete.

```shell
$ cd /home/rpki/rpstir2/bin
$ ./rpstir2.sh state   | jq .
```

When you get the following JSON message, if "isRunning" is "true", it means that sync and validation are still running; if it is "false", sync and validation complete. At this time, the router can obtain rpki data through RTR port.

```JSON
{
	"result": "ok",
	"msg": "",
	"data": {
		"startTime": "2026-01-01 01:01:01 CST",
		"isRunning": "false",
		"runningState": "idle"
	}
}
```
Note: jq can format JSON for output

### 3.6 Get sync results
You can get results of synchronization and validation. It shows the valid, warning and invalid number of cer, roa, mft and crl respectively.

```shell
$ cd /home/rpki/rpstir2/bin
$./rpstir2.sh results  | jq .
```
```JSON
{
	"cerResult": {
		"fileType": "cer",
		"validCount": 16920,
		"warningCount": 0,
		"invalidCount": 6
	},
	"crlResult": {
		"fileType": "crl",
		"validCount": 16916,
		"warningCount": 0,
		"invalidCount": 51
	},
	"mftResult": {
		"fileType": "mft",
		"validCount": 16914,
		"warningCount": 0,
		"invalidCount": 71
	},
	"roaResult": {
		"fileType": "roa",
		"validCount": 31779,
		"warningCount": 0,
		"invalidCount": 288
	},
	"moaResult": {
		"fileType": "moa",
		"validCount": 1,
		"warningCount": 0,
		"invalidCount": 0
	}
}
```



### 3.7 Export Roas
You can get all valid roas after sync.

```shell
$ cd /home/rpki/rpstir2/bin
$./rpstir2.sh exportroas | jq .
```
```
[
  {
    "repo": "rpki.afrinic.net",
    "rir": "AFRINIC",
    "maxLength": 20,
    "addressPrefix": "102.128.144/20",
    "asn": 328210
  },
  {
    "repo": "rpki.afrinic.net",
    "rir": "AFRINIC",
    "maxLength": 24,
    "addressPrefix": "102.128.144/20",
    "asn": 328210
  },
  ....
```  


### 3.8 Parse file
You can parse cer/mft/crl/roa/asa/moa file.


```shell
$ cd /home/rpki/rpstir2/bin
$./rpstir2.sh parse ../doc/AS211321.asa | jq .
```
```JSON
{
  "data": {
    "signerInfoModel": {
      "messageDigest": "c7388599a2d43808f7bb13a0d095ffcfbba880ae850c2699f67b764538e66ca0",
      "signingTime": "2021-11-11T11:19:00Z",
      "contentType": "1.2.840.113549.1.9.16.1.49",
      "digestAlgorithm": "sha256",
      "version": 3
    },
    "eeCertModel": {
      "eeCertEnd": 1308,
      "eeCertStart": 102,
      "crldpModel": {
        "critical": false,
        "crldps": [
          "rsync://rsync.accept.krill.cloud/repo/accept/0/7088BE00CA85327CA016C9074ED007C3FA919991.crl"
        ]
      },
      "cerIpAddressModel": {
        "critical": false,
        "cerIpAddresses": null
      },
      "siaModel": {
        "critical": false,
        "signedObject": "rsync://rsync.accept.krill.cloud/repo/accept/0/AS211321.asa",
        "caRepository": "",
        "rpkiNotify": "",
        "rpkiManifest": ""
      },
      "issuerAll": "CN=7088be00ca85327ca016c9074ed007c3fa919991",
      "subjectAll": "CN=37CA1DDE4D094734AB3B048269E12FDDAEAA691B",
      "isCa": false,
      "version": 3,
      "digestAlgorithm": "SHA256-RSA",
      "sn": "2de71e5b974c86a28d6bb3c1e1b4ece5091f12c",
      "notBefore": "2021-11-11T19:14:00+08:00",
      "notAfter": "2022-11-10T19:19:00+08:00",
      "keyUsageModel": {
        "keyUsageValue": "Certificate Sign, CRL Sign",
        "critical": true,
        "keyUsage": 1
      },
      "extKeyUsages": [],
      "basicConstraintsValid": false
    },
    "siaModel": {
      "critical": false,
      "signedObject": "rsync://rsync.accept.krill.cloud/repo/accept/0/AS211321.asa",
      "caRepository": "",
      "rpkiNotify": "",
      "rpkiManifest": ""
    },
    "aiaModel": {
      "critical": false,
      "caIssuers": "rsync://localcert.ripe.net/repository/DEFAULT/cIi-AMqFMnygFskHTtAHw_qRmZE.cer"
    },
    "version": 0,
    "ski": "37ca1dde4d094734ab3b048269e12fddaeaa691b",
    "aki": "7088be00ca85327ca016c9074ed007c3fa919991",
    "filePath": "/tmp/ParseFile219618154/",
    "fileName": "AS211321.asa",
    "fileHash": "e2f896d37de277e93309ef52f01fe3131b762b1383e86cd7933887ad7eaf1257",
    "customerAsns": [
      {
        "ProviderAsnOwners": null,
        "providerAsns": [
          65000,
          65001,
          65002
        ],
        "customerAsnOwner": "",
        "customerAsn": 211321,
        "version": 0
      }
    ],
    "eContentType": "1.2.840.113549.1.9.16.1.49"
  },
  "msg": "",
  "result": "ok"
}
```


```shell
$ cd /home/rpki/rpstir2/bin
$./rpstir2.sh parse ../doc/rpki-fzca-bottom-1784704862072011507.moa | jq .
```
```JSON
{
  "data": {
    "signerInfoModel": {
      "messageDigest": "8f47c533085798537794a9d33e7ec108a60c635edf1c810c2423b2888afaa2cd",
      "signingTime": "2026-07-22T07:21:02Z",
      "contentType": "1.2.840.113549.1.9.16.1.24",
      "digestAlgorithm": "sha256",
      "version": 3
    },
    "eeCertModel": {
      "eeCertEnd": 1241,
      "eeCertStart": 97,
      "crldpModel": {
        "critical": false,
        "crldps": [
          "rsync://rpki.fzca.com/repository/bottom/rpki-fzca-bottom.crl"
        ]
      },
      "cerIpAddressModel": {
        "critical": true,
        "cerIpAddresses": [
          {
            "addressPrefixRange": "\"10.2.0.0/16\"",
            "rangeEnd": "0a.02.ff.ff",
            "rangeStart": "0a.02.00.00",
            "max": "",
            "min": "",
            "addressPrefix": "10.2/16",
            "addressFamily": 1
          },
          {
            "addressPrefixRange": "\"110.112.114.0/24\"",
            "rangeEnd": "6e.70.72.ff",
            "rangeStart": "6e.70.72.00",
            "max": "",
            "min": "",
            "addressPrefix": "110.112.114/24",
            "addressFamily": 1
          },
          {
            "addressPrefixRange": "\"2001:0:200:101::/64\"",
            "rangeEnd": "2001:0000:0200:0101:ffff:ffff:ffff:ffff",
            "rangeStart": "2001:0000:0200:0101:0000:0000:0000:0000",
            "max": "",
            "min": "",
            "addressPrefix": "2001:0:200:101/64",
            "addressFamily": 2
          }
        ]
      },
      "siaModel": {
        "critical": false,
        "signedObject": "rsync://rpki.fzca.com/repository/bottom/rpki-fzca-bottom-1784704862072011507.moa",
        "caRepository": "",
        "rpkiNotify": "",
        "rpkiManifest": ""
      },
      "issuerAll": "CN=rpki-fzca-middle",
      "subjectAll": "CN=rpki-fzca-bottom",
      "isCa": false,
      "version": 3,
      "digestAlgorithm": "SHA256-RSA",
      "sn": "6a606f55",
      "notBefore": "2026-07-22T15:21:02+08:00",
      "notAfter": "2036-07-19T15:21:02+08:00",
      "keyUsageModel": {
        "keyUsageValue": "Certificate Sign, CRL Sign",
        "critical": true,
        "keyUsage": 1
      },
      "extKeyUsages": [],
      "basicConstraintsValid": false
    },
    "aiaModel": {
      "critical": false,
      "caIssuers": "rsync://rpki.fzca.com/repository/middle/rpki-fzca-middle.cer"
    },
    "siaModel": {
      "critical": false,
      "signedObject": "rsync://rpki.fzca.com/repository/bottom/rpki-fzca-bottom-1784704862072011507.moa",
      "caRepository": "",
      "rpkiNotify": "",
      "rpkiManifest": ""
    },
    "ipv4Prefixes": [
      "110.112.114.0/24",
      "10.2.0.0/16"
    ],
    "version": 0,
    "ski": "f142b67e3c0d555d71cc5199d91a60d5f4c7b82b",
    "aki": "295dafa87fe33f8036c4b51fae51fd582e8eb8f2",
    "filePath": "/tmp/ParseFile1184945333/",
    "fileName": "rpki-fzca-bottom-1784704862072011507.moa",
    "fileHash": "f20173a4f3c2a142299174c7e5866666230476304099293549f5b0ea39096763",
    "eContentType": "1.2.840.113549.1.9.16.1.24",
    "ipv6MappingPrefix": "2001:0:200:101::/64"
  },
  "msg": "",
  "result": "ok"
}
```

### 3.9 Rebuild
You can compile the program by yourself if you have installed GoLang.

```shell
$ cd /home/rpki/rpstir2/bin
$./rpstir2.sh rebuild
```

### 3.10 Help

```shell
$ cd /home/rpki/rpstir2/bin
$./rpstir2.sh help
```


## 4 Reporting bugs and getting help
Please open an issue on our [GitHub page](https://github.com/bgpsecurity/rpstir2/issues) or mail to [shaoqing@zdns.cn](mailto:shaoqing@zdns.cn) with any problems or bugs you encounter.





