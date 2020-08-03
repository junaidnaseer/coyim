package gui

import (
	"fmt"

	"github.com/coyim/coyim/i18n"
	"github.com/coyim/coyim/xmpp/jid"
	"github.com/coyim/gotk3adapter/gtki"
)

type createMUCRoom struct {
	accountManager *accountManager
	errorBox       *errorNotification
	builder        *builder

	gtki.Dialog `gtk-widget:"create-chat-dialog"`

	notification         gtki.InfoBar
	notificationArea     gtki.Box          `gtk-widget:"notification-area"`
	form                 gtki.Grid         `gtk-widget:"form"`
	account              gtki.ComboBox     `gtk-widget:"accounts"`
	chatServices         gtki.ComboBoxText `gtk-widget:"chatServices"`
	chatServiceEntry     gtki.Entry        `gtk-widget:"chatServiceEntry"`
	room                 gtki.Entry        `gtk-widget:"room"`
	cancelButton         gtki.Button       `gtk-widget:"button-cancel"`
	createButton         gtki.Button       `gtk-widget:"button-ok"`
	createButtonPrevText string

	model       gtki.ListStore `gtk-widget:"accounts-model"`
	accountList []*account
	cancel      chan bool

	u *gtkUI
}

func (v *createMUCRoom) clearErrors() {
	v.errorBox.Hide()
}

func (v *createMUCRoom) notifyOnError(err string) {
	doInUIThread(func() {
		if v.notification != nil {
			v.notificationArea.Remove(v.notification)
		}

		v.errorBox.ShowMessage(err)
	})
}

func (u *gtkUI) newMUCRoomView(accountManager *accountManager) *createMUCRoom {
	view := &createMUCRoom{
		accountManager: accountManager,
		u:              u,
	}

	view.builder = newBuilder("MUCCreateRoom")
	panicOnDevError(view.builder.bindObjects(view))
	view.errorBox = newErrorNotification(view.notificationArea)

	accountsInput := view.builder.get("accounts").(gtki.ComboBox)
	ac := u.createConnectedAccountsComponent(accountsInput, view,
		func(acc *account) {
			go view.updateChatServices(acc)
		},
		func() {
			view.chatServices.RemoveAll()
		},
	)

	view.builder.ConnectSignals(map[string]interface{}{
		"create_room_handler": func() {
			go view.createRoomHandler()
		},
		"cancel_handler":              view.Destroy,
		"on_close_window_signal":      ac.onDestroy,
		"on_room_changed":             view.disableCreationIfFieldsAreEmpty,
		"on_chatServiceEntry_changed": view.disableCreationIfFieldsAreEmpty,
	})

	return view
}

func (v *createMUCRoom) updateChatServices(ac *account) {
	enteredService, _ := v.chatServiceEntry.GetText()
	v.chatServices.RemoveAll()

	items, err := ac.session.GetChatServices(jid.Parse(ac.Account()).Host())

	if err != nil {
		v.u.log.WithError(err).Debug("something went wrong trying to get chat services")
		return
	}

	for _, i := range items {
		v.chatServices.AppendText(i.Jid)
	}

	if len(items) > 0 && enteredService == "" {
		v.chatServices.SetActive(0)
	}
}

func (v *createMUCRoom) updateFields(f bool) {
	v.cancelButton.SetSensitive(f)
	v.createButton.SetSensitive(f)
	v.account.SetSensitive(f)
	v.room.SetSensitive(f)
	v.chatServices.SetSensitive(f)
}

func (v *createMUCRoom) createRoomHandler() {
	account := v.getCurrentConnectedAcount()
	if account == nil {
		return
	}

	roomName, _ := v.room.GetText()
	service := v.chatServices.GetActiveText()

	if roomName == "" || service == "" {
		v.errorBox.ShowMessage(i18n.Local("Please fill the required fields to create the room."))
		return
	}

	doInUIThread(func() {
		v.updateFields(false)

		v.createButtonPrevText, _ = v.createButton.GetLabel()
		_ = v.createButton.SetProperty("label", i18n.Local("Creating room..."))
	})

	v.cancel = make(chan bool, 1)

	ec := account.session.CreateRoom(jid.Parse(fmt.Sprintf("%s@%s", roomName, service)).(jid.Bare))

	go func() {
		shouldUpdateUI := false
		isRoomCreated := false

		defer func() {
			if shouldUpdateUI {
				doInUIThread(func() {
					if isRoomCreated {
						v.errorBox.ShowMessage(i18n.Local("Room created with success"))
					} else {
						v.errorBox.ShowMessage(i18n.Local("Could not create the new room"))
					}
					v.updateFields(true)
					_ = v.createButton.SetProperty("label", v.createButtonPrevText)
				})
			}
		}()

		for {
			select {
			case err, ok := <-ec:
				if !ok {
					return
				}

				if err != nil {
					v.u.log.WithError(err).Debug("something went wrong trying to create the room")
				} else {
					isRoomCreated = true
				}
				shouldUpdateUI = true
				return
			case <-v.cancel:
				return
			}
		}
	}()
}

func (v *createMUCRoom) getCurrentConnectedAcount() *account {
	v.errorBox.Hide()
	idAcc := v.getSelectedAccountID()
	if idAcc == "" {
		v.errorBox.ShowMessage(i18n.Local("No account selected, please select one account from the list or connect to one."))
		return nil
	}

	account, found := v.accountManager.getAccountByID(idAcc)
	if !found {
		v.errorBox.ShowMessage(i18n.Localf("The given account %s is not connected.", idAcc))
		return nil
	}

	return account
}

func (v *createMUCRoom) getSelectedAccountID() string {
	iter, _ := v.account.GetActiveIter()

	val, err := v.model.GetValue(iter, 1)
	if err != nil {
		return ""
	}

	account, err := val.GetString()
	if err != nil {
		return ""
	}
	return account
}

func (v *createMUCRoom) disableCreationIfFieldsAreEmpty() {
	accountVal := v.getSelectedAccountID()
	serviceVal := v.chatServices.GetActiveText()
	roomVal, _ := v.room.GetText()

	if accountVal == "" || serviceVal == "" || roomVal == "" {
		v.createButton.SetSensitive(false)
	} else {
		v.createButton.SetSensitive(true)
	}
}

func (u *gtkUI) mucCreateChatRoom() {
	view := u.newMUCRoomView(u.accountManager)
	doInUIThread(func() {
		view.SetTransientFor(u.window)
		view.Show()
	})
}
