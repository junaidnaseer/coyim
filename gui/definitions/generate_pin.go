package definitions

func init() {
	add(`GeneratePIN`, &defGeneratePIN{})
}

type defGeneratePIN struct{}

func (*defGeneratePIN) String() string {
	return `<interface>
  <object class="GtkDialog" id="dialog">
    <property name="window-position">GTK_WIN_POS_CENTER</property>
    <child internal-child="vbox">
        <object class="GtkBox" id="notification-area">
          <property name="border-width">10</property>
          <property name="homogeneous">false</property>
          <property name="orientation">GTK_ORIENTATION_VERTICAL</property>
          <child>
            <object class="GtkBox">
              <property name="homogeneous">false</property>
              <property name="orientation">GTK_ORIENTATION_VERTICAL</property>
              <property name="spacing">6</property>
              <child>
                  <object class="GtkImage">
                      <property name="file">build/images/Pin_1.png</property>
                  </object>
              </child>
              <child>
                  <object class="GtkLabel" id="SharePinLabel">
                      <property name="label" translatable="yes"></property>
                  </object>
              </child>
              <child>
                  <object class="GtkLabel" id="PinLabel">
                      <property name="visible">True</property>
                      <property name="selectable">True</property>
                  </object>
              </child>
              <child>
                  <object class="GtkButton" id="GenPin">
                      <property name="label" translatable="yes">Generate New PIN</property>
                      <property name="receives-default">True</property>
                      <signal name="clicked" handler="on_gen_pin"/>
                  </object>
                  <packing>
                      <property name="padding">10</property>
                  </packing>
              </child>
              <child>
                  <object class="GtkGrid">
                  <property name="column-spacing">6</property>
                  <property name="row-spacing">2</property>
                  <child>
                      <object class="GtkImage">
                          <property name="file">build/images/padlock.png</property>
                      </object>
                      <packing>
                          <property name="left-attach">0</property>
                          <property name="top-attach">0</property>
                      </packing>
                  </child>
                  <child>
                      <object class="GtkLabel">
                          <property name="visible">True</property>
                          <property name="label" translatable="yes">Share in person</property>
                          <property name="justify">GTK_JUSTIFY_LEFT</property>
                          <property name="halign">GTK_ALIGN_START</property>
                      </object>
                      <packing>
                          <property name="left-attach">1</property>
                          <property name="top-attach">0</property>
                      </packing>
                  </child>
                  <child>
                      <object class="GtkImage">
                          <property name="file">build/images/padlock.png</property>
                      </object>
                      <packing>
                          <property name="left-attach">0</property>
                          <property name="top-attach">1</property>
                      </packing>
                  </child>
                  <child>
                      <object class="GtkLabel">
                          <property name="visible">True</property>
                          <property name="label" translatable="yes">Share through an encrypted channel</property>
                          <property name="justify">GTK_JUSTIFY_LEFT</property>
                          <property name="halign">GTK_ALIGN_START</property>
                      </object>
                      <packing>
                          <property name="left-attach">1</property>
                          <property name="top-attach">1</property>
                      </packing>
                  </child>
                  <child>
                      <object class="GtkImage">
                          <property name="file">build/images/alert.png</property>
                      </object>
                      <packing>
                          <property name="left-attach">0</property>
                          <property name="top-attach">2</property>
                      </packing>
                  </child>
                  <child>
                      <object class="GtkLabel">
                          <property name="visible">True</property>
                          <property name="label" translatable="yes">Share in a phone call</property>
                          <property name="justify">GTK_JUSTIFY_LEFT</property>
                          <property name="halign">GTK_ALIGN_START</property>
                      </object>
                      <packing>
                          <property name="left-attach">1</property>
                          <property name="top-attach">2</property>
                      </packing>
                  </child>
                  </object>
              </child>
            </object>
          </child>
          <child internal-child="action_area">
            <object class="GtkButtonBox" id="button_box">
              <property name="orientation">GTK_ORIENTATION_HORIZONTAL</property>
              <child>
                <object class="GtkButton" id="button_finished">
                  <property name="can-default">true</property>
                  <property name="label" translatable="yes">Finish</property>
                  <signal name="clicked" handler="close_share_pin"/>
                </object>
              </child>
            </object>
          </child>
        </object>
    </child>
  </object>
</interface>
`
}