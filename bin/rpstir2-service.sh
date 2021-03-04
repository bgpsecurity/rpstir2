#!/bin/bash



function startFunc()
{
  cd "$(dirname "$0")";
  configFile="../conf/project.conf"
  source $(pwd)/read-conf.sh

  rpstir2_program_dir=$(ReadINIfile $configFile rpstir2 programDir) 
  cd $rpstir2_program_dir/bin
  ./rpstir2 &

  
  echo -e "\nyou can view the running status through the log files in rpstir2/log.\n"
  return 0
}
function stopFunc()
{
  pidhttp=`ps -ef|grep 'rpstir2'|grep -v grep|grep -v 'rpstir2-service.sh' |awk '{print $2}'`
  echo "The current rpstir2 process id is $pidhttp"
  for pid in $pidhttp
  do
    if [ "$pidhttp" = "" ]; then
      echo "pidhttp is null"
    else
      kill  $pidhttp
      echo "shutdown rpstir2 success"
 	fi
  done
  return 0
} 
function deployFunc()
{
  # program/data dir
  cd "$(dirname "$0")";
  curpath=$(pwd)
  configFile="../conf/project.conf"
  source $(pwd)/read-conf.sh
  rpstir2_program_dir=$(ReadINIfile $configFile rpstir2 programDir)
  rpstir2_data_dir=$(ReadINIfile $configFile rpstir2 dataDir)
  echo "program directory is " $rpstir2_program_dir
  echo "data directory is " $rpstir2_data_dir
  mkdir -p ${rpstir2_program_dir} ${rpstir2_program_dir}/bin    ${rpstir2_program_dir}/conf  ${rpstir2_program_dir}/log  
  mkdir -p ${rpstir2_data_dir}    ${rpstir2_data_dir}/rsyncrepo ${rpstir2_data_dir}/rrdprepo ${rpstir2_data_dir}/slurm  ${rpstir2_data_dir}/tal 
     
  # go mod   
  cd ${rpstir2_program_dir}/src
  go get -u github.com/golang/protobuf@v1.4.3
  go get -u github.com/astaxie/beego@v1.12.3
  go get -u github.com/go-xorm/xorm@v0.7.3
  go get -u github.com/go-xorm/core@v0.6.2
  go get -u github.com/ant0ine/go-json-rest@v3.3.3-0.20170913041208-ebb33769ae01+incompatible
  go get -u github.com/cpusoft/go-json-rest@ecdd1cf
  go get -u github.com/cpusoft/goutil@latest
  go mod tidy
  
  # go install: go tool compile -help
  go env -w CGO_ENABLED=0
  go env -w GOOS=linux
  go env -w GOARCH=amd64
  go install -v -gcflags "-N -l" .
  mv $GOPATH/bin/rpstir2 $rpstir2_program_dir/bin/rpstir2
  cp -r ${rpstir2_program_dir}/build/tal/*           ${rpstir2_data_dir}/tal/
  chmod +x ${rpstir2_program_dir}/bin/*
  
  # start
  cd ${rpstir2_program_dir}/bin
  ./rpstir2 &

  # init
  serverHost=$(ReadINIfile $configFile rpstir2 serverHost) 
  serverHttpsPort=$(ReadINIfile $configFile rpstir2 serverHttpsPort) 
  echo "curl -k -d '{\"sysStyle\": \"init\"}'  -H \"Content-type: application/json\" -X POST https://$serverHost:$serverHttpsPort/sys/initreset"
  curl -k -d '{"sysStyle": "init"}'  -H "Content-type: application/json" -X POST https://$serverHost:$serverHttpsPort/sys/initreset
  cd $curpath
  echo -e "\nNow, you can call './rpstir2-command.sh sync' to start RPKI sync.\n"
  return 0
}


function updateFunc()
{
  # program/data dir
  cd "$(dirname "$0")";
  curpath=$(pwd)
  configFile="../conf/project.conf"
  source $(pwd)/read-conf.sh
  rpstir2_program_dir=$(ReadINIfile $configFile rpstir2 programDir)
  rpstir2_data_dir=$(ReadINIfile $configFile rpstir2 dataDir)
  echo "program directory is " $rpstir2_program_dir
  echo "data directory is " $rpstir2_data_dir
  mkdir -p ${rpstir2_program_dir} ${rpstir2_program_dir}/bin    ${rpstir2_program_dir}/conf  ${rpstir2_program_dir}/log  
  mkdir -p ${rpstir2_data_dir}    ${rpstir2_data_dir}/rsyncrepo ${rpstir2_data_dir}/rrdprepo ${rpstir2_data_dir}/slurm  ${rpstir2_data_dir}/tal 
 
  # svn/git update
  cd ${rpstir2_program_dir}
  git_dir="${rpstir2_program_dir}/.git"
  # save local project.conf
  oldConfigFile=$(date +%Y%m%d%H%M%S)
  echo "it will save local conf/project.conf to conf/project.conf.$oldConfigFile.bak, that you can copy your local configuration to new prject.conf, and then start rpstir2."
  if [ -d ${git_dir} ];then
    cp ${rpstir2_program_dir}/conf/project.conf ${rpstir2_program_dir}/conf/project.conf.$oldConfigFile.bak
    git checkout .
    git pull
  else
    cp ${rpstir2_program_dir}/conf/project.conf ${rpstir2_program_dir}/conf/project.conf.$oldConfigFile.bak
    svn update --accept tf 
  fi

  # go mod
  cd ${rpstir2_program_dir}/src
  go get -u github.com/cpusoft/goutil@latest
  go get -u github.com/cpusoft/go-json-rest@ecdd1cf
  go mod tidy


  # go install: go tool compile -help
  go env -w CGO_ENABLED=0
  go env -w GOOS=linux
  go env -w GOARCH=amd64
  go install -v -gcflags "-N -l" .
  mv $GOPATH/bin/rpstir2 $rpstir2_program_dir/bin/rpstir2
  cp -r ${rpstir2_program_dir}/build/tal/*           ${rpstir2_data_dir}/tal/
  chmod +x ${rpstir2_program_dir}/bin/*  
  cd $curpath

  echo -e "it saved local conf/project.conf to conf/project.conf.$oldConfigFile.bak, that you can copy your local configuration to new prject.conf, and then start rpstir2.\n"
  return 0
}

function helpFunc()
{
    echo "rpstir2-service.sh help:"
    echo "1) ./rpstir2-service.sh deploy: deploy rpstir2, just run once"
    echo "2) ./rpstir2-service.sh update: update rpstir2. It will stop rpstir2, update source code (not update project.conf) , and rebuild. But it does not start rpstir2 automatically."     
    echo "3) ./rpstir2-service.sh start:  start rpstir2 service"
    echo "4) ./rpstir2-service.sh stop:   stop rpstir2 service" 
    echo "*) ./rpstir2-service.sh:        it will show this help"
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
  deploy)
    echo "deploy rpstir2"
    stopFunc
    deployFunc
    ;; 
  update | rebuild)
    echo "deploy rpstir2"
    stopFunc
    updateFunc
    ;; 
  help)
    helpFunc
    ;;      
  *)
    helpFunc
    ;;
esac
echo -e "\n"

