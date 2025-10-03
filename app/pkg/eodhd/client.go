package eodhd

import (
	"backend/service-platform/app/internal/runtime"
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type DefaultEODHDClient struct {
	httpClient *resty.Client
	res        *runtime.Resource
	logger     *zap.Logger
}

func NewEODHDClient(httpClient *resty.Client, logger *zap.Logger, res *runtime.Resource) EODHDClient {
	return &DefaultEODHDClient{
		httpClient: httpClient,
		logger:     logger,
		res:        res,
	}
}

func (c *DefaultEODHDClient) GetFundamental(ctx context.Context, ticker string) (FundamentalData, error) {
	var result FundamentalData

	url := fmt.Sprintf("%s/fundamentals/%s", c.res.Config.EodhdConfig.BaseAPI, ticker)

	resp, err := c.res.HttpClient.R().
		SetContext(ctx).
		SetQueryParam("api_token", c.res.Config.EodhdConfig.Token).
		SetQueryParam("fmt", "json").
		Get(url)

	if err != nil {
		c.logger.Error("failed to fetch company data for ticker %s: %w", zap.String("ticker", ticker), zap.Error(err))
		return FundamentalData{}, err
	}

	if resp.StatusCode() != 200 {
		c.res.Logger.Warn("Non-200 response from API",
			zap.String("ticker", ticker),
			zap.Int("status_code", resp.StatusCode()),
			zap.String("response", string(resp.Body())),
		)
		return FundamentalData{}, fmt.Errorf("API returned status %d for ticker %s", resp.StatusCode(), ticker)
	}

	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		c.logger.Error("failed to unmarshal company data for ticker %s: %w", zap.String("ticker", ticker), zap.Error(err))
		return FundamentalData{}, err
	}

	return result, nil
}

func (c *DefaultEODHDClient) GetRealtime(ctx context.Context, ticker string) (RealtimeData, error) {
	url := fmt.Sprintf("%s/real-time/%s", c.res.Config.EodhdConfig.BaseAPI, ticker)

	resp, err := c.res.HttpClient.R().
		SetContext(ctx).
		SetQueryParam("api_token", c.res.Config.EodhdConfig.Token).
		SetQueryParam("fmt", "json").
		Get(url)

	if err != nil {
		c.res.Logger.Error("Failed to fetch real-time data in concurrent call", zap.Error(err))
		return RealtimeData{}, err
	}

	if resp.StatusCode() != 200 {
		err := fmt.Errorf("real-time API returned status %d", resp.StatusCode())
		c.res.Logger.Warn("Non-200 response from real-time API in concurrent call",
			zap.Int("status_code", resp.StatusCode()),
			zap.String("response", string(resp.Body())),
		)
		return RealtimeData{}, err
	}

	var result RealtimeData
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		c.res.Logger.Error("Failed to unmarshal real-time data in concurrent call", zap.Error(err))
		return RealtimeData{}, err
	}

	return result, nil
}

func (c *DefaultEODHDClient) GetNews(ctx context.Context, ticker string) (NewsData, error) {
	var result NewsData
	url := fmt.Sprintf("%s/news", c.res.Config.EodhdConfig.BaseAPI)

	resp, err := c.res.HttpClient.R().
		SetContext(ctx).
		SetQueryParam("s", ticker).
		SetQueryParam("offset", "0").
		SetQueryParam("limit", "10").
		SetQueryParam("api_token", c.res.Config.EodhdConfig.Token).
		SetQueryParam("fmt", "json").
		Get(url)

	if err != nil {
		c.res.Logger.Error("Failed to fetch news for ticker %s: %w", zap.String("ticker", ticker), zap.Error(err))
		return NewsData{}, fmt.Errorf("failed to fetch news for ticker %s: %w", ticker, err)
	}

	if resp.StatusCode() != 200 {
		c.res.Logger.Warn("Non-200 response from API",
			zap.String("ticker", ticker),
			zap.Int("status_code", resp.StatusCode()),
			zap.String("response", string(resp.Body())),
		)
		return NewsData{}, fmt.Errorf("API returned status %d for ticker %s", resp.StatusCode(), ticker)
	}

	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		c.res.Logger.Error("Failed to unmarshal news for ticker %s: %w", zap.String("ticker", ticker), zap.Error(err))
		return NewsData{}, err
	}

	return result, nil
}
