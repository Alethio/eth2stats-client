package commands

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/alethio/eth2stats-client/core"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Connect to the eth2stats server and start sending data",
	Run: func(cmd *cobra.Command, args []string) {
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, syscall.SIGINT)
		signal.Notify(stopChan, syscall.SIGTERM)

		c := core.New(core.Config{
			Eth2stats: core.Eth2statsConfig{
				ServerAddr: viper.GetString("eth2stats.addr"),
				TLS:        viper.GetBool("eth2stats.tls"),
				NodeName:   viper.GetString("eth2stats.node-name"),
			},
			BeaconNodeType: viper.GetString("beacon.type"),
			BeaconNodeAddr: viper.GetString("beacon.addr"),
		})
		go c.Run()

		select {
		case <-stopChan:
			log.Info("Got stop signal. Finishing work.")

			c.Close()

			log.Info("Work done. Goodbye!")
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

	runCmd.Flags().String("beacon.type", "", "Beacon node type [prysm]")
	viper.BindPFlag("beacon.type", runCmd.Flag("beacon.type"))

	runCmd.Flags().String("beacon.addr", "", "Beacon node endpoint address")
	viper.BindPFlag("beacon.addr", runCmd.Flag("beacon.addr"))
}
