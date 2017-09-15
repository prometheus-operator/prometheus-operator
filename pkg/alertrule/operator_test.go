package alertrule

import "testing"

func TestAlertruleKeyToConfigMapKey(t *testing.T) {
	expected := "default/alertrule-node-exporter-down"
	cfgMapKey := alertruleKeyToConfigMapKey("default/node-exporter-down")
	if cfgMapKey != expected {
		t.Fatalf("Expected: %s, Actual %s", expected, cfgMapKey)
	}
}

func TestConfigMapKeyToAlertruleKey(t *testing.T) {
	expected := "default/danger-danger"
	arKey := configMapKeyToAlertruleKey("default/alertrule-danger-danger")
	if arKey != expected {
		t.Fatalf("Expected: %s, Actual %s", expected, arKey)
	}
}

func TestAlertruleNameToAlertrulePath(t *testing.T) {
	expected := "danger-danger.rules"
	path := alertruleNameToAlertrulePath("danger-danger")
	if path != expected {
		t.Fatalf("Expected: %s, Actual %s", expected, path)
	}
}