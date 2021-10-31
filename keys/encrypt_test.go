package keys_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"testing"

	"github.com/ejfhp/ddb/keys"
)

func TestEncryptDecryptText(t *testing.T) {
	key := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2}
	tests := [][]byte{
		[]byte("tanto va la gatta al lardo che ci lascia lo zampino"),
		[]byte(`Nel mezzo del cammin di nostra vita
		mi ritrovai per una selva oscura,
		ché la diritta via era smarrita.
		
		Ahi quanto a dir qual era è cosa dura
		esta selva selvaggia e aspra e forte
		che nel pensier rinova la paura!
		
		Tant’è amara che poco è più morte;
		ma per trattar del ben ch’i’ vi trovai,
		dirò de l’altre cose ch’i’ v’ ho scorte.
		
		Io non so ben ridir com’i’ v’intrai,
		tant’era pien di sonno a quel punto
		che la verace via abbandonai.
		
		Ma poi ch’i’ fui al piè d’un colle giunto,
		là dove terminava quella valle
		che m’avea di paura il cor compunto,
		
		guardai in alto e vidi le sue spalle
		vestite già de’ raggi del pianeta
		che mena dritto altrui per ogne calle.`),
	}
	for i, txt := range tests {
		crypted, err := keys.AESEncrypt(key, []byte(txt))
		if err != nil {
			t.Logf("first encryption has failed: %v", err)
			t.Fail()
		}
		decrypted, err := keys.AESDecrypt(key, crypted)
		if err != nil {
			t.Logf("first decryption has failed: %v", err)
			t.Fail()
		}
		if string(decrypted) != string(txt) {
			t.Logf("%d: encryption/decription failed '%s' != '%s'", i, string(decrypted), txt)
			t.Fail()

		}
	}
}

func TestEncryptDecryptFile(t *testing.T) {
	key := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2}
	file := "../testdata/image.png"
	image, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("error reading test file %s: %v", file, err)
	}
	imageSha := sha256.Sum256(image)
	imageHash := hex.EncodeToString(imageSha[:])
	tests := [][]byte{
		image,
	}
	for i, txt := range tests {
		crypted, err := keys.AESEncrypt(key, []byte(txt))
		if err != nil {
			t.Logf("first encryption has failed: %v", err)
			t.Fail()
		}
		decrypted, err := keys.AESDecrypt(key, crypted)
		if err != nil {
			t.Logf("first decryption has failed: %v", err)
			t.Fail()
		}
		decSha := sha256.Sum256(decrypted)
		decHash := hex.EncodeToString(decSha[:])
		if decHash != imageHash {
			t.Logf("%d: encryption/decription failed '%s' != '%s'", i, decHash, imageHash)
			t.Fail()
		}
	}
}
