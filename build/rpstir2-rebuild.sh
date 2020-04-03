#!/bin/bash
configFile="../conf/project.conf"
source ./read-conf.sh

cur=$(pwd)
rpstir2_build_dir=$(pwd)
cd ..
rpstir2_source_dir=$(pwd)
cd ${rpstir2_build_dir}
echo "source directory is " $rpstir2_source_dir
echo "build directory is " $rpstir2_build_dir



rpstir2_program_dir=$(ReadINIfile $configFile rpstir2 programdir)
rpstir2_data_dir=$(ReadINIfile $configFile rpstir2 datadir)
echo "program directory is " $rpstir2_program_dir
echo "data directory is " $rpstir2_data_dir

#######################################
echo "1. create program data dir, and recreate mysql tables"
# reset program and data dir, and copy default project.conf
cd ${rpstir2_source_dir}
svn update --accept tf 
cd ${rpstir2_source_dir}/build
chmod +x *.sh
cd ${rpstir2_program_dir}/bin/
./rpstir2-service.sh stop
cp ${rpstir2_build_dir}/rpstir2-command.sh ${rpstir2_program_dir}/bin/
cp ${rpstir2_build_dir}/rpstir2-service.sh ${rpstir2_program_dir}/bin/
cp ${rpstir2_build_dir}/read-conf.sh ${rpstir2_program_dir}/bin/
cd ${rpstir2_program_dir}/bin/
chmod +x *




########################################
echo "2. update some go depencies"
go get -u github.com/cpusoft/goutil
echo -e "\n"


########################################
echo "3. start to build rpstir2"
cd $rpstir2_source_dir
echo "build dir is" $rpstir2_source_dir
oldgopath=$GOPATH
echo "old gopath is " $oldgopath
# linux / windows
CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
GOPATH=$GOPATH:$rpstir2_source_dir
echo "new gopath is " $GOPATH
# see: go tool compile -help
go install -v -gcflags "-N -l" ./...
export GOPATH=$oldgopath
echo "reset old gopath is " $GOPATH
cd $cur
echo "it will cp ${rpstir2_source_dir}/bin/* ${rpstir2_program_dir}/bin/"
cp ${rpstir2_source_dir}/bin/* ${rpstir2_program_dir}/bin/
chmod +x ${rpstir2_program_dir}/bin/*
echo "build compelete and cd $cur"
cd  ${rpstir2_program_dir}/bin/
./rpstir2-service.sh start
echo -e "\n"


echo -e "\n"
echo "rebuild complete"
echo -e "\n"

