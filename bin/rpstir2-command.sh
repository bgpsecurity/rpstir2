#!/bin/bash
cd "$(dirname "$0")";
configFile="../conf/project.conf"
source $(pwd)/read-conf.sh
# `ReadINIfile "file" "[section]" "item" `
serverHost=`ReadINIfile "$configFile" "rpstir2" "serverHost" `
serverHttpsPort=`ReadINIfile "$configFile" "rpstir2" "serverHttpsPort" `
serverHttpPort=`ReadINIfile "$configFile" "rpstir2" "serverHttpPort" `
echo $serverHost":"$serverHttpsPort

function helpFunc()
{
    echo "rpstir2-command.sh usage:"
    echo "./rpstir2-command.sh <command> [arguments]"
    echo "The commands are:"
    echo -e " sync                     it downloads rpki data by sync or rrdp, and need use ' states' to get result "
    echo -e " rsync                    it downloads rpki data only by rsync, and need use ' states' to get result "
    echo -e " rrdp                     it downloads rpki data only by rrdp, and need use ' states' to get result "
    echo -e " crontab                  it just uses in crontab, other functions are similar to sync"
    echo -e " fullsync                 it forces full sync data, other functions are similar to sync" 
    echo -e " state                    when it shows 'isRunning:false', it means that synchronization and validation processes are completed" 
    echo -e " results                  it shows the valid, warning and invalid number of cer, roa, mft and crl respectively."
    echo -e " resetall                 it resets all data in mysql and in local cache" 
    echo -e " parse *.cer/crl/mft/roa  it uploads file(*.cer/*.crl/*.mft/*.roa) to parse"
    echo -e " help                     it shows this help"
}

function checkFile()
{
    if [ $# != 1 ] ; then
        echo "file is empty"
        exit 1;
    fi

    if [ ! -f $1 ]; then
        echo "$1 does not exist"
        exit 1;
    fi
}


case $1 in
  sync)
    echo "start rpstir2 sync"
    curl -s -k -d '{"syncStyle": "sync"}'  -H "Content-type: application/json" -X POST https://$serverHost:$serverHttpsPort/sync/start
    echo -e "\n"
    ;;
  rsync)
    echo "start rpstir2 rsync"
    curl -s -k -d '{"syncStyle": "rsync"}'  -H "Content-type: application/json" -X POST https://$serverHost:$serverHttpsPort/sync/start
    echo -e "\n"
    ;;
  rrdp)
    echo "start rpstir2 rrdp"
    curl -s -k -d '{"syncStyle": "rrdp"}'  -H "Content-type: application/json" -X POST https://$serverHost:$serverHttpsPort/sync/start
    echo -e "\n"
    ;;        
  crontab )
    source /etc/profile
    source /root/.bashrc
    echo "start rpstir2 crontab sync"
    curl -s -k -d '{"syncStyle": "sync"}'  -H "Content-type: application/json" -X POST https://$serverHost:$serverHttpsPort/sync/start
    echo -e "\n"
    ;; 
  fullsync ) 
    echo "start rpstir2 fullsync"
    curl -s -k -d '{"sysStyle": "fullsync"}'  -H "Content-type: application/json" -X POST https://$serverHost:$serverHttpsPort/sys/initreset
    echo -e "\n"
    ;;  
   
  state )    
    echo "start rpstir2 states"
    curl -s -k -d '{"operate":"get"}'  -H "Content-type: application/json" -X POST https://$serverHost:$serverHttpsPort/sys/servicestate
    echo -e "\n"
    ;;   
  results )    
    echo "start rpstir2 results"
    curl -s -k -d "" -X POST https://$serverHost:$serverHttpsPort/sys/results
    echo -e "\n"
    ;;      
  resetall ) 
    echo "start rpstir2 resetall"
    curl -s -k -d '{"sysStyle": "resetall"}'  -H "Content-type: application/json" -X POST https://$serverHost:$serverHttpsPort/sys/initreset
    echo -e "\n"
    ;;     
   parse) 
    echo "start rpstir2 parse"
    checkFile $2
    curl -s -k -F  "file=@${2}" http://$serverHost:$serverHttpPort/parsevalidate/parsefile
    echo -e "\n"
    ;;         
   
  help)
    helpFunc
    ;;  
  *)
    helpFunc
    ;;
 esac
 



