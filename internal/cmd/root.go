/*
Copyright © 2024 Giacomo Grangia
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cc-memcached-go",
	Short: "Coding challenges: build your own memcached in golang.",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		port := viper.GetInt("port")
		listenSoc := &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
		}
		tcpConn, err := net.ListenTCP("tcp", listenSoc)
		if err != nil {
			fmt.Println("Error listening: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Listening on port", port)

		defer tcpConn.Close()

		for {
			conn, err := tcpConn.Accept()
			if err != nil {
				fmt.Println("Error accepting connections: ", err.Error())
				os.Exit(1)
			}
			go handleRequest(conn)
		}

	},
}

func handleRequest(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			// Check for EOF
			if err == io.EOF {
				fmt.Println("Client closed the connection")
			} else {
				fmt.Println("Error reading: ", err.Error())
			}
			break
		}
		fmt.Println(string(message))
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cc-memcached-go.yaml)")
	rootCmd.PersistentFlags().IntP("port", "p", 11211, "listening port")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cc-memcached-go" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cc-memcached-go")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
