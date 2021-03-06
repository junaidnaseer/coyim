package gui

import (
	"fmt"

	"github.com/coyim/coyim/i18n"

	"github.com/coyim/coyim/coylog"
	"github.com/coyim/coyim/session/muc"
	"github.com/coyim/gotk3adapter/gtki"
	log "github.com/sirupsen/logrus"
)

type mucRoomConfigPage interface {
	pageView() gtki.Overlay
	isNotValid() bool
	showValidationErrors()
	collectData()
	refresh()
	notifyError(string)
	onConfigurationApply()
	onConfigurationApplyError()
	onConfigurationCancel()
	onConfigurationCancelError()
}

type roomConfigPageBase struct {
	u    *gtkUI
	form *muc.RoomConfigForm

	page              gtki.Overlay `gtk-widget:"room-config-page-overlay"`
	content           gtki.Box     `gtk-widget:"room-config-page-content"`
	notificationsArea gtki.Box     `gtk-widget:"notifications-box"`

	notifications  *notificationsComponent
	loadingOverlay *loadingOverlayComponent
	onRefresh      *callbacksSet

	log coylog.Logger
}

func (c *mucRoomConfigComponent) newConfigPage(pageID, pageTemplate string, page interface{}, signals map[string]interface{}) *roomConfigPageBase {
	p := &roomConfigPageBase{
		u:              c.u,
		loadingOverlay: c.u.newLoadingOverlayComponent(),
		onRefresh:      newCallbacksSet(),
		form:           c.form,
		log: c.log.WithFields(log.Fields{
			"page":     pageID,
			"template": pageTemplate,
		}),
	}

	builder := newBuilder("MUCRoomConfigPage")
	panicOnDevError(builder.bindObjects(p))

	p.notifications = c.u.newNotificationsComponent()
	p.loadingOverlay = c.u.newLoadingOverlayComponent()
	p.notificationsArea.Add(p.notifications.contentBox())

	p.page.AddOverlay(p.loadingOverlay.getOverlay())
	p.page.SetHExpand(true)
	p.page.SetVExpand(true)

	builder = newBuilder(pageTemplate)
	panicOnDevError(builder.bindObjects(page))
	builder.ConnectSignals(signals)

	pc, err := builder.GetObject(fmt.Sprintf("room-config-%s-page", pageID))
	if err != nil {
		panic(fmt.Sprintf("developer error: the ID for \"%s\" page doesn't exists", pageID))
	}

	pageContent := pc.(gtki.Box)
	pageContent.SetHExpand(false)
	p.content.Add(pageContent)

	mucStyles.setRoomConfigPageStyle(pageContent)

	return p
}

// pageView implements the "mucRoomConfigPage" interface
func (p *roomConfigPageBase) pageView() gtki.Overlay {
	return p.page
}

// isNotValid implements the "mucRoomConfigPage" interface
func (p *roomConfigPageBase) isNotValid() bool {
	return false
}

// validate implements the "mucRoomConfigPage" interface
func (p *roomConfigPageBase) showValidationErrors() {
}

// Nothing to do, just implement the "mucRoomConfigPage" interface
func (p *roomConfigPageBase) collectData() {}

// refresh MUST be called from the UI thread
func (p *roomConfigPageBase) refresh() {
	p.page.ShowAll()
	p.hideLoadingOverlay()
	p.clearErrors()
	p.onRefresh.invokeAll()
}

// clearErrors MUST be called from the ui thread
func (p *roomConfigPageBase) clearErrors() {
	p.notifications.clearErrors()
}

// notifyError MUST be called from the ui thread
func (p *roomConfigPageBase) notifyError(m string) {
	p.notifications.notifyOnError(m)
}

// onConfigurationApply MUST be called from the ui thread
func (p *roomConfigPageBase) onConfigurationApply() {
	p.showLoadingOverlay(i18n.Local("Saving room configuration"))
}

// onConfigurationApplyError MUST be called from the ui thread
func (p *roomConfigPageBase) onConfigurationApplyError() {
	p.hideLoadingOverlay()
}

// onConfigurationCancel MUST be called from the ui thread
func (p *roomConfigPageBase) onConfigurationCancel() {
	p.showLoadingOverlay(i18n.Local("Aborting room configuration"))
}

// onConfigurationCancelError MUST be called from the ui thread
func (p *roomConfigPageBase) onConfigurationCancelError() {
	p.hideLoadingOverlay()
}

// showLoadingOverlay MUST be called from the ui thread
func (p *roomConfigPageBase) showLoadingOverlay(m string) {
	p.loadingOverlay.setSolid()
	p.loadingOverlay.showWithMessage(m)
}

// hideLoadingOverlay MUST be called from the ui thread
func (p *roomConfigPageBase) hideLoadingOverlay() {
	p.loadingOverlay.hide()
}
