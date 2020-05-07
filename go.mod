module github.com/alethio/eth2stats-client

require (
	github.com/alethio/eth2stats-proto v0.0.0-20200122120216-4625b646ae41
	github.com/davecgh/go-spew v1.1.1
	github.com/dghubble/sling v1.3.0
	github.com/elazarl/goproxy v0.0.0-20191011121108-aa519ddbe484 // indirect
	github.com/gin-gonic/gin v1.5.0
	github.com/golang/protobuf v1.3.5
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kwix/logrus-module-formatter v0.0.0-20190702125859-070a70371a97
	github.com/magiconair/properties v1.8.1
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pkg/errors v0.8.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4
	github.com/prometheus/common v0.4.0
	github.com/prysmaticlabs/ethereumapis v0.0.0-20200414190010-6607cc86ddb7
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.5.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8 // indirect
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/genproto v0.0.0-20200323114720-3f67cca34472 // indirect
	google.golang.org/grpc v1.28.0
	gopkg.in/go-playground/validator.v9 v9.29.1
	gopkg.in/yaml.v2 v2.2.8 // indirect
	moul.io/http2curl v1.0.0 // indirect
)

go 1.13

replace github.com/ugorji/go/codec => github.com/ugorji/go/codec v1.1.7

replace github.com/alethio/eth2stats-proto => /Users/alex/Documents/Work/go-mod/github.com/alethio/eth2stats-proto
