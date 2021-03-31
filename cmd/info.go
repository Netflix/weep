package cmd

import (
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/netflix/weep/metadata"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	infoCmd.PersistentFlags().BoolVarP(&infoDecode, "decode", "d", false, "decode weep info output")
	infoCmd.PersistentFlags().BoolVarP(&infoRaw, "raw", "R", false, "print raw info output")
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:    "info",
	Short:  infoShortHelp,
	Long:   infoLongHelp,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if infoDecode {
			return DecodeWeepInfo(args, cmd.OutOrStdout())
		} else {
			return PrintWeepInfo(cmd.OutOrStdout())
		}
	},
}

func marshalStruct(obj interface{}) []byte {
	out, err := yaml.Marshal(obj)
	if err != nil {
		log.Errorf("failed to marshal struct: %v", err)
		return nil
	}
	return out
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func PrintWeepInfo(w io.Writer) error {
	var writer io.Writer
	if infoRaw {
		writer = w
	} else {
		b64encoder := base64.NewEncoder(base64.StdEncoding, w)
		defer b64encoder.Close()
		writer = zlib.NewWriter(b64encoder)
	}
	if closeable, ok := writer.(io.WriteCloser); ok {
		defer closeable.Close()
	}

	roles, err := roleList(true)
	if err != nil {
		log.Errorf("failed to retrieve role list from ConsoleMe: %v", err)
	} else {
		_, _ = writer.Write([]byte(roles))
	}

	// We're ignoring errors here in the interest of best-effort information gathering
	_, _ = writer.Write([]byte("\nVersion\n"))
	_, _ = writer.Write(marshalStruct(metadata.GetVersion()))

	_, _ = writer.Write([]byte("\nConfiguration\n"))
	_, _ = writer.Write(marshalStruct(viper.AllSettings()))

	_, _ = writer.Write([]byte("\nHost Info\n"))
	_, _ = writer.Write(marshalStruct(metadata.GetInstanceInfo()))

	return nil
}

func DecodeWeepInfo(args []string, w io.Writer) error {
	var r io.Reader
	if isInputFromPipe() {
		// Input is being piped in to weep
		r = os.Stdin
	} else {
		// Input should be the first arg, but we can't trust that
		if len(args) > 0 {
			r = strings.NewReader(args[0])
		} else {
			return fmt.Errorf("must pass arg") // TODO: handle error better
		}
	}
	b64decoder := base64.NewDecoder(base64.StdEncoding, r)
	zreader, err := zlib.NewReader(b64decoder)
	defer zreader.Close()
	if err != nil {
		return err
	}

	io.Copy(w, zreader)

	return nil
}
