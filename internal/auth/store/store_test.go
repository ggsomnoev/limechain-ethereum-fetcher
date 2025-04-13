package store_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ethfetcher/internal/auth/model"
	"ethfetcher/internal/auth/store"
)

var _ = Describe("Store", func() {
	When("created", func() {
		It("exists", func() {
			Expect(store.NewStore(nil)).NotTo(BeNil())
		})
	})

	Describe("instance", func() {
		var (
			st    *store.Store
			token model.Token
		)

		BeforeEach(func() {
			token = model.Token{
				Value:     "sample-token",
				Username:  "alice",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}

			st = store.NewStore(pool)
		})

		Describe("ValidateUser", func() {
			// User is already seeded from the migration.
			Context("when user exists and credentials are valid", func() {
				It("should return true", func() {
					valid, err := st.ValidateUser(ctx, "alice", "alice")
					Expect(err).NotTo(HaveOccurred())
					Expect(valid).To(BeTrue())
				})
			})

			Context("when user does not exist or credentials are invalid", func() {
				It("should return false", func() {
					valid, err := st.ValidateUser(ctx, "nonexistentuser", "wrongpassword")
					Expect(err).NotTo(HaveOccurred())
					Expect(valid).To(BeFalse())
				})
			})
		})

		Describe("StoreToken", func() {
			var errAction error
			JustBeforeEach(func() {
				errAction = st.StoreToken(ctx, token)
			})

			It("succeeds", func() {
				Expect(errAction).To(BeNil())
			})

			Context("when storing a token", func() {
				var err error
				BeforeEach(func() {
					token.Value = "new-token"
					err = st.StoreToken(ctx, token)
				})
				It("should store the token and update if username exists", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Describe("IsTokenValid", func() {
			var (
				valid     bool
				username  string
				errAction error
			)

			JustBeforeEach(func() {
				valid, username, errAction = st.IsTokenValid(ctx, token.Value)
			})

			Context("when token is valid", func() {
				BeforeEach(func() {
					Expect(st.StoreToken(ctx, token)).To(Succeed())
				})
				It("should return the correct user", func() {
					Expect(errAction).NotTo(HaveOccurred())
					Expect(valid).To(BeTrue())
					Expect(username).To(Equal("alice"))
				})
			})

			Context("when token is expired", func() {
				BeforeEach(func() {
					token.ExpiresAt = time.Now().Add(-25 * time.Hour)
					Expect(st.StoreToken(ctx, token)).To(Succeed())
				})
				It("should return false", func() {
					Expect(errAction).NotTo(HaveOccurred())
					Expect(valid).To(BeFalse())
					Expect(username).To(BeEmpty())
				})
			})

			Context("when token is not found", func() {
				BeforeEach(func() {
					token.Value = "non-existent"
				})
				It("should return false", func() {
					Expect(errAction).NotTo(HaveOccurred())
					Expect(valid).To(BeFalse())
					Expect(username).To(BeEmpty())
				})
			})
		})

		Describe("DeleteExpiredTokens", func() {
			Context("when the token is expired", func() {
				BeforeEach(func() {
					token.ExpiresAt = time.Now().Add(-25 * time.Hour)
					Expect(st.StoreToken(ctx, token)).To(Succeed())
				})

				JustBeforeEach(func() {
					Expect(st.DeleteExpiredTokens(ctx)).To(Succeed())
				})

				It("should be deleted", func() {
					valid, username, err := st.IsTokenValid(ctx, token.Value)
					Expect(err).NotTo(HaveOccurred())
					Expect(valid).To(BeFalse())
					Expect(username).To(BeEmpty())
				})
			})
		})
	})

})
