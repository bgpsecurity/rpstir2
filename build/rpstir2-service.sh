 #!/bin/sh
 
case $1 in
  start | begin)
    echo "start rpstir2 http and tcp server"
    ./rpstir2-http &
    ./rpstir2-rtr-tcp &
    ;;
  stop | end | shutdown | shut)
    echo "stop rpstir2 http and tcp server"
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
    ;;
  *)
    echo "rpstir2-service.sh help:"
    echo "1). start service: rpstir2-service.sh start"
    echo "2). stop service: rpstir2-service.sh stop" 
    ;;
 esac
 


