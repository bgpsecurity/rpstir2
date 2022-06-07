module rpstir2

go 1.16

replace (
	rpstir2-chainvalidate => ./rpstir2-chainvalidate
	rpstir2-clear => ./rpstir2-clear
	rpstir2-model => ./rpstir2-model
	rpstir2-parsevalidate => ./rpstir2-parsevalidate
	rpstir2-parsevalidate-openssl => ./rpstir2-parsevalidate-openssl
	rpstir2-parsevalidate-packet => ./rpstir2-parsevalidate-packet
	rpstir2-rtrclient => ./rpstir2-rtrclient
	rpstir2-rtrproducer => ./rpstir2-rtrproducer
	rpstir2-rtrserver => ./rpstir2-rtrserver
	rpstir2-sync => ./rpstir2-sync
	rpstir2-sync-core => ./rpstir2-sync-core
	rpstir2-sync-entire => ./rpstir2-sync-entire
	rpstir2-sync-tal => ./rpstir2-sync-tal
	rpstir2-sys => ./rpstir2-sys
)

require (
	github.com/cpusoft/goutil v1.0.33-0.20220607031057-949adfc35ea5
	github.com/gin-gonic/gin v1.8.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	rpstir2-chainvalidate v0.0.0-00010101000000-000000000000
	rpstir2-clear v0.0.0-00010101000000-000000000000
	rpstir2-model v0.0.0-00010101000000-000000000000 // indirect
	rpstir2-parsevalidate v0.0.0-00010101000000-000000000000
	rpstir2-parsevalidate-openssl v0.0.0-00010101000000-000000000000 // indirect
	rpstir2-parsevalidate-packet v0.0.0-00010101000000-000000000000 // indirect
	rpstir2-rtrclient v0.0.0-00010101000000-000000000000
	rpstir2-rtrproducer v0.0.0-00010101000000-000000000000
	rpstir2-rtrserver v0.0.0-00010101000000-000000000000
	rpstir2-sync-core v0.0.0-00010101000000-000000000000 // indirect
	rpstir2-sync-entire v0.0.0-00010101000000-000000000000
	rpstir2-sync-tal v0.0.0-00010101000000-000000000000
	rpstir2-sys v0.0.0-00010101000000-000000000000

)
