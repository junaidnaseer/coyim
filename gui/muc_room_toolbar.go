package gui

import (
	"github.com/coyim/gotk3adapter/gtki"
)

type roomViewToolbar struct {
	view             gtki.Box            `gtk-widget:"room-view-toolbar"`
	roomNameLabel    gtki.Label          `gtk-widget:"room-name-label"`
	roomSubjectLabel gtki.Label          `gtk-widget:"room-subject-label"`
	roomStatusIcon   gtki.Image          `gtk-widget:"room-status-icon"`
	menu             gtki.MenuToolButton `gtk-widget:"room-menu"`
	destroyItem      gtki.MenuItem       `gtk-widget:"destroy-item"`
}

func (v *roomView) newRoomViewToolbar() *roomViewToolbar {
	t := &roomViewToolbar{}

	t.initBuilder(v)
	t.initDefaults(v)
	t.initSubscribers(v)

	return t
}

func (t *roomViewToolbar) initBuilder(v *roomView) {
	builder := newBuilder("MUCRoomToolbar")
	panicOnDevError(builder.bindObjects(t))

	builder.ConnectSignals(map[string]interface{}{
		"on_leave_room":   v.onLeaveRoom,
		"on_destroy_room": v.onDestroyRoom,
	})
}

func (t *roomViewToolbar) initDefaults(v *roomView) {
	t.roomStatusIcon.SetFromPixbuf(getMUCIconPixbuf("room"))

	t.roomNameLabel.SetText(v.roomID().String())
	mucStyles.setRoomToolbarNameLabelStyle(t.roomNameLabel)

	t.displayRoomSubject(v.room.GetSubject())
	mucStyles.setRoomToolbarSubjectLabelStyle(t.roomSubjectLabel)
}

func (t *roomViewToolbar) initSubscribers(v *roomView) {
	v.subscribe("toolbar", func(ev roomViewEvent) {
		switch e := ev.(type) {
		case subjectReceivedEvent:
			t.subjectReceivedEvent(e.subject)
		case subjectUpdatedEvent:
			t.subjectReceivedEvent(e.subject)
		case roomDestroyedEvent:
			t.roomDestroyedEvent()
		case selfOccupantRemovedEvent:
			t.selfOccupantRemovedEvent()
		case occupantSelfJoinedEvent:
			t.selfOccupantJoinedEvent(v.isSelfOccupantAnOwner())
		}
	})
}

func (t *roomViewToolbar) subjectReceivedEvent(subject string) {
	doInUIThread(func() {
		t.displayRoomSubject(subject)
	})
}

func (t *roomViewToolbar) roomDestroyedEvent() {
	doInUIThread(t.disable)
}

func (t *roomViewToolbar) selfOccupantRemovedEvent() {
	doInUIThread(t.disable)
}

func (t *roomViewToolbar) selfOccupantJoinedEvent(owner bool) {
	t.destroyItem.SetVisible(owner)
}

// disable MUST be called from UI Thread
func (t *roomViewToolbar) disable() {
	mucStyles.setRoomToolbarNameLabelDisabledStyle(t.roomNameLabel)
	t.roomStatusIcon.SetFromPixbuf(getMUCIconPixbuf("room-offline"))
	t.menu.Hide()
}

// displayRoomSubject MUST be called from the UI thread
func (t *roomViewToolbar) displayRoomSubject(subject string) {
	t.roomSubjectLabel.SetText(subject)
	if subject == "" {
		t.roomSubjectLabel.Hide()
	} else {
		t.roomSubjectLabel.Show()
	}
}
