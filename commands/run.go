package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/alethio/eth2stats-client/core"
)

const RetryInterval = time.Second * 12

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Connect to the eth2stats server and start sending data",
	Run: func(cmd *cobra.Command, args []string) {
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, syscall.SIGINT)
		signal.Notify(stopChan, syscall.SIGTERM)

		go func() {
			for {
				c := core.New(core.Config{
					Eth2stats: core.Eth2statsConfig{
						Version:    fmt.Sprintf("eth2stats-client/%s", RootCmd.Version),
						ServerAddr: viper.GetString("eth2stats.addr"),
						TLS:        viper.GetBool("eth2stats.tls"),
						NodeName:   viper.GetString("eth2stats.node-name"),
					},
					BeaconNode: core.BeaconNodeConfig{
						Type:        viper.GetString("beacon.type"),
						Addr:        viper.GetString("beacon.addr"),
						MetricsAddr: viper.GetString("beacon.metrics-addr"),
					},
					DataFolder: viper.GetString("data.folder"),
				})

				err := c.Run()
				if err != nil {
					log.Error(err)
				}

				// we're only getting here if there's been no error during set up
				time.Sleep(time.Second * 12)
				log.Info("retrying...")
			}
		}()

		select {
		case <-stopChan:
			log.Info("got stop signal. finishing work.")
			log.Info("work done. goodbye!")
		}
	},
}

func init() {
	runCmd.Flags().String("eth2stats.addr", "", "Eth2stats server address")
	viper.BindPFlag("eth2stats.addr", runCmd.Flag("eth2stats.addr"))

	runCmd.Flags().String("eth2stats.node-name", "", "The name this node will have on Eth2stats")
	viper.BindPFlag("eth2stats.node-name", runCmd.Flag("eth2stats.node-name"))

	runCmd.Flags().Bool("eth2stats.tls", true, "Enable/disable TLS for eth2stats server connection")
	viper.BindPFlag("eth2stats.tls", runCmd.Flag("eth2stats.tls"))

	runCmd.Flags().String("beacon.type", "", "Beacon node type [prysm, lighthouse]")
	viper.BindPFlag("beacon.type", runCmd.Flag("beacon.type"))

	runCmd.Flags().String("beacon.addr", "", "Beacon node endpoint address")
	viper.BindPFlag("beacon.addr", runCmd.Flag("beacon.addr"))

	runCmd.Flags().String("beacon.metrics-addr", "", "The url where the beacon client exposes metrics (used for memory usage)")
	viper.BindPFlag("beacon.metrics-addr", runCmd.Flag("beacon.metrics-addr"))

	runCmd.Flags().String("data.folder", "./data", "Folder in which to persist data")
	viper.BindPFlag("data.folder", runCmd.Flag("data.folder"))
}
