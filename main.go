package main

import (
	"fmt"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/registry"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func pushChart(chartPath, href, host, username, password string) error {

	// helm registry login --username "" --password ""
	client, err := registry.NewClient()
	if err != nil {
		return err
	}
	err = client.Login(host,
		registry.LoginOptBasicAuth(username, password), registry.LoginOptInsecure(true))
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

	// add chartname and version to url
	ref := fmt.Sprintf("%s:%s",
		path.Join(strings.TrimPrefix(href, fmt.Sprintf("%s://", registry.OCIScheme)), meta.Name()),
		meta.Metadata.Version)
	_, err = client.Push(chartBytes, ref)
	if err != nil {
		println(err)
		return err
	}
	return nil
}

func main() {

	host := ""
	username := ""
	password := ""
	chartPath := ""
	href := ""

	err := pushChart(chartPath, href, host, username, password)
	if err != nil {
		return
	}
	println("Chart successfully pushed to helm registry")

}
