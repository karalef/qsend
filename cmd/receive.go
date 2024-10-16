package cmd

import (
	"context"
	"fmt"
	"qsend/app"

	"github.com/spf13/cobra"
)

var receiveCmd = &cobra.Command{
	Use:     "receive",
	Aliases: []string{"r"},
	Short:   "Receive files",
	Long:    "Receive files. The destination directory can be set with the config, or by passing the --output flag. If none of the above are set, the current working directory will be used as a destination directory.",
	Example: `# Receive files in the current directory (if not set in the config)
qsend receive
or
qsend
# Receive files in a specific directory
qsend --output /tmp
`,
	RunE: receiveRun,
}

func receiveRun(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	a, err := app.New(appName, cfg)
	if err != nil {
		return err
	}

	u, srv, err := a.Receive(context.Background())
	if err != nil {
		return err
	}

	fmt.Println(`Scan the following URL with a QR reader to start the file transfer, press CTRL+C to exit:`)
	fmt.Println(u)
	fmt.Println(app.QRString(u, cfg.Reversed.Value))
	srv.Wait()
	return nil
}
