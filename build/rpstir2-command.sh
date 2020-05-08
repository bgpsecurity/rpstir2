#!/bin/bash
cd "$(dirname "$0")";
configFile="../conf/project.conf"
source $(pwd)/read-conf.sh

function helpFunc()
{
    echo "rpstir2-command.sh help:"
    echo -e "1) ./rpstir2-command.sh rsync:                     it downloads rpki data by rsync, and need use '3)' to get result "
    echo -e "2) ./rpstir2-command.sh crontab:                   it just uses in crontab, other functions are similar to rsync"
    echo -e "3) ./rpstir2-command.sh rrdp:                      it downloads rpki data by rrdp(delta), and need use '3)' to get result " 
    echo -e "4) ./rpstir2-command.sh states:                    when it shows 'state:end', it means rsync/rrdp is end" 
    echo -e "5) ./rpstir2-command.sh results:                   it shows the valid, warning and invalid number of cer, roa, mft and crl respectively."
    echo -e "6) ./rpstir2-command.sh reset:                     it resets all data in mysql and in local cache" 
    echo -e "7) ./rpstir2-command.sh parse *.cer/crl/mft/roa:   it uploads file(*.cer/*.crl/*.mft/*.roa) to parse"
    echo -e "8) ./rpstir2-command.sh slurm *.json:              it uploads slurm file(*.json)"
    echo -e "*) ./rpstir2-command.sh:                           it shows this help"
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
  init ) 
    echo "start rpstir2 init"
    # `ReadINIfile "file" "[section]" "item" `
    sysserver=`ReadINIfile "$configFile" "rpstir2" "sysserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d "" http://$sysserver:$httpport/sys/init
    ;; 
  rsyncstart | rsync)
    echo "start rpstir2 rsync"
    # `ReadINIfile "file" "[section]" "item" `
    rsyncserver=`ReadINIfile "$configFile" "rpstir2" "rsyncserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d "" http://$rsyncserver:$httpport/rsync/start
    ;;
  crontab )
    source /etc/profile
    source /root/.bashrc
    echo "start rpstir2 crontab rsync"
    # `ReadINIfile "file" "[section]" "item" `
    rsyncserver=`ReadINIfile "$configFile" "rpstir2" "rsyncserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d "" http://$rsyncserver:$httpport/rsync/start
    ;;  
  rrdpstart | rrdp | delta)
    echo "start rpstir2 rrdp"
    # `ReadINIfile "file" "[section]" "item" `
    rrdpserver=`ReadINIfile "$configFile" "rpstir2" "rrdpserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d "" http://$rrdpserver:$httpport/rrdp/start
    ;;
  states | sumstates)    
    echo "start rpstir2 states"
    # `ReadINIfile "file" "[section]" "item" `
    sysserver=`ReadINIfile "$configFile" "rpstir2" "sysserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d "" http://$sysserver:$httpport/sys/summarystates
    ;;   
  results|result )    
    echo "start rpstir2 results"
    # `ReadINIfile "file" "[section]" "item" `
    sysserver=`ReadINIfile "$configFile" "rpstir2" "sysserver" `
    #echo $sysserver 
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    #echo $httpport
    # curl
    #echo "curl -d \"\" http://$sysserver:$httpport/sys/results"
    curl -d "" http://$sysserver:$httpport/sys/results
    ;;      
  reset ) 
    echo "start rpstir2 reset"
    # `ReadINIfile "file" "[section]" "item" `
    sysserver=`ReadINIfile "$configFile" "rpstir2" "sysserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d "" http://$sysserver:$httpport/sys/reset
    ;;  
  parsefile | parse) 
    echo "start rpstir2 parse"
    checkFile $2
    # `ReadINIfile "file" "[section]" "item" `
    parsevalidateserver=`ReadINIfile "$configFile" "rpstir2" "parsevalidateserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -F  "file=@${2}" http://$parsevalidateserver:$httpport/parsevalidate/parsefile
    ;;         
  slurmupload | slurm) 
    echo "start rpstir2 slurm"
    checkFile $2
    # `ReadINIfile "file" "[section]" "item" `
    slurmserver=`ReadINIfile "$configFile" "rpstir2" "slurmserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -F  "file=@${2}" http://$slurmserver:$httpport/slurm/upload
    ;;   
  help)
    helpFunc
    ;;  
  *)
    helpFunc
    ;;
 esac
 


