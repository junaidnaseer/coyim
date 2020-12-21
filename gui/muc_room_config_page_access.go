package gui

import "github.com/coyim/gotk3adapter/gtki"

type roomConfigAccessPage struct {
	*roomConfigPageBase

	configAccessBox  gtki.Box    `gtk-widget:"room-config-access-page"`
	notificationBox  gtki.Box    `gtk-widget:"notification-box"`
	roomPassword     gtki.Entry  `gtk-widget:"room-password"`
	roomMembersOnly  gtki.Switch `gtk-widget:"room-membersonly"`
	roomAllowInvites gtki.Switch `gtk-widget:"room-allowinvites"`
}

func (c *mucRoomConfigComponent) newRoomConfigAccessPage() mucRoomConfigPage {
	p := &roomConfigAccessPage{}

	builder := newBuilder("MUCRoomConfigPageAccess")
	panicOnDevError(builder.bindObjects(p))

	p.roomConfigPageBase = c.newConfigPage(p.configAccessBox, p.notificationBox)

	p.initDefaultValues()

	return p
}

func (p *roomConfigAccessPage) initDefaultValues() {
	setEntryText(p.roomPassword, p.form.Password)
	setSwitchActive(p.roomMembersOnly, p.form.MembersOnly)
	setSwitchActive(p.roomAllowInvites, p.form.OccupantsCanInvite)
}

func (p *roomConfigAccessPage) collectData() {
	p.form.Password = getEntryText(p.roomPassword)
	p.form.PasswordProtected = p.form.Password != ""
	p.form.MembersOnly = getSwitchActive(p.roomMembersOnly)
	p.form.OccupantsCanInvite = getSwitchActive(p.roomAllowInvites)
}