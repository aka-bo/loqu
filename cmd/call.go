/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/aka-bo/loqu/pkg/client"
)

var clientOptions = &client.Options{
	Host: "localhost",
	Port: 80,
}

// clientCmd represents the client command
var callCmd = &cobra.Command{
	Use:   "call",
	Short: "Execute calls against a web server",
	Run: func(cmd *cobra.Command, args []string) {
		glog.Infoln("call called")
		defer glog.Flush()
		if cmd.Flags().Changed("data") {
			data, err := cmd.Flags().GetString("data")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			clientOptions.Data = &data
		}
		client.Run(clientOptions)
	},
}

func init() {
	rootCmd.AddCommand(callCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	callCmd.Flags().StringVarP(&clientOptions.Host, "host", "H", clientOptions.Host, "The target host")
	callCmd.Flags().IntVarP(&clientOptions.Port, "port", "p", clientOptions.Port, "The port the target host is listening on")
	callCmd.Flags().BoolVar(&clientOptions.UseWebSocket, "ws", false, "Use the websocket protocol will be used for server communications")
	callCmd.Flags().IntVarP(&clientOptions.IntervalSeconds, "interval", "i", 0, "if interval is greater than 0, requests will be sent continuously spaced at specified intervals in seconds. When used in conjuction with the --ws flag, a single websocket connection will be used for all writes")
	callCmd.Flags().StringP("data", "d", "", "Data to send to the target web server")
}
