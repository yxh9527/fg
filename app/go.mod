module app

go 1.24.0

require (
	github.com/alicebob/miniredis/v2 v2.35.0
	github.com/redis/go-redis/v9 v9.17.3
	github.com/shopspring/decimal v1.3.1
	gorm.io/gorm v1.31.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	golang.org/x/text v0.20.0 // indirect
)

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go v1.2.6
