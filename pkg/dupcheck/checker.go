package dupcheck

import (
	"encoding/json"
	"fmt"

	appErrors "github.com/PoorMercymain/bashrun/errors"
)

func CheckDuplicatesInJSON(d *json.Decoder, path []string) error {
	t, err := d.Token()
	if err != nil {
		return err
	}

	switch delim := t.(type) {
	case json.Delim:
		if delim == '{' {
			return checkObject(d, path)
		} else if delim == '[' {
			return checkArray(d, path)
		}
	}

	return nil
}

func checkObject(d *json.Decoder, path []string) error {
	keys := make(map[string]bool)
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return err
		}
		key := t.(string)

		if keys[key] {
			return fmt.Errorf("%w: %v", appErrors.ErrDuplicateInJSON, append(path, key))
		}
		keys[key] = true

		if err := CheckDuplicatesInJSON(d, append(path, key)); err != nil {
			return err
		}
	}

	_, err := d.Token()

	return err
}

func checkArray(d *json.Decoder, path []string) error {
	values := make(map[string]bool)
	index := 0
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return err
		}

		valStr := fmt.Sprintf("%v", t)

		if values[valStr] {
			return fmt.Errorf("%w: duplicate value %v at index %d in path %v", appErrors.ErrDuplicateInJSON, t, index, path)
		}
		values[valStr] = true

		index++
	}

	_, err := d.Token()

	return err
}
