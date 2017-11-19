// torparse.go - parse various documents produced by Tor
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionutil, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package torparse

import (
	"bytes"
	"encoding/pem"
	"fmt"
)

type TorEntry [][]byte
type TorEntries []TorEntry
type TorEntriesMap map[string]TorEntries

type TorDocument TorEntriesMap

func (te TorEntry) Joined() (joined []byte) {
	for index, subentry := range te {
		if index != 0 {
			joined = append(joined, byte(' '))
		}
		joined = append(joined, subentry...)
	}
	return joined
}

func ExactlyOnce(e TorEntries) bool {
	return len(e) == 1
}

func AtMostOnce(e TorEntries) bool {
	return len(e) <= 1
}

func (entries TorEntries) FJoined() (joined []byte) {
	return entries[0].Joined()
}


func ParseOutNextField(data []byte) (field string, content TorEntry, rest []byte, err error) {
	pemStart := []byte("-----BEGIN ")
	nl_split := bytes.SplitN(data, []byte("\n"), 2)
	if len(nl_split) != 2 {
		return field, content, data,
			fmt.Errorf("Cannot split by newline")
	}
	/* Overwrite with the rest */
	rest = nl_split[1]
	sp_split := bytes.SplitN(nl_split[0], []byte(" "), -1)
	if len(sp_split) <= 0 { /* We have no data left */
		return field, content, data,
			fmt.Errorf("No data left")
	}

	field = string(sp_split[0])
	content = sp_split[1:]
	/* test if we have pem data now. if so append to previous field */
	if bytes.HasPrefix(rest, pemStart) {
		block, pem_rest := pem.Decode(data)
		content = append(content, block.Bytes)
		rest = pem_rest
	}
	return field, content, rest, err
}

// TODO: trim/skip empty strings/separators
func ParseTorDocument(doc_data []byte) (docs []TorDocument, rest []byte) {
	var doc TorDocument
	var field string
	var content TorEntry
	var firstField string

	var parse_err error
	for {
		field, content, doc_data, parse_err = ParseOutNextField(doc_data)
		//log.Printf("parsed: %v : %v", field, content)
		if parse_err != nil {
			//log.Printf("Error parsing document: %v", parse_err)
			break
		}
		if firstField == "" { /* We're just in the begining - doc name */
			firstField = field
		}
		if field == firstField {
			if doc != nil {
				/* Append previous doc */
				docs = append(docs, doc)
			}
			doc = make(TorDocument)
		}
		doc[field] = append(doc[field], content)
	}
	if doc != nil {
		docs = append(docs, doc) /* Append a doc */
	}

	return docs, doc_data
}
