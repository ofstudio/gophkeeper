package models

import "github.com/awnumar/memguard"

// Wipe wipes filed values of the ItemData with zeros.
func (d *ItemData) Wipe() {
	for _, field := range d.Fields {
		if field != nil && field.Value != nil {
			memguard.WipeBytes(field.Value)
		}
	}
}
