#!/bin/bash
source /etc/profile
source /root/.bashrc
cd /root/rpki/rpstir2/bin/
run="$1"
rsync="rsync"
if [ $run -eq $rsync ] ; then
./rpstir2-command.sh rsync
else
./rpstir2-service.sh start
fi
echo -e "\n"
