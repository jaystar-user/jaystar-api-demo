package kintoneAPI

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SeanZhenggg/go-utils/logger"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"jaystar/internal/config"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/httpUtil"
	"net/http"
	"time"
)

type KintoneClient struct {
	*httpUtil.HttpClient
	cfg    config.IConfigEnv
	logger logger.ILogger
}

func ProvideKintoneClient(cfg config.IConfigEnv, logger logger.ILogger) *KintoneClient {
	return &KintoneClient{
		HttpClient: httpUtil.ProvideHttpClient(60 * time.Second),
		cfg:        cfg,
		logger:     logger,
	}
}

func (kc *KintoneClient) Get(ctx context.Context, authKey string, path string, query map[string]string, respBody any) (err error) {
	var (
		code int
		resp []byte
	)

	defer func() {
		r := recover()
		if r != nil || err != nil {
			kc.logger.Info(ctx, "KintoneClient Get api response",
				zap.Int("http_status_code", code),
				zap.String("http_response", string(resp)),
				zap.Error(err),
			)
		}
	}()

	_url := kc.cfg.GetKintoneConfig().Url + path
	headers := map[string]string{
		kintone.HeaderUserAuthorization: authKey,
	}
	code, resp, err = kc.HttpClient.Get(_url, query, headers)
	if err != nil {
		return xerrors.Errorf("HttpClient.Get error: %w", err)
	}

	if code >= http.StatusBadRequest {
		baseResp := &dto.KintoneApiBaseResponse{}
		if err = json.Unmarshal(resp, baseResp); err != nil {
			return xerrors.Errorf("json.Unmarshal error: %w", err)
		}

		return errors.New(fmt.Sprintf(
			"status: %d, id: %s, code: %s, message: %s",
			code,
			baseResp.Id,
			baseResp.Code,
			baseResp.Message,
		))
	}

	err = json.Unmarshal(resp, respBody)
	if err != nil {
		return xerrors.Errorf("json.Unmarshal error: %w", err)
	}

	return nil
}

func (kc *KintoneClient) Post(ctx context.Context, authKey string, path string, body any, respBody any) (err error) {
	var (
		code      int
		resp      []byte
		bodyBytes []byte
		headers   map[string]string
	)

	defer func() {
		r := recover()
		if r != nil || err != nil {
			kc.logger.Info(ctx, "KintoneClient Post api response",
				zap.String("http_request_body", string(bodyBytes)),
				zap.Any("http_request_header", headers),
				zap.Int("http_status_code", code),
				zap.String("http_response_body", string(resp)),
				zap.Error(err),
			)
		}
	}()

	_url := kc.cfg.GetKintoneConfig().Url + path

	headers = map[string]string{
		kintone.HeaderUserAuthorization: authKey,
		"Content-Type":                  "application/json",
	}

	bodyBytes, err = json.Marshal(body)
	if err != nil {
		err = xerrors.Errorf("KintoneAPI Post json.Marshal error: %w", err)
		return
	}

	code, resp, err = kc.HttpClient.Post(_url, bodyBytes, headers)
	if err != nil {
		return xerrors.Errorf("KintoneAPI kc.HttpClient.Post error: %w", err)
	}

	if code >= http.StatusBadRequest {
		errResp := &dto.KintoneApiBaseResponse{}
		if err := json.Unmarshal(resp, errResp); err != nil {
			return xerrors.Errorf("KintoneAPI response json.Unmarshal error: %w", err)
		}

		return errors.New(fmt.Sprintf(
			"KintoneAPI response send error : status: %d, id: %s, code: %s, message: %s",
			code,
			errResp.Id,
			errResp.Code,
			errResp.Message,
		))
	}

	err = json.Unmarshal(resp, respBody)
	if err != nil {
		return xerrors.Errorf("KintoneAPI kc.HttpClient.Post json.Unmarshal error: %w", err)
	}

	return nil
}

func (kc *KintoneClient) Put(ctx context.Context, authKey string, path string, body any, respBody any) (err error) {
	var (
		code      int
		resp      []byte
		bodyBytes []byte
		headers   map[string]string
	)

	defer func() {
		r := recover()
		if r != nil || err != nil {
			kc.logger.Info(ctx, "KintoneClient Put api response",
				zap.String("http_request_body", string(bodyBytes)),
				zap.Any("http_request_header", headers),
				zap.Int("http_status_code", code),
				zap.String("http_response_body", string(resp)),
				zap.Error(err),
			)
		}
	}()

	_url := kc.cfg.GetKintoneConfig().Url + path

	headers = map[string]string{
		kintone.HeaderUserAuthorization: authKey,
		"Content-Type":                  "application/json",
	}

	bodyBytes, err = json.Marshal(body)
	if err != nil {
		err = xerrors.Errorf("KintoneAPI Put json.Marshal error: %w", err)
		return
	}

	code, resp, err = kc.HttpClient.Put(_url, bodyBytes, headers)
	if err != nil {
		return xerrors.Errorf("KintoneAPI kc.HttpClient.Put error: %w", err)
	}

	if code >= http.StatusBadRequest {
		errResp := &dto.KintoneApiBaseResponse{}
		if err := json.Unmarshal(resp, errResp); err != nil {
			return xerrors.Errorf("KintoneAPI response json.Unmarshal error: %w", err)
		}

		return errors.New(fmt.Sprintf(
			"KintoneAPI response send error : status: %d, id: %s, code: %s, message: %s",
			code,
			errResp.Id,
			errResp.Code,
			errResp.Message,
		))
	}

	err = json.Unmarshal(resp, respBody)
	if err != nil {
		return xerrors.Errorf("KintoneAPI kc.HttpClient.Put json.Unmarshal error: %w", err)
	}

	return nil
}
