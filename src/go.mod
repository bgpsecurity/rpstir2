module rpstir2

go 1.18

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
	rpstir2-chainvalidate v0.0.0-00010101000000-000000000000
	rpstir2-clear v0.0.0-00010101000000-000000000000
	rpstir2-parsevalidate v0.0.0-00010101000000-000000000000
	rpstir2-rtrclient v0.0.0-00010101000000-000000000000
	rpstir2-rtrproducer v0.0.0-00010101000000-000000000000
	rpstir2-rtrserver v0.0.0-00010101000000-000000000000
	rpstir2-sync-entire v0.0.0-00010101000000-000000000000
	rpstir2-sync-tal v0.0.0-00010101000000-000000000000
	rpstir2-sys v0.0.0-00010101000000-000000000000
)

require (
	github.com/ant0ine/go-json-rest v3.3.2+incompatible // indirect
	github.com/cpusoft/go-json-rest v4.0.0+incompatible // indirect
	github.com/cpusoft/goutil v1.0.33-0.20221115093718-436ad686fb84
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.8.1
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.11.1 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/guregu/null v4.0.0+incompatible // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/onsi/ginkgo v1.15.0 // indirect
	github.com/onsi/gomega v1.10.5 // indirect
	github.com/parnurzeal/gorequest v0.2.17-0.20200918112808-3a0cb377f571 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	golang.org/x/crypto v0.0.0-20220926161630-eccd6366d1be // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20220926192436-02166a98028e // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	golang.org/x/sys v0.0.0-20220926163933-8cfa568d3c25 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/protobuf v1.28.2-0.20220920080600-7a48e2b66218 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	moul.io/http2curl v1.0.0 // indirect
	rpstir2-model v0.0.0-00010101000000-000000000000 // indirect
	rpstir2-parsevalidate-openssl v0.0.0-00010101000000-000000000000 // indirect
	rpstir2-parsevalidate-packet v0.0.0-00010101000000-000000000000 // indirect
	rpstir2-sync-core v0.0.0-00010101000000-000000000000 // indirect
	xorm.io/builder v0.3.12 // indirect
	xorm.io/xorm v1.3.2 // indirect
)
