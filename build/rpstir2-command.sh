#!/bin/bash
cd "$(dirname "$0")";
configFile="../conf/project.conf"
source $(pwd)/read-conf.sh

function helpFunc()
{
    echo "rpstir2-command.sh help:"
    echo -e "1) ./rpstir2-command.sh init:                 it will init all data in mysql and in local cache, just run once" 
    echo -e "1) ./rpstir2-command.sh rsync:                it will download rpki data by rsync, and need use '3)' to get result "
    echo -e "2) ./rpstir2-command.sh rrdp:                 it will download rpki data by rrdp(delta), and need use '3)' to get result " 
    echo -e "3) ./rpstir2-command.sh states:               when result shows 'state:end', it means rsync/rrdp is end" 
    echo -e "4) ./rpstir2-command.sh reset:                it will reset all data in mysql and in local cache" 
    echo -e "5) ./rpstir2-command.sh parsefile /***/file:  it will parse and validate the $file"
    echo -e "6) ./rpstir2-command.sh slurm /***/file:      it will upload slurm $file"
    echo -e "*) ./rpstir2-command.sh:                      it will show this help"
}


case $1 in
  init ) 
    # `ReadINIfile "file" "[section]" "item" `
    sysserver=`ReadINIfile "$configFile" "rpstir2" "sysserver" `
    echo $sysserver 
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    echo $httpport
    # curl
    echo "curl -d \"\" http://$sysserver:$httpport/sys/init"
    curl -d "" http://$sysserver:$httpport/sys/init
    ;; 
  rsyncstart | rsync)
    echo "start rpstir2 rsync"
    # `ReadINIfile "file" "[section]" "item" `
    rsyncserver=`ReadINIfile "$configFile" "rpstir2" "rsyncserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    echo $rsyncserver $httpport
    # curl
    echo "curl -d \"\" http://$rsyncserver:$httpport/rsync/start"
    curl -d "" http://$rsyncserver:$httpport/rsync/start
    ;;
  crontab )
    source /etc/profile
    source /root/.bashrc
    echo "start rpstir2 crontab rsync"
    # `ReadINIfile "file" "[section]" "item" `
    rsyncserver=`ReadINIfile "$configFile" "rpstir2" "rsyncserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    echo $rsyncserver $httpport
    # curl
    echo "curl -d \"\" http://$rsyncserver:$httpport/rsync/start"
    curl -d "" http://$rsyncserver:$httpport/rsync/start
    ;;  
  rrdpstart | rrdp | delta)
    echo "start rpstir2 rrdp"
    # `ReadINIfile "file" "[section]" "item" `
    rrdpserver=`ReadINIfile "$configFile" "rpstir2" "rrdpserver" `
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    echo $rrdpserver $httpport
    # curl
    echo "curl -d \"\" http://$rrdpserver:$httpport/rrdp/start"
    curl -d "" http://$rrdpserver:$httpport/rrdp/start
    ;;

  rtrupdate | rtr)
    # `ReadINIfile "file" "[section]" "item" `
    rtrserver=`ReadINIfile "$configFile" "rpstir2" "rtrserver" `
    echo $rtrserver 
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    echo $httpport
    # curl
    echo "curl -d \"\" http://$rtrserver:$httpport/rtr/update"
    curl -d "" http://$rtrserver:$httpport/rtr/update
    ;;   
  states | sumstates)    
    # `ReadINIfile "file" "[section]" "item" `
    sysserver=`ReadINIfile "$configFile" "rpstir2" "sysserver" `
    echo $sysserver 
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    echo $httpport
    # curl
    echo "curl -d \"\" http://$sysserver:$httpport/sys/summarystates"
    curl -d "" http://$sysserver:$httpport/sys/summarystates
    ;;     
  reset ) 
    # `ReadINIfile "file" "[section]" "item" `
    sysserver=`ReadINIfile "$configFile" "rpstir2" "sysserver" `
    echo $sysserver 
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    echo $httpport
    # curl
    echo "curl -d \"\" http://$sysserver:$httpport/sys/reset"
    curl -d "" http://$sysserver:$httpport/sys/reset
    ;;      
  parsevalidatefile | parsefile) 
    # `ReadINIfile "file" "[section]" "item" `
    parsevalidateserver=`ReadINIfile "$configFile" "rpstir2" "parsevalidateserver" `
    echo $parsevalidateserver 
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    echo $httpport
    # curl
    echo "curl -F  \"file=@${2}\" http://$parsevalidateserver:$httpport/parsevalidate/file"
    curl -F  "file=@${2}" http://$parsevalidateserver:$httpport/parsevalidate/file
    ;;   
  slurmupload | slurm) 
    # `ReadINIfile "file" "[section]" "item" `
    slurmserver=`ReadINIfile "$configFile" "rpstir2" "slurmserver" `
    echo $slurmserver 
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    echo $httpport
    # curl
    echo "curl -F  \"file=@${2}\" http://$slurmserver:$httpport/slurm/upload"
    curl -F  "file=@${2}" http://$slurmserver:$httpport/slurm/upload
    ;;   
  roacomp | roacmp) 
    # `ReadINIfile "file" "[section]" "item" `
    roacompserver=`ReadINIfile "$configFile" "rpstir2" "roacompserver" `
    echo $roacompserver 
    httpport=`ReadINIfile "$configFile" "rpstir2" "httpport" `
    echo $httpport
    # curl
    echo "curl -d \"\" http://$roacompserver:$httpport/roacomp/roacompstart"
    curl -d "" http://$roacompserver:$httpport/roacomp/roacompstart
    ;;          
  help)
    helpFunc
    ;;  
  *)
    helpFunc
    ;;
 esac
 


