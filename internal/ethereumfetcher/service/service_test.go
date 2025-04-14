package service_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ethfetcher/internal/ethereumfetcher/api"
	"ethfetcher/internal/ethereumfetcher/service"
	"ethfetcher/internal/ethereumfetcher/service/servicefakes"
)

var ErrStoreFailure = errors.New("store failure")

var _ = Describe("Service", func() {
	When("created", func() {
		It("exists", func() {
			Expect(service.NewService(nil, nil)).NotTo(BeNil())
		})
	})

	Describe("instance", func() {
		var (
			ctx       context.Context
			svc       *service.Service
			store     *servicefakes.FakeStore
			ethClient *servicefakes.FakeEthereumClient

			txHash string
			tx     api.Transaction
		)

		BeforeEach(func() {
			ctx = context.Background()
			store = &servicefakes.FakeStore{}
			ethClient = &servicefakes.FakeEthereumClient{}

			svc = service.NewService(store, ethClient)

			txHash = "0xabc"
			tx = api.Transaction{
				TransactionHash: txHash,
				From:            "0xfrom",
				To:              "0xto",
				Value:           "100",
			}
		})

		Describe("GetEthTransactions", func() {
			When("transaction is found in the store", func() {
				BeforeEach(func() {
					store.GetByHashReturns(tx, nil)
				})

				It("returns it without calling the ethereum client", func() {
					txs, err := svc.GetEthTransactions(ctx, []string{txHash})
					Expect(err).NotTo(HaveOccurred())
					Expect(txs).To(HaveLen(1))
					Expect(txs[0]).To(Equal(tx))
					Expect(ethClient.FetchTransactionByHashCallCount()).To(Equal(0))
				})
			})

			When("transaction is not in store but fetched by the ethereum client", func() {
				BeforeEach(func() {
					store.GetByHashReturns(api.Transaction{}, nil)
					ethClient.FetchTransactionByHashReturns(tx, nil)
				})

				It("fetches from ethereum and stores it", func() {
					txs, err := svc.GetEthTransactions(ctx, []string{txHash})
					Expect(err).NotTo(HaveOccurred())
					Expect(txs).To(ContainElement(tx))
					Expect(ethClient.FetchTransactionByHashCallCount()).To(Equal(1))
					Expect(store.InsertCallCount()).To(Equal(1))
				})
			})

			When("ethereum client does not find the transaction", func() {
				BeforeEach(func() {
					store.GetByHashReturns(api.Transaction{}, nil)
					ethClient.FetchTransactionByHashReturns(api.Transaction{}, nil)
				})

				It("skips", func() {
					txs, err := svc.GetEthTransactions(ctx, []string{txHash})
					Expect(err).NotTo(HaveOccurred())
					Expect(txs).To(BeEmpty())
				})
			})

			When("store returns an error", func() {
				BeforeEach(func() {
					store.GetByHashReturns(api.Transaction{}, errors.New("db failure"))
				})

				It("returns the error", func() {
					txs, err := svc.GetEthTransactions(ctx, []string{txHash})
					Expect(err).To(MatchError(ContainSubstring("failed to get transaction")))
					Expect(txs).To(BeNil())
				})
			})

			When("ethereum client returns an error", func() {
				BeforeEach(func() {
					store.GetByHashReturns(api.Transaction{}, nil)
					ethClient.FetchTransactionByHashReturns(api.Transaction{}, errors.New("eth failure"))
				})

				It("returns the error", func() {
					txs, err := svc.GetEthTransactions(ctx, []string{txHash})
					Expect(err).To(MatchError(ContainSubstring("eth failure")))
					Expect(txs).To(BeNil())
				})
			})
		})

		Describe("GetEthTransactionsByRLP", func() {
			When("RLP decoding succeeds", func() {
				var rlp string
				var decodedHashes []string

				BeforeEach(func() {
					rlp = "0xabcdefasddsa"
					decodedHashes = []string{"0xabc", "0xdef"}

					ethClient.RlpHexToHashListReturns(decodedHashes, nil)
					store.GetByHashReturns(tx, nil)
				})

				It("returns the transactions", func() {
					txs, err := svc.GetEthTransactionsByRLP(ctx, rlp)
					Expect(err).NotTo(HaveOccurred())
					Expect(txs).To(HaveLen(2))
					Expect(ethClient.RlpHexToHashListCallCount()).To(Equal(1))
					Expect(store.GetByHashCallCount()).To(Equal(2))
				})
			})

			When("RLP decoding fails", func() {
				BeforeEach(func() {
					ethClient.RlpHexToHashListReturns(nil, errors.New("some decoding error"))
				})
				It("returns an error", func() {
					txs, err := svc.GetEthTransactionsByRLP(ctx, "invalid_rlp")
					Expect(err).To(HaveOccurred())
					Expect(txs).To(HaveLen(0))
					Expect(ethClient.RlpHexToHashListCallCount()).To(Equal(1))
					Expect(store.GetByHashCallCount()).To(Equal(0))
				})
			})
		})

		Describe("GetAllEthTransactions", func() {
			It("succeeds", func() {
				store.GetAllReturns([]api.Transaction{tx}, nil)

				txs, err := svc.GetAllEthTransactions(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(txs).To(HaveLen(1))
				Expect(txs[0]).To(Equal(tx))
			})

			It("returns an error", func() {
				store.GetAllReturns(nil, ErrStoreFailure)

				txs, err := svc.GetAllEthTransactions(ctx)
				Expect(err).To(MatchError(ErrStoreFailure))
				Expect(txs).To(HaveLen(0))
			})
		})

		Describe("GetAllUserTransactions", func() {
			It("succeeds", func() {
				store.GetAllByUserReturns([]api.Transaction{tx}, nil)

				txs, err := svc.GetAllUserTransactions(ctx, "user123")
				Expect(err).NotTo(HaveOccurred())
				Expect(txs).To(HaveLen(1))
				Expect(txs[0]).To(Equal(tx))

				Expect(store.GetAllByUserCallCount()).To(Equal(1))
				_, user := store.GetAllByUserArgsForCall(0)
				Expect(user).To(Equal("user123"))
			})

			It("returns an error", func() {
				store.GetAllByUserReturns(nil, ErrStoreFailure)

				txs, err := svc.GetAllUserTransactions(ctx, "user123")
				Expect(err).To(MatchError(ErrStoreFailure))
				Expect(txs).To(BeNil())
			})
		})

		Describe("SetUserTransactions", func() {
			It("succeeds", func() {
				err := svc.SetUserTransactions(ctx, "user123", []api.Transaction{tx})
				Expect(err).NotTo(HaveOccurred())
				Expect(store.InsertUserTransactionsCallCount()).To(Equal(1))

				_, user, txs := store.InsertUserTransactionsArgsForCall(0)
				Expect(user).To(Equal("user123"))
				Expect(txs).To(HaveLen(1))
				Expect(txs[0]).To(Equal(tx))
			})

			It("returns an error", func() {
				store.InsertUserTransactionsReturns(ErrStoreFailure)

				err := svc.SetUserTransactions(ctx, "user123", []api.Transaction{tx})
				Expect(err).To(MatchError(ErrStoreFailure))
			})
		})
	})
})
