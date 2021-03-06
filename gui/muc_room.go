package gui

import (
	"github.com/coyim/coyim/coylog"
	"github.com/coyim/coyim/i18n"

	"github.com/coyim/coyim/session/muc"
	"github.com/coyim/coyim/session/muc/data"
	"github.com/coyim/coyim/xmpp/jid"
	"github.com/coyim/gotk3adapter/gtki"
)

type roomViewDataProvider interface {
	passwordProvider() string
	returnTo() func()
}

type roomViewData struct {
	passsword string
	onReturn  func()
}

func (rvd *roomViewData) passwordProvider() string {
	return rvd.passsword
}

func (rvd *roomViewData) returnTo() func() {
	return rvd.onReturn
}

func newRoomViewData() *roomViewData {
	return &roomViewData{}
}

type roomView struct {
	u       *gtkUI
	account *account
	builder *builder

	room *muc.Room

	cancel chan bool

	opened           bool
	passwordProvider func() string
	returnTo         func()

	window                 gtki.Window   `gtk-widget:"room-window"`
	overlay                gtki.Overlay  `gtk-widget:"room-overlay"`
	messagesOverlay        gtki.Overlay  `gtk-widget:"room-messages-overlay"`
	messagesOverlayBox     gtki.Box      `gtk-widget:"room-messages-overlay-box"`
	messagesBox            gtki.Box      `gtk-widget:"room-messages-box"`
	privacityWarningBox    gtki.Box      `gtk-widget:"room-privacity-warnings-box"`
	loadingNotificationBox gtki.Box      `gtk-widget:"room-loading-notification-box"`
	content                gtki.Box      `gtk-widget:"room-main-box"`
	notificationsArea      gtki.Revealer `gtk-widget:"room-notifications-revealer"`
	roomInfoErrorBar       gtki.InfoBar  `gtk-widget:"room-info-error-bar"`

	notifications *roomNotifications

	warnings           *roomViewWarningsOverlay
	warningsInfoBar    *roomViewWarningsInfoBar
	loadingViewOverlay *roomViewLoadingOverlay

	subscribers *roomViewSubscribers

	main    *roomViewMain
	toolbar *roomViewToolbar
	roster  *roomViewRoster
	conv    *roomViewConversation
	lobby   *roomViewLobby

	log coylog.Logger
}

func newRoomView(u *gtkUI, a *account, roomID jid.Bare) *roomView {
	view := &roomView{
		u:       u,
		account: a,
	}

	// TODO: We already know this need to change
	view.room = a.newRoomModel(roomID)
	view.log = a.log.WithField("room", roomID)

	view.room.Subscribe(view.handleRoomEvent)

	view.subscribers = newRoomViewSubscribers(view.roomID(), view.log)

	view.initBuilderAndSignals()
	view.initDefaults()
	view.initSubscribers()
	view.initNotifications()

	view.toolbar = view.newRoomViewToolbar()
	view.roster = view.newRoomViewRoster()
	view.conv = view.newRoomViewConversation()

	view.warnings = view.newRoomViewWarningsOverlay()
	view.warningsInfoBar = view.newRoomViewWarningsInfoBar()
	view.loadingViewOverlay = view.newRoomViewLoadingOverlay()

	view.requestRoomDiscoInfo()

	return view
}

func (v *roomView) initBuilderAndSignals() {
	v.builder = newBuilder("MUCRoomWindow")

	panicOnDevError(v.builder.bindObjects(v))

	v.builder.ConnectSignals(map[string]interface{}{
		"on_destroy_window":       v.onDestroyWindow,
		"on_room_info_load_retry": v.requestRoomDiscoInfo,
	})
}

func (v *roomView) initDefaults() {
	v.setTitle(i18n.Localf("%s [%s]", v.roomID(), v.account.Account()))
}

func (v *roomView) initSubscribers() {
	v.subscribe("room", func(ev roomViewEvent) {
		doInUIThread(func() {
			v.onEventReceived(ev)
		})
	})
}

func (v *roomView) onEventReceived(ev roomViewEvent) {
	switch t := ev.(type) {
	case roomDiscoInfoReceivedEvent:
		v.roomDiscoInfoReceivedEvent(t.info)
	case roomConfigRequestTimeoutEvent:
		v.roomConfigRequestTimeoutEvent()
	case selfOccupantAffiliationUpdatedEvent:
		v.selfOccupantAffiliationUpdatedEvent(t.affiliationUpdate, t.actor, t.reason)
	}
}

func (v *roomView) requestRoomDiscoInfo() {
	v.loadingViewOverlay.onRoomDiscoInfoLoad()
	v.roomInfoErrorBar.Hide()
	go v.account.session.RequestRoomDiscoInfo(v.roomID())
}

// roomDiscoInfoReceivedEvent MUST be called from the UI thread
func (v *roomView) roomDiscoInfoReceivedEvent(di data.RoomDiscoInfo) {
	v.loadingViewOverlay.hide()

	v.warnings.clear()
	v.showRoomWarnings(di)
	v.privacityWarningBox.PackStart(v.warningsInfoBar.view(), true, false, 0)
}

// roomConfigRequestTimeoutEvent MUST be called from the UI thread
func (v *roomView) roomConfigRequestTimeoutEvent() {
	v.loadingViewOverlay.hide()
	v.warnings.clear()

	v.roomInfoErrorBar.Show()
}

// selfOccupantAffiliationUpdatedEvent MUST be called from the UI thread
func (v *roomView) selfOccupantAffiliationUpdatedEvent(affiliationUpdate data.AffiliationUpdate, actor, reason string) {
	m := displaySelfOccupantAffiliationUpdate(affiliationUpdate, actor, reason)
	v.notifications.info(m)
}

func (v *roomView) showRoomWarnings(info data.RoomDiscoInfo) {
	v.warnings.add(i18n.Local("Please be aware that communication in chat rooms is not encrypted - anyone that can intercept communication between you and the server - and the server itself - will be able to see what you are saying in this chat room. Only join this room and communicate here if you trust the server to not be hostile."))

	switch info.AnonymityLevel {
	case "semi":
		v.warnings.add(i18n.Local("This room is partially anonymous. This means that only moderators can connect your nickname with your real username (your JID)."))
	case "no":
		v.warnings.add(i18n.Local("This room is not anonymous. This means that any person in this room can connect your nickname with your real username (your JID)."))
	default:
		v.log.WithField("anonymityLevel", info.AnonymityLevel).Warn("Unknown anonymity setting for room")
	}

	if info.Logged {
		v.warnings.add(i18n.Local("This room is publicly logged, meaning that everything you and the others in the room say or do can be made public on a website."))
	}
}

func (v *roomView) showWarnings() {
	mucStyles.setRoomMessagesBoxStyle(v.messagesBox)

	v.warnings.show()
	v.showNotificationsOverlay()
}

func (v *roomView) showNotificationsOverlay() {
	mucStyles.setRoomOverlayMessagesBoxStyle(v.messagesOverlayBox)
	v.messagesOverlay.Show()
}

func (v *roomView) closeNotificationsOverlay() {
	v.messagesOverlay.Hide()
}

func (v *roomView) onDestroyWindow() {
	v.opened = false
	v.account.removeRoomView(v.roomID())
	go v.cancelActiveRequests()
}

// cancelActiveRequests MUST NOT be called from the UI thread
func (v *roomView) cancelActiveRequests() {
	if v.cancel != nil {
		v.cancel <- true
		v.cancel = nil
	}
}

func (v *roomView) setTitle(t string) {
	v.window.SetTitle(t)
}

func (v *roomView) isOpen() bool {
	return v.opened
}

func (v *roomView) isSelfOccupantInTheRoom() bool {
	return v.room.IsSelfOccupantInTheRoom()
}

func (v *roomView) isSelfOccupantAnOwner() bool {
	return v.room.IsSelfOccupantAnOwner()
}

func (v *roomView) present() {
	if v.isOpen() {
		v.window.Present()
	}
}

func (v *roomView) show() {
	v.opened = true
	v.window.Show()
}

func (v *roomView) onLeaveRoom() {
	// TODO: Implement the logic behind leaving this room and
	// how the view will interact with the user during this process
	v.tryLeaveRoom(nil, nil)
}

// tryLeaveRoom MUST be called from the UI thread.
// Please note that "onSuccess" and "onError" will be called from another thread.
func (v *roomView) tryLeaveRoom(onSuccess func(), onError func(error)) {
	onSuccessFinal := func() {
		doInUIThread(v.window.Destroy)
		if onSuccess != nil {
			onSuccess()
		}
	}

	onErrorFinal := func(err error) {
		v.log.WithError(err).Error("An error occurred when trying to leave the room")
		if onError != nil {
			onError(err)
		}
	}

	go v.account.leaveRoom(
		v.roomID(),
		v.room.SelfOccupantNickname(),
		onSuccessFinal,
		onErrorFinal,
		nil,
	)
}

func (v *roomView) publishDestroyEvent(reason string, alternativeRoomID jid.Bare, password string) {
	v.publishEvent(roomDestroyedEvent{
		reason:      reason,
		alternative: alternativeRoomID,
		password:    password,
	})
}

// tryDestroyRoom MUST be called from the UI thread, but please, note that
// the "onSuccess" and "onError" callbacks will be called from another thread
func (v *roomView) tryDestroyRoom(reason string, alternativeRoomID jid.Bare, password string) {
	v.loadingViewOverlay.onRoomDestroy()

	sc, ec := v.account.session.DestroyRoom(v.roomID(), reason, alternativeRoomID, password)
	go func() {
		select {
		case <-sc:
			v.log.Info("The room has been destroyed")
			v.publishDestroyEvent(reason, alternativeRoomID, password)
			doInUIThread(func() {
				v.notifications.info(i18n.Local("The room has been destroyed"))
				v.loadingViewOverlay.hide()
			})
		case err := <-ec:
			v.log.WithError(err).Error("An error occurred when trying to destroy the room")
			doInUIThread(func() {
				v.loadingViewOverlay.hide()
				dr := createDialogErrorComponent(
					i18n.Localf("Room [%s] destroy error", v.roomID().String()),
					i18n.Local("The room couldn't be destroyed"), "",
					func() {
						v.tryDestroyRoom(reason, alternativeRoomID, password)
					})
				dr.updateMessageForDestroyError(err)
				dr.show()
			})
		}
	}()
}

func (v *roomView) tryUpdateOccupantAffiliation(o *muc.Occupant, affiliation data.Affiliation, reason string) {
	v.loadingViewOverlay.onOccupantAffiliationUpdate()
	sc, ec := v.account.session.UpdateOccupantAffiliation(v.roomID(), o.RealJid, affiliation, reason)

	select {
	case <-sc:
		v.log.Info("The affiliation has been changed")
		v.onOccupantAffiliationUpdateSuccess(o, affiliation, reason)
	case err := <-ec:
		v.log.WithError(err).Error("An error occurred in the affiliation update process")
		v.onOccupantAffiliationUpdateError(o, affiliation, reason)
	}
}

func (v *roomView) onOccupantAffiliationUpdateSuccess(o *muc.Occupant, affiliation data.Affiliation, reason string) {
	affiliationUpdate := data.AffiliationUpdate{
		New:      affiliation,
		Previous: o.Affiliation,
	}
	v.publishOccupantAffiliationUpdatedEvent(o.Nickname, affiliationUpdate, v.room.SelfOccupantNickname(), reason)

	o.UpdateAffiliation(affiliation)

	doInUIThread(func() {
		v.notifications.info(i18n.Localf("The position of %s was updated successfully", o.Nickname))
		v.loadingViewOverlay.hide()
	})
}

func (v *roomView) onOccupantAffiliationUpdateError(o *muc.Occupant, affiliation data.Affiliation, reason string) {
	doInUIThread(func() {
		v.loadingViewOverlay.hide()
		v.notifications.info(i18n.Local("The affiliation update process failed"))
		dr := createDialogErrorComponent(
			i18n.Local("Update occupant position error"),
			i18n.Localf("The position of %s couldn't be updated", o.Nickname),
			i18n.Local("An error occurred in the position update proccess."),
			func() {
				v.tryUpdateOccupantAffiliation(o, affiliation, reason)
			})
		dr.show()
	})
}

func (v *roomView) tryUpdateOccupantRole(o *muc.Occupant, role data.Role, reason string) {
	// TODO: implements loading overlay
	sc, ec := v.account.session.UpdateOccupantRole(v.roomID(), o.RealJid, role, reason)

	select {
	case <-sc:
		v.log.Info("The role has been changed")
		v.onOccupantRoleUpdateSuccess(o, role, reason)
	case err := <-ec:
		v.log.WithError(err).Error("An error occurred in the role update process")
		v.onOccupantRoleUpdateError(o, role, reason)
	}
}

func (v *roomView) onOccupantRoleUpdateSuccess(o *muc.Occupant, role data.Role, reason string) {
	// TODO: publish update occupant role event
	o.UpdateRole(role)
	doInUIThread(func() {
		v.notifications.info(i18n.Localf("The role of %s was updated successfully", o.Nickname))
	})
}

func (v *roomView) onOccupantRoleUpdateError(o *muc.Occupant, role data.Role, reason string) {
	doInUIThread(func() {
		// TODO: Call to error dialog component
		v.notifications.info(i18n.Local("The role update process failed"))
	})
}

func (v *roomView) switchToLobbyView() {
	v.initRoomLobby()

	if v.shouldReturnOnCancel() {
		v.lobby.switchToReturnOnCancel()
	} else {
		v.lobby.switchToCancel()
	}

	v.lobby.show()
}

func (v *roomView) switchToMainView() {
	v.initRoomMain()
	v.main.show()
}

func (v *roomView) onJoined() {
	doInUIThread(func() {
		v.lobby.destroy()
		v.content.Remove(v.lobby.content)
		v.switchToMainView()
	})
}

func (v *roomView) shouldReturnOnCancel() bool {
	return v.returnTo != nil
}

func (v *roomView) onJoinCancel() {
	v.window.Destroy()

	if v.shouldReturnOnCancel() {
		v.returnTo()
	}
}

// messageForbidden MUST NOT be called from the UI thread
func (v *roomView) messageForbidden() {
	v.publishEvent(messageForbidden{})
}

// messageNotAccepted MUST NOT be called from the UI thread
func (v *roomView) messageNotAccepted() {
	v.publishEvent(messageNotAcceptable{})
}

// nicknameConflict MUST NOT be called from the UI thread
func (v *roomView) nicknameConflict(nickname string) {
	v.publishEvent(nicknameConflictEvent{nickname})
}

// serviceUnavailableEvent MUST NOT be called from the UI thread
func (v *roomView) serviceUnavailable() {
	v.publishEvent(serviceUnavailableEvent{})
}

// unknownError MUST NOT be called from the UI thread
func (v *roomView) unknownError() {
	v.publishEvent(unknownErrorEvent{})
}

// registrationRequired MUST NOT be called from the UI thread
func (v *roomView) registrationRequired(nickname string) {
	v.publishEvent(registrationRequiredEvent{nickname})
}

// notAuthorized MUST NOT be called from the UI thread
func (v *roomView) notAuthorized() {
	v.publishEvent(notAuthorizedEvent{})
}

// occupantForbidden MUST NOT be called from the UI thread
func (v *roomView) occupantForbidden() {
	v.publishEvent(occupantForbiddenEvent{})
}

// publishOccupantAffiliationUpdatedEvent MUST NOT be called from the UI thread
func (v *roomView) publishOccupantAffiliationUpdatedEvent(nickname string, affiliationUpdate data.AffiliationUpdate, actor, reason string) {
	v.publishEvent(occupantAffiliationUpdatedEvent{
		nickname:          nickname,
		affiliationUpdate: affiliationUpdate,
		actor:             actor,
		reason:            reason,
	})
}

// publishSelfOccupantAffiliationUpdatedEvent MUST NOT be called from the UI thread
func (v *roomView) publishSelfOccupantAffiliationUpdatedEvent(nickname string, affiliationUpdate data.AffiliationUpdate, actor, reason string) {
	ev := selfOccupantAffiliationUpdatedEvent{}
	ev.nickname = nickname
	ev.affiliationUpdate = affiliationUpdate
	ev.actor = actor
	ev.reason = reason
	v.publishEvent(ev)
}

// mainWindow MUST be called from the UI thread
func (v *roomView) mainWindow() gtki.Window {
	return v.window
}

func (v *roomView) roomID() jid.Bare {
	return v.room.ID
}

func (v *roomView) roomDisplayName() string {
	return v.roomID().Local().String()
}
