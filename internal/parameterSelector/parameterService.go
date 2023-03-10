package parameterSelector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/travelaudience/go-promhttp"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

const (
	CACHE_TTL              = 10 * time.Second
	CCY_MULTIPLIER_MIN     = 0.001
	CCY_MULTIPLIER_MAX     = 10000
	CCY_MULTIPLIER_EPS     = 0.001
	LOCAL_DEFAULT_COMPANY  = "default"
	REMOTE_DEFAULT_COMPANY = "default"
)

type ParameterService interface {
	CurrencyMultiplier(ccy string, company string) (float32, bool)
}

type LocalParameterService struct {
	ccyMultipliers map[ccyKey]ccyValue
}
type RemoteParameterService struct {
	Url string

	cache  map[cachedKey]cachedValue
	client *promhttp.Client
}

type cachedKey interface {
	Key() string
}
type cachedValue interface {
	Expires() time.Time
}

type ccyKey struct {
	Ccy     string
	Company string
}
type ccyValue float32
type ccyCache struct {
	ccyValue
	expires time.Time
}

func (k ccyKey) Key() string {
	return k.Ccy + k.Company
}
func (v ccyCache) Expires() time.Time {
	return v.expires
}

var (
	paramService      ParameterService
	localParamService *LocalParameterService
)

func GetParameterService() ParameterService {
	if paramService == nil {
		localParamService = CreateLocalParameterService()
		if config.GlobalConfig.ExtParamService == "" {
			paramService = localParamService
		} else {
			paramService = CreateRemoteParameterService(config.GlobalConfig.ExtParamService)
		}
	}
	return paramService
}

func CreateLocalParameterService() *LocalParameterService {
	betConf, err := parseBetConfig()
	if err != nil {
		logger.Errorf("could not parse bet config")
		return nil
	}
	multipliers := make(map[ccyKey]ccyValue, len(betConf.CcyMultipliers))
	for c, m := range betConf.CcyMultipliers {
		for k, v := range m {
			multipliers[ccyKey{Ccy: k, Company: c}] = ccyValue(v)
		}
	}
	return &LocalParameterService{
		ccyMultipliers: multipliers,
	}
}

func CreateRemoteParameterService(url string) *RemoteParameterService {
	return &RemoteParameterService{
		Url:   url,
		cache: map[cachedKey]cachedValue{},
		client: &promhttp.Client{
			Client:     http.DefaultClient,
			Registerer: prometheus.DefaultRegisterer,
		},
	}
}

func (i *LocalParameterService) CurrencyMultiplier(ccy string, company string) (float32, bool) {
	if company == "" {
		company = LOCAL_DEFAULT_COMPANY
	}
	value, ok := i.ccyMultipliers[ccyKey{Ccy: ccy, Company: company}]
	if !ok && company != LOCAL_DEFAULT_COMPANY {
		value, ok = i.ccyMultipliers[ccyKey{Ccy: ccy, Company: LOCAL_DEFAULT_COMPANY}]
		logger.Infof("using local default ccy multiplier for currency [%s]", ccy)
	} else {
		logger.Debugf("using local override ccy multiplier for currency and company [%s]-[%s]", ccy, company)
	}
	if !ok {
		logger.Errorf("unknown currency and company pair [%s]-[%s]", ccy, company)
		return 0.0, false
	}
	return float32(value), true
}

func (i *RemoteParameterService) CurrencyMultiplier(ccy string, company string) (float32, bool) {
	multiplier, ok := localParamService.CurrencyMultiplier(ccy, company)
	if !ok {
		return 0.0, false
	} else {
		key := ccyKey{
			Ccy:     ccy,
			Company: company,
		}
		value, ok := i.cache[key]
		if !ok || time.Now().After(value.Expires()) {
			var remoteMultiplier ccyValue
			remoteMultiplier, ok = i.FetchCurrencyMultiplier(ccy, company)
			if ok {
				multiplier = float32(remoteMultiplier)
			}
			i.cache[key] = ccyCache{
				ccyValue: ccyValue(multiplier),
				expires:  time.Now().Add(CACHE_TTL),
			}
		} else {
			var cached ccyCache
			cached, ok = value.(ccyCache)
			if ok {
				multiplier = float32(cached.ccyValue)
			} else {
				logger.Errorf("Ignoring cached value for [%s]-[%s] not of type currency multiplier: %#v",
					ccy, company, value)
			}
		}
	}
	return multiplier, true
}

func (i *RemoteParameterService) FetchCurrencyMultiplier(ccy string, company string) (ccyValue, bool) {
	if company == "" {
		company = REMOTE_DEFAULT_COMPANY
	}
	logger.Debugf("requesting currency and company [%s][%s] from remote service", ccy, company)
	api := fmt.Sprintf("betlevels?filter[currency]=%s&filter[company]=%s", ccy, company)
	b := new(bytes.Buffer)
	resp, err := i.request(api, "GET", b)
	if err != nil {
		logger.Errorf("failed request to [%s]", api)
		return 0.0, false
	}
	multiplier, ok := i.validateCurrencyMultiplierResp(ccy, company, resp)
	if !ok {
		return 0.0, false
	}
	return multiplier, true
}

type restBetLevels struct {
	Status int64          `json:"status"`
	Data   []restBetLevel `json:"data"`
}

type restBetLevel struct {
	Uuid       string  `json:"status"`
	Company    string  `json:"company"`
	Currency   string  `json:"currency"`
	Multiplier float64 `json:"multiplier"`
}

func (i *RemoteParameterService) validateCurrencyMultiplierResp(ccy string, company string, resp *http.Response) (ccyValue, bool) {
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	logger.Debugf("resp: %s", string(body))
	var betLevelsResp restBetLevels
	err := json.Unmarshal(body, &betLevelsResp)
	if err != nil {
		logger.Errorf("could not decode json response: %s", string(body))
		return ccyValue(0.0), false
	}
	if len(betLevelsResp.Data) != 1 {
		logger.Debugf("remote service returned no match for currency and company [%s]-[%s]", ccy, company)
		return ccyValue(0.0), false
	}
	betLevel := betLevelsResp.Data[0]
	if betLevel.Currency != ccy {
		logger.Errorf("response currency [%s] is not equal to request [%s]", betLevel.Currency, ccy)
		return ccyValue(0.0), false
	}
	if betLevel.Company != company {
		logger.Errorf("response company [%s] is not equal to request [%s]", betLevel.Company, company)
		return ccyValue(0.0), false
	}
	if betLevel.Multiplier < CCY_MULTIPLIER_MIN || betLevel.Multiplier > CCY_MULTIPLIER_MAX {
		logger.Errorf("multiplier [%f] is out of range [%d-%d]", betLevel.Multiplier, CCY_MULTIPLIER_MIN, CCY_MULTIPLIER_MAX)
		return ccyValue(0.0), false
	}
	nearestValid := math.Trunc(betLevel.Multiplier/CCY_MULTIPLIER_EPS) * CCY_MULTIPLIER_EPS
	if math.Abs(betLevel.Multiplier-nearestValid) > CCY_MULTIPLIER_EPS {
		logger.Errorf("multiplier [%f] is not close enough to a nearest valid value [%f]", betLevel.Multiplier, nearestValid)
		return ccyValue(0.0), false
	}
	return ccyValue(betLevel.Multiplier), true
}

func (i *RemoteParameterService) request(api string, method string, body *bytes.Buffer) (resp *http.Response, err error) {
	url := i.Url + "/" + api
	logger.Debugf("request to [%s]", url)
	req, _ := http.NewRequest(method, url, body)
	req.Header.Add("Content-Type", "application/json")
	client, _ := i.client.ForRecipient("restapi")
	resp, err = client.Do(req)
	return
}
