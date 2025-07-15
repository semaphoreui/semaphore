package util

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	log "github.com/sirupsen/logrus"
)

func GetRDSAuthToken(hostname string, username string, region string) (string, error) {
	hostname, port := ExtractHostPort(hostname, "5432")
	log.Debugf("Generating RDS auth token for %s:%s, user %s in region %s", hostname, port, username, region)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Errorf("Failed to load AWS config: %v", err)
		return "", err
	}

	credProvider := cfg.Credentials

	creds, err := credProvider.Retrieve(context.TODO())
	if err != nil {
		log.Errorf("Failed to retrieve AWS credentials: %v", err)
		return "", err
	}

	log.Debugf("AWS credentials retrieved successfully. Access key ID: %s, Provider: %s",
		creds.AccessKeyID[:4]+"...", creds.Source)

	log.Debugf("Building auth token for endpoint %s:%s in region %s", hostname, port, region)
	authToken, err := auth.BuildAuthToken(
		context.TODO(),
		hostname+":"+port,
		region,
		username,
		cfg.Credentials,
	)
	if err != nil {
		log.Errorf("Failed to generate RDS auth token: %v", err)
		return "", err
	}

	log.Debugf("Successfully generated RDS auth token (length: %d)", len(authToken))
	return authToken, nil
}

func ExtractHostPort(hostPort string, defaultPort string) (host string, port string) {
	parts := strings.Split(hostPort, ":")
	if len(parts) == 1 {
		return parts[0], defaultPort
	}
	return parts[0], parts[1]
}
