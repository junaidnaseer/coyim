package session

import (
	"github.com/coyim/coyim/session/muc/data"
	xmppData "github.com/coyim/coyim/xmpp/data"
	"github.com/coyim/coyim/xmpp/jid"
	log "github.com/sirupsen/logrus"
)

func (s *session) UpdateOccupantRole(roomID jid.Bare, occupantID jid.Full, role data.Role, reason string) (<-chan bool, <-chan error) {
	return s.muc.updateOccupantRole(roomID, occupantID, role, reason)
}

func (m *mucManager) updateOccupantRole(roomID jid.Bare, occupantID jid.Full, role data.Role, reason string) (<-chan bool, <-chan error) {
	l := m.log.WithFields(log.Fields{
		"room":     roomID,
		"occupant": occupantID,
		"role":     role.Name(),
	})

	rc := make(chan bool)
	ec := make(chan error)

	go func() {
		reply, _, err := m.conn().SendIQ(roomID.String(), "set", &xmppData.MUCAdmin{
			Item: &xmppData.MUCItem{
				Role:   role.Name(),
				Jid:    occupantID.String(),
				Reason: reason,
			},
		})

		if err != nil {
			l.WithError(err).Error("An error occurred when updating the occupant role")
			ec <- ErrUpdateOccupantRequest
			return
		}

		err = validateIqResponse(reply)
		if err != nil {
			l.WithError(err).Error("An error occurred when trying to read the response from the room configuration rollback request")
			ec <- ErrUpdateOccupantResponse
			return
		}

		rc <- true
	}()

	return rc, ec
}
