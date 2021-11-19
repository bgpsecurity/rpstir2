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
	github.com/cpusoft/goutil v1.0.31-0.20211028100045-b5d09dc4ca71
	github.com/gin-gonic/gin v1.7.4
	golang.org/x/net v0.0.0-20210907225631-ff17edfbf26d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	rpstir2-chainvalidate v1.0.1-0.20210907091654-302e80da62b3
	rpstir2-clear v1.0.1-0.20210907092121-ede3bfd569d4
	rpstir2-parsevalidate v1.0.1-0.20210907091644-c8bbfa85950e
	rpstir2-rtrclient v1.0.1-0.20210907092048-b3b2557be2c8
	rpstir2-rtrproducer v1.0.1-0.20210907091704-ed730a09b29d
	rpstir2-rtrserver v1.0.1-0.20211115084130-aebaf4d06ef3
	rpstir2-sync-entire v1.0.1-0.20210907091501-0265fc20bce5
	rpstir2-sync-tal v1.0.1-0.20210907091140-9ebcbfd1e8a6
	rpstir2-sys v1.0.1-0.20210907092151-11f7870a16ef
)
