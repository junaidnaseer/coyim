package gui

import (
	"github.com/coyim/coyim/i18n"
	"github.com/coyim/coyim/session"
	"github.com/coyim/coyim/xmpp/jid"
	"github.com/coyim/gotk3adapter/gtki"
)

type roomDestroyErrorView struct {
	roomID            jid.Bare
	reason            string
	alternativeRoomID jid.Bare
	password          string
	destroyError      error

	onRetry func(reason string, alternativeID jid.Bare, password string)

	dialog            gtki.Dialog `gtk-widget:"destroy-room-error-dialog"`
	errorMessage      gtki.Label  `gtk-widget:"destroy-room-error-message"`
	titleErrorMessage gtki.Label  `gtk-widget:"title-error-message"`
}

func (v *roomView) newDestroyError(reason string, alternativeRoomID jid.Bare, password string, err error) *roomDestroyErrorView {
	rd := &roomDestroyErrorView{
		roomID:            v.roomID(),
		reason:            reason,
		alternativeRoomID: alternativeRoomID,
		password:          password,
		destroyError:      err,
		onRetry:           v.tryDestroyRoom,
	}

	rd.initBuilder()
	rd.initDefaults()

	return rd
}

func (rd *roomDestroyErrorView) initBuilder() {
	b := newBuilder("MUCRoomDestroyErrorDialog")
	panicOnDevError(b.bindObjects(rd))

	b.ConnectSignals(map[string]interface{}{
		"on_retry":  rd.retry,
		"on_cancel": rd.close,
	})
}

func (rd *roomDestroyErrorView) initDefaults() {
	mucStyles.setLabelBoldStyle(rd.titleErrorMessage)

	rd.dialog.SetTitle(i18n.Localf("Room [%s] destroy error", rd.roomID))
	rd.errorMessage.SetLabel(rd.friendlyMessageForDestroyError(rd.destroyError))
}

func (rd *roomDestroyErrorView) retry() {
	go rd.onRetry(rd.reason, rd.alternativeRoomID, rd.password)
	rd.close()
}

func (rd *roomDestroyErrorView) close() {
	rd.dialog.Destroy()
}

func (rd *roomDestroyErrorView) show() {
	rd.dialog.Show()
}

func (rd *roomDestroyErrorView) friendlyMessageForDestroyError(err error) string {
	switch err {
	case session.ErrDestroyRoomInvalidIQResponse:
		return i18n.Local("We were able to connect to the room service, " +
			"but we received an invalid response from it. Please try again.")
	case session.ErrDestroyRoomForbidden:
		return i18n.Local("You don't have the permission to destroy this room. " +
			"Please get in contact with one of the room owners.")
	case session.ErrDestroyRoomUnknown:
		return i18n.Local("The room's service responded with an unknow error, " +
			"so, the room can't be destroyed. Please try again.")
	case session.ErrDestroyRoomNoResult:
		return i18n.Local("We were able to send the request to destroy the room, " +
			"but the service responded with an unknow result. Please contact the " +
			"room's administrator.")
	case session.ErrDestroyRoomDoesntExist:
		return i18n.Local("We couldn't find the room that you are trying to destroy.")
	default:
		return i18n.Local("An error occurred when destroying the room. Please try again.")
	}
}
