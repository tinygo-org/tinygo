package hmacdrbg

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// Vectors from NIST CAVP drbgvectors_no_reseed.zip, HMAC_DRBG.rsp, SHA-256, COUNT = 0.
// https://csrc.nist.gov/projects/cryptographic-algorithm-validation-program/random-number-generators
var hmacDRBGTests = []struct {
	entropy, nonce, personalization, add1, add2, returned string
}{
	{
		entropy:  "ca851911349384bffe89de1cbdc46e6831e44d34a4fb935ee285dd14b71a7488",
		nonce:    "659ba96c601dc69fc902940805ec0ca8",
		returned: "e528e9abf2dece54d47c7e75e5fe302149f817ea9fb4bee6f4199697d04d5b89d54fbb978a15b5c443c9ec21036d2460b6f73ebad0dc2aba6e624abf07745bc107694bb7547bb0995f70de25d6b29e2d3011bb19d27676c07162c8b5ccde0668961df86803482cb37ed6d5c0bb8d50cf1f50d476aa0458bdaba806f48be9dcb8",
	},
	{
		entropy:  "d3cc4d1acf3dde0c4bd2290d262337042dc632948223d3a2eaab87da44295fbd",
		nonce:    "0109b0e729f457328aa18569a9224921",
		add1:     "3c311848183c9a212a26f27f8c6647e40375e466a0857cc39c4e47575d53f1f6",
		add2:     "fcb9abd19ccfbccef88c9c39bfb3dd7b1c12266c9808992e305bc3cff566e4e4",
		returned: "9c7b758b212cd0fcecd5daa489821712e3cdea4467b560ef5ddc24ab47749a1f1ffdbbb118f4e62fcfca3371b8fbfc5b0646b83e06bfbbab5fac30ea09ea2bc76f1ea568c9be0444b2cc90517b20ca825f2d0eccd88e7175538b85d90ab390183ca6395535d34473af6b5a5b88f5a59ee7561573337ea819da0dcc3573a22974",
	},
	{
		entropy:         "5cacc68165a2e2ee20812f35ec73a79dbf30fd475476ac0c44fc6174cdac2b55",
		nonce:           "6f885496c1e63af620becd9e71ecb824",
		personalization: "e72dd8590d4ed5295515c35ed6199e9d211b8f069b3058caa6670b96ef1208d0",
		returned:        "f1012cf543f94533df27fedfbf58e5b79a3dc517a9c402bdbfc9a0c0f721f9d53faf4aafdc4b8f7a1b580fcaa52338d4bd95f58966a243cdcd3f446ed4bc546d9f607b190dd69954450d16cd0e2d6437067d8b44d19a6af7a7cfa8794e5fbd728e8fb2f2e8db5dd4ff1aa275f35886098e80ff844886060da8b1e7137846b23b",
	},
	{
		entropy:         "5d3286bc53a258a53ba781e2c4dcd79a790e43bbe0e89fb3eed39086be34174b",
		nonce:           "c5422294b7318952ace7055ab7570abf",
		personalization: "2dba094d008e150d51c4135bb2f03dcde9cbf3468a12908a1b025c120c985b9d",
		add1:            "793a7ef8f6f0482beac542bb785c10f8b7b406a4de92667ab168ecc2cf7573c6",
		add2:            "2238cdb4e23d629fe0c2a83dd8d5144ce1a6229ef41dabe2a99ff722e510b530",
		returned:        "d04678198ae7e1aeb435b45291458ffde0891560748b43330eaf866b5a6385e74c6fa5a5a44bdb284d436e98d244018d6acedcdfa2e9f499d8089e4db86ae89a6ab2d19cb705e2f048f97fb597f04106a1fa6a1416ad3d859118e079a0c319eb95686f4cbcce3b5101c7a0b010ef029c4ef6d06cdfac97efb9773891688c37cf",
	},
}

func TestHMACDRBG(t *testing.T) {
	unhex := func(s string) []byte {
		b, err := hex.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	for i, tc := range hmacDRBGTests {
		d := New(unhex(tc.entropy), unhex(tc.nonce), unhex(tc.personalization))
		want := unhex(tc.returned)
		got := make([]byte, len(want))
		if err := d.Generate(got, unhex(tc.add1)); err != nil {
			t.Fatal(err)
		}
		if err := d.Generate(got, unhex(tc.add2)); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("vector %d: got %x, want %x", i, got, want)
		}
	}
}

func TestHMACDRBGLimits(t *testing.T) {
	d := New(make([]byte, 32), make([]byte, 16), nil)
	if err := d.Generate(make([]byte, MaxRequest+1), nil); err != ErrRequest {
		t.Errorf("large request: got %v, want %v", err, ErrRequest)
	}
	d.reseedCounter = reseedInterval + 1
	if err := d.Generate(make([]byte, 1), nil); err != ErrReseed {
		t.Errorf("reseed interval: got %v, want %v", err, ErrReseed)
	}
}
