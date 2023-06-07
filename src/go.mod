module rpstir2

go 1.19

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
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	rpstir2-model v1.0.1-0.20230602021126-da8e9d252004 // indirect
	rpstir2-parsevalidate-openssl v0.0.0-00010101000000-000000000000 // indirect
	rpstir2-parsevalidate-packet v0.0.0-00010101000000-000000000000 // indirect
	rpstir2-sync-core v0.0.0-00010101000000-000000000000 // indirect
)

require (
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/cpusoft/goutil v1.0.33-0.20230602020845-83903abb3d93
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.9.1
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.1 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/guregu/null v4.0.0+incompatible // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/parnurzeal/gorequest v0.2.17-0.20200918112808-3a0cb377f571 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b // indirect
	github.com/shiena/ansicolor v0.0.0-20230509054315-a9deabde6e02 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sync v0.2.0
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	moul.io/http2curl v1.0.0 // indirect
	xorm.io/builder v0.3.12 // indirect
	xorm.io/xorm v1.3.2 // indirect
)
