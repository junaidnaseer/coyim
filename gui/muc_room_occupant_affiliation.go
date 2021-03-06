package gui

import (
	"github.com/coyim/coyim/i18n"
	"github.com/coyim/coyim/session"
	"github.com/coyim/coyim/session/muc"
	"github.com/coyim/coyim/session/muc/data"
	"github.com/coyim/coyim/xmpp/jid"
	"github.com/coyim/gotk3adapter/gdki"
	"github.com/coyim/gotk3adapter/gtki"
)

func (r *roomViewRosterInfo) onChangeAffiliation() {
	av := r.newOccupantAffiliationUpdateView(r.account, r.roomID, r.occupant)
	av.showDialog()
}

type occupantAffiliationUpdateView struct {
	account        *account
	roomID         jid.Bare
	occupant       *muc.Occupant
	rosterInfoView *roomViewRosterInfo

	dialog           gtki.Dialog      `gtk-widget:"affiliation-dialog"`
	affiliationLabel gtki.Label       `gtk-widget:"affiliation-type-label"`
	adminRadio       gtki.RadioButton `gtk-widget:"affiliation-admin"`
	memberRadio      gtki.RadioButton `gtk-widget:"affiliation-member"`
	noneRadio        gtki.RadioButton `gtk-widget:"affiliation-none"`
	reasonLabel      gtki.Label       `gtk-widget:"affiliation-reason-label"`
	reasonEntry      gtki.TextView    `gtk-widget:"affiliation-reason-entry"`
	applyButton      gtki.Button      `gtk-widget:"affiliation-apply-button"`
}

func (r *roomViewRosterInfo) newOccupantAffiliationUpdateView(a *account, roomID jid.Bare, o *muc.Occupant) *occupantAffiliationUpdateView {
	av := &occupantAffiliationUpdateView{
		account:        a,
		roomID:         roomID,
		rosterInfoView: r,
		occupant:       o,
	}

	av.initBuilder()
	av.initDefaults()

	return av
}

func (av *occupantAffiliationUpdateView) initBuilder() {
	builder := newBuilder("MUCRoomAffiliationDialog")
	panicOnDevError(builder.bindObjects(av))

	builder.ConnectSignals(map[string]interface{}{
		"on_cancel":    av.closeDialog,
		"on_apply":     av.onApply,
		"on_key_press": av.onKeyPress,
		"on_toggled":   av.onRadioButtonToggled,
	})
}

// onRadioButtonToggled MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) onRadioButtonToggled() {
	av.applyButton.SetSensitive(av.occupant.Affiliation.Name() != av.getAffiliationBasedOnRadioSelected().Name())
}

func (av *occupantAffiliationUpdateView) onKeyPress(_ gtki.Widget, ev gdki.Event) {
	if isNormalEnter(g.gdk.EventKeyFrom(ev)) {
		av.onApply()
	}
}

func (av *occupantAffiliationUpdateView) initDefaults() {
	av.dialog.SetTransientFor(av.rosterInfoView.parentWindow())
	mucStyles.setFormSectionLabelStyle(av.affiliationLabel)

	switch av.occupant.Affiliation.(type) {
	case *data.AdminAffiliation:
		av.adminRadio.SetActive(true)
	case *data.MemberAffiliation:
		av.memberRadio.SetActive(true)
	case *data.NoneAffiliation:
		av.noneRadio.SetActive(true)
	}
}

// onApply MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) onApply() {
	go av.rosterInfoView.updateOccupantAffiliation(av.occupant, av.getAffiliationBasedOnRadioSelected(), getTextViewText(av.reasonEntry))
	av.closeDialog()
}

func (av *occupantAffiliationUpdateView) getAffiliationBasedOnRadioSelected() data.Affiliation {
	switch {
	case av.adminRadio.GetActive():
		return &data.AdminAffiliation{}
	case av.memberRadio.GetActive():
		return &data.MemberAffiliation{}
	default:
		return &data.NoneAffiliation{}
	}
}

// show MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) showDialog() {
	av.dialog.Show()
}

// close MUST be called from the UI thread
func (av *occupantAffiliationUpdateView) closeDialog() {
	av.dialog.Destroy()
}

func occupantAffiliationName(a data.Affiliation) string {
	switch a.(type) {
	case *data.OwnerAffiliation:
		return i18n.Local("Owner")
	case *data.AdminAffiliation:
		return i18n.Local("Admin")
	case *data.MemberAffiliation:
		return i18n.Local("Member")
	case *data.OutcastAffiliation:
		return i18n.Local("Outcast")
	case *data.NoneAffiliation:
		return i18n.Local("None")
	default:
		// This should not be possible but we need it to not complain with golang
		return ""
	}
}

func occupantRoleName(a data.Role) string {
	switch a.(type) {
	case *data.ModeratorRole:
		return i18n.Local("Moderator")
	case *data.ParticipantRole:
		return i18n.Local("Participant")
	case *data.VisitorRole:
		return i18n.Local("Visitor")
	case *data.NoneRole:
		return i18n.Local("None")
	default:
		return ""
	}
}

func affiliationUpdateErrorMessage(err error) string {
	switch err {
	case session.ErrUpdateOccupantResponse:
		return i18n.Local("We couldn't update the occupant affiliation because, either you don't have permission to do it or the server is busy. Please try again.")
	default:
		return i18n.Local("An error occurred when updating the occupant affiliation. Please try again.")
	}
}
