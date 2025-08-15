package s3client

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	awsv4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	ionoscloud "github.com/ionos-cloud/sdk-go-object-storage"
)

func NewIonosSignerMw(region, service, accessKey, secretKey string) ionoscloud.MiddlewareFunctionWithError {
	signer := awsv4.NewSigner(credentials.NewStaticCredentials(accessKey, secretKey, ""))

	// Define default values for region and service to maintain backward compatibility
	if region == "" {
		region = "eu-central-3"
	}
	if service == "" {
		service = "s3"
	}
	return func(r *http.Request) error {
		var reader io.ReadSeeker
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				return err
			}
			reader = bytes.NewReader(bodyBytes)
		}
		_, err := signer.Sign(r, reader, service, region, time.Now())
		return err
	}
}

func GenerateIonosClient(endpoint, accessKey, secretKey string) (*ionoscloud.APIClient, error) {
	config := ionoscloud.NewConfiguration(endpoint)
	endpointType := strings.Split(endpoint, ".")[0]
	endpointName := strings.Split(endpoint, ".")[1]
	config.MiddlewareWithError = NewIonosSignerMw(endpointName, endpointType, accessKey, secretKey)
	client := ionoscloud.NewAPIClient(config)
	log.Println("Generated Ionos client with endpoint:", endpoint, "accessKey:", accessKey, "secretKey:", secretKey, "endpointName:", endpointName, "endpointType:", endpointType)

	return client, nil
}
