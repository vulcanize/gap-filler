package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vulcanize/gap-filler-service/pkg/mux"
)

var (
	proxyCmd = &cobra.Command{
		Use:     "proxy",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			addr := fmt.Sprintf("%s:%s", viper.GetString("http.host"), viper.GetString("http.port"))
			router := mux.NewServeMux(&mux.Options{
				PostgraphileAddr: viper.GetString("postgraphile"),
			})
			return http.ListenAndServe(addr, router)
		},
	}
)

func init() {
	rootCmd.AddCommand(proxyCmd)

	// flags
	proxyCmd.PersistentFlags().String("http-host", "127.0.0.1", "http host")
	proxyCmd.PersistentFlags().String("http-port", "8080", "http port")
	proxyCmd.PersistentFlags().String("postgraphile", "http://127.0.0.1:5020/graphql", "postgraphile address")

	// and their .toml config bindings
	viper.BindPFlag("http.host", rootCmd.PersistentFlags().Lookup("http-host"))
	viper.BindPFlag("http.port", rootCmd.PersistentFlags().Lookup("http-port"))
	viper.BindPFlag("postgraphile", rootCmd.PersistentFlags().Lookup("postgraphile"))
}
