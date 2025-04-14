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
		txHash, rlpHex     string
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
		rlpHex = "f123123dsffasdsaqweqwesasdasda"
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
			var req *http.Request

			JustBeforeEach(func() {
				e.ServeHTTP(recorder, req)
			})

			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/lime/eth?transactionHashes="+txHash, nil)
				svc.GetEthTransactionsReturns([]api.Transaction{tx}, nil)
			})

			It("returns the transactions", func() {
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["transactions"]).To(HaveLen(1))
				ItReturnsTheTx(response)
			})

			Context("and an authentication token is provided", func() {
				BeforeEach(func() {
					req.Header.Set("Authorization", "Bearer validtoken")

					svc.SetUserTransactionsReturns(nil)
					tokenValidationSvc.ValidateTokenReturns(true, "user123", nil)
				})

				It("succeeds", func() {
					Expect(recorder.Code).To(Equal(http.StatusOK))
					req.Header.Set("Authorization", "Bearer validtoken")

					var response map[string]interface{}
					err := json.Unmarshal(recorder.Body.Bytes(), &response)
					Expect(err).NotTo(HaveOccurred())

					Expect(response["transactions"]).To(HaveLen(1))
					ItReturnsTheTx(response)

					Expect(svc.SetUserTransactionsCallCount()).To(Equal(1))
					_, userID, txs := svc.SetUserTransactionsArgsForCall(0)
					Expect(userID).To(Equal("user123"))
					Expect(txs).To(HaveLen(1))
				})

				Context("and an error occurs while storing user transactions", func() {
					BeforeEach(func() {
						svc.SetUserTransactionsReturns(ErrStoreError)
					})

					It("returns an internal server error", func() {
						Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

						var response map[string]string
						err := json.Unmarshal(recorder.Body.Bytes(), &response)
						Expect(err).NotTo(HaveOccurred())
						Expect(response["error"]).To(Equal(
							fmt.Sprintf("%s: %v", process.ErrStoreUserTransactions.Error(), ErrStoreError),
						))
					})
				})
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
			var req *http.Request

			JustBeforeEach(func() {
				e.ServeHTTP(recorder, req)
			})

			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/lime/eth/"+rlpHex, nil)
				svc.GetEthTransactionsByRLPReturns([]api.Transaction{tx}, nil)
			})

			It("returns the transactions", func() {
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["transactions"]).To(HaveLen(1))
				ItReturnsTheTx(response)
			})

			Context("and an authentication token is provided", func() {
				BeforeEach(func() {
					req.Header.Set("Authorization", "Bearer validtoken")

					svc.SetUserTransactionsReturns(nil)
					tokenValidationSvc.ValidateTokenReturns(true, "user123", nil)
				})

				It("succeeds", func() {
					Expect(recorder.Code).To(Equal(http.StatusOK))
					req.Header.Set("Authorization", "Bearer validtoken")

					var response map[string]interface{}
					err := json.Unmarshal(recorder.Body.Bytes(), &response)
					Expect(err).NotTo(HaveOccurred())

					Expect(response["transactions"]).To(HaveLen(1))
					ItReturnsTheTx(response)

					Expect(svc.SetUserTransactionsCallCount()).To(Equal(1))
					_, userID, txs := svc.SetUserTransactionsArgsForCall(0)
					Expect(userID).To(Equal("user123"))
					Expect(txs).To(HaveLen(1))
				})

				Context("and an error occurs while storing user transactions", func() {
					BeforeEach(func() {
						svc.SetUserTransactionsReturns(ErrStoreError)
					})

					It("returns an internal server error", func() {
						Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

						var response map[string]string
						err := json.Unmarshal(recorder.Body.Bytes(), &response)
						Expect(err).NotTo(HaveOccurred())
						Expect(response["error"]).To(Equal(
							fmt.Sprintf("%s: %v", process.ErrStoreUserTransactions.Error(), ErrStoreError),
						))
					})
				})
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
		var req *http.Request

		JustBeforeEach(func() {
			e.ServeHTTP(recorder, req)
		})

		BeforeEach(func() {
			req = httptest.NewRequest(http.MethodGet, "/lime/all", nil)
			svc.GetAllEthTransactionsReturns([]api.Transaction{tx}, nil)
		})

		When("all transactions are fetched successfully", func() {
			It("returns all transactions", func() {
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["transactions"]).To(HaveLen(1))
				ItReturnsTheTx(response)
			})

			Context("and an authentication token is provided", func() {
				BeforeEach(func() {
					req.Header.Set("Authorization", "Bearer validtoken")

					svc.SetUserTransactionsReturns(nil)
					tokenValidationSvc.ValidateTokenReturns(true, "user123", nil)
				})

				It("succeeds", func() {
					Expect(recorder.Code).To(Equal(http.StatusOK))
					req.Header.Set("Authorization", "Bearer validtoken")

					var response map[string]interface{}
					err := json.Unmarshal(recorder.Body.Bytes(), &response)
					Expect(err).NotTo(HaveOccurred())

					Expect(response["transactions"]).To(HaveLen(1))
					ItReturnsTheTx(response)

					Expect(svc.SetUserTransactionsCallCount()).To(Equal(1))
					_, userID, txs := svc.SetUserTransactionsArgsForCall(0)
					Expect(userID).To(Equal("user123"))
					Expect(txs).To(HaveLen(1))
				})

				Context("and an error occurs while storing user transactions", func() {
					BeforeEach(func() {
						svc.SetUserTransactionsReturns(ErrStoreError)
					})

					It("returns an internal server error", func() {
						Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

						var response map[string]string
						err := json.Unmarshal(recorder.Body.Bytes(), &response)
						Expect(err).NotTo(HaveOccurred())
						Expect(response["error"]).To(Equal(
							fmt.Sprintf("%s: %v", process.ErrStoreUserTransactions.Error(), ErrStoreError),
						))
					})
				})
			})
		})

		When("an error occurs while fetching all transactions", func() {
			BeforeEach(func() {
				svc.GetAllEthTransactionsReturns(nil, ErrStoreError)
			})

			It("returns an internal server error", func() {
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal(fmt.Sprintf("failed to fetch all transactions: %s", ErrStoreError.Error())))
			})
		})
	})

	Describe("GET /lime/my", func() {
		var req *http.Request

		JustBeforeEach(func() {
			e.ServeHTTP(recorder, req)
		})

		BeforeEach(func() {
			req = httptest.NewRequest(http.MethodGet, "/lime/my", nil)
			svc.GetAllUserTransactionsReturns([]api.Transaction{tx}, nil)
		})

		When("token is valid and user has transactions", func() {
			BeforeEach(func() {
				req.Header.Set("Authorization", "Bearer validtoken")
				tokenValidationSvc.ValidateTokenReturns(true, "user123", nil)
			})
			It("succeeds", func() {
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["transactions"]).To(HaveLen(1))
				ItReturnsTheTx(response)
			})
		})

		When("token is invalid", func() {
			BeforeEach(func() {
				req.Header.Set("Authorization", "Bearer invalidtoken")
				tokenValidationSvc.ValidateTokenReturns(false, "", nil)
			})
			It("returns unauthorized", func() {
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal("invalid or expired token"))
			})
		})

		When("no token is provided", func() {
			It("returns unauthorized", func() {
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal("authorization token is required"))
			})
		})

		When("an error occurs while getting user transactions", func() {
			BeforeEach(func() {
				req.Header.Set("Authorization", "Bearer validtoken")
				tokenValidationSvc.ValidateTokenReturns(true, "user123", nil)
				svc.GetAllUserTransactionsReturns(nil, ErrStoreError)
			})
			It("returns internal server error", func() {
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal(fmt.Sprintf("%s: %v", process.ErrGetUserTransactions.Error(), ErrStoreError)))
			})
		})
	})
})
