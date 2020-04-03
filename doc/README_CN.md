


# Rpstir2
(Relying Party Security Technology for Internet Routing 2)

## 简介
在RPKI体系中，RP软件扮演者重要的角色，是CA与边界路由器之间的桥梁。RPSTIR做为现有RP软件（rpki.net，RPSTIR，RIPE NCC，Routinator）之一，提供包括同步、验证、指导BGP路由在内的所有RP端的标准功能。

RPSTIR2基于原有RPSTIR设计思想重新开发，采用Go语言，对RPKI相关协议支持更加完善，性能得到优化。

## 功能模块
RPSTIR2核心模块包括：同步模块，验证模块，RTR模块和数据库模块。

RPSTIR2支持的功能包括：
* 同步协议支持Rsync和RRDP(rfc8182)
* 支持Slurm(rfc8416)
* 支持RTR
* 支持Validation Reconsidered(rfc8360)


## 安装

* ### 操作系统要求
  CentOS8或Ubuntu18及以上，均需64位系统



* ### 安装OpenSSL 
  版本需要 1.1.1b及以上，在编译OpenSSL时，需要设置"enable-rfc3779"
```
	wget --no-verbose --inet4-only https://www.openssl.org/source/openssl-1.1.1d.tar.gz 
    tar xzvf openssl-1.1.1d.tar.gz 
    cd openssl-1.1.1d 
    config shared enable-rfc3779
	make
	make install
	echo "export PATH=/usr/local/ssl/bin:$PATH" >> /root/.bashrc
    source /root/.bashrc
```
 
    
* ### 安装MySQL
   版本需要8及以上，需要支持json。
   请登陆mysql按如下脚本，创建用户和数据表：
```
	CREATE USER 'rpstir2'@'localhost' IDENTIFIED BY 'Rpstir-123';
	CREATE USER 'rpstir2'@'%' IDENTIFIED BY 'Rpstir-123';
	CREATE DATABASE rpstir2;
	GRANT ALL PRIVILEGES ON rpstir2.* TO 'rpstir2'@'localhost'  with grant option;
	GRANT ALL PRIVILEGES ON rpstir2.* TO 'rpstir2'@'%'  with grant option;
	flush privileges;
```
 * ### 安装RPSTIR2
    首先规划好三个目录：  源代码目录为/root/rpki/source/rpstir2，  程序运行目录为/root/rpki/rpstir2， 数据缓存目录为/root/rpki/data，后文均按此配置安装和运行，可根据实际情况修改。
    RPSTIR2支持多种安装方式，建议采用方法1，简单修改配置后，即可运行。
   
  1. 安装预编译版本
      从https://github.com/bgpsecruity/rpstir2/releases/ 下载最新预编译版本到/root/rpki/rpstir2。 解压后进入/root/rpki/rpstir2/conf目录，修改配置文件project.conf，根据实际修改程序目录和数据目录，和MySQL参数
```
    [rpstir2]
    programdir=/root/rpki/rpstir2
    datadir=/root/rpki/data
	[mysql]
    server=127.0.0.1:3306
    user=rpstir2
    password=Rpstir-123
    database=rpstir2
```
  修改完毕后，即可按后文的“运行RPSTIR2”启动运行


2. 下载源代码手动编译
    需要安装开发环境，且过程比较复杂，不建议一般用户使用。
   （1） 本机安装GoLang开发环境
   版本需要1.13及以上
   （2）下载源代码
   从https://github.com/bgpsecurity/rpstir2.git 下载源代码到/root/rpki/source/rpstir2。进入/root/rpki/source/rpstir2/conf目录，修改配置文件project.conf，根据实际修改程序目录和数据目录，和MySQL参数
```
    [rpstir2]
    programdir=/root/rpki/rpstir2
    datadir=/root/rpki/data
    [mysql]
    server=127.0.0.1:3306
    user=rpstir2
    password=Rpstir-123
    database=rpstir2
 ```
 （3）执行编译
   进入/root/rpki/source/rpstir2/build目录，执行
   ```
    chmod +x *.sh 
    ./rpstir2-deploy.sh，
   ```
将完成自动化的部署，然后即可按后文的“运行RPSTIR2”启动运行。
  注：如果部署时，编译失败，提示Go Lang缺少依赖包，则请根据提示，自行下载所需的依赖。
   
 3. 使用docker
   即将发布
    
## 运行RPSTIR2
 * ### 启停服务
    执行所有RPSTIR2命令，都必须先启动RPSTIR2服务。进入bin目录，下述命令分别为起停RPSTIR2服务
 ```
  ./rpstir-serverice.sh start
  ./rpstir-serverice.sh stop
  ```

 * ### 初始化
  在安装部署完RPSTIR2执行一次，从而完成数据表的初始化
  ```
   ./rpstir-command init  
 ```

 * ### 启动rsync同步验证
   起动rsync同步，并完成后续的验证，最后生成RTR数据
  ```
   ./rpstir-command rsync  
 ```
 
 * ### 启动RRDP(delta)同步验证
   起动RRDP(delta)同步，并完成后续的验证，最后生成RTR数据
   注意，当前不是所有RIR都提供了RRDP服务，所以数据会比Rsync同步验证方式要少
  ```
   ./rpstir-command rrdp  
 ```

 * ### 获取rsync或RRDP验证状态
   由于rsync和RRDP命令是后台执行，运行时间较长。所以需要单独的命令，查看同步验证过程是否已经都完成
  ```
   ./rpstir-command states  
 ```
  当返回
   {
    "result": "ok",
    "msg": "",
    "state": {
        "endTime": "2019-12-19 14:07:11",
        "startTime": "2019-12-18 16:29:06",
        "state": "end"
    }
 }
 表明已经同步、验证、RTR均已完成

 * ### 重置
   当希望rsync或RRDP完全重新同步时，可以直行重置命令，将清空数据表和本地缓存，恢复到刚安装的状态
  ```
   ./rpstir-command reset  
 ```

## 技术咨询和Bug报告

 如果发现任何bug或者有任何问题，欢迎提出issue，或者发邮箱到 shaoqing@zdns.cn


