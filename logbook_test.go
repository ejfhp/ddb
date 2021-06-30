package ddb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

func TestRecordShortText(t *testing.T) {
	log.SetWriter(os.Stdout)
	// toAddress := "1PGh5YtRoohzcZF7WX8SJeZqm6wyaCte7X"
	fromKey := "L4ZaBkP1UTyxdEM7wysuPd1scHMLLf8sf8B2tcEcssUZ7ujrYWcQ"
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	logbook, err := ddb.NewLogbook(fromKey, password, taal, woc)
	if err != nil {
		t.Logf("failed to create new Logbook: %v", err)
		t.Fail()
	}
	filename := "Inferno"
	file := `Nel mezzo del cammin di nostra vita
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
		che mena dritto altrui per ogne calle.`
	txs, err := logbook.RecordFile(filename, []byte(file))
	if err != nil {
		t.Logf("failed to log text: %v", err)
		t.Fail()
	}
	for i, tx := range txs {
		fmt.Printf("%d - ID: %s  TXHEX: %s", i, tx[0], tx[1])
	}
}

func TestAddHeader(t *testing.T) {
	data := "rosso di sera bel tempo si spera"
	payload := ddb.AddHeader(ddb.APP_NAME, ddb.VER_AES, []byte(data))
	if len(payload) != 9+len(data) {
		t.Logf("wrong header len: %d", len(payload))
	}
	fmt.Printf("%s\n", payload)
	if string(payload) != "ddb;0001;"+data {
		t.Logf("wrong header: '%s'", payload)
	}
}
