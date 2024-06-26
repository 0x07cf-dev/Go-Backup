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
	"github.com/0x07cf-dev/go-backup/internal/backup"
	"github.com/spf13/cobra"
)

// TODO uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Transfers from the machine to the remote",
	Args:  remoteArg,
	Run: func(cmd *cobra.Command, args []string) {
		session := backup.NewSession(ctx,
			backup.WithRemote(remoteDest),
			backup.WithRemoteRoot(remoteRoot),
			backup.WithSimulation(simulate),
			backup.WithInteractivity(!unattended),
			backup.WithDebug(debug),
			backup.WithLanguage(language),
		)
		session.Backup()
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uploadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uploadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
