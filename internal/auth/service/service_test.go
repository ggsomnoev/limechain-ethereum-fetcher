package service_test

import (
	"context"
	"errors"
	"ethfetcher/internal/auth/service"
	"ethfetcher/internal/auth/service/servicefakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	ErrStoreFailed = errors.New("store failed")
)

var _ = Describe("Auth Service", func() {
	var (
		ctx   context.Context
		store *servicefakes.FakeStore
		svc   *service.Service

		secret = "mysecret"
	)

	BeforeEach(func() {
		ctx = context.Background()
		store = &servicefakes.FakeStore{}
		svc = service.NewService(store, secret)
	})

	Describe("Authenticate", func() {
		It("returns a token for valid credentials", func() {
			store.ValidateUserReturns(true, nil)
			store.StoreTokenReturns(nil)

			token, err := svc.Authenticate(ctx, "alice", "alice")
			Expect(err).ToNot(HaveOccurred())
			Expect(token).ToNot(BeEmpty())
		})

		It("returns error if the user is invalid", func() {
			store.ValidateUserReturns(false, nil)

			token, err := svc.Authenticate(ctx, "wrong", "user")
			Expect(err).To(MatchError(service.ErrInvalidCredentials))
			Expect(token).To(BeEmpty())
		})

		It("returns error if store returns error on validate", func() {
			store.ValidateUserReturns(false, ErrStoreFailed)

			token, err := svc.Authenticate(ctx, "bob", "bob")
			Expect(err).To(MatchError(ErrStoreFailed))
			Expect(token).To(BeEmpty())
		})

		It("returns error if storing token fails", func() {
			store.ValidateUserReturns(true, nil)
			store.StoreTokenReturns(ErrStoreFailed)

			token, err := svc.Authenticate(ctx, "alice", "alice")
			Expect(err).To(MatchError(ErrStoreFailed))
			Expect(token).To(BeEmpty())
		})
	})

	Describe("ValidateToken", func() {
		It("returns true for valid token", func() {
			store.IsTokenValidReturns(true, nil)
			valid, err := svc.ValidateToken(ctx, "sometoken")
			Expect(err).ToNot(HaveOccurred())
			Expect(valid).To(BeTrue())
		})

		It("returns false if token is not valid", func() {
			store.IsTokenValidReturns(false, nil)
			valid, err := svc.ValidateToken(ctx, "badtoken")
			Expect(err).ToNot(HaveOccurred())
			Expect(valid).To(BeFalse())
		})

		It("returns error if store fails", func() {
			store.IsTokenValidReturns(false, ErrStoreFailed)
			valid, err := svc.ValidateToken(ctx, "token")
			Expect(err).To(MatchError(ErrStoreFailed))
			Expect(valid).To(BeFalse())
		})
	})

	Describe("DeleteExpiredTokens", func() {
		It("succeeds", func() {
			store.DeleteExpiredTokensReturns(nil)
			err := svc.DeleteExpiredTokens(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if store fails", func() {
			store.DeleteExpiredTokensReturns(ErrStoreFailed)
			err := svc.DeleteExpiredTokens(ctx)
			Expect(err).To(MatchError(ErrStoreFailed))
		})
	})
})
