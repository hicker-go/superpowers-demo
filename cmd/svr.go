/*
Copyright © 2026 qinzj
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// svrCmd represents the svr command
var svrCmd = &cobra.Command{
	Use:   "svr",
	Short: "Start the SSO OIDC server",
	Long:  `Start the HTTP server for SSO OIDC login and authentication.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		v := viper.New()
		v.SetConfigFile("configs/settings.yaml")
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("read config: %w", err)
		}
		cfg := v.AllSettings()
		b, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(svrCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// svrCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// svrCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
