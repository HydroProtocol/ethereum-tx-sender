module git.ddex.io/infrastructure/ethereum-launcher

go 1.12

require (
	cloud.google.com/go v0.43.0 // indirect
	git.ddex.io/lib/log v0.0.0-20190729100049-f91fdcf0b05c
	git.ddex.io/lib/monitor v0.0.0-20190730042322-bfe12938b7da
	github.com/HydroProtocol/hydro-sdk-backend v0.0.41
	github.com/HydroProtocol/nights-watch v0.1.8
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/golang/protobuf v1.3.2
	github.com/jinzhu/gorm v1.9.10
	github.com/labstack/echo v3.3.10+incompatible
	github.com/onrik/ethrpc v0.0.0-20190305112807-6b8e9c0e9a8f
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.3.0
	github.com/tidwall/gjson v1.3.2 // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/sys v0.0.0-20190726091711-fc99dfbffb4e // indirect
	google.golang.org/grpc v1.22.1
)

replace golang.org/x/net => github.com/golang/net v0.0.0-20190620200207-3b0461eec859
