
package main

import "testing"

type BondAPI struct {
}

func New() *BondAPI {
   return &BondAPI{}
}

func TestNew(t *testing.T) {
    t.Run("should return a new instance of BondAPI", func(t *testing.T) {
        api := New()
        if api == nil {
            t.Error("expected a new instance of BondAPI, got nil")
        }
    })
}