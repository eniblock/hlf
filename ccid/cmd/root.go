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
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
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
	Args: cobra.MinimumNArgs(1),
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
	command := args[1:]
	label, err := cmd.Flags().GetString("label")
	if err != nil {
		log.WithError(err).Fatal("Getting flag 'label' failed")
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		log.WithError(err).Fatal("Getting flag 'name' failed")
	}
	tarBytes := generateTar(address, label)
	shaSum := sha256.Sum256(tarBytes)
	shaSumStr := hex.EncodeToString(shaSum[:])
	// save the final tar, if needed
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		log.WithError(err).Fatal("Getting flag 'output' failed")
	}
	if output != "" {
		err := os.WriteFile(output, tarBytes, 0644)
		if err != nil {
			log.WithError(err).Fatal("Can't write final tar")
		}
	}

	if len(command) > 0 {
		env := getEnv()
		if name != "" {
			env["CCID"] = name + ":" + shaSumStr
			env["CHAINCODE_ID"] = env["CCID"]
			env["CHAINCODE_CCID"] = env["CCID"]
			log.Info("ccid: " + env["CCID"])
		}
		env["CHAINCODE_NAME"] = name
		env["CHAINCODE_LABEL"] = label
		env["CHAINCODE_CONNECTION_ADDRESS"] = address
		addressSplit := strings.SplitN(address, ":", 2)
		if len(addressSplit) == 2 {
			env["CHAINCODE_HOST"] = addressSplit[0]
			env["CHAINCODE_PORT"] = addressSplit[1]
		}
		execPath := command[0]
		if !filepath.IsAbs(execPath) {
			execPath, err = exec.LookPath(command[0])
			if err != nil {
				log.WithError(err).Fatal("Can't find the command in the PATH")
			}
		}
		err := syscall.Exec(execPath, command, envToList(env))
		// we should never get there, unless we can even run the command
		if err != nil {
			log.WithError(err).Fatal("Can't run the command")
		}
	} else {
		if name != "" {
			fmt.Println(name + ":" + shaSumStr)
		} else {
			// show the sha256 of the tar
			fmt.Println(shaSumStr)
		}
	}
}

func getEnv() map[string]string {
	env := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.SplitN(item, "=", 2)
		env[splits[0]] = splits[1]
	}
	return env
}

func envToList(env map[string]string) []string {
	res := []string{}
	for key, value := range env {
		res = append(res, key+"="+value)
	}
	return res
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
		gw := gzip.NewWriter(&codeBuf)
		tw := tar.NewWriter(gw)
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
			log.WithError(err).Fatal("Can't add connection.json header")
		}
		if _, err := tw.Write(connectionBytes); err != nil {
			log.WithError(err).Fatal("Can't add connection.json content")
		}
		if err := tw.Close(); err != nil {
			log.WithError(err).Fatal("Can't generate code.tar.gz")
		}
		if err := gw.Close(); err != nil {
			log.WithError(err).Fatal("Can't generate code.tar.gz")
		}
	}
	// create final tar
	var tarBuf bytes.Buffer
	{
		gw := gzip.NewWriter(&tarBuf)
		tw := tar.NewWriter(gw)
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
			log.WithError(err).Fatal("Can't add code.tar.gz header")
		}
		if _, err := tw.Write(codeBuf.Bytes()); err != nil {
			log.WithError(err).Fatal("Can't add code.tar.gz content")
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
			log.WithError(err).Fatal("Can't add metadata.json header")
		}
		if _, err := tw.Write(metadataBytes); err != nil {
			log.WithError(err).Fatal("Can't add metadata.json content")
		}
		if err := tw.Close(); err != nil {
			log.WithError(err).Fatal("Can't generate final tar")
		}
		if err := gw.Close(); err != nil {
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
	rootCmd.Flags().StringP("label", "l", "", "Labels of the external chaincode")
	rootCmd.Flags().StringP("name", "n", "", "External chaincode name")
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
