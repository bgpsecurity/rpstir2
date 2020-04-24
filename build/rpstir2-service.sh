#!/bin/sh



function startFunc()
{
  cd "$(dirname "$0")";
  configFile="../conf/project.conf"
  source $(pwd)/read-conf.sh

  rpstir2_program_dir=$(ReadINIfile $configFile rpstir2 programdir) 
  cd $rpstir2_program_dir/bin
  ./rpstir2-http &
  ./rpstir2-rtr-tcp &
  return 0
}
function stopFunc()
{
  pidhttp=`ps -ef|grep 'rpstir2-http'|grep -v grep|awk '{print $2}'`
  echo "The current rpstir2-http process id is $pidhttp"
  for pid in $pidhttp
  do
    if [ "$pidhttp" = "" ]; then
      echo "pidhttp is null"
    else
      kill  $pidhttp
      echo "shutdown rpstir2-http success"
 	fi
  done

  pidtcp=`ps -ef|grep 'rpstir2-rtr-tcp'|grep -v grep|awk '{print $2}'`
  echo "The current rpstir2-rtr-tcp process id is $pidtcp"
  for pid in $pidtcp
  do
    if [ "$pidtcp" = "" ]; then
      echo "pidtcp is null"
    else
      kill  $pidtcp
      echo "shutdown rpstir2-rtr-tcp success"
 	fi
  done
  return 0
} 
function deployFunc()
{
  cd "$(dirname "$0")";
  curpath=$(pwd)
  configFile="../conf/project.conf"
  source $(pwd)/read-conf.sh

  rpstir2_program_dir=$(ReadINIfile $configFile rpstir2 programdir)
  rpstir2_source_dir=$(ReadINIfile $configFile rpstir2 sourcedir)
  rpstir2_data_dir=$(ReadINIfile $configFile rpstir2 datadir)

  echo "program directory is " $rpstir2_program_dir
  echo "source directory is " $rpstir2_source_dir
  echo "data directory is " $rpstir2_data_dir

  mkdir -p ${rpstir2_program_dir} ${rpstir2_program_dir}/bin    ${rpstir2_program_dir}/conf  ${rpstir2_program_dir}/log  
  mkdir -p ${rpstir2_data_dir}    ${rpstir2_data_dir}/rsyncrepo ${rpstir2_data_dir}/rrdprepo ${rpstir2_data_dir}/slurm  ${rpstir2_data_dir}/tal 
     
  mkdir -p $GOPATH/src/golang.org/x
  cd  $GOPATH/src/golang.org/x
  git clone https://github.com/golang/crypto.git
  git clone https://github.com/golang/net.git
  go get -u github.com/astaxie/beego/logs
  go get -u github.com/go-xorm/xorm
  go get -u github.com/go-sql-driver/mysql
  go get -u github.com/go-xorm/core
  go get -u github.com/parnurzeal/gorequest
  go get -u github.com/ant0ine/go-json-rest
  go get -u github.com/satori/go.uuid
  go get -u github.com/cpusoft/go-json-rest
  go get -u github.com/cpusoft/goutil

  cd ${rpstir2_source_dir}
  oldgopath=$GOPATH
  CGO_ENABLED=0
  GOOS=linux
  GOARCH=amd64
  GOPATH=$GOPATH:$rpstir2_source_dir
  # see: go tool compile -help
  go install -v -gcflags "-N -l" ./...
  export GOPATH=$oldgopath
  cp ${rpstir2_source_dir}/bin/*                    ${rpstir2_program_dir}/bin/
  cp ${rpstir2_source_dir}/build/rpstir2-command.sh ${rpstir2_program_dir}/bin/
  cp ${rpstir2_source_dir}/build/rpstir2-service.sh ${rpstir2_program_dir}/bin/
  cp ${rpstir2_source_dir}/build/read-conf.sh       ${rpstir2_program_dir}/bin/
  cp -r ${rpstir2_source_dir}/build/tal/*           ${rpstir2_data_dir}/tal/
  cp -r ${rpstir2_source_dir}/conf/*                ${rpstir2_program_dir}/conf/
  

  chmod +x ${rpstir2_source_dir}/build/*.sh
  chmod +x ${rpstir2_program_dir}/bin/*
  
  # init
  cd ${rpstir2_program_dir}/bin
  ./rpstir2-http &
  ./rpstir2-rtr-tcp &
  sysserver=$(ReadINIfile $configFile rpstir2 sysserver) 
  echo $sysserver 
  httpport=$(ReadINIfile $configFile rpstir2 httpport) 
  echo $httpport
  # curl
  echo "curl -d \"\" http://$sysserver:$httpport/sys/init"
  curl -d "" http://$sysserver:$httpport/sys/init
  
  cd $curpath
  return 0
}


function updateFunc()
{
  cd "$(dirname "$0")";
  curpath=$(pwd)
  configFile="../conf/project.conf"
  source $(pwd)/read-conf.sh

  rpstir2_program_dir=$(ReadINIfile $configFile rpstir2 programdir)
  rpstir2_source_dir=$(ReadINIfile $configFile rpstir2 sourcedir)
  rpstir2_data_dir=$(ReadINIfile $configFile rpstir2 datadir)

  echo "program directory is " $rpstir2_program_dir
  echo "source directory is " $rpstir2_source_dir
  echo "data directory is " $rpstir2_data_dir

  mkdir -p ${rpstir2_program_dir} ${rpstir2_program_dir}/bin    ${rpstir2_program_dir}/conf  ${rpstir2_program_dir}/log  
  mkdir -p ${rpstir2_data_dir}    ${rpstir2_data_dir}/rsyncrepo ${rpstir2_data_dir}/rrdprepo ${rpstir2_data_dir}/slurm  ${rpstir2_data_dir}/tal 
 
 
  cd ${rpstir2_source_dir}
  go get -u github.com/cpusoft/goutil
  git_dir="${rpstir2_source_dir}/.git"
  if [ -d ${git_dir} ];then
    git checkout .
    git pull
  else
    svn update --accept tf 
  fi


  oldgopath=$GOPATH
  CGO_ENABLED=0
  GOOS=linux
  GOARCH=amd64
  GOPATH=$GOPATH:$rpstir2_source_dir
  # see: go tool compile -help
  go install -v -gcflags "-N -l" ./...
  export GOPATH=$oldgopath

  # will not copy conf/project.conf
  cp ${rpstir2_source_dir}/bin/*                    ${rpstir2_program_dir}/bin/
  cp ${rpstir2_source_dir}/build/rpstir2-command.sh ${rpstir2_program_dir}/bin/
  cp ${rpstir2_source_dir}/build/rpstir2-service.sh ${rpstir2_program_dir}/bin/
  cp ${rpstir2_source_dir}/build/read-conf.sh       ${rpstir2_program_dir}/bin/
  cp -r ${rpstir2_source_dir}/build/tal/*           ${rpstir2_data_dir}/tal/
  
  chmod +x ${rpstir2_source_dir}/build/*.sh
  chmod +x ${rpstir2_program_dir}/bin/*  
  cd $curpath
  return 0
}

function helpFunc()
{
    echo "rpstir2-service.sh help:"
    echo "1) ./rpstir2-service.sh deploy: deploy rpstir2, just run once"
    echo "2) ./rpstir2-service.sh update: update rpstir2. It will stop rpstir2, and update source code (not update project.conf) and rebuild, then restart rpstir2"     
    echo "3) ./rpstir2-service.sh start:  start rpstir2 service"
    echo "4) ./rpstir2-service.sh stop:   stop rpstir2 service" 
    echo "*) ./rpstir2-service.sh:        it will show this help"
}

case $1 in
  start | begin)
    echo "start rpstir2 http and tcp server"
    startFunc
    ;;
  stop | end | shutdown | shut)
    echo "stop rpstir2 http and tcp server"
    stopFunc
    ;;
  deploy)
    echo "deploy rpstir2"
    stopFunc
    deployFunc
    ;; 
  update | rebuild)
    echo "deploy rpstir2"
    stopFunc
    updateFunc
    startFunc
    ;; 
  help)
    helpFunc
    ;;      
  *)
    helpFunc
    ;;
esac
echo -e "\n"

