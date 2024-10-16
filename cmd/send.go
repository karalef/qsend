package cmd

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"qsend/app"

	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:     "send",
	Short:   "Send a file(s) or directory(s) from this host",
	Long:    "Send a file(s) or directory(s) from this host",
	Aliases: []string{"s"},
	Example: `# Send /path/file.txt
qsend send /path/file.txt
or
qsend /path/file.txt

# Zip file1.txt, file2.txt and directory, then send the zip package
qsend ./file1.txt ./file2.txt ./directory
`,
	Args: cobra.MinimumNArgs(1),
	RunE: sendRun,
}

func sendRun(command *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	a, err := app.New(appName, cfg)
	if err != nil {
		return err
	}
	file, temp, err := fileFromArgs(args)
	if err != nil {
		return err
	}
	if temp {
		defer os.Remove(file)
	}
	u, srv, err := a.Send(context.Background(), file)
	if err != nil {
		return err
	}

	fmt.Println(`Scan the following URL with a QR reader to start the file transfer, press CTRL+C to exit:`)
	fmt.Println(u)
	fmt.Println(app.QRString(u, cfg.Reversed.Value))
	srv.Wait()
	return nil
}

func fileFromArgs(args []string) (path string, temp bool, err error) {
	temp = len(args) > 1
	files := make(map[string]fs.FileInfo, len(args))
	for _, arg := range args {
		fi, err := os.Stat(arg)
		if err != nil {
			return "", false, err
		}
		if fi.IsDir() {
			temp = true
		}
		files[arg] = fi
	}
	if !temp {
		return args[0], temp, nil
	}

	tmpfile, err := os.CreateTemp("", "qsend*.zip")
	if err != nil {
		return "", false, err
	}
	defer tmpfile.Close()

	writer := zip.NewWriter(tmpfile)
	for filepath, fi := range files {
		if fi.IsDir() {
			err := writer.AddFS(os.DirFS(filepath))
			if err != nil {
				return "", temp, err
			}
		} else {
			file, err := os.Open(filepath)
			if err != nil {
				return "", temp, err
			}
			defer file.Close()

			fh, err := zip.FileInfoHeader(fi)
			if err != nil {
				return "", temp, err
			}
			w, err := writer.CreateHeader(fh)
			if err != nil {
				return "", temp, err
			}

			if _, err = io.Copy(w, file); err != nil {
				return "", temp, err
			}
		}
	}
	return tmpfile.Name(), temp, writer.Close()
}
