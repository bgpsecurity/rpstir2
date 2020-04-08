#!/bin/bash
cd "$(dirname "$0")";
configFile="../conf/project.conf"
source $(pwd)/read-conf.sh

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
  *)
    echo "rpstir2-command.sh help:"
    echo -e "1). system init: rpstir2-command.sh init\n  will init all data in mysql and in local cache" 
    echo -e "1). start rsync: rpstir2-command.sh rsync\n  and need use '3)' periodically to get result "
    echo -e "2). start rrdp: rpstir2-command.sh rrdp\n  and need use '3)' periodically to get result " 
    echo -e "3). get rsync/rrdp state: rpstir2-command.sh sumstates\n  when result shows 'state:end', it means rsync/rrdp is end" 
    echo -e "4). system reset: rpstir2-command.sh resetall\n  will reset all data in mysql and in local cache" 
    echo -e "5). parse cer/crl/mft/roa file: rpstir2-command.sh parsefile $file\n   $file is the acutal file name"
    echo -e "6). upload slurm file: rpstir2-command.sh slurm $file\n   $file is the acutal file name"
    echo -e "7). start roacomp: rpstir2-command.sh roacomp\n"
    ;;
 esac
 


