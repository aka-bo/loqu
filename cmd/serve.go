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
	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/aka-bo/loquaciousd/pkg/server"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts an HTTP server which logs all lifecycle events",
	Run: func(cmd *cobra.Command, args []string) {
		glog.Info("serve called")

		glog.Infof("calling server.Run()")
		server.Run(options)
	},
}

var options = &server.Options{
	ShutdownDelaySeconds: 15,
	ListenPort:           80,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serveCmd.Flags().IntVarP(&options.ListenPort, "port", "p", options.ListenPort, "The port the service will bind to.")
	serveCmd.Flags().IntVarP(&options.ShutdownDelaySeconds, "shutdown-delay", "s", options.ShutdownDelaySeconds, "The amount of time in seconds to delay on shutdown. Useful for testing graceful termination.")
}
