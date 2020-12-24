

# RPSTIR2
## 1. Introduction
RPKI is a hierarchical Public Key Infrastructure(PKI) that binds Internet Number Resources(INRs) such as Autonomous System Numbers(ASNs) and IP addresses to public keys via certificates. RPKI allows INR holder(certificate holder) to allocate certain IP prefix to their customers via issuing resource certificates(RCs) and authorizing an ASN to announce certain IP prefixes via issuing ROAs, and all of these RPKI objects are published in RPKI repository.

As the bridge between inter-domain routing system and RPKI repository, RPKI Relying Party(RP) is designed to assist BGP Speakers in synchronization of RPKI objects, validation of certificate chain, cache management and transmission of Validated ROA Payloads(VRPs).

RPSTIR2 is a kind of RP software written in GO, which based on design idea of RPSTIR, provides all the standard functions mentioned above. RPSTIR2 also supports more RPKI-related protocols and optimizes performance.

RPSTIR2 is capable of running on CentOS8(64bit)/Ubuntu18(64bit) or higher.
&nbsp;

## 2. Install RPSTIR2
There are two ways to install RPSTIR2, including installing from source code and using docker.

### 2.1 Install from source code

#### 2.1.1 Install OpenSSL
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

#### 2.1.2 Install MySQL
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

Note: You also can use docker to run MySQL as shown in section 2.2.1. 

#### 2.1.3 Install GoLang
The GoLang version must be 1.13 or higher.

```shell
$ wget --no-verbose --inet4-only https://dl.google.com/go/go1.14.1.linux-amd64.tar.gz
$ tar -C /usr/local -xzf go1.14.1.linux-amd64.tar.gz
$ echo "export GOROOT=/usr/local/go" >> /root/.bashrc 
$ echo "export GOPATH=/usr/local/goext" >> /root/.bashrc 
$ echo "export PATH=$PATH:/usr/local/go/bin:/usr/local/goext/bin" >> /root/.bashrc 
$ source  /root/.bashrc
```

#### 2.1.4 Create RPSTIR2 directories
Before installing RPSTIR2, you should create directories in advance, one of which is for program and the other is for the cache data. you can be modified in locations of your choice as shown in section 2.1.6. 

```shell
$ mkdir -p /root/rpki/ /root/rpki/rpstir2  /root/rpki/data 
```

| Directory  | Path                      |
| :--------: | ------------------------- |
| programDir | /root/rpki/rpstir2        |
| dataDir    | /root/rpki/data           |


#### 2.1.5 Download RPSTIR2 

```shell
$ cd /root/rpki/
$ git clone https://github.com/bgpsecurity/rpstir2.git 
```

#### 2.1.6 Configure RPSTIR2
You can modify configuration parameters of programDir, dataDir, mysql, and tcpport of rtr in configuration file(/root/rpki/rpstir2/conf/project.conf). 




##### 2.1.7 Deploy RPSTIR2
The RPSTIR2 will build and deploy automatically to /root/rpki/rpstir2. 

```shell
$ cd /root/rpki/rpstir2/bin
$ chmod +x *.sh 
$ ./rpstir2-service.sh deploy
$ ./rpstir2-service.sh update
```

#### 2.1.8 Configure scheduled task
You can use crontab to perform scheduled synchronization tasks. Then RPSTIR2 will download RPKI objects, and complete the subsequent validation procedure according to the schedule you set. 

```shell
$ crontab -e
1 1 * * *  /root/rpki/rpstir2/bin/rpstir2-command.sh crontab
```

Note: The RPSTIR2 service must be started first as shown in section 2.3.1. 

### 2.2 Install from Docker
#### 2.2.1 Pull RPSTIR2 MySQL docker image
You can pull mysql docker image and login in MySQL as root.

```shell
$ docker pull mysql
$ docker run -itd --name rpstir2_mysql -p 13306:3306 -e MYSQL_ROOT_PASSWORD=Rpstir-123 mysql
$ docker exec -it rpstir2_mysql /bin/bash
$ mysql -uroot -p
Rpstir-123
```


After that, you should create user accounts and database of RPSTIR2 as shown in section 2.1.2. 

```SQL
CREATE USER 'rpstir2'@'localhost' IDENTIFIED WITH mysql_native_password BY 'Rpstir-123';
CREATE USER 'rpstir2'@'%' IDENTIFIED WITH mysql_native_password BY 'Rpstir-123';
flush privileges;

CREATE DATABASE rpstir2;
GRANT ALL PRIVILEGES ON rpstir2.* TO 'rpstir2'@'localhost'  with grant option;
GRANT ALL PRIVILEGES ON rpstir2.* TO 'rpstir2'@'%'  with grant option;
flush privileges;
quit;
```

#### 2.2.2 Pull RPSTIR2 docker image
On the host, the cache data is stored in "/root/rpki/rpstir2data/data/", and the logs of rpstir2 are saved in "/root/rpki/rpstir2data/log", and tcpport of rtr is 18082.

```shell
$ cd /root/rpki/
$ mkdir -p /root/rpki/rpstir2data /root/rpki/rpstir2data/data  /root/rpki/rpstir2data/log
$ docker pull cpusoft/rpstir2_centos8
$ docker run -itd --privileged -p 18080-18090:8080-8090   -v /root/rpki/rpstir2data/data:/root/rpki/data  -v /root/rpki/rpstir2data/log:/root/rpki/rpstir2/log --name rpstir2_centos8 cpusoft/rpstir2_centos8  /usr/sbin/init
```

#### 2.2.3 Configure RPSTIR2
Then, you should login in RPSTIR2 container, and run deploy and update. 

```shell
$ docker exec -it rpstir2_centos8 /bin/bash
$ cd /root/rpki/rpstir2/bin 
$ chmod +x *.sh
$ ./rpstir2-service.sh deploy
$ ./rpstir2-service.sh update
```
Note1: You can change synchronization schedule task in crontab as shown in section 2.1.8.

Note2: Because RPSTIR2 uses the docker's bridge network (172.17.0.1) to link MySQL in other container, the configuration of mysql server is changed to "172.17.0.1:13306" in /root/rpki/rpstir2/conf/project.conf.


## 3 Running RPSTIR2
All functions of RPSTIR2 are accessible on the command line via sub-commands.

(1) rpstir2-serverice.sh: Execute system commands such as system start stop and upgrade.

(2) rpstir2-command.sh: To execute specific synchronization, view status and results, and other program commands.

### 3.1 Start and stop the RPSTIR2 service
To execute all RPSTIR2 commands, the RPSTIR2 service must be started first. 
You can check for errors by looking at the log files in ./log/ directory.

```shell
$ cd /root/rpki/rpstir2/bin
$./rpstir2-serverice.sh start 
$./rpstir2-serverice.sh stop 
```

### 3.2 Sync and validate RPKI objects
You can download RPKI objects with rsync or RRDP protocol, and complete the subsequent validation procedure. Or when the parameter is sync, the system will automatically perform hybrid synchronization. 


#### (1) rsync

```shell
$ cd /root/rpki/rpstir2/bin
$ ./rpstir2-command.sh rsync 
```

#### (2) rrdp

```shell
$ cd /root/rpki/rpstir2/bin
$ ./rpstir2-command.sh rrdp  
```

#### (3) sync

```shell
$ cd /root/rpki/rpstir2/bin
$ ./rpstir2-command.sh sync  
```

### 3.3 Get sync and validation status
Because rsync and RRDP take long time to run, they are executed in the background. So you need a command to determine if the synchronization and validation process is complete.

```shell
$ cd /root/rpki/rpstir2/bin
$ ./rpstir2-command.sh states  
```

When you get the following JSON message, if "isRunning" is "true", it means that sync and validation are still running; if it is "false", sync and validation complete. At this time, the router can obtain rpki data through RTR port.

```JSON
{
	"result": "ok",
	"msg": "",
	"serviceState": {
		"startTime": "2020-01-01 01:01:01 CST",
		"isRunning": "false",
		"runningState": "idle"
	}
}

```
### 3.4 Results
You can get results of synchronization and validation. It shows the valid, warning and invalid number of cer, roa, mft and crl respectively.

```shell
$ cd /root/rpki/rpstir2/bin
$./rpstir2-command.sh results  
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



### 3.5 Help

```shell
$ cd /root/rpki/rpstir2/bin
$./rpstir2-service.sh help
$./rpstir2-command.sh help
```


## 4 Reporting bugs and getting help
Please open an issue on our [GitHub page](https://github.com/bgpsecurity/rpstir2/issues) or mail to [shaoqing@zdns.cn](mailto:shaoqing@zdns.cn) with any problems or bugs you encounter.





