package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/armosec/kubescape/cautils"
	"github.com/armosec/kubescape/cautils/getter"
	"github.com/armosec/kubescape/clihandler"
	"github.com/armosec/opa-utils/reporthandling"
	"github.com/spf13/cobra"
)

var frameworkCmd = &cobra.Command{
	Use:       fmt.Sprintf("framework <framework names list> [`<glob pattern>`/`-`] [flags]\nExamples:\n$ kubescape scan framework nsa [flags]\n$ kubescape scan framework mitre,nsa [flags]\n$ kubescape scan framework 'nsa, mitre' [flags]\nSupported frameworks: %s", getter.NativeFrameworks),
	Short:     fmt.Sprintf("The framework you wish to use. Supported frameworks: %s", strings.Join(getter.NativeFrameworks, ", ")),
	Long:      "Execute a scan on a running Kubernetes cluster or `yaml`/`json` files (use glob) or `-` for stdin",
	ValidArgs: getter.NativeFrameworks,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("requires at least one framework name")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		flagValidationFramework()
		scanInfo.PolicyIdentifier = []reporthandling.PolicyIdentifier{}
		// If no framework provided, use all
		if len(args) == 0 {
			scanInfo.SetPolicyIdentifierForGivenFrameworks(getter.NativeFrameworks)
			scanInfo.ScanAll = true
		} else {
			// Read frameworks from input args
			scanInfo.PolicyIdentifier = []reporthandling.PolicyIdentifier{}
			frameworks := strings.Split(strings.Join(strings.Fields(args[0]), ""), ",")
			scanInfo.PolicyIdentifier = SetScanForFirstFramework(frameworks)
			if len(frameworks) > 1 {
				scanInfo.SetPolicyIdentifierForGivenFrameworks(frameworks[1:])
			}

			if len(args) > 1 {
				// expected yaml/url input
				if err := scanInfo.SetInputPatterns(args); err != nil {
					return err
				}
			}
		}
		scanInfo.Init()
		cautils.SetSilentMode(scanInfo.Silent)
		err := clihandler.ScanCliSetup(&scanInfo)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	scanCmd.AddCommand(frameworkCmd)
	scanInfo = cautils.ScanInfo{}
	scanInfo.FrameworkScan = true
}

func SetScanForFirstFramework(frameworks []string) []reporthandling.PolicyIdentifier {
	newPolicy := reporthandling.PolicyIdentifier{}
	newPolicy.Kind = reporthandling.KindFramework
	newPolicy.Name = frameworks[0]
	scanInfo.PolicyIdentifier = append(scanInfo.PolicyIdentifier, newPolicy)
	return scanInfo.PolicyIdentifier
}

func flagValidationFramework() {
	if scanInfo.Submit && scanInfo.Local {
		fmt.Println("You can use `keep-local` or `submit`, but not both")
		os.Exit(1)
	}
	if 100 < scanInfo.FailThreshold {
		fmt.Println("bad argument: out of range threshold")
		os.Exit(1)
	}
}
