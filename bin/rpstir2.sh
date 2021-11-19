#!/bin/bash

configFile="../conf/project.conf"
source ./read-conf.sh
# `ReadINIfile "file" "[section]" "item" `
programDir=`ReadINIfile "$configFile" "rpstir2-rp" "programDir" `
serverHost=`ReadINIfile "$configFile" "rpstir2-rp" "serverHost" `
serverHttpsPort=`ReadINIfile "$configFile" "rpstir2-rp" "serverHttpsPort" `
serverHttpPort=`ReadINIfile "$configFile" "rpstir2-rp" "serverHttpPort" `
echo  ${serverHost}
echo  ${serverHttpsPort}


function startFunc()
{
  nohup ./rpstir2  >> ../log/nohup.log 2>&1 &
  echo -e "\nyou can view the running status through the log files in ../log/.\n"
  return 0
}


function stopFunc()
{
  pidhttp=`ps -ef|grep 'rpstir2'|grep -v grep|grep -v 'rpstir2.sh' |awk '{print $2}'`
  echo "The current rpstir2 process id is $pidhttp"
  for pid in $pidhttp
  do
    if [ "$pid" = "" ]; then
      echo "rpstir2 is not running"
    else
      kill  $pid
      echo "shutdown rpstir2 success"
 	fi
  done
  return 0
}


function buildSrc()
{
  # go mod   
  cd ../src
  go mod tidy
  
  # go install: go tool compile -help
  export CGO_ENABLED=0
  export GOOS=linux
  export GOARCH=amd64
  go build 
  mv ./rpstir2 ../bin/rpstir2
  chmod +x ../bin/*
  return 0
}


function helpFunc()
{
    echo "rpstir2.sh help:"
    echo -e "./rpstir2.sh start\t\t\tstart rpstir2."
    echo -e "./rpstir2.sh stop\t\t\tstop rpstir2."  
    echo -e "./rpstir2.sh rebuild\t\t\trebuild rpstir2 by yourself(need have Go language compilation environment)."

    echo -e "./rpstir2.sh init\t\t\t(need start first) (re)initialize database."
    echo -e "./rpstir2.sh sync\t\t\t(need start first) download rpki data by sync or rrdp, and need use 'states' to get result."
    echo -e "./rpstir2.sh fullsync\t\t\t(need start first) force full sync data, other functions are similar to sync." 
    echo -e "./rpstir2.sh state\t\t\t(need start first) when it shows 'isRunning:false', it means that synchronization and validation processes are completed." 
    echo -e "./rpstir2.sh results\t\t\t(need start first) shows the valid, warning and invalid number of cer, roa, mft and crl respectively."
    echo -e "./rpstir2.sh exportroas\t\t\t(need start first) exports all roas which are valid or warning."
    echo -e "./rpstir2.sh parse *.cer/crl/mft/roa\t(need start first) uploads file(*.cer/*.crl/*.mft/*.roa) to parse."
    echo -e "./rpstir2.sh help\t\t\tshow this help."
}

case $1 in
  start | begin)
    echo "start rpstir2 server"
    startFunc
    ;;
  stop | end | shutdown | shut)
    echo "stop rpstir2 server"
    stopFunc
    ;;
  
  init)
    echo "initialize rpstir2 database"
    echo ${serverHost}
    echo ${serverHttpsPort}
    curl -s -k -d '{"sysStyle": "init"}'  -H "Content-type: application/json" -X POST https://${serverHost}:${serverHttpsPort}/sys/initreset
    echo -e "\n"
    ;;
  rebuild)
    echo "rebuild rpstir2 by yourself"  
    buildSrc
    ;;


  sync)
    echo "start rpstir2 sync"
    echo ${serverHost}
    echo ${serverHttpsPort}
    curl -s -k -d '{"syncStyle": "sync"}'  -H "Content-type: application/json" -X POST https://${serverHost}:${serverHttpsPort}/entiresync/syncstart
    echo -e "\n"
    ;;
  crontab )
    source /etc/profile
    source /root/.bashrc
    echo "start rpstir2 crontab sync"
    echo ${serverHost}
    echo ${serverHttpsPort}
    curl -s -k -d '{"syncStyle": "sync"}'  -H "Content-type: application/json" -X POST https://${serverHost}:${serverHttpsPort}/entiresync/syncstart
    echo -e "\n"
    ;; 
  fullsync ) 
    echo "start rpstir2 fullsync"
    echo ${serverHost}
    echo ${serverHttpsPort}
    echo "POST https://www.baidu.com/"
    curl -v -k -d '{"sysStyle": "fullsync","syncPolicy":"entire"}'  -H "Content-type: application/json" -X POST https://${serverHost}:${serverHttpsPort}/sys/initreset
    echo -e "\n"
    ;;  
   
  state )    
    echo "start rpstir2 states"
    curl -s -k -d '{"operate":"get"}'  -H "Content-type: application/json" -X POST https://${serverHost}:${serverHttpsPort}/sys/servicestate
    echo -e "\n"
    ;;   
  results )    
    echo "start rpstir2 results"
    curl -s -k -d "" -X POST https://${serverHost}:${serverHttpsPort}/sys/results
    echo -e "\n"
    ;;     
  exportroas )    
    echo "start rpstir2 export roas"
    curl -s -k -d "" -X POST https://${serverHost}:${serverHttpsPort}/sys/exportroas
    echo -e "\n"
    ;;     
   
  parse) 
    echo "start rpstir2 parse"
    checkFile $2
    curl -s -k -F  "file=@${2}" http://${serverHost}:${serverHttpPort}/parsevalidate/parsefile
    echo -e "\n"
    ;;   

  help)
    helpFunc
    ;;      
  *)
    helpFunc
    ;;
esac
echo -e "\n"

