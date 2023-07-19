package main

import (
	"bytes"
	"fmt"
	"log"
	"context"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	orasRegistry "oras.land/oras-go/pkg/registry"
	"oras.land/oras-go/pkg/registry/remote"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var settings = cli.New()

func pushChart(chartPath, href, host, username, password, chartname string) error {

	// helm registry login --username "" --password ""
	client, err := registry.NewClient()
	if err != nil {
		return err
	}
	err = client.Login(host,
		registry.LoginOptBasicAuth(username, password))
	if err != nil {
		return err
	}

	// load chart in bytes
	stat, err := os.Stat(chartPath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return err
	}
	if stat.IsDir() {
		return err
	}
	// chart metadata
	meta, err := loader.Load(chartPath)
	if err != nil {
		return err
	}
	chartBytes, err := ioutil.ReadFile(chartPath)
	if err != nil {
		return err
	}

	var pushOpts []registry.PushOption
	provRef := fmt.Sprintf("%s.prov", chartBytes)
	if _, err := os.Stat(provRef); err == nil {
		provBytes, err := ioutil.ReadFile(provRef)
		if err != nil {
			return err
		}
		pushOpts = append(pushOpts, registry.PushOptProvData(provBytes))
	}
	// disable strict mode for configuring chartName in repo
	withStrictMode := registry.PushOptStrictMode(false)
	// add chartname and version to url
	ref := fmt.Sprintf("%s:%s",
		path.Join(strings.TrimPrefix(href, fmt.Sprintf("%s://", registry.OCIScheme)), chartname),
		meta.Metadata.Version)
	_, err = client.Push(chartBytes, ref, withStrictMode)
	if err != nil {
		println(err)
		return err
	}
	return nil
}

func pullChart(href, host, username, password, chartname, version string) error {

	// helm registry login --username "" --password ""
	client, err := registry.NewClient()
	if err != nil {
		return err
	}
	err = client.Login(host,
		registry.LoginOptBasicAuth(username, password))
	if err != nil {
		return err
	}

	// add chartname and version to url
	ref := fmt.Sprintf("%s:%s",
		path.Join(strings.TrimPrefix(href, fmt.Sprintf("%s://", registry.OCIScheme)), chartname),
		version)
	chartDetails, err := client.Pull(
		ref,
		registry.PullOptWithChart(true),
		registry.PullOptWithProv(true),
		registry.PullOptIgnoreMissingProv(true),
	)
	if err != nil {
		println(err)
		return err
	}
	log.Print(bytes.NewBuffer(chartDetails.Prov.Data))
	return nil
}

func helmShowReadme(chartURL, version string) (string, error) {
	actionConfig := new(action.Configuration)
	err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf)
	if err != nil {
		return "", err
	}

	client := action.NewShowWithConfig(action.ShowReadme, actionConfig)
	client.OutputFormat = action.ShowReadme
	client.Version = version
	err = addRegistryClient(client)
	if err != nil {
		return "", err
	}
	output, err := runShow(chartURL, client)
	if err != nil {
		return "", err
	}
	return output, nil
}

func helmShowValues(chartURL, version string) (string, error) {
	actionConfig := new(action.Configuration)
	err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf)
	if err != nil {
		return "", err
	}

	client := action.NewShowWithConfig(action.ShowValues, actionConfig)
	client.OutputFormat = action.ShowValues
	client.Version = version
	err = addRegistryClient(client)
	if err != nil {
		return "", err
	}
	output, err := runShow(chartURL, client)
	if err != nil {
		return "", err
	}
	return output, nil
}

// fetchOCIChartTagsList gives example snippets for listing tags in the repository without pagination.
func fetchOCIChartTagsList (repoURL string) ([]string, error) {
	repo, err := remote.NewRepository(repoURL)
	if err != nil {
		return nil, err // Handle error
	}

	ctx := context.Background()
	tags, err := orasRegistry.Tags(ctx, repo)
	if err != nil {
		return nil, err // Handle error
	}
	return tags, nil
}

func addRegistryClient(client *action.Show) error {
	registryClient, err := newRegistryClient(client.CertFile, client.KeyFile, client.CaFile, client.InsecureSkipTLSverify)
	if err != nil {
		return fmt.Errorf("missing registry client: %w", err)
	}
	client.SetRegistryClient(registryClient)
	return nil
}

func newRegistryClient(certFile, keyFile, caFile string, insecureSkipTLSverify bool) (*registry.Client, error) {
	if certFile != "" && keyFile != "" || caFile != "" || insecureSkipTLSverify {
		registryClient, err := newRegistryClientWithTLS(certFile, keyFile, caFile, insecureSkipTLSverify)
		if err != nil {
			return nil, err
		}
		return registryClient, nil
	}
	registryClient, err := newDefaultRegistryClient()
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

func newRegistryClientWithTLS(certFile, keyFile, caFile string, insecureSkipTLSverify bool) (*registry.Client, error) {
	// Create a new registry client
	registryClient, err := registry.NewRegistryClientWithTLS(os.Stderr, certFile, keyFile, caFile, insecureSkipTLSverify,
		settings.RegistryConfig, settings.Debug,
	)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

func newDefaultRegistryClient() (*registry.Client, error) {
	// Create a new registry client
	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stderr),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

func runShow(chartURL string, client *action.Show) (string, error) {
	log.Print("Original chart version: %q", client.Version)
	if client.Version == "" && client.Devel {
		log.Print("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	cp, err := client.ChartPathOptions.LocateChart(chartURL, settings)
	if err != nil {
		return "", err
	}
	return client.Run(cp)
}

func main() {
	host := "https://registry-1.docker.io"
	username := "ashexp"
	password := "dckr_pat_0t9-gLzcn_8fPdf10QNzzaOYNog"
	chartPath := "chart-1.0.2410-DEPLOY.tgz"
	version:= "1.0.2410-DEPLOY"
	chartname:= "chart"
	href := "registry-1.docker.io/ashexp"

	err := pushChart(chartPath, href, host, username, password, chartname)
	if err != nil {
		return
	}
	println("Chart successfully pushed to helm registry")
	err = pullChart(href, host, username, password, chartname, version)
	if err != nil {
		return
	}
	println("Chart successfully pulled from helm registry")
	// add chartname and version to url
	ref := fmt.Sprintf("%s://%s/%s", registry.OCIScheme, strings.TrimSpace(href), strings.TrimSpace(chartname))
	readme, err := helmShowReadme(ref, version)
	if err != nil {
		fmt.Printf("Failed to retrieve the readme: %v\n", err)
	}
	fmt.Println("Readme:")
	fmt.Println(readme)

	values, err := helmShowValues(ref, version)
	if err != nil {
		fmt.Printf("Failed to retrieve the values: %v\n", err)
	}
	fmt.Println("Values:")
	fmt.Println(values)
	ref = fmt.Sprintf("%s/%s", strings.TrimSpace(href), strings.TrimSpace(chartname))
	tagsList, err := fetchOCIChartTagsList(ref)
	if err != nil {
		fmt.Printf("Failed to retrieve the tags: %v\n", err)
	}
	fmt.Println("Tags:")
	for _, tag := range tagsList {
		fmt.Println(tag)
	}
}