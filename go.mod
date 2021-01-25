module github.com/vulcanize/gap-filler

go 1.15

require (
	github.com/ethereum/go-ethereum v1.9.11
	github.com/friendsofgo/graphiql v0.2.2
	github.com/graphql-go/graphql v0.7.9
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.0
	github.com/valyala/fastjson v1.6.3
)

replace github.com/ethereum/go-ethereum v1.9.11 => github.com/vulcanize/go-ethereum v1.9.11-statediff-0.0.8
