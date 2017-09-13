package alertrule

import "testing"

func TestAlertruleKeyToConfigMapKey(t *testing.T) {
	expected := "default/alertrule-node-exporter-down"
	cfgMapKey := alertruleKeyToConfigMapKey("default/node-exporter-down")
	if cfgMapKey != expected {
		t.Fatalf("Expected: %v, Actual %v", expected, cfgMapKey)
	}
}