package store_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"ethfetcher/internal/ethereumfetcher/api"
	"ethfetcher/internal/ethereumfetcher/store"
)

var _ = Describe("Store", func() {
	When("created", func() {
		It("exists", func() {
			Expect(store.NewStore(nil)).NotTo(BeNil())
		})
	})

	Describe("instance", Serial, func() {
		var (
			st *store.Store

			tx1, tx2 api.Transaction
		)

		BeforeEach(func() {
			st = store.NewStore(pool)

			tx1 = api.Transaction{
				TransactionHash: "0xaaa",
				BlockNumber:     100,
				From:            "0xfrom1",
				To:              "0xto1",
				Value:           "1000",
			}

			tx2 = api.Transaction{
				TransactionHash: "0xbbb",
				BlockNumber:     200,
				From:            "0xfrom2",
				To:              "0xto2",
				Value:           "2000",
			}
		})

		When("tx are added", func() {
			JustBeforeEach(func() {
				Expect(st.Insert(ctx, tx1)).To(Succeed())
				Expect(st.Insert(ctx, tx2)).To(Succeed())
			})

			JustAfterEach(func() {
				Expect(st.DeleteTxsByHash(ctx, tx1.TransactionHash)).To(Succeed())
				Expect(st.DeleteTxsByHash(ctx, tx2.TransactionHash)).To(Succeed())
			})

			Context("and tx1 is found", func() {
				It("returns tx1", func() {
					foundTx, err := st.GetByHash(ctx, tx1.TransactionHash)
					Expect(err).NotTo(HaveOccurred())
					Expect(foundTx).To(MatchFields(IgnoreExtras, Fields{
						"TransactionHash": Equal(tx1.TransactionHash),
						"BlockNumber":     Equal(tx1.BlockNumber),
						"From":            Equal(tx1.From),
						"To":              Equal(tx1.To),
						"Value":           Equal(tx1.Value),
					}))
				})
			})

			Context("and all txs are found", func() {
				It("returns all txs", func() {
					foundTxs, err := st.GetAll(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(foundTxs).To(ContainElements(
						MatchFields(IgnoreExtras, Fields{
							"TransactionHash": Equal(tx1.TransactionHash),
							"BlockNumber":     Equal(tx1.BlockNumber),
							"From":            Equal(tx1.From),
							"To":              Equal(tx1.To),
							"Value":           Equal(tx1.Value),
						}),
						MatchFields(IgnoreExtras, Fields{
							"TransactionHash": Equal(tx2.TransactionHash),
							"BlockNumber":     Equal(tx2.BlockNumber),
							"From":            Equal(tx2.From),
							"To":              Equal(tx2.To),
							"Value":           Equal(tx2.Value),
						}),
					))
				})
			})

			Context("and a tx already exists", func() {
				It("fails with duplicate insert error", func() {
					Expect(st.Insert(ctx, tx1)).To(MatchError(store.ErrDuplicateTransaction))
				})
			})

			Context("and a tx does not exist", func() {
				It("returns empty tx without an error", func() {
					tx, err := st.GetByHash(ctx, "0xnonexistent")
					Expect(err).NotTo(HaveOccurred())
					Expect(tx.TransactionHash).To(BeEmpty())
				})
			})
		})

		When("user tx are added", func() {
			var username string

			BeforeEach(func() {
				// Username already seeded into the DB from the initial migration.
				username = "alice"

				Expect(st.Insert(ctx, tx1)).To(Succeed())
				Expect(st.Insert(ctx, tx2)).To(Succeed())
			})

			AfterEach(func() {
				Expect(st.DeleteTxsByHash(ctx, tx1.TransactionHash)).To(Succeed())
				Expect(st.DeleteTxsByHash(ctx, tx2.TransactionHash)).To(Succeed())
			})

			JustBeforeEach(func() {
				Expect(st.InsertUserTransactions(ctx, username, []api.Transaction{tx1, tx2})).To(Succeed())
			})

			JustAfterEach(func() {
				Expect(st.DeleteTxsByUsername(ctx, username)).To(Succeed())
			})

			It("returns all transactions for that user", func() {
				foundTxs, err := st.GetAllByUser(ctx, username)
				Expect(err).NotTo(HaveOccurred())
				Expect(foundTxs).To(HaveLen(2))

				Expect(foundTxs).To(ContainElements(
					MatchFields(IgnoreExtras, Fields{
						"TransactionHash": Equal(tx1.TransactionHash),
						"BlockNumber":     Equal(tx1.BlockNumber),
						"From":            Equal(tx1.From),
						"To":              Equal(tx1.To),
						"Value":           Equal(tx1.Value),
					}),
					MatchFields(IgnoreExtras, Fields{
						"TransactionHash": Equal(tx2.TransactionHash),
						"BlockNumber":     Equal(tx2.BlockNumber),
						"From":            Equal(tx2.From),
						"To":              Equal(tx2.To),
						"Value":           Equal(tx2.Value),
					}),
				))
			})

			It("returns an empty list for a user with no transactions", func() {
				foundTxs, err := st.GetAllByUser(ctx, "bob")
				Expect(err).NotTo(HaveOccurred())
				Expect(foundTxs).To(BeEmpty())
			})
		})
	})
})
