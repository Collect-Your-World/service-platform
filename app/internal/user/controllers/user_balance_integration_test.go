package controllers_test

import (
	integrationtest "backend/service-platform/app/internal/integrationtest"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	authmodels "backend/service-platform/app/internal/auth/models"
	commonmodels "backend/service-platform/app/internal/common/models"
	testutil "backend/service-platform/app/internal/testutil"
	"backend/service-platform/app/internal/user/constants/currency"
	txconst "backend/service-platform/app/internal/user/constants/transaction"
	usermodels "backend/service-platform/app/internal/user/models"
)

type UserBalanceIntegrationSuite struct {
	integrationtest.RouterSuite
}

func TestUserBalanceIntegrationSuite(t *testing.T) {
	suite.Run(t, new(UserBalanceIntegrationSuite))
}

func (s *UserBalanceIntegrationSuite) Test_GetBalances_All_And_ByCurrency() {
	email := "ub@example.com"
	password := "password123"

	_, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[string]](s.Echo, http.MethodPost, "/api/v1/auth/register", nil, authmodels.RegisterRequest{Email: email, Password: password})
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, code)

	loginResp, loginCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](s.Echo, http.MethodPost, "/api/v1/auth/login", nil, authmodels.AuthUserRequest{Email: email, Password: password})
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, loginCode)
	token := loginResp.Data.AccessToken

	// Initially no balances -> expect empty map
	allResp, allCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[map[string]int64]](s.Echo, http.MethodGet, "/api/v1/users/balances", &token, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, allCode)
	s.A.Equal(0, len(allResp.Data))

	// Fetch user id via /me
	meResp, meCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.MeResponse]](s.Echo, http.MethodGet, "/api/v1/auth/me", &token, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, meCode)
	// Record a change via manager directly
	_, _, err = s.Managers.UserBalanceManager.RecordChange(s.Ctx, meResp.Data.ID, currency.COIN, 100, txconst.DAILY_REWARD, txconst.DAILY_LOGIN, txconst.COMPLETED)
	s.R.NoError(err)

	// Query all balances
	allResp2, allCode2, err := testutil.RequestHTTP[commonmodels.GeneralResponse[map[string]int64]](s.Echo, http.MethodGet, "/api/v1/users/balances", &token, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, allCode2)
	s.A.Equal(int64(100), allResp2.Data["COIN"])

	// Query by currency
	coinResp, coinCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[map[string]int64]](s.Echo, http.MethodGet, "/api/v1/users/balances?currency=COIN", &token, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, coinCode)
	s.A.Equal(int64(100), coinResp.Data["COIN"])

	spinResp, spinCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[map[string]int64]](s.Echo, http.MethodGet, "/api/v1/users/balances?currency=SPIN", &token, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, spinCode)
	s.A.Equal(int64(0), spinResp.Data["SPIN"])

	// Record a change via manager directly
	_, _, err = s.Managers.UserBalanceManager.RecordChange(s.Ctx, meResp.Data.ID, currency.SPIN, 200, txconst.AD_WATCH, txconst.VIDEO_AD, txconst.COMPLETED)
	s.R.NoError(err)

	allResp3, allCode3, err := testutil.RequestHTTP[commonmodels.GeneralResponse[map[string]int64]](s.Echo, http.MethodGet, "/api/v1/users/balances", &token, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, allCode3)
	s.A.Equal(int64(100), allResp3.Data["COIN"])
	s.A.Equal(int64(200), allResp3.Data["SPIN"])
}

func (s *UserBalanceIntegrationSuite) Test_GetBalanceHistory_Filtered() {
	email := "ub2@example.com"
	password := "password123"

	_, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[string]](s.Echo, http.MethodPost, "/api/v1/auth/register", nil, authmodels.RegisterRequest{Email: email, Password: password})
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, code)

	loginResp, loginCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](s.Echo, http.MethodPost, "/api/v1/auth/login", nil, authmodels.AuthUserRequest{Email: email, Password: password})
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, loginCode)
	token := loginResp.Data.AccessToken

	meResp, meCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.MeResponse]](s.Echo, http.MethodGet, "/api/v1/auth/me", &token, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, meCode)

	_, _, err = s.Managers.UserBalanceManager.RecordChange(s.Ctx, meResp.Data.ID, currency.COIN, 100, txconst.DAILY_REWARD, txconst.DAILY_LOGIN, txconst.COMPLETED)
	s.R.NoError(err)
	_, _, err = s.Managers.UserBalanceManager.RecordChange(s.Ctx, meResp.Data.ID, currency.SPIN, 50, txconst.AD_WATCH, txconst.VIDEO_AD, txconst.COMPLETED)
	s.R.NoError(err)

	// Missing currency should fail
	bad, badCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](s.Echo, http.MethodGet, "/api/v1/users/balances/history", &token, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusBadRequest, badCode)
	s.R.Equal(http.StatusBadRequest, bad.Code)

	// Filter by currency=COIN
	coinHist, coinCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[[]usermodels.BalanceTransactionResponse]](s.Echo, http.MethodGet, "/api/v1/users/balances/history?currency=COIN", &token, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, coinCode)
	s.A.GreaterOrEqual(len(coinHist.Data), 1)
	s.A.Equal(string(currency.COIN), coinHist.Data[0].Currency)
	s.A.Equal(string(txconst.DAILY_REWARD), coinHist.Data[0].Type)
	s.A.Equal(string(txconst.DAILY_LOGIN), coinHist.Data[0].Source)
	s.A.Equal(string(txconst.COMPLETED), coinHist.Data[0].Status)

	// Filter by type=AD_WATCH
	adWatchHist, adCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[[]usermodels.BalanceTransactionResponse]](s.Echo, http.MethodGet, "/api/v1/users/balances/history?currency=SPIN&type=AD_WATCH", &token, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, adCode)
	s.A.GreaterOrEqual(len(adWatchHist.Data), 1)
	s.A.Equal(string(currency.SPIN), adWatchHist.Data[0].Currency)
	s.A.Equal(string(txconst.AD_WATCH), adWatchHist.Data[0].Type)
	s.A.Equal(string(txconst.VIDEO_AD), adWatchHist.Data[0].Source)
	s.A.Equal(string(txconst.COMPLETED), adWatchHist.Data[0].Status)

	// Filter by type=AD_WATCH and currency=COIN, should return empty
	completedHist, completedCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[[]usermodels.BalanceTransactionResponse]](s.Echo, http.MethodGet, "/api/v1/users/balances/history?currency=COIN&type=AD_WATCH", &token, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, completedCode)
	s.A.Equal(0, len(completedHist.Data))
}
