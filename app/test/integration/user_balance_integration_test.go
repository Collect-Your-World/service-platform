package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"backend/service-platform/app/api/client/request"
	"backend/service-platform/app/api/client/response"
	"backend/service-platform/app/database/constant/currency"
	txconst "backend/service-platform/app/database/constant/transaction"
	httputil "backend/service-platform/app/test/util"
)

type UserBalanceIntegrationSuite struct {
	RouterSuite
}

func TestUserBalanceIntegrationSuite(t *testing.T) {
	suite.Run(t, new(UserBalanceIntegrationSuite))
}

func (s *UserBalanceIntegrationSuite) Test_GetBalances_All_And_ByCurrency() {
	email := "ub@example.com"
	password := "password123"

	_, code, err := httputil.RequestHTTP[response.GeneralResponse[string]](s.e, http.MethodPost, "/api/v1/auth/register", nil, request.RegisterRequest{Email: email, Password: password})
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, code)

	loginResp, loginCode, err := httputil.RequestHTTP[response.GeneralResponse[response.AuthResponse]](s.e, http.MethodPost, "/api/v1/auth/login", nil, request.AuthUserRequest{Email: email, Password: password})
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, loginCode)
	token := loginResp.Data.AccessToken

	// Initially no balances -> expect empty map
	allResp, allCode, err := httputil.RequestHTTP[response.GeneralResponse[map[string]int64]](s.e, http.MethodGet, "/api/v1/users/balances", &token, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, allCode)
	s.a.Equal(0, len(allResp.Data))

	// Fetch user id via /me
	meResp, meCode, err := httputil.RequestHTTP[response.GeneralResponse[response.MeResponse]](s.e, http.MethodGet, "/api/v1/auth/me", &token, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, meCode)
	// Record a change via manager directly
	_, _, err = s.managers.UserBalanceManager.RecordChange(s.ctx, meResp.Data.ID, currency.COIN, 100, txconst.DAILY_REWARD, txconst.DAILY_LOGIN, txconst.COMPLETED)
	s.r.NoError(err)

	// Query all balances
	allResp2, allCode2, err := httputil.RequestHTTP[response.GeneralResponse[map[string]int64]](s.e, http.MethodGet, "/api/v1/users/balances", &token, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, allCode2)
	s.a.Equal(int64(100), allResp2.Data["COIN"])

	// Query by currency
	coinResp, coinCode, err := httputil.RequestHTTP[response.GeneralResponse[map[string]int64]](s.e, http.MethodGet, "/api/v1/users/balances?currency=COIN", &token, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, coinCode)
	s.a.Equal(int64(100), coinResp.Data["COIN"])

	spinResp, spinCode, err := httputil.RequestHTTP[response.GeneralResponse[map[string]int64]](s.e, http.MethodGet, "/api/v1/users/balances?currency=SPIN", &token, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, spinCode)
	s.a.Equal(int64(0), spinResp.Data["SPIN"])

	// Record a change via manager directly
	_, _, err = s.managers.UserBalanceManager.RecordChange(s.ctx, meResp.Data.ID, currency.SPIN, 200, txconst.AD_WATCH, txconst.VIDEO_AD, txconst.COMPLETED)
	s.r.NoError(err)

	allResp3, allCode3, err := httputil.RequestHTTP[response.GeneralResponse[map[string]int64]](s.e, http.MethodGet, "/api/v1/users/balances", &token, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, allCode3)
	s.a.Equal(int64(100), allResp3.Data["COIN"])
	s.a.Equal(int64(200), allResp3.Data["SPIN"])
}

func (s *UserBalanceIntegrationSuite) Test_GetBalanceHistory_Filtered() {
	email := "ub2@example.com"
	password := "password123"

	_, code, err := httputil.RequestHTTP[response.GeneralResponse[string]](s.e, http.MethodPost, "/api/v1/auth/register", nil, request.RegisterRequest{Email: email, Password: password})
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, code)

	loginResp, loginCode, err := httputil.RequestHTTP[response.GeneralResponse[response.AuthResponse]](s.e, http.MethodPost, "/api/v1/auth/login", nil, request.AuthUserRequest{Email: email, Password: password})
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, loginCode)
	token := loginResp.Data.AccessToken

	meResp, meCode, err := httputil.RequestHTTP[response.GeneralResponse[response.MeResponse]](s.e, http.MethodGet, "/api/v1/auth/me", &token, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, meCode)

	_, _, err = s.managers.UserBalanceManager.RecordChange(s.ctx, meResp.Data.ID, currency.COIN, 100, txconst.DAILY_REWARD, txconst.DAILY_LOGIN, txconst.COMPLETED)
	s.r.NoError(err)
	_, _, err = s.managers.UserBalanceManager.RecordChange(s.ctx, meResp.Data.ID, currency.SPIN, 50, txconst.AD_WATCH, txconst.VIDEO_AD, txconst.COMPLETED)
	s.r.NoError(err)

	// Missing currency should fail
	bad, badCode, err := httputil.RequestHTTP[response.GeneralResponse[any]](s.e, http.MethodGet, "/api/v1/users/balances/history", &token, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusBadRequest, badCode)
	s.r.Equal(http.StatusBadRequest, bad.Code)

	// Filter by currency=COIN
	coinHist, coinCode, err := httputil.RequestHTTP[response.GeneralResponse[[]response.BalanceTransactionResponse]](s.e, http.MethodGet, "/api/v1/users/balances/history?currency=COIN", &token, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, coinCode)
	s.a.GreaterOrEqual(len(coinHist.Data), 1)
	s.a.Equal(currency.COIN, coinHist.Data[0].Currency)
	s.a.Equal(txconst.DAILY_REWARD, coinHist.Data[0].Type)
	s.a.Equal(txconst.DAILY_LOGIN, coinHist.Data[0].Source)
	s.a.Equal(txconst.COMPLETED, coinHist.Data[0].Status)

	// Filter by type=AD_WATCH
	adWatchHist, adCode, err := httputil.RequestHTTP[response.GeneralResponse[[]response.BalanceTransactionResponse]](s.e, http.MethodGet, "/api/v1/users/balances/history?currency=SPIN&type=AD_WATCH", &token, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, adCode)
	s.a.GreaterOrEqual(len(adWatchHist.Data), 1)
	s.a.Equal(currency.SPIN, adWatchHist.Data[0].Currency)
	s.a.Equal(txconst.AD_WATCH, adWatchHist.Data[0].Type)
	s.a.Equal(txconst.VIDEO_AD, adWatchHist.Data[0].Source)
	s.a.Equal(txconst.COMPLETED, adWatchHist.Data[0].Status)

	// Filter by type=AD_WATCH and currency=COIN, should return empty
	completedHist, completedCode, err := httputil.RequestHTTP[response.GeneralResponse[[]response.BalanceTransactionResponse]](s.e, http.MethodGet, "/api/v1/users/balances/history?currency=COIN&type=AD_WATCH", &token, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, completedCode)
	s.a.Equal(0, len(completedHist.Data))
}
