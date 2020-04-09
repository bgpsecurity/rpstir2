

# RPSTIR2
## Introduction

RPKI is a hierarchical Public Key Infrastructure(PKI) that binds Internet Number Resources(INRs) such as Autonomous System Numbers(ASNs) and IP addresses to public keys via certificates. RPKI allows INR holder(certificate holder) to allocate certain IP prefix to their customers via issuing resource certificates(RCs) and authorizing an ASN to announce certain IP prefixes via issuing ROAs, and all of these RPKI objects are published in RPKI repository.

As the bridge between inter-domain routing system and RPKI repository, RPKI Relying Party(RP) is designed to assist BGP Speakers in synchronization of RPKI objects, validation of certificate chain, cache management and transmission of Validated ROA Payloads(VRPs).

RPSTIR2 is a kind of RP software written in GO, which based on design idea of RPSTIR, provides all the standard functions mentioned above. RPSTIR2 also supports more RPKI-related protocols and optimizes performance.

RPSTIR2 is capable of running on CentOS8(64bit)/Ubuntu18(64bit) or higher.
&nbsp;
## Getting started

### Install OpenSSL

OpenSSL version must be 1.1.1b or higher, and  "enable-rfc3779" needs to be set when compiling OpenSSL.

```shell
$ wget --no-verbose --inet4-only
https://www.openssl.org/source/openssl-1.1.1f.tar.gz 
$ tar xzvf openssl-1.1.1f.tar.gz 
$ cd openssl-1.1.1f
$ config shared enable-rfc3779
$ make
$ make install
$ echo "export PATH=/usr/local/ssl/bin:$PATH" >> /root/.bashrc
$ source /root/.bashrc
```
&nbsp;
### Configure MySQL

MySQL version must be 8 or higher and should support JSON. After MySQL has been installed, please login to MySQL and create user accounts and data tables according to the following script.

```mysql
ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY 'Rpstir-123';
CREATE USER 'rpstir2'@'localhost' IDENTIFIED WITH mysql_native_password BY 'Rpstir-123';
CREATE USER 'rpstir2'@'%' IDENTIFIED WITH mysql_native_password BY 'Rpstir-123';
flush privileges;

CREATE DATABASE rpstir2;
GRANT ALL PRIVILEGES ON rpstir2.* TO 'rpstir2'@'localhost'  with grant option;
GRANT ALL PRIVILEGES ON rpstir2.* TO 'rpstir2'@'%'  with grant option;
flush privileges;
```
&nbsp;
## Install RPSTIR2

Before installing RPSTIR2, you should create three directories in advance, one of which is for RPSTIR2 source code, and one is for program and the other is for the cache data. The following documents are explained according to the configuration given in the following table, which can be modified in locations of your choice.

| Directory  | Path                      |
| :--------: | ------------------------- |
| sourcedir  | /root/rpki/source/rpstir2 |
| programdir | /root/rpki/rpstir2        |
| datadir    | /root/rpki/data           |

```shell
$ mkdir -p /root/rpki/source/ /root/rpki/rpstir2  /root/rpki/data 
```
There are two ways to install RPSTIR2, including installing from source code and using docker.

### 1. Install from source code

##### (1) Install GoLang
Before install RPSTIR2 from source code, you should install the GoLang development environment, and the version must be 1.13 or higher

```shell
$ wget --no-verbose --inet4-only https://dl.google.com/go/go1.14.1.linux-amd64.tar.gz
$ tar -C /usr/local -xzf go1.14.1.linux-amd64.tar.gz
$ echo "export GOROOT=/usr/local/go" >> /root/.bashrc 
$ echo "export GOPATH=/usr/local/goext" >> /root/.bashrc 
$ echo "export PATH=$PATH:/usr/local/go/bin:/usr/local/goext/bin" >> /root/.bashrc 
$ source  /root/.bashrc
```
##### (2) Download RPSTIR2 
You can download source code of RPSTIR2.

```shell
$ cd /root/rpki/source/
$ git clone https://github.com/bgpsecurity/rpstir2.git 
```
##### (3) Configure RPSTIR2
RPSTIR2 can generate a basic configuration file(**/root/rpki/source/rpstir2/conf/project.conf**) in sourcedir.  You can modify configuration parameters of programdir, datadir and mysql.

```shell
$ cd /root/rpki/source/rpstir2/conf
$ vim project.conf
[rpstir2]
programdir=/root/rpki/rpstir2
sourcedir=/root/rpki/source/rpstir2
datadir=/root/rpki/data
[mysql]
server=127.0.0.1:3306
user=rpstir2
password=Rpstir-123
database=rpstir2
```

##### (4) Compile and install RPSTIR2
Now, you can compile source code. The RPSTIR2 will be installed automatically in /root/rpki/rpstir2

```shell
$ cd /root/rpki/source/rpstir2/build
$ chmod +x *.sh 
$ ./rpstir2-service.sh deploy
```

Note: if you get building errors, indicating GoLang is lack of dependent packages, you can download all the missing dependent packages.



##### (5) Set scheduled task
You can use crontab to perform scheduled synchronization tasks every day.

```shell
crontab -e
1 1 * * *  /root/rpki/rpstir2/bin/rpstir2-command.sh crontab
```

### 2. Install from Docker
##### (1) Pull RPSTIR2 docker image
The RPSTIR2 images is based on centos8, you can pull docker image and run RPSTIR2 as rpstir2_centos8

```shell
docker pull cpusoft/rpstir2_centos8
docker volume create --name=rpstir2data
mkdir -p /root/rpki/rpstir2data /root/rpki/rpstir2data/data  /root/rpki/rpstir2data/log 
docker run -itd --privileged -p 13306:3306 -p 18080-18090:8080-8090  -v /root/rpki/rpstir2data/data:/root/rpki/data  -v /root/rpki/rpstir2data/log:/root/rpki/rpstir2/log    --name=rpstir2_centos8   cpusoft/rpstir2_centos8 /usr/sbin/init
```

##### (2) Login and deploy RPSTIR2
Then, you should login in rpstir2_centos8, and run deploy.  

```shell
docker exec -it rpstir2_centos8 /bin/bash
cd /root/rpki/source/rpstir2/build 
chmod +x *.sh
./rpstir2-service.sh deploy
```
##### (3) Set scheduled task
Scheduled task has been set with crontab, and the synchronization task is executed at 1:1 am every day. You can change it.

```shell
crontab -e
1 1 * * *  /root/rpki/rpstir2/bin/rpstir2-command.sh crontab
```

##### (4)  Login out and see cache data
"Ctrl-D" to exist rpstir2_centos8, and you can see cache data in "/root/rpki/rpstir2data/data/" and log of rpstir2 in "/root/rpki/rpstir2data/log"
&nbsp;
## Running

All functions of RPSTIR2 are accessible on the command line via sub-commands.

### Start and stop the RPSTIR2 service

To execute all RPSTIR2 commands, the RPSTIR2 service must be started first. 

```shell
$ cd /bin
$./rpstir2-serverice.sh start # start the daemon
$./rpstir2-serverice.sh stop # stop the daemon
```
When RPSTIR2  service starts,   RPSTIR2 http server is listening on port 8080, and an RTR server is listening on port 8082.  You can change these two ports in project.conf.

```shell
[rpstir2]
httpport=8080
[rtr]
tcpport=8082
```

### Initing

After installing RPSTIR2, you should execute the following command to complete the initialization of the data tables.

```shell
$ ./rpstir2-command.sh init  
```

### Sync and validate RPKI objects

You can download RPKI objects with rsync or RRDP protocol and complete the subsequent validation procedure.

​1. rsync

```shell
$ ./rpstir2-command.sh rsync 
```

​2. RRDP

```shell
$ ./rpstir2-command.sh rrdp  
```

### Get sync and validation status

Because rsync and RRDP are executed in the background, and they will take long time to run. So you need a command to determine if the synchronization and validation process is complete.

```shell
$ ./rpstir2-command.sh states  
```

When you get the following JSON message, it indicates that synchronization and validation of RPKI objects and information transmission via RPKI-RTR protocol to routers have been completed. And all cache data are stored in "/root/rpki/rpstir2data/". 

```JSON
{ "result": "ok",
  "msg": "",
  "state":
  	{ "endTime": "2019-12-19 14:07:11", 
     	"startTime": "2019-12-18 16:29:06",
     	"state": "end" 
    } 
 }
```

### Reset

When you need to re-synchronize and re-validate RPKI objects, you can clean the tables in MySQL and cached data by executing the following command.

```shell
$./rpstir2-command.sh reset  
```
&nbsp;
## Technical consultation and bug report

Please open an issue on our [GitHub page](https://github.com/bgpsecurity/rpstir2/issues) or mail to [shaoqing@zdns.cn](mailto:shaoqing@zdns.cn) with any problems or bugs you encounter.





