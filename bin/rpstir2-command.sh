#!/bin/bash
cd "$(dirname "$0")";
configFile="../conf/project.conf"
source $(pwd)/read-conf.sh

function helpFunc()
{
    echo "rpstir2-command.sh usage:"
    echo "./rpstir2-command.sh <command> [arguments]"
    echo "The commands are:"
    echo -e "\t sync                     it downloads rpki data by sync or rrdp, and need use ' states' to get result "
    echo -e "\t rsync                    it downloads rpki data only by rsync, and need use ' states' to get result "
    echo -e "\t rrdp                     it downloads rpki data only by rrdp, and need use ' states' to get result "
    echo -e "\t crontab                  it just uses in crontab, other functions are similar to sync"
    echo -e "\t fullsync                 it forece full sync data, other functions are similar to sync" 
    echo -e "\t states                   when it shows 'state:end', it means rsync/rrdp is end" 
    echo -e "\t results                  it shows the valid, warning and invalid number of cer, roa, mft and crl respectively."
    echo -e "\t resetall                 it resets all data in mysql and in local cache" 
    echo -e "\t parse *.cer/crl/mft/roa  it uploads file(*.cer/*.crl/*.mft/*.roa) to parse"
    echo -e "\t slurm *.json             it uploads slurm file(*.json)"
    echo -e "\t help                     it shows this help"
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
    # `ReadINIfile "file" "[section]" "item" `
    syncserver=`ReadINIfile "$configFile" "rpstir2" "syncserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d '{"syncStyle": "sync"}'  -H "Content-type: application/json" -X POST http://$syncserver:$httpport/sync/start
    ;;
  rsync)
    echo "start rpstir2 rsync"
    # `ReadINIfile "file" "[section]" "item" `
    rsyncserver=`ReadINIfile "$configFile" "rpstir2" "rsyncserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d '{"syncStyle": "rsync"}'  -H "Content-type: application/json" -X POST http://$rsyncserver:$httpport/sync/start
    ;;
  rrdp)
    echo "start rpstir2 rrdp"
    # `ReadINIfile "file" "[section]" "item" `
    rrdpserver=`ReadINIfile "$configFile" "rpstir2" "rrdpserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d '{"syncStyle": "rrdp"}'  -H "Content-type: application/json" -X POST http://$rrdpserver:$httpport/sync/start
    ;;        
  crontab )
    source /etc/profile
    source /root/.bashrc
    echo "start rpstir2 crontab sync"
    # `ReadINIfile "file" "[section]" "item" `
    syncserver=`ReadINIfile "$configFile" "rpstir2" "syncserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d '{"syncStyle": "sync"}'  -H "Content-type: application/json" -X POST http://$syncserver:$httpport/sync/start
    ;; 
  fullsync ) 
    echo "start rpstir2 fullsync"
    # `ReadINIfile "file" "[section]" "item" `
    sysserver=`ReadINIfile "$configFile" "rpstir2" "sysserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d '{"syncStyle": "fullsync"}'  -H "Content-type: application/json" -X POST http://$sysserver:$httpport/sys/initreset
    ;;  
  states )    
    echo "start rpstir2 states"
    # `ReadINIfile "file" "[section]" "item" `
    sysserver=`ReadINIfile "$configFile" "rpstir2" "sysserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d "" -X POST http://$sysserver:$httpport/sys/summarystates
    ;;   
  results )    
    echo "start rpstir2 results"
    # `ReadINIfile "file" "[section]" "item" `
    sysserver=`ReadINIfile "$configFile" "rpstir2" "sysserver" `
    #echo $sysserver 
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    #echo $httpport
    # curl
    #echo "curl -d \"\" http://$sysserver:$httpport/sys/results"
    curl -d "" -X POST http://$sysserver:$httpport/sys/results
    ;;      
  resetall ) 
    echo "start rpstir2 resetall"
    # `ReadINIfile "file" "[section]" "item" `
    sysserver=`ReadINIfile "$configFile" "rpstir2" "sysserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -d '{"sysStyle": "resetall"}'  -H "Content-type: application/json" -X POST http://$sysserver:$httpport/sys/initreset
    ;;     
   parse) 
    echo "start rpstir2 parse"
    checkFile $2
    # `ReadINIfile "file" "[section]" "item" `
    parsevalidateserver=`ReadINIfile "$configFile" "rpstir2" "parsevalidateserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    # curl
    curl -F  "file=@${2}" http://$parsevalidateserver:$httpport/parsevalidate/parsefile
    ;;         
  slurm) 
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
 
 echo -e "Now, you can view the running status through the log files in $rpstir2_program_dir/log\n"


