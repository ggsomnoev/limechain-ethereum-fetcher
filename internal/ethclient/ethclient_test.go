package ethclient_test

import (
	"encoding/hex"
	"ethfetcher/internal/ethclient"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TODO: test other eth client methods
var _ = Describe("RlpHexToHashList", func() {
	It("decodes RLP hex into list of hashes", func() {
		hash1 := common.Hex2Bytes("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		hash2 := common.Hex2Bytes("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
		rlpBytes, err := rlp.EncodeToBytes([][]byte{hash1, hash2})
		Expect(err).NotTo(HaveOccurred())

		rlpHex := "0x" + hex.EncodeToString(rlpBytes)

		client := ethclient.NewEthereumClient(nil)
		hashes, err := client.RlpHexToHashList(rlpHex)
		Expect(err).NotTo(HaveOccurred())

		Expect(hashes).To(HaveLen(2))
		Expect(hashes[0]).To(Equal("0x" + hex.EncodeToString(hash1)))
		Expect(hashes[1]).To(Equal("0x" + hex.EncodeToString(hash2)))
	})

	It("returns error on invalid hex string", func() {
		client := ethclient.NewEthereumClient(nil)
		_, err := client.RlpHexToHashList("not-a-hex")
		Expect(err).To(MatchError(ContainSubstring("failed to decode hex string")))
	})

	It("returns error on invalid RLP", func() {
		client := ethclient.NewEthereumClient(nil)
		_, err := client.RlpHexToHashList("0xdeaded69")
		Expect(err).To(MatchError(ContainSubstring("failed to decode RLP")))
	})
})
