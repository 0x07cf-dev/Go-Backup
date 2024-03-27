/*
Copyright © 2024 0x07cf-dev <0x07cf@pm.me>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TODO downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Transfers from the remote to the machine",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: remoteArg,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("download called")
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
