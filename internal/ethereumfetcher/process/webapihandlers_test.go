package process_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ethfetcher/internal/ethereumfetcher/api"
	"ethfetcher/internal/ethereumfetcher/process"
	"ethfetcher/internal/ethereumfetcher/process/processfakes"

	"github.com/labstack/echo/v4"
)

var (
	ErrInternalServerError = errors.New("internal error")
	ErrRlpError            = errors.New("RLP error")
	ErrStoreError          = errors.New("store failure")
)

var _ = Describe("Web API", func() {
	var (
		e                  *echo.Echo
		svc                *processfakes.FakeService
		tokenValidationSvc *processfakes.FakeTokenValidationService
		ctx                context.Context
		txHash             string
		tx                 api.Transaction
		recorder           *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		e = echo.New()
		svc = &processfakes.FakeService{}
		tokenValidationSvc = &processfakes.FakeTokenValidationService{}
		ctx = context.Background()
		recorder = httptest.NewRecorder()

		txHash = "0xabc"
		tx = api.Transaction{
			TransactionHash: txHash,
			From:            "0xfrom",
			To:              "0xto",
			Value:           "100",
		}

		process.RegisterEthHandlers(ctx, e, svc, tokenValidationSvc)
	})

	ItReturnsTheTx := func(response map[string]interface{}) {
		txs := response["transactions"].([]interface{})
		txFromResponse := txs[0].(map[string]interface{})

		Expect(txFromResponse["transactionHash"]).To(Equal(tx.TransactionHash))
		Expect(txFromResponse["from"]).To(Equal(tx.From))
		Expect(txFromResponse["to"]).To(Equal(tx.To))
		Expect(txFromResponse["value"]).To(Equal(tx.Value))
	}

	Describe("GET /lime/eth", func() {
		When("transactionHashes are provided", func() {
			It("returns the transactions", func() {
				svc.GetEthTransactionsReturns([]api.Transaction{tx}, nil)

				req := httptest.NewRequest(http.MethodGet, "/lime/eth?transactionHashes="+txHash, nil)
				e.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["transactions"]).To(HaveLen(1))
				ItReturnsTheTx(response)
			})
		})

		When("transactionHashes are missing", func() {
			It("returns a bad request error", func() {
				req := httptest.NewRequest(http.MethodGet, "/lime/eth", nil)
				e.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal(process.ErrTxHashesParamIsRequired.Error()))
			})
		})

		When("an error occurs while fetching transactions", func() {
			It("returns an internal server error", func() {
				svc.GetEthTransactionsReturns(nil, ErrInternalServerError)

				req := httptest.NewRequest(http.MethodGet, "/lime/eth?transactionHashes="+txHash, nil)
				e.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal(ErrInternalServerError.Error()))
			})
		})
	})

	Describe("GET /lime/eth/:rlphex", func() {
		When("RLP hex is valid", func() {
			It("returns the transactions", func() {
				svc.GetEthTransactionsByRLPReturns([]api.Transaction{tx}, nil)

				req := httptest.NewRequest(http.MethodGet, "/lime/eth/"+txHash, nil)
				e.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["transactions"]).To(HaveLen(1))
				ItReturnsTheTx(response)
			})
		})

		When("an error occurs while fetching transactions by RLP", func() {
			It("returns an internal server error", func() {
				svc.GetEthTransactionsByRLPReturns(nil, ErrRlpError)

				req := httptest.NewRequest(http.MethodGet, "/lime/eth/invalidRLP", nil)
				e.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal(ErrRlpError.Error()))
			})
		})
	})

	Describe("GET /lime/all", func() {
		When("all transactions are fetched successfully", func() {
			It("returns all transactions", func() {
				svc.GetAllEthTransactionsReturns([]api.Transaction{tx}, nil)

				req := httptest.NewRequest(http.MethodGet, "/lime/all", nil)
				e.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["transactions"]).To(HaveLen(1))
				ItReturnsTheTx(response)
			})
		})

		When("an error occurs while fetching all transactions", func() {
			It("returns an internal server error", func() {
				svc.GetAllEthTransactionsReturns(nil, ErrStoreError)

				req := httptest.NewRequest(http.MethodGet, "/lime/all", nil)
				e.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal(fmt.Sprintf("failed to fetch all transactions: %s", ErrStoreError.Error())))
			})
		})
	})
})
