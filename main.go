package main

import (
	"dredge/app/cmd"
	"dredge/app/util"
	"dredge/app/util/mylog"
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.szostok.io/version/extension"
)

func main() {
	mylog.Preinit()

	fmt.Fprintln(os.Stderr, util.Banner)

	rootCmd := &cobra.Command{Use: "dredge"}
	rootCmd.AddCommand(cmd.Run)
	rootCmd.AddCommand(extension.NewVersionCobraCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
		return
	}
}
