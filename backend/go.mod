module github.com/Mousa96/chatting-service

go 1.21

toolchain go1.24.3

// Force dependencies to use versions compatible with Go 1.21
replace (
	github.com/aws/aws-sdk-go-v2 => github.com/aws/aws-sdk-go-v2 v1.21.2
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream => github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.14
	github.com/aws/aws-sdk-go-v2/config => github.com/aws/aws-sdk-go-v2/config v1.18.45
	github.com/aws/aws-sdk-go-v2/internal/configsources => github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.38
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 => github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.32
	github.com/aws/aws-sdk-go-v2/internal/v4a => github.com/aws/aws-sdk-go-v2/internal/v4a v1.1.6
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding => github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.15
	github.com/aws/aws-sdk-go-v2/service/internal/checksum => github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.33
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url => github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.32
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared => github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.15.6
	github.com/aws/aws-sdk-go-v2/service/s3 => github.com/aws/aws-sdk-go-v2/service/s3 v1.40.2
	golang.org/x/crypto => golang.org/x/crypto v0.15.0
	golang.org/x/net => golang.org/x/net v0.17.0
	golang.org/x/sys => golang.org/x/sys v0.15.0
	golang.org/x/text => golang.org/x/text v0.14.0
	golang.org/x/tools => golang.org/x/tools v0.13.0
)

require (
	github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3
	github.com/golang-jwt/jwt/v4 v4.4.2
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/gorilla/websocket v1.5.3
	github.com/lib/pq v1.10.9
	github.com/stretchr/testify v1.10.0
	github.com/swaggo/http-swagger v1.3.4
	github.com/swaggo/swag v1.16.4
	golang.org/x/crypto v0.17.0
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/smithy-go v1.22.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/swaggo/files v0.0.0-20220610200504-28940afbdbfe // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/tools v0.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
