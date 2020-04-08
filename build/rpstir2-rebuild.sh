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
chmod +x ${rpstir2_program_dir}/bin/*


