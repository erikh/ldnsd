package config

import (
	"reflect"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	t.Run("empty equals the defaults", func(t *testing.T) {
		c := Empty()
		if err := c.validateAndFix(); err != nil {
			t.Fatalf("empty configuration errored out on validation: %v", err)
		}

		c2 := Empty()
		if !reflect.DeepEqual(c, c2) {
			t.Fatal("empty is not equal to the empty after validation")
		}

		c3 := &Config{}
		if err := c3.validateAndFix(); err != nil {
			t.Fatal("empty configuration did not validate properly")
		}

		if !reflect.DeepEqual(c2, c3) {
			t.Fatal("empty, validated configuration was not equal to empty configuration")
		}
	})

	t.Run("defaults are overridable before validation", func(t *testing.T) {
		c := Empty()
		c.Domain = "home"
		if err := c.validateAndFix(); err != nil {
			t.Fatalf("errored out changing the hostname pre-validation: %v", err)
		}

		if c.Domain != "home" {
			t.Fatalf("previously set value was not preserved post-validation")
		}
	})
}
