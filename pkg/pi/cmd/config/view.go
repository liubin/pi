/*
Copyright 2014 The Kubernetes Authors.

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

package config

import (
	"fmt"
	"io"

	"github.com/hyperhq/client-go/tools/clientcmd"
	clientcmdapi "github.com/hyperhq/client-go/tools/clientcmd/api"
	"github.com/hyperhq/client-go/tools/clientcmd/api/latest"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"
	"github.com/hyperhq/pi/pkg/printers"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/util/flag"
)

type ViewOptions struct {
	ConfigAccess clientcmd.ConfigAccess
	Merge        flag.Tristate
	Flatten      bool
	Minify       bool
	RawByteData  bool
}

var (
	view_long = templates.LongDesc(`
		Display pi config file.`)

	view_example = templates.Examples(`
		# Show pi config settings(available flag --minify --output).
		pi config view

		# Show current user current-context info
		pi config view --minify=true

		# Output as json
		pi config view --output=json
		`)
)

func NewCmdConfigView(out, errOut io.Writer, ConfigAccess clientcmd.ConfigAccess) *cobra.Command {
	options := &ViewOptions{ConfigAccess: ConfigAccess}
	// Default to yaml
	defaultOutputFormat := "yaml"

	cmd := &cobra.Command{
		Use:     "view",
		Short:   i18n.T("Display pi config file"),
		Long:    view_long,
		Example: view_example,
		Run: func(cmd *cobra.Command, args []string) {
			options.Complete()
			outputFormat := cmdutil.GetFlagString(cmd, "output")
			if outputFormat == "wide" || outputFormat == "name" {
				fmt.Fprintf(errOut, "--output '%v' is not available in pi config view; reset to default output format (%s)\n\n", outputFormat, defaultOutputFormat)
				// TODO: once printing is abstracted, this should be handled at flag declaration time
				//cmd.Flags().Set("output", defaultOutputFormat)
			}
			if outputFormat == "" {
				fmt.Fprintf(errOut, "Reset to default output format (%s) as --output is empty\n", defaultOutputFormat)
				// TODO: once printing is abstracted, this should be handled at flag declaration time
				cmd.Flags().Set("output", defaultOutputFormat)
			}

			printOpts := cmdutil.ExtractCmdPrintOptions(cmd, false)
			printer, err := cmdutil.PrinterForOptions(meta.NewDefaultRESTMapper(nil, nil), latest.Scheme, nil, []runtime.Decoder{latest.Codec}, printOpts)
			cmdutil.CheckErr(err)
			printer = printers.NewVersionedPrinter(printer, latest.Scheme, latest.ExternalVersion)

			cmdutil.CheckErr(options.Run(out, printer))
		},
	}

	cmdutil.AddPrinterFlags(cmd)
	cmd.Flags().Set("output", defaultOutputFormat)

	//options.Merge.Default(true)
	//f := cmd.Flags().VarPF(&options.Merge, "merge", "", "Merge the full hierarchy of pi config files")
	//f.NoOptDefVal = "true"
	//cmd.Flags().BoolVar(&options.RawByteData, "raw", false, "Display raw byte data")
	//cmd.Flags().BoolVar(&options.Flatten, "flatten", false, "Flatten the resulting pi config file into self-contained output (useful for creating portable pi config files)")
	cmd.Flags().BoolVar(&options.Minify, "minify", false, "Remove all information not used by current-context from the output")
	return cmd
}

func (o ViewOptions) Run(out io.Writer, printer printers.ResourcePrinter) error {
	config, err := o.loadConfig()
	if err != nil {
		return err
	}

	if o.Minify {
		if err := clientcmdapi.MinifyConfig(config); err != nil {
			return err
		}
	}

	//if o.Flatten {
	//	if err := clientcmdapi.FlattenConfig(config); err != nil {
	//		return err
	//	}
	//} else if !o.RawByteData {
	//	clientcmdapi.ShortenConfig(config)
	//}

	err = printer.PrintObj(config, out)
	if err != nil {
		return err
	}

	return nil
}

func (o *ViewOptions) Complete() bool {
	//if o.ConfigAccess.IsExplicitFile() {
	//	if !o.Merge.Provided() {
	//		o.Merge.Set("false")
	//	}
	//}

	return true
}

func (o ViewOptions) loadConfig() (*clientcmdapi.Config, error) {
	err := o.Validate()
	if err != nil {
		return nil, err
	}

	config, err := o.getStartingConfig()
	return config, err
}

func (o ViewOptions) Validate() error {
	//if !o.Merge.Value() && !o.ConfigAccess.IsExplicitFile() {
	//	return errors.New("if merge==false a precise file must to specified")
	//}

	return nil
}

// getStartingConfig returns the Config object built from the sources specified by the options, the filename read (only if it was a single file), and an error if something goes wrong
func (o *ViewOptions) getStartingConfig() (*clientcmdapi.Config, error) {
	switch {
	//case !o.Merge.Value():
	//	return clientcmd.LoadFromFile(o.ConfigAccess.GetExplicitFile())

	default:
		return o.ConfigAccess.GetStartingConfig()
	}
}
