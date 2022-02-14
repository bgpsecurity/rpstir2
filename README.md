

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
$ wget --no-verbose --inet4-only https://www.openssl.org/source/openssl-1.1.1f.tar.gz 
$ tar xzvf openssl-1.1.1f.tar.gz 
$ cd openssl-1.1.1f
$ ./config shared enable-rfc3779
$ make
$ make install
$ echo "export PATH=/usr/local/ssl/bin:$PATH" >> /root/.bashrc
$ source /root/.bashrc
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

Note: You also can use docker to run MySQL. 

### 2.3 Install GoLang(Optional)
If you plan to compile the program by yourself, you need to install a version of Golang higher than 1.17. Otherwise you don't need to install it.


### 2.4 Create RPSTIR2 directories
Before installing RPSTIR2, you should create directories in advance, one of which is for program and the other is for the cache data. you can modify the shell, and change "conf/project.conf". 

| Directory  | Path                      |
| :--------: | ------------------------- |
| programDir | /root/rpki/rpstir2        |
| dataDir    | /root/rpki/data           |


```shell
$ mkdir -p /root/rpki/ /root/rpki/rpstir2  /root/rpki/data  /root/rpki/data/rrdprepo  /root/rpki/data/rsyncrepo /root/rpki/data/tal
```

### 2.5 Download RPSTIR2 

```shell
$ cd /root/rpki/
$ git clone https://github.com/bgpsecurity/rpstir2.git 
$ cd /root/rpki/rpstir2/bin
$ chmod +x *
$ cp /root/rpki/rpstir2/build/tal/*  /root/rpki/data/tal/
```

### 2.6 Configure RPSTIR2
You can modify configuration parameters of programDir, dataDir, mysql, and  port in configuration file(/root/rpki/rpstir2/conf/project.conf). 

## 3 Running RPSTIR2

### 3.1 Initialize the RPSTIR2

```shell
$ cd /root/rpki/rpstir2/bin
$./rpstir2.sh start 
$./rpstir2.sh init 
```

### 3.2 Start and stop the RPSTIR2
The RPSTIR2 must be started first, you can check for errors by looking at the log files in ./log/ directory.

```shell
$ cd /root/rpki/rpstir2/bin
$./rpstir2.sh start 
```

```shell
$ cd /root/rpki/rpstir2/bin
$./rpstir2.sh stop 
```

### 3.3 Configure scheduled task
You can use crontab to perform scheduled synchronization tasks. Then RPSTIR2 will download RPKI objects, and complete the subsequent validation procedure according to the schedule you set. 

```shell
$ crontab -e
1 1 * * *  /root/rpki/rpstir2/bin/rpstir2.sh crontab
```
Note: The RPSTIR2 service must be started first. 

### 3.4 Sync and validate RPKI objects
You can download RPKI objects with rsync or RRDP protocol, and complete the subsequent validation procedure. 

```shell
$ cd /root/rpki/rpstir2/bin
$ ./rpstir2.sh sync  
```

### 3.5 Get sync and validation status
Because rsync and RRDP take long time to run, they are executed in the background. So you need a command to determine if the synchronization and validation process is complete.

```shell
$ cd /root/rpki/rpstir2/bin
$ ./rpstir2.sh state   | jq .
```

When you get the following JSON message, if "isRunning" is "true", it means that sync and validation are still running; if it is "false", sync and validation complete. At this time, the router can obtain rpki data through RTR port.

```JSON
{
	"result": "ok",
	"msg": "",
	"data": {
		"startTime": "2020-01-01 01:01:01 CST",
		"isRunning": "false",
		"runningState": "idle"
	}
}
```
Note: jq can format JSON for output

### 3.6 Get sync results
You can get results of synchronization and validation. It shows the valid, warning and invalid number of cer, roa, mft and crl respectively.

```shell
$ cd /root/rpki/rpstir2/bin
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
    }
}
```



### 3.7 Export Roas
You can get all valid roas after sync.

```shell
$ cd /root/rpki/rpstir2/bin
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
You can parse cer/mft/crl/roa/sig file.

```shell
$ cd /root/rpki/rpstir2/bin
$./rpstir2.sh parse ../doc/checklist.sig | jq .
```
```JSON
{
  "data": {
    "signerInfoModel": {
      "messageDigest": "BD32690504277FE2D1CCEF127174F0ACA0A1785170452C472BE631839425400D",
      "signingTime": "2021-02-10T14:50:25Z",
      "contentType": "1.3.6.1.4.1.41948.49",
      "digestAlgorithm": "sha256",
      "version": 3
    },
    "eeCertModel": {
      "eeCertEnd": 1279,
      "eeCertStart": 224,
      "crldpModel": {
        "critical": false,
        "crldps": [
          "rsync://chloe.sobornost.net/rpki/RIPE-nljobsnijders/LMq8Kl3LkWGqticaaLl6IAGSsJ4.crl"
        ]
      },
      "cerIpAddressModel": {
        "critical": false,
        "cerIpAddresses": null
      },
      "siaModel": {
        "critical": false,
        "signedObject": "",
        "caRepository": "",
        "rpkiNotify": "",
        "rpkiManifest": ""
      },
      "issuerAll": "CN=2ccabc2a5dcb9161aab6271a68b97a200192b09e",
      "subjectAll": "CN=EE",
      "isCa": false,
      "version": 3,
      "digestAlgorithm": "SHA256-RSA",
      "sn": "9",
      "notBefore": "2021-02-10T22:50:10+08:00",
      "notAfter": "2022-02-10T22:50:10+08:00",
      "keyUsageModel": {
        "keyUsageValue": "Digital Signature",
        "critical": true,
        "keyUsage": 1
      },
      "extKeyUsages": [],
      "basicConstraintsValid": false
    },
    "aiaModel": {
      "critical": false,
      "caIssuers": "rsync://rpki.ripe.net/repository/DEFAULT/LMq8Kl3LkWGqticaaLl6IAGSsJ4.cer"
    },
    "version": 0,
    "ski": "41ca827f3de666e9f7323f3059f6a7bb8b671175",
    "aki": "2ccabc2a5dcb9161aab6271a68b97a200192b09e",
    "filePath": "",
    "fileName": "checklist.sig",
    "fileHash": "fe44eb4ef1e389c1879f000f31485bac43e3c51a66040337625aa887a20d9556",
    "rpkiSignedChecklist": {
      "fileHashModels": [
        {
          "hash": "9516dd64be7c1725b9fca117120e58e8d842a5206873399b3ddffc91c4b6acf0",
          "file": "b42_ipv6_loa.png"
        },
        {
          "hash": "0ae1394722005cd92f4c6aa024d5d6b3e2e67d629f11720d9478a633a117a1c7",
          "file": "b42_service_definition.json"
        }
      ],
      "digestAlgorithm": "2.16.840.1.101.3.4.2.1",
      "cerIpAddresses": [
        {
          "addressPrefixRange": "",
          "rangeEnd": "",
          "rangeStart": "",
          "max": "",
          "min": "",
          "addressPrefix": "2001:67c:208c::/48",
          "addressFamily": 2
        }
      ]
    },
    "eContentType": "1.3.6.1.4.1.41948.49"
  },
  "msg": "",
  "result": "ok"
}
```


```shell
$ cd /root/rpki/rpstir2/bin
$./rpstir2.sh parse ../doc/AS211321.asa | jq .
```
```JSON
{
  "data": {
    "signerInfoModel": {
      "messageDigest": "C7388599A2D43808F7BB13A0D095FFCFBBA880AE850C2699F67B764538E66CA0",
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
        "keyUsageValue": "Digital Signature",
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
    "filePath": "",
    "fileName": "AS211321.asa",
    "fileHash": "e2f896d37de277e93309ef52f01fe3131b762b1383e86cd7933887ad7eaf1257",
    "asProviderAttestations": [
      {
        "ProviderAsIds": [
          {
            "addressFamilyIdentifier": "",
            "providerAsId": 65000
          },
          {
            "addressFamilyIdentifier": "0x0001",
            "providerAsId": 65001
          },
          {
            "addressFamilyIdentifier": "0x0002",
            "providerAsId": 65002
          }
        ],
        "customerAsId": 211321
      }
    ],
    "eContentType": "1.2.840.113549.1.9.16.1.49"
  },
  "msg": "",
  "result": "ok"
}
```

### 3.9 Rebuild
You can compile the program by yourself if you have installed GoLang.

```shell
$ cd /root/rpki/rpstir2/bin
$./rpstir2.sh rebuild
```

### 3.10 Help

```shell
$ cd /root/rpki/rpstir2/bin
$./rpstir2.sh help
```


## 4 Reporting bugs and getting help
Please open an issue on our [GitHub page](https://github.com/bgpsecurity/rpstir2/issues) or mail to [shaoqing@zdns.cn](mailto:shaoqing@zdns.cn) with any problems or bugs you encounter.





