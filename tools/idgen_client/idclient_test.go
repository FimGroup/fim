package idgen_client

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed test_amqp_connaddr
var connAddr string

func TestIdClient_GetId(t *testing.T) {
	client := NewIdClient(strings.TrimSpace(connAddr), "general")
	id, err := client.BlockingGetId()
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("no valid id acquired")
	}
	t.Log(id)
}
