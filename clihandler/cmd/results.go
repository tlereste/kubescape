package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/armosec/k8s-interface/k8sinterface"
	"github.com/armosec/kubescape/clihandler"
	"github.com/armosec/kubescape/clihandler/cliinterfaces"
	"github.com/armosec/kubescape/resultshandling/reporter"
	"github.com/armosec/opa-utils/reporthandling"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

type ResultsObject struct {
	filePath     string
	customerGUID string
	clusterName  string
}

func NewResultsObject(customerGUID, clusterName, filePath string) *ResultsObject {
	return &ResultsObject{
		filePath:     filePath,
		customerGUID: customerGUID,
		clusterName:  clusterName,
	}
}

func (resultsObject *ResultsObject) SetResourcesReport() (*reporthandling.PostureReport, error) {
	// load framework results from json file
	frameworkReports, err := loadResultsFromFile(resultsObject.filePath)
	if err != nil {
		return nil, err
	}

	return &reporthandling.PostureReport{
		FrameworkReports:     frameworkReports,
		ReportID:             uuid.NewV4().String(),
		ReportGenerationTime: time.Now().UTC(),
		CustomerGUID:         resultsObject.customerGUID,
		ClusterName:          resultsObject.clusterName,
	}, nil
}

var resultsCmd = &cobra.Command{
	Use:   "results <json file>\nExample:\n$ kubescape submit results path/to/results.json",
	Short: "Submit a pre scanned results file. The file must be in json format",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing results file")
		}

		k8s := k8sinterface.NewKubernetesApi()

		// get config
		clusterConfig, err := getSubmittedClusterConfig(k8s)
		if err != nil {
			return err
		}

		clusterName := clusterConfig.GetClusterName()
		customerGUID := clusterConfig.GetCustomerGUID()

		resultsObjects := NewResultsObject(customerGUID, clusterName, args[0])

		// submit resources
		r := reporter.NewReportEventReceiver(customerGUID, clusterName)

		submitInterfaces := cliinterfaces.SubmitInterfaces{
			ClusterConfig: clusterConfig,
			SubmitObjects: resultsObjects,
			Reporter:      r,
		}

		if err := clihandler.Submit(submitInterfaces); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	submitCmd.AddCommand(resultsCmd)
}

func loadResultsFromFile(filePath string) ([]reporthandling.FrameworkReport, error) {
	frameworkReports := []reporthandling.FrameworkReport{}
	f, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(f, &frameworkReports); err != nil {
		frameworkReport := reporthandling.FrameworkReport{}
		if err = json.Unmarshal(f, &frameworkReport); err != nil {
			return frameworkReports, err
		}
		frameworkReports = append(frameworkReports, frameworkReport)
	}
	return frameworkReports, nil
}
