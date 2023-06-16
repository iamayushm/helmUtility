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

	host := "445808685819.dkr.ecr.us-east-2.amazonaws.com"
	username := "AWS"
	password := "eyJwYXlsb2FkIjoiNHZqZjFWT2dHSXZpTjU2OFlDcitKK3NHSmt3dWhjWEJnbHIzMXhyZ0pzeW5qeUNkemdKNUZMSHVUYlNVSDRnRFp2eWlvZG02TmhtRkpxbm5Oblg5cHBaWTNRSnBqOE1nS2NWM0ZDOHVUZzBwNFhaZ0xCdmwrYmtUSFY1OHJ6VnhUQTNFb1BFTnltckhETEVSa3ZNcFVKeEIzcTFjeEsvSFdOMkhVaXZGaEZvOEtVRjBjRFMrTU5ONGlqdXJRMng1Mm5HSFBMRHBwS2xGMkZZMjVBRFVhdExRL2k2SE5JOVlISXFtZk5la3ZMak5hTVB1TE9hQ0cyQVQrbkNOcW5jVnRxWXFONnZrU0hicVZXbzA5YmVWWUgzRTJnMzdhcjVvN2VNbTZPUTVKYkkvWlR0YitiZC82TXp3b1R6UWIrM2VhRmtxQTlMUGNBTFN0QUJRZTBvVmtmVUFrQUNPZjJjdkQvWTVRaTV6WTFyS2U3b0tRRkhETWUyTEtBQkswQkNLSkNTODU1QUJWSVFpL0xNcHRuNmFlRHlIQjZMdVA4VHduNHJDUmF0Vnc0ME16NHFFWnFUWTl1dmt2SXRyaUJsa0cvZEhhN281WWFNVW4vZm1TUHhPRktqajR2bUxtOVpQeklCN0dydktvYTBLc1FiUlA1WkxCU0FEdVVhdHl3dmhLWDlGOE55TkhvTC9aZXgrUzkwTStTbkhKRDJDT2VYRWNFMzRNdTdUUElmOU03ajN4VDlwaFZjcWVhQzZwRGxnWWlLTUdhS2FKbWErdFdjTGxaZ3B3eDRkU2xTQ3MyMjVudUZGeU1uTXZHTi9WR0MxRU5mUnc4K2p0YStxYWlsS3d0WG0zVXVDSVdhNmFaMm5kNUhhdGs5V2dFQW9NRFQvK09oamhJR0hjVzUwWWkzcC9EOEtaTkovS3dUV1NNa21zK01zdDFnWjlybjA2OXZmd0pXZ2tLNjhLdENPL2JJbktXVjhtS201VHl3S1c5R2E4SE5MdDB5NitEdW83azFhc0NWVEgyS0V4VE9ITzlGMVJ6dTVKRnhqMXpRQ1kwcmZYZTQyWXpjRVoxdjBCNkxkZGdPTkg1QkxFN21UdzdrNS9JdzZJT0RpWHJoRVRYY2dqeTJ2VGg0dzFRUjE4aUF0dENtR1RxVTE5MENZOUE0QXpab3kzRzRLQ1pDeWZaSHJuRFJNRGo2ZXpaN3N3OEJRTGxBcUpWYjBTZHkxSWR2d2Jhb2lQa1cwUXJEWXlLclZJdjE3QUVtZzJnOUk5SlZqRDRDUnRVcjFCMFE5RjB5R1FWa0pCUXd0bWxBVm9lbkJ0dWVQdUNDai9jRXV6emc4NWZFclRoeWNXWGhSZTVDUjZuclVQdGtnaE1aTG9UT1V4Y0RjL1hRNlNhRTRhQjJGRkE9PSIsImRhdGFrZXkiOiJBUUVCQUhqQjcvaWd3TWc0TlB3YXVyeFNJWXg0SGZueHVHYy80OGJEd3Z3RHBOWVdaZ0FBQUg0d2ZBWUpLb1pJaHZjTkFRY0dvRzh3YlFJQkFEQm9CZ2txaGtpRzl3MEJCd0V3SGdZSllJWklBV1VEQkFFdU1CRUVERUZuck1mYmQ0dXgvQ1lEeEFJQkVJQTc1MDFUT3lxRlBMb1hkTXJHU3ZybmpCcUNmaXllRHpCdEZ0QjMyTzAwa3lyUTFJVi9oNFBZQnhvekRWUXZadk80REVwR2w4UERZYnkzbkJFPSIsInZlcnNpb24iOiIyIiwidHlwZSI6IkRBVEFfS0VZIiwiZXhwaXJhdGlvbiI6MTY4Njk0MTYzOX0="
	chartPath := "chart-0.1.0.tgz"
	href := "oci://445808685819.dkr.ecr.us-east-2.amazonaws.com/"

	err := pushChart(chartPath, href, host, username, password)
	if err != nil {
		return
	}
	println("Chart successfully pushed to helm registry")

}
