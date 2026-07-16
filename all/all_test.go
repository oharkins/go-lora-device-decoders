package all_test

import (
	"testing"

	decoders "github.com/oharkins/go-lora-device-decoders"
	_ "github.com/oharkins/go-lora-device-decoders/all"
)

func TestEveryDecoderDeclaresOffers(t *testing.T) {
	keys := decoders.List()
	if len(keys) == 0 {
		t.Fatal("no decoders registered")
	}
	for _, k := range keys {
		d, ok := decoders.GetByKey(k)
		if !ok {
			t.Fatalf("missing decoder for %s", k)
		}
		offers := d.Offers()
		if len(offers) == 0 {
			t.Errorf("%s: Offers() is empty — pipeline consumers need declared measurements", k)
			continue
		}
		seen := map[string]struct{}{}
		for _, o := range offers {
			if o.Name == "" {
				t.Errorf("%s: offering with empty name", k)
			}
			if _, dup := seen[o.Name]; dup {
				t.Errorf("%s: duplicate offering %q", k, o.Name)
			}
			seen[o.Name] = struct{}{}
		}
	}
}
