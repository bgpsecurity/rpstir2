#!/bin/bash
configFile="../conf/project.conf"
source ./read-conf.sh

cur=$(pwd)
rpstir2_build_dir=$(pwd)
cd ..
rpstir2_source_dir=$(pwd)
cd ${rpstir2_build_dir}
rpstir2_program_dir=$(ReadINIfile $configFile rpstir2 programdir)
rpstir2_data_dir=$(ReadINIfile $configFile rpstir2 datadir)
echo "source directory is " $rpstir2_source_dir
echo "build directory is " $rpstir2_build_dir
echo "program directory is " $rpstir2_program_dir

mkdir -p ${rpstir2_program_dir} ${rpstir2_program_dir}/bin    ${rpstir2_program_dir}/conf  ${rpstir2_program_dir}/log  
mkdir -p ${rpstir2_data_dir}    ${rpstir2_data_dir}/rsyncrepo ${rpstir2_data_dir}/rrdprepo ${rpstir2_data_dir}/slurm  ${rpstir2_data_dir}/tal 

     
mkdir -p $GOPATH/src/golang.org/x
cd  $GOPATH/src/golang.org/x
git clone https://github.com/golang/crypto.git
git clone https://github.com/golang/net.git
go get -u github.com/astaxie/beego/logs
go get -u github.com/go-xorm/xorm
go get -u github.com/go-sql-driver/mysql
go get -u github.com/go-xorm/core
go get -u github.com/parnurzeal/gorequest
go get -u github.com/ant0ine/go-json-rest
go get -u github.com/satori/go.uuid
go get -u github.com/cpusoft/go-json-rest
go get -u github.com/cpusoft/goutil

cd $rpstir2_source_dir
oldgopath=$GOPATH
# linux / windows
CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
GOPATH=$GOPATH:$rpstir2_source_dir
# see: go tool compile -help
go install -v -gcflags "-N -l" ./...
export GOPATH=$oldgopath
cd $cur
cp ${rpstir2_source_dir}/bin/* ${rpstir2_program_dir}/bin/
cp ${rpstir2_build_dir}/rpstir2-command.sh ${rpstir2_program_dir}/bin/
cp ${rpstir2_build_dir}/rpstir2-service.sh ${rpstir2_program_dir}/bin/
cp ${rpstir2_build_dir}/read-conf.sh ${rpstir2_program_dir}/bin/
cp ${rpstir2_build_dir}/rpstir2-crontab.sh ${rpstir2_program_dir}/bin/
cp -r ${rpstir2_source_dir}/conf/* ${rpstir2_program_dir}/conf/
cp ${rpstir2_source_dir}/build/tal/*   ${rpstir2_data_dir}/tal
chmod +x ${rpstir2_program_dir}/bin/*


