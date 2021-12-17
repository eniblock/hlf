/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ccid",
	Short: "Hyper Ledger Fabric External Chaincode Identifier Generator",
	// Long: ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Args: cobra.ExactArgs(2),
	Run:  mainRun,
}

type Connection struct {
	Address     string `json:"address"`
	DialTimeout string `json:"dial_timeout"`
	TlsRequired bool   `json:"tls_required"`
}

type Metadata struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

func mainRun(cmd *cobra.Command, args []string) {
	address := args[0]
	label := args[1]
	tarBytes := generateTar(address, label)
	// save the final tar, if needed
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		log.WithError(err).Fatal("Getting flag 'retention' failed")
	}
	if output != "" {
		err := os.WriteFile(output, tarBytes, 0644)
		if err != nil {
			log.WithError(err).Fatal("Can't write final tar")
		}
	}
	// show the sha256 of the tar
	shaSum := sha256.Sum256(tarBytes)
	fmt.Println(hex.EncodeToString(shaSum[:]))
}

func generateTar(address string, label string) []byte {
	metadata := Metadata{
		Path:  "",
		Type:  "external",
		Label: label,
	}
	connection := Connection{
		Address:     address,
		DialTimeout: "10s",
		TlsRequired: false,
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		log.WithError(err).Fatal("Can't serialize metadata")
	}
	connectionBytes, err := json.Marshal(connection)
	if err != nil {
		log.WithError(err).Fatal("Can't serialize connection")
	}
	// Create code.tar.gz
	var codeBuf bytes.Buffer
	{
		tw := tar.NewWriter(&codeBuf)
		if err := tw.WriteHeader(&tar.Header{
			Name:       "connection.json",
			Mode:       0600,
			Size:       int64(len(connectionBytes)),
			Uid:        0,
			Gid:        0,
			ModTime:    time.Unix(0, 0),
			AccessTime: time.Unix(0, 0),
			ChangeTime: time.Unix(0, 0),
		}); err != nil {
			log.WithError(err).Fatal("Can't add connection.json")
		}
		if _, err := tw.Write(connectionBytes); err != nil {
			log.WithError(err).Fatal("Can't add connection.json")
		}
		if err := tw.Close(); err != nil {
			log.WithError(err).Fatal("Can't generate code.tar.gz")
		}
	}
	// create final tar
	var tarBuf bytes.Buffer
	{
		tw := tar.NewWriter(&tarBuf)
		// add code.tar.gz
		if err := tw.WriteHeader(&tar.Header{
			Name:       "code.tar.gz",
			Mode:       0600,
			Size:       int64(codeBuf.Len()),
			Uid:        0,
			Gid:        0,
			ModTime:    time.Unix(0, 0),
			AccessTime: time.Unix(0, 0),
			ChangeTime: time.Unix(0, 0),
		}); err != nil {
			log.WithError(err).Fatal("Can't add code.tar.gz")
		}
		if _, err := tw.Write(codeBuf.Bytes()); err != nil {
			log.WithError(err).Fatal("Can't add code.tar.gz")
		}
		// add metadata.json
		if err := tw.WriteHeader(&tar.Header{
			Name:       "metadata.json",
			Mode:       0600,
			Size:       int64(len(metadataBytes)),
			Uid:        0,
			Gid:        0,
			ModTime:    time.Unix(0, 0),
			AccessTime: time.Unix(0, 0),
			ChangeTime: time.Unix(0, 0),
		}); err != nil {
			log.WithError(err).Fatal("Can't add code.tar.gz")
		}
		if _, err := tw.Write(metadataBytes); err != nil {
			log.WithError(err).Fatal("Can't add code.tar.gz")
		}
		if err := tw.Close(); err != nil {
			log.WithError(err).Fatal("Can't generate final tar")
		}
	}
	return tarBuf.Bytes()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringP("output", "o", "", "Output path for the final tar")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ccid" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ccid")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
