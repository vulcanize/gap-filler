package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vulcanize/gap-filler/pkg/mux"
)

var (
	proxyCmd = &cobra.Command{
		Use: "proxy",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			fmt.Println()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			gqlAddr, err := url.Parse(viper.GetString("gql.target"))
			if err != nil {
				return err
			}

			gqlTracingAPIAddr, err := url.Parse(viper.GetString("gql.target"))
			if err != nil {
				return err
			}

			rpcClient, err := rpc.Dial(viper.GetString("eth.rpc"))
			if err != nil {
				return err
			}

			tracingClient, err := rpc.Dial(viper.GetString("tracingapi.rpc"))
			if err != nil {
				return err
			}

			router, err := mux.NewServeMux(&mux.Options{
				BasePath:       viper.GetString("http.path"),
				EnableGraphiQL: viper.GetBool("gql.gui"),
				Postgraphile: mux.PostgraphileOptions{
					Default:    gqlAddr,
					TracingAPI: gqlTracingAPIAddr,
				},
				RPC: mux.RPCOptions{
					Default: rpcClient,
					Tracing: tracingClient,
				},
			})
			if err != nil {
				logrus.Info(err)
				return err
			}

			addr := fmt.Sprintf("%s:%s", viper.GetString("http.host"), viper.GetString("http.port"))
			return http.ListenAndServe(addr, router)
		},
	}
)

func init() {
	rootCmd.AddCommand(proxyCmd)

	// flags
	proxyCmd.PersistentFlags().String("http-host", "127.0.0.1", "http host")
	proxyCmd.PersistentFlags().String("http-port", "8080", "http port")
	proxyCmd.PersistentFlags().String("http-path", "/", "http base path")

	proxyCmd.PersistentFlags().String("eth-rpc", "http://127.0.0.1:8545", "ethereum rpc address")
	proxyCmd.PersistentFlags().String("tracing-rpc", "http://127.0.0.1:8545", "traicing api address")

	proxyCmd.PersistentFlags().String("gql-target", "http://127.0.0.1:5020/graphql", "postgraphile address")
	proxyCmd.PersistentFlags().String("gql-target-tracingapi", "http://127.0.0.1:5020/graphql", "tracing api postgraphile address")
	proxyCmd.PersistentFlags().Bool("gql-gui", false, "enable graphiql interface")

	// and their .toml config bindings
	viper.BindPFlag("http.host", proxyCmd.PersistentFlags().Lookup("http-host"))
	viper.BindPFlag("http.port", proxyCmd.PersistentFlags().Lookup("http-port"))
	viper.BindPFlag("http.path", proxyCmd.PersistentFlags().Lookup("http-path"))

	viper.BindPFlag("eth.rpc", proxyCmd.PersistentFlags().Lookup("eth-rpc"))
	viper.BindPFlag("tracing.rpc", proxyCmd.PersistentFlags().Lookup("tracing-rpc"))

	viper.BindPFlag("gql.target", proxyCmd.PersistentFlags().Lookup("gql-target"))
	viper.BindPFlag("gql.target.tracingapi", proxyCmd.PersistentFlags().Lookup("gql-target-tracingapi"))
	viper.BindPFlag("gql.gui", proxyCmd.PersistentFlags().Lookup("gql-gui"))
}
