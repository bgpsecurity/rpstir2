#!/bin/bash
source /etc/profile
source /root/.bashrc

abpath=$(readlink -f  "$0")
rpstir2_program_bin_dir=$(dirname "$abpath")  
cd $rpstir2_program_bin_dir

run="$1"
echo $run
if [ $run == "rsync" ] ; then
./rpstir2-command.sh rsync
fi
if [ $run == "start" ] ; then
./rpstir2-service.sh start
fi
if [ $run == "stop" ] ; then
./rpstir2-service.sh stop
fi
echo -e "\n"
