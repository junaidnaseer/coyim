package session

import (
	"bytes"
	"encoding/xml"

	"github.com/coyim/coyim/session/access"
	"github.com/coyim/coyim/session/filetransfer"
	"github.com/coyim/coyim/xmpp/data"
)

func init() {
	registerKnownIQ("set", "http://jabber.org/protocol/si si", streamInitIQ)
}

var supportedSIProfiles = map[string]func(access.Session, *data.ClientIQ, data.SI) (interface{}, string, bool){
	"http://jabber.org/protocol/si/profile/file-transfer":           filetransfer.InitIQ,
	"http://jabber.org/protocol/si/profile/directory-transfer":      filetransfer.InitIQ,
	"http://jabber.org/protocol/si/profile/encrypted-data-transfer": filetransfer.InitIQ,
}

func streamInitIQ(s access.Session, stanza *data.ClientIQ) (ret interface{}, iqtype string, ignore bool) {
	s.Log().Info("IQ: stream initiation")
	var si data.SI
	if err := xml.NewDecoder(bytes.NewBuffer(stanza.Query)).Decode(&si); err != nil {
		s.Log().WithError(err).Warn("Failed to parse stream initiation")
	}

	prof, ok := supportedSIProfiles[si.Profile]
	if !ok {
		s.Log().WithField("profile", si.Profile).Warn("Unsupported SI profile")
		return nil, "", false
	}

	return prof(s, stanza, si)
}
