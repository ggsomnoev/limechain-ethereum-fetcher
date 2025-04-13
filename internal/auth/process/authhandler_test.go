package process_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ethfetcher/internal/auth/model"
	"ethfetcher/internal/auth/process"
	"ethfetcher/internal/auth/process/processfakes"
	"ethfetcher/internal/auth/service"
)

var _ = Describe("Auth Handler", func() {
	var (
		e        *echo.Echo
		svc      *processfakes.FakeService
		ctx      context.Context
		recorder *httptest.ResponseRecorder

		validToken = "valid.jwt.token"
	)

	BeforeEach(func() {
		e = echo.New()
		svc = &processfakes.FakeService{}
		ctx = context.Background()
		recorder = httptest.NewRecorder()

		process.RegisterAuthHandlers(ctx, e, svc)
	})

	Describe("POST /lime/authenticate", func() {
		When("valid credentials are provided", func() {
			It("returns a valid token", func() {
				svc.AuthenticateReturns(validToken, nil)

				reqBody := model.AuthRequest{
					Username: "alice",
					Password: "alice",
				}
				bodyBytes, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/lime/authenticate", bytes.NewReader(bodyBytes))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

				e.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var res model.AuthResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &res)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.Token).To(Equal(validToken))
			})
		})

		When("invalid JSON is sent", func() {
			It("returns an internal server error", func() {
				req := httptest.NewRequest(http.MethodPost, "/lime/authenticate", bytes.NewBufferString("{invalid json"))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

				e.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var res map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &res)
				Expect(err).NotTo(HaveOccurred())
				Expect(res["error"]).To(ContainSubstring("invalid character"))
			})
		})

		When("authentication fails", func() {
			It("returns an unauthorized server error", func() {
				svc.AuthenticateReturns("", service.ErrInvalidCredentials)

				reqBody := model.AuthRequest{
					Username: "alice",
					Password: "wrongpass",
				}
				bodyBytes, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/lime/authenticate", bytes.NewReader(bodyBytes))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

				e.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var res map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &res)
				Expect(err).NotTo(HaveOccurred())
				Expect(res["error"]).To(Equal(service.ErrInvalidCredentials.Error()))
			})
		})
	})
})
