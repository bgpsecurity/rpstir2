#!/bin/bash
source /etc/profile
source /root/.bashrc
cd /root/rpki/rpstir2/bin/
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
