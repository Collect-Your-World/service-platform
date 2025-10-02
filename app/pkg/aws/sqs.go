package aws

import (
	appConfig "backend/service-platform/app/internal/config"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func NewSQSClient(ctx context.Context, ac appConfig.ApplicationConfig) (*sqs.Client, error) {
	var err error
	var cfg aws.Config

	if ac.AwsConfig.Endpoint != "" {
		cfg, err = config.LoadDefaultConfig(
			ctx,
			config.WithRegion(ac.AwsConfig.Region),
		)
		if err != nil {
			return nil, err
		}

		return sqs.NewFromConfig(cfg, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(ac.AwsConfig.Endpoint)
		}), nil
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(ac.AwsConfig.Region))
		if err != nil {
			return nil, err
		}

		return sqs.NewFromConfig(cfg), nil
	}
}
