package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// slugWords, kısa ID üretiminde kullanılan, anlamlı ve telaffuzu kolay kelime havuzu.
// Küçük, çakışma riskini azaltmak için yeterince büyük (100+ kelime).
var slugWords = []string{
	"able", "acid", "aged", "also", "area", "army", "away",
	"baby", "back", "ball", "band", "bank", "base", "bath",
	"bear", "beat", "been", "bell", "best", "bird", "blow",
	"blue", "boat", "body", "bold", "bone", "book", "born",
	"both", "bowl", "burn", "busy", "cake", "calm", "came",
	"card", "care", "case", "cash", "cave", "chip", "cite",
	"city", "clam", "clay", "clip", "club", "coal", "coat",
	"code", "coil", "cold", "come", "cook", "cool", "cope",
	"copy", "core", "corn", "cost", "cozy", "crew", "crop",
	"cube", "cure", "cute", "dark", "data", "dawn", "days",
	"dead", "deal", "dear", "deck", "deep", "dew",  "dice",
	"diet", "disc", "dish", "disk", "dive", "dock", "does",
	"dome", "door", "dove", "down", "draw", "drip", "drop",
	"drum", "duck", "dune", "dusk", "dust", "duty", "each",
	"earn", "ease", "east", "edge", "emit", "epic", "even",
	"ever", "exit", "face", "fact", "fade", "fair", "fall",
	"farm", "fast", "fate", "feel", "feet", "felt", "file",
	"fill", "film", "find", "fire", "firm", "fish", "fist",
	"flag", "flat", "flew", "flip", "flow", "foam", "fold",
	"folk", "fond", "font", "food", "fool", "ford", "fork",
}

// GenerateSlugID, "kap-{kelime}-{kelime}-{2 haneli sayı}" biçiminde
// kriptografik olarak rastgele bir kısa ID döner.
//
// Örnek çıktılar: "kap-fast-apple-47", "kap-bold-river-03"
//
// crypto/rand kullanır; math/rand ile kıyaslandığında tahmin edilemez.
func GenerateSlugID() (string, error) {
	w1, err := randomWord()
	if err != nil {
		return "", fmt.Errorf("utils.GenerateSlugID: birinci kelime: %w", err)
	}

	w2, err := randomWord()
	if err != nil {
		return "", fmt.Errorf("utils.GenerateSlugID: ikinci kelime: %w", err)
	}

	// 00–99 arası iki haneli sayı (önde sıfır korunur).
	num, err := cryptoRandInt(100)
	if err != nil {
		return "", fmt.Errorf("utils.GenerateSlugID: sayı üretimi: %w", err)
	}

	return fmt.Sprintf("kap-%s-%s-%02d", w1, w2, num), nil
}

// randomWord, slugWords havuzundan kriptografik olarak rastgele bir kelime seçer.
func randomWord() (string, error) {
	idx, err := cryptoRandInt(len(slugWords))
	if err != nil {
		return "", fmt.Errorf("utils.randomWord: %w", err)
	}
	return slugWords[idx], nil
}

// cryptoRandInt, [0, max) aralığında kriptografik olarak güvenli rastgele int döner.
func cryptoRandInt(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, fmt.Errorf("utils.cryptoRandInt: rand.Int: %w", err)
	}
	return int(n.Int64()), nil
}
