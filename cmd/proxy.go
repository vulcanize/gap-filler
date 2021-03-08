package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vulcanize/gap-filler/pkg/mux"
)

var (
	ErrNoRpcEndpoints = errors.New("no rpc endpoints is available")

	proxyCmd = &cobra.Command{
		Use: "proxy",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			fmt.Println()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			gqlDefaultAddr, err := url.Parse(viper.GetString("gql.default"))
			if err != nil {
				return err
			}

			gqlTracingAPIAddr, err := url.Parse(viper.GetString("gql.tracing"))
			if err != nil {
				return err
			}

			rpcClients, err := parseRpcAddresses(viper.GetString("rpc.eth"))
			if err != nil {
				logrus.Error("bad rpc.eth addresses")
				return err
			}

			tracingClients, err := parseRpcAddresses(viper.GetString("rpc.tracing"))
			if err != nil {
				logrus.Error("bad rpc.tracing addresses")
				return err
			}

			router, err := mux.NewServeMux(&mux.Options{
				BasePath:       viper.GetString("http.path"),
				EnableGraphiQL: viper.GetBool("gql.gui"),
				Postgraphile: mux.PostgraphileOptions{
					Default:    gqlDefaultAddr,
					TracingAPI: gqlTracingAPIAddr,
				},
				RPC: mux.RPCOptions{
					DefaultClients: rpcClients,
					TracingClients: tracingClients,
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

func parseRpcAddresses(value string) ([]*rpc.Client, error) {
	rpcAddresses := strings.Split(value, ",")
	rpcClients := make([]*rpc.Client, 0, len(rpcAddresses))
	for _, address := range rpcAddresses {
		rpcClient, err := rpc.Dial(address)
		if err != nil {
			logrus.Errorf("couldn't connect to %s. Error: %s", address, err)
			continue
		}

		rpcClients = append(rpcClients, rpcClient)
	}

	if len(rpcClients) == 0 {
		logrus.Error(ErrNoRpcEndpoints)
		return nil, ErrNoRpcEndpoints
	}

	return rpcClients, nil
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	// flags
	proxyCmd.PersistentFlags().String("http-host", "127.0.0.1", "http host")
	proxyCmd.PersistentFlags().String("http-port", "8080", "http port")
	proxyCmd.PersistentFlags().String("http-path", "/", "http base path")

	proxyCmd.PersistentFlags().String("rpc-eth", "http://127.0.0.1:8545", "comma separated ethereum rpc addresses. Example http://127.0.0.1:8545,http://127.0.0.2:8545")
	proxyCmd.PersistentFlags().String("rpc-tracing", "http://127.0.0.1:8000", "comma separated traicing api addresses")

	proxyCmd.PersistentFlags().String("gql-default", "http://127.0.0.1:5020/graphql", "postgraphile address")
	proxyCmd.PersistentFlags().String("gql-tracing", "http://127.0.0.1:5020/graphql", "tracing api postgraphile address")
	proxyCmd.PersistentFlags().Bool("gql-gui", false, "enable graphiql interface")

	// and their .toml config bindings
	viper.BindPFlag("http.host", proxyCmd.PersistentFlags().Lookup("http-host"))
	viper.BindPFlag("http.port", proxyCmd.PersistentFlags().Lookup("http-port"))
	viper.BindPFlag("http.path", proxyCmd.PersistentFlags().Lookup("http-path"))

	viper.BindPFlag("rpc.eth", proxyCmd.PersistentFlags().Lookup("rpc-eth"))
	viper.BindPFlag("rpc.tracing", proxyCmd.PersistentFlags().Lookup("rpc-tracing"))

	viper.BindPFlag("gql.default", proxyCmd.PersistentFlags().Lookup("gql-default"))
	viper.BindPFlag("gql.tracing", proxyCmd.PersistentFlags().Lookup("gql-tracing"))
	viper.BindPFlag("gql.gui", proxyCmd.PersistentFlags().Lookup("gql-gui"))
}
