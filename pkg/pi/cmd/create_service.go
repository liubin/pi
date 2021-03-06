/*
Copyright 2016 The Kubernetes Authors.

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
	"io"

	"github.com/spf13/cobra"

	"github.com/hyperhq/pi/pkg/pi"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"
	"k8s.io/api/core/v1"
)

// NewCmdCreateService is a macro command to create a new service
func NewCmdCreateService(f cmdutil.Factory, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"svc"},
		Short:   i18n.T("Create a service using specified subcommand"),
		Long:    "Create a service using specified subcommand",
		Run:     cmdutil.DefaultSubCommandRun(errOut),
	}
	cmd.AddCommand(NewCmdCreateServiceClusterIP(f, cmdOut))
	//cmd.AddCommand(NewCmdCreateServiceNodePort(f, cmdOut))
	cmd.AddCommand(NewCmdCreateServiceLoadBalancer(f, cmdOut))
	//cmd.AddCommand(NewCmdCreateServiceExternalName(f, cmdOut))

	return cmd
}

var (
	serviceClusterIPLong = templates.LongDesc(i18n.T(`
    Create a ClusterIP service with the specified name.`))

	serviceClusterIPExample = templates.Examples(i18n.T(`
    # Create a new ClusterIP service named my-cs
    pi create service clusterip my-cs --tcp=5678:8080

    # Create a new ClusterIP service named my-cs (in headless mode)
    pi create service clusterip my-cs --clusterip="None"`))
)

func addPortFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("tcp", []string{}, "Port pairs can be specified as '<port>:<targetPort>'.")
}

// NewCmdCreateServiceClusterIP is a command to create a ClusterIP service
func NewCmdCreateServiceClusterIP(f cmdutil.Factory, cmdOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clusterip NAME [--tcp=<port>:<targetPort>]",
		Short:   i18n.T("Create a ClusterIP service."),
		Long:    serviceClusterIPLong,
		Example: serviceClusterIPExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := CreateServiceClusterIP(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	//cmdutil.AddApplyAnnotationFlags(cmd)
	//cmdutil.AddValidateFlags(cmd)
	//cmdutil.AddPrinterFlags(cmd)
	//cmdutil.AddGeneratorFlags(cmd, cmdutil.ServiceClusterIPGeneratorV1Name)
	addPortFlags(cmd)
	cmd.Flags().String("clusterip", "", i18n.T("Assign your own ClusterIP or set to 'None' for a 'headless' service (no loadbalancing)."))
	return cmd
}

func errUnsupportedGenerator(cmd *cobra.Command, generatorName string) error {
	return cmdutil.UsageErrorf(cmd, "Generator %s not supported. ", generatorName)
}

// CreateServiceClusterIP is the implementation of the create service clusterip command
func CreateServiceClusterIP(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}
	var generator pi.StructuredGenerator
	switch generatorName := cmdutil.ServiceClusterIPGeneratorV1Name; generatorName {
	case cmdutil.ServiceClusterIPGeneratorV1Name:
		generator = &pi.ServiceCommonGeneratorV1{
			Name:      name,
			TCP:       cmdutil.GetFlagStringSlice(cmd, "tcp"),
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: cmdutil.GetFlagString(cmd, "clusterip"),
		}
	default:
		return errUnsupportedGenerator(cmd, generatorName)
	}
	return RunCreateSubcommand(f, cmd, cmdOut, &CreateSubcommandOptions{
		Name:                name,
		StructuredGenerator: generator,
	})
}

var (
	serviceNodePortLong = templates.LongDesc(i18n.T(`
    Create a NodePort service with the specified name.`))

	serviceNodePortExample = templates.Examples(i18n.T(`
    # Create a new NodePort service named my-ns
    pi create service nodeport my-ns --tcp=5678:8080`))
)

// NewCmdCreateServiceNodePort is a macro command for creating a NodePort service
func NewCmdCreateServiceNodePort(f cmdutil.Factory, cmdOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nodeport NAME [--tcp=port:targetPort] [--dry-run]",
		Short:   i18n.T("Create a NodePort service."),
		Long:    serviceNodePortLong,
		Example: serviceNodePortExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := CreateServiceNodePort(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddPrinterFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, cmdutil.ServiceNodePortGeneratorV1Name)
	cmd.Flags().Int("node-port", 0, "Port used to expose the service on each node in a cluster.")
	addPortFlags(cmd)
	return cmd
}

// CreateServiceNodePort is the implementation of the create service nodeport command
func CreateServiceNodePort(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}
	var generator pi.StructuredGenerator
	switch generatorName := cmdutil.GetFlagString(cmd, "generator"); generatorName {
	case cmdutil.ServiceNodePortGeneratorV1Name:
		generator = &pi.ServiceCommonGeneratorV1{
			Name:      name,
			TCP:       cmdutil.GetFlagStringSlice(cmd, "tcp"),
			Type:      v1.ServiceTypeNodePort,
			ClusterIP: "",
			NodePort:  cmdutil.GetFlagInt(cmd, "node-port"),
		}
	default:
		return errUnsupportedGenerator(cmd, generatorName)
	}
	return RunCreateSubcommand(f, cmd, cmdOut, &CreateSubcommandOptions{
		Name:                name,
		StructuredGenerator: generator,
	})
}

var (
	serviceLoadBalancerLong = templates.LongDesc(i18n.T(`
    Create a LoadBalancer service with the specified name.`))

	serviceLoadBalancerExample = templates.Examples(i18n.T(`
    # Create a new LoadBalancer service named my-lbs (x.x.x.x is fip)
    pi create service loadbalancer my-lbs --tcp=5678:8080 -f=x.x.x.x -l=role=web,zone=gcp-us-central1-a`))
)

// NewCmdCreateServiceLoadBalancer is a macro command for creating a LoadBalancer service
func NewCmdCreateServiceLoadBalancer(f cmdutil.Factory, cmdOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "loadbalancer NAME [--tcp=port:targetPort] --loadbalancerip=fip --selector=key=val",
		Short:   i18n.T("Create a LoadBalancer service."),
		Long:    serviceLoadBalancerLong,
		Example: serviceLoadBalancerExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := CreateServiceLoadBalancer(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	//cmdutil.AddApplyAnnotationFlags(cmd)
	//cmdutil.AddValidateFlags(cmd)
	//cmdutil.AddPrinterFlags(cmd)
	//cmdutil.AddGeneratorFlags(cmd, cmdutil.ServiceLoadBalancerGeneratorV1Name)
	addPortFlags(cmd)
	cmd.Flags().StringP("loadbalancerip", "f", "", "Set fip as LoadBalancerIP")
	cmd.Flags().StringSliceP("selector", "l", []string{}, "Labels selectors for pods")
	return cmd
}

// CreateServiceLoadBalancer is the implementation of the create service loadbalancer command
func CreateServiceLoadBalancer(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}
	var generator pi.StructuredGenerator
	switch generatorName := cmdutil.ServiceLoadBalancerGeneratorV1Name; generatorName {
	case cmdutil.ServiceLoadBalancerGeneratorV1Name:
		generator = &pi.ServiceCommonGeneratorV1{
			Name:           name,
			TCP:            cmdutil.GetFlagStringSlice(cmd, "tcp"),
			Type:           v1.ServiceTypeLoadBalancer,
			ClusterIP:      "",
			LoadBalancerIP: cmdutil.GetFlagString(cmd, "loadbalancerip"),
			Selector:       cmdutil.GetFlagStringSlice(cmd, "selector"),
		}
	default:
		return errUnsupportedGenerator(cmd, generatorName)
	}
	return RunCreateSubcommand(f, cmd, cmdOut, &CreateSubcommandOptions{
		Name:                name,
		StructuredGenerator: generator,
	})
}

var (
	serviceExternalNameLong = templates.LongDesc(i18n.T(`
	Create an ExternalName service with the specified name.

	ExternalName service references to an external DNS address instead of
	only pods, which will allow application authors to reference services
	that exist off platform, on other clusters, or locally.`))

	serviceExternalNameExample = templates.Examples(i18n.T(`
	# Create a new ExternalName service named my-ns 
	pi create service externalname my-ns --external-name bar.com`))
)

// NewCmdCreateServiceExternalName is a macro command for creating an ExternalName service
func NewCmdCreateServiceExternalName(f cmdutil.Factory, cmdOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "externalname NAME --external-name external.name [--dry-run]",
		Short:   i18n.T("Create an ExternalName service."),
		Long:    serviceExternalNameLong,
		Example: serviceExternalNameExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := CreateExternalNameService(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddPrinterFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, cmdutil.ServiceExternalNameGeneratorV1Name)
	addPortFlags(cmd)
	cmd.Flags().String("external-name", "", i18n.T("External name of service"))
	cmd.MarkFlagRequired("external-name")
	return cmd
}

// CreateExternalNameService is the implementation of the create service externalname command
func CreateExternalNameService(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}
	var generator pi.StructuredGenerator
	switch generatorName := cmdutil.GetFlagString(cmd, "generator"); generatorName {
	case cmdutil.ServiceExternalNameGeneratorV1Name:
		generator = &pi.ServiceCommonGeneratorV1{
			Name:         name,
			Type:         v1.ServiceTypeExternalName,
			ExternalName: cmdutil.GetFlagString(cmd, "external-name"),
			ClusterIP:    "",
		}
	default:
		return errUnsupportedGenerator(cmd, generatorName)
	}
	return RunCreateSubcommand(f, cmd, cmdOut, &CreateSubcommandOptions{
		Name:                name,
		StructuredGenerator: generator,
	})
}
