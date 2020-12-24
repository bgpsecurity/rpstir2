module rpstir2

go 1.15

// replace .. => .. latest
replace (
	chainvalidate/chainvalidate => ./chainvalidate/chainvalidate
	chainvalidate/chainvalidatehttp => ./chainvalidate/chainvalidatehttp
	chainvalidate/db => ./chainvalidate/db
	chainvalidate/model => ./chainvalidate/model
	github.com/cpusoft/beego => github.com/astaxie/beego v1.12.3
	model => ./model
	parsevalidate/db => ./parsevalidate/db
	parsevalidate/model => ./parsevalidate/model
	parsevalidate/openssl => ./parsevalidate/openssl
	parsevalidate/packet => ./parsevalidate/packet
	parsevalidate/parsevalidate => ./parsevalidate/parsevalidate
	parsevalidate/parsevalidateconsole => ./parsevalidate/parsevalidateconsole
	parsevalidate/parsevalidatehttp => ./parsevalidate/parsevalidatehttp
	parsevalidate/util => ./parsevalidate/util
	rrdp/db => ./rrdp/db
	rrdp/model => ./rrdp/model
	rrdp/rrdp => ./rrdp/rrdp
	rrdp/rrdphttp => ./rrdp/rrdphttp
	rsync/db => ./rsync/db
	rsync/model => ./rsync/model
	rsync/rsync => ./rsync/rsync
	rsync/rsynchttp => ./rsync/rsynchttp
	rtr/db => ./rtr/db
	rtr/model => ./rtr/model
	rtr/redis => ./rtr/redis
	rtr/rtr => ./rtr/rtr
	rtr/rtrhttp => ./rtr/rtrhttp
	rtr/rtrtcp => ./rtr/rtrtcp
	rtr/rtrtcpclient => ./rtr/rtrtcpclient
	rtr/rtrtcpserver => ./rtr/rtrtcpserver
	rtrproducer/db => ./rtrproducer/db
	rtrproducer/model => ./rtrproducer/model
	rtrproducer/rtr => ./rtrproducer/rtr
	rtrproducer/rtrhttp => ./rtrproducer/rtrhttp
	sync/db => ./sync/db
	sync/model => ./sync/model
	sync/sync => ./sync/sync
	sync/synchttp => ./sync/synchttp
	sys/db => ./sys/db
	sys/model => ./sys/model
	sys/sys => ./sys/sys
	sys/syshttp => ./sys/syshttp
	tal/tal => ./tal/tal
	tal/talhttp => ./tal/talhttp

)

// git log  to see commit hash
// get first 7 hash as latest version
require (
	chainvalidate/chainvalidate v0.0.0-00010101000000-000000000000 // indirect
	chainvalidate/chainvalidatehttp v0.0.0-00010101000000-000000000000
	chainvalidate/db v0.0.0-00010101000000-000000000000 // indirect
	chainvalidate/model v0.0.0-00010101000000-000000000000 // indirect
	github.com/astaxie/beego v1.12.3
	github.com/cpusoft/go-json-rest v4.0.0+incompatible
	github.com/cpusoft/goutil latest
	github.com/google/go-cmp v0.5.0 // indirect
	github.com/guregu/null v4.0.0+incompatible // indirect
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	model v0.0.0-00010101000000-000000000000 // indirect
	parsevalidate/db v0.0.0-00010101000000-000000000000 // indirect
	parsevalidate/model v0.0.0-00010101000000-000000000000 // indirect
	parsevalidate/openssl v0.0.0-00010101000000-000000000000 // indirect
	parsevalidate/packet v0.0.0-00010101000000-000000000000 // indirect
	parsevalidate/parsevalidate v0.0.0-00010101000000-000000000000 // indirect
	parsevalidate/parsevalidatehttp v0.0.0-00010101000000-000000000000
	parsevalidate/util v0.0.0-00010101000000-000000000000 // indirect
	rrdp/db v0.0.0-00010101000000-000000000000 // indirect
	rrdp/model v0.0.0-00010101000000-000000000000 // indirect
	rrdp/rrdp v0.0.0-00010101000000-000000000000 // indirect
	rrdp/rrdphttp v0.0.0-00010101000000-000000000000
	rsync/db v0.0.0-00010101000000-000000000000 // indirect
	rsync/model v0.0.0-00010101000000-000000000000 // indirect
	rsync/rsync v0.0.0-00010101000000-000000000000 // indirect
	rsync/rsynchttp v0.0.0-00010101000000-000000000000
	rtr/db v0.0.0-00010101000000-000000000000 // indirect
	rtr/model v0.0.0-00010101000000-000000000000 // indirect
	rtr/rtrhttp v0.0.0-00010101000000-000000000000
	rtr/rtrtcp v0.0.0-00010101000000-000000000000 // indirect
	rtr/rtrtcpclient v0.0.0-00010101000000-000000000000 // indirect
	rtr/rtrtcpserver v0.0.0-00010101000000-000000000000
	rtrproducer/db v0.0.0-00010101000000-000000000000 // indirect
	rtrproducer/rtr v0.0.0-00010101000000-000000000000 // indirect
	rtrproducer/rtrhttp v0.0.0-00010101000000-000000000000
	sync/db v0.0.0-00010101000000-000000000000 // indirect
	sync/sync v0.0.0-00010101000000-000000000000 // indirect
	sync/synchttp v0.0.0-00010101000000-000000000000
	sys/db v0.0.0-00010101000000-000000000000 // indirect
	sys/model v0.0.0-00010101000000-000000000000 // indirect
	sys/sys v0.0.0-00010101000000-000000000000 // indirect
	sys/syshttp v0.0.0-00010101000000-000000000000
	tal/tal v0.0.0-00010101000000-000000000000 // indirect
	tal/talhttp v0.0.0-00010101000000-000000000000
)
