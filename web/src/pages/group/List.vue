<template>
  <div>
    <el-card>
      <el-form :inline="true">
        <el-form-item>
          <el-button
              size="small"
              type="primary"
              icon="el-icon-plus"
              @click="handleEdit('')">Add
          </el-button>
        </el-form-item>
      </el-form>

      <el-table
          ref="multipleTable"
          :data="tableData"
          border>

        <el-table-column
            sortable="true"
            prop="id"
            label="ID"
            width="60">
        </el-table-column>

        <el-table-column
            prop="name"
            label="Name">
        </el-table-column>

        <el-table-column
            prop="note"
            label="Description">
        </el-table-column>

        <el-table-column
            prop="allow_lan"
            label="LAN">
          <template slot-scope="scope">
            <el-switch
                v-model="scope.row.allow_lan"
                disabled>
            </el-switch>
          </template>
        </el-table-column>

        <el-table-column
            prop="bandwidth"
            label="Bandwidth"
            width="90">
          <template slot-scope="scope">
            <el-row v-if="scope.row.bandwidth > 0">{{ convertBandwidth(scope.row.bandwidth, 'BYTE', 'Mbps') }} Mbps
            </el-row>
            <el-row v-else>Unlimited</el-row>
          </template>
        </el-table-column>

        <el-table-column
            prop="client_dns"
            label="Client DNS"
            width="150">
          <template slot-scope="scope">
            <el-row v-for="(item,inx) in scope.row.client_dns" :key="inx">{{ item.val }}</el-row>
          </template>
        </el-table-column>

        <el-table-column
            prop="route_include"
            label="Route include"
            width="180">
          <template slot-scope="scope">
            <el-row v-for="(item,inx) in scope.row.route_include.slice(0, readMinRows)" :key="inx">{{
                item.val
              }}
            </el-row>
            <div v-if="scope.row.route_include.length > readMinRows">
              <div v-if="readMore[`ri_${ scope.row.id }`]">
                <el-row v-for="(item,inx) in scope.row.route_include.slice(readMinRows)" :key="inx">{{
                    item.val
                  }}
                </el-row>
              </div>
              <el-button size="mini" type="text" @click="toggleMore(`ri_${ scope.row.id }`)">
                {{ readMore[`ri_${scope.row.id}`] ? "▲ Collapse" : "▼ More" }}
              </el-button>
            </div>
          </template>
        </el-table-column>

        <el-table-column
            prop="route_exclude"
            label="Route exclude"
            width="180">
          <template slot-scope="scope">
            <el-row v-for="(item,inx) in scope.row.route_exclude.slice(0, readMinRows)" :key="inx">{{
                item.val
              }}
            </el-row>
            <div v-if="scope.row.route_exclude.length > readMinRows">
              <div v-if="readMore[`re_${ scope.row.id }`]">
                <el-row v-for="(item,inx) in scope.row.route_exclude.slice(readMinRows)" :key="inx">{{
                    item.val
                  }}
                </el-row>
              </div>
              <el-button size="mini" type="text" @click="toggleMore(`re_${ scope.row.id }`)">
                {{ readMore[`re_${scope.row.id}`] ? "▲ Collapse" : "▼ More" }}
              </el-button>
            </div>
          </template>
        </el-table-column>

        <el-table-column
            prop="link_acl"
            label="ACL"
            min-width="160">
          <template slot-scope="scope">
            <el-row v-for="(item,inx) in scope.row.link_acl.slice(0, readMinRows)" :key="inx">
              {{ item.action }} => {{ item.val }} : {{ item.port }}
            </el-row>
            <div v-if="scope.row.link_acl.length > readMinRows">
              <div v-if="readMore[`la_${ scope.row.id }`]">
                <el-row v-for="(item,inx) in scope.row.link_acl.slice(readMinRows)" :key="inx">
                  {{ item.action }} => {{ item.val }} : {{ item.port }}
                </el-row>
              </div>
              <el-button size="mini" type="text" @click="toggleMore(`la_${ scope.row.id }`)">
                {{ readMore[`la_${scope.row.id}`] ? "▲ Collapse" : "▼ More" }}
              </el-button>
            </div>
          </template>
        </el-table-column>

        <el-table-column
            prop="status"
            label="Status"
            width="80">
          <template slot-scope="scope">
            <el-tag v-if="scope.row.status === 1" type="success">Enabled</el-tag>
            <el-tag v-else type="danger">Disabled</el-tag>
          </template>

        </el-table-column>

        <el-table-column
            prop="updated_at"
            label="Update time"
            :formatter="tableDateFormat">
        </el-table-column>

        <el-table-column
            label="Actions"
            width="160">
          <template slot-scope="scope">
            <el-button
                size="mini"
                type="primary"
                @click="handleEdit(scope.row)">Edit
            </el-button>

            <el-popconfirm
                style="margin-left: 10px"
                @confirm="handleDel(scope.row)"
                title="Are you sure you want to delete the user group?">
              <el-button
                  slot="reference"
                  size="mini"
                  type="danger">Delete
              </el-button>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div class="sh-20"></div>

      <el-pagination
          background
          layout="prev, pager, next"
          :pager-count="11"
          @current-change="pageChange"
          :current-page="page"
          :total="count">
      </el-pagination>

    </el-card>

    <!--Add and modify pop-up boxes-->
    <el-dialog
        :close-on-click-modal="false"
        title="Adding user group"
        :visible.sync="user_edit_dialog"
        width="850px"
        @close='closeDialog'
        center>

      <el-form :model="ruleForm" :rules="rules" ref="ruleForm" label-width="100px" class="ruleForm">
        <el-tabs v-model="activeTab" :before-leave="beforeTabLeave">
          <el-tab-pane label="General" name="general">
            <el-form-item label="User group ID" prop="id">
              <el-input v-model="ruleForm.id" disabled></el-input>
            </el-form-item>

            <el-form-item label="Group name" prop="name">
              <el-input v-model="ruleForm.name" :disabled="ruleForm.id > 0"></el-input>
            </el-form-item>

            <el-form-item label="Remark" prop="note">
              <el-input v-model="ruleForm.note"></el-input>
            </el-form-item>

            <el-form-item label="Bandwidth limitation" prop="bandwidth_format" style="width:260px;">
              <el-input v-model="ruleForm.bandwidth_format"
                        oninput="value= value.match(/\d+(\.\d{0,2})?/) ? value.match(/\d+(\.\d{0,2})?/)[0] : ''">
                <template slot="append">Mbps</template>
              </el-input>
            </el-form-item>
            <el-form-item label="Exclude local network" prop="allow_lan">
              <!--  active-text="After it is turned on, the user's local network segment will not be encrypted via anylink." -->
              <el-switch v-model="ruleForm.allow_lan"></el-switch>
              <div class="msg-info">
                Note: Local network refers to:
                The network where the PC running the anyconnect client is located, that is, the local routing segment.
                After it is enabled, data in the local routing segment of the PC will not be forwarded via the tunnel link.
                At the same time, the anyconnect client needs to check the local network (Allow Local Lan) switch for the function to take effect.
              </div>
            </el-form-item>

            <el-form-item label="Client DNS" prop="client_dns">
              <el-row class="msg-info">
                <el-col :span="20">Enter the IP format such as: 192.168.0.10</el-col>
                <el-col :span="4">
                  <el-button size="mini" type="success" icon="el-icon-plus" circle
                             @click.prevent="addDomain(ruleForm.client_dns)"></el-button>
                </el-col>
              </el-row>
              <el-row v-for="(item,index) in ruleForm.client_dns"
                      :key="index" style="margin-bottom: 5px" :gutter="10">
                <el-col :span="10">
                  <el-input v-model="item.val"></el-input>
                </el-col>
                <el-col :span="12">
                  <el-input v-model="item.note" placeholder="Note"></el-input>
                </el-col>
                <el-col :span="2">
                  <el-button size="mini" type="danger" icon="el-icon-minus" circle
                             @click.prevent="removeDomain(ruleForm.client_dns,index)"></el-button>
                </el-col>
              </el-row>
            </el-form-item>

            <el-form-item label="Intranet domain name" prop="split_dns">
              <el-row class="msg-info">
                <el-col :span="20">(Split DNS) is usually left blank. If you enter a domain name, only the configured domain name (including subdomains) will go through the configured dns</el-col>
                <el-col :span="4">
                  <el-button size="mini" type="success" icon="el-icon-plus" circle
                             @click.prevent="addDomain(ruleForm.split_dns)"></el-button>
                </el-col>
              </el-row>
              <el-row v-for="(item,index) in ruleForm.split_dns"
                      :key="index" style="margin-bottom: 5px" :gutter="10">
                <el-col :span="10">
                  <el-input v-model="item.val"></el-input>
                </el-col>
                <el-col :span="12">
                  <el-input v-model="item.note" placeholder="Note"></el-input>
                </el-col>
                <el-col :span="2">
                  <el-button size="mini" type="danger" icon="el-icon-minus" circle
                             @click.prevent="removeDomain(ruleForm.split_dns,index)"></el-button>
                </el-col>
              </el-row>
            </el-form-item>

            <el-form-item label="State" prop="status">
              <el-radio-group v-model="ruleForm.status">
                <el-radio :label="1" border>Enable</el-radio>
                <el-radio :label="0" border>Disable</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-tab-pane>

          <el-tab-pane label="Authentication" name="authtype">
            <el-form-item label="Certification" prop="authtype">
              <el-radio-group v-model="ruleForm.auth.type" @change="authTypeChange">
                <el-radio label="local" border>Local</el-radio>
                <el-radio label="radius" border>Radius</el-radio>
                <el-radio label="ldap" border>LDAP</el-radio>
              </el-radio-group>
            </el-form-item>
            <template v-if="ruleForm.auth.type == 'radius'">
              <el-form-item label="Server address" prop="auth.radius.addr"
                            :rules="this.ruleForm.auth.type== 'radius' ? this.rules['auth.radius.addr'] : [{ required: false }]">
                <el-input v-model="ruleForm.auth.radius.addr" placeholder="For example ip:1812"></el-input>
              </el-form-item>
              <el-form-item label="Key" prop="auth.radius.secret"
                            :rules="this.ruleForm.auth.type== 'radius' ? this.rules['auth.radius.secret'] : [{ required: false }]">
                <el-input v-model="ruleForm.auth.radius.secret" placeholder=""></el-input>
              </el-form-item>
              <el-form-item label="Nasip" prop="auth.radius.nasip">
                <el-input v-model="ruleForm.auth.radius.nasip" placeholder=""></el-input>
              </el-form-item>
            </template>

            <template v-if="ruleForm.auth.type == 'ldap'">
              <el-form-item label="Server address" prop="auth.ldap.addr"
                            :rules="this.ruleForm.auth.type== 'ldap' ? this.rules['auth.ldap.addr'] : [{ required: false }]">
                <el-input v-model="ruleForm.auth.ldap.addr" placeholder="For example, ip:389 / domain name:389"></el-input>
              </el-form-item>
              <el-form-item label="Enable TLS" prop="auth.ldap.tls">
                <el-switch v-model="ruleForm.auth.ldap.tls"></el-switch>
              </el-form-item>
              <el-form-item label="Administrator DN" prop="auth.ldap.bind_name"
                            :rules="this.ruleForm.auth.type== 'ldap' ? this.rules['auth.ldap.bind_name'] : [{ required: false }]">
                <el-input v-model="ruleForm.auth.ldap.bind_name"
                          placeholder="For example: CN=bindadmin,DC=abc,DC=COM"></el-input>
              </el-form-item>
              <el-form-item label="Administrator password" prop="auth.ldap.bind_pwd"
                            :rules="this.ruleForm.auth.type== 'ldap' ? this.rules['auth.ldap.bind_pwd'] : [{ required: false }]">
                <el-input type="password" v-model="ruleForm.auth.ldap.bind_pwd" placeholder=""></el-input>
              </el-form-item>
              <el-form-item label="Base DN" prop="auth.ldap.base_dn"
                            :rules="this.ruleForm.auth.type== 'ldap' ? this.rules['auth.ldap.base_dn'] : [{ required: false }]">
                <el-input v-model="ruleForm.auth.ldap.base_dn" placeholder="例如 DC=abc,DC=com"></el-input>
              </el-form-item>
              <el-form-item label="User object class" prop="auth.ldap.object_class"
                            :rules="this.ruleForm.auth.type== 'ldap' ? this.rules['auth.ldap.object_class'] : [{ required: false }]">
                <el-input v-model="ruleForm.auth.ldap.object_class"
                          placeholder="For example: person / user / posixAccount"></el-input>
              </el-form-item>
              <el-form-item label="User unique ID" prop="auth.ldap.search_attr"
                            :rules="this.ruleForm.auth.type== 'ldap' ? this.rules['auth.ldap.search_attr'] : [{ required: false }]">
                <el-input v-model="ruleForm.auth.ldap.search_attr"
                          placeholder="For example: sAMAccountName / uid / cn"></el-input>
              </el-form-item>
              <el-form-item label="Restricted user groups" prop="auth.ldap.member_of">
                <el-input v-model="ruleForm.auth.ldap.member_of"
                          placeholder="Optional, only allow the specified group to log in, for example: CN=HomeWork,DC=abc,DC=com"></el-input>
              </el-form-item>
            </template>
          </el-tab-pane>

          <el-tab-pane label="Routing settings" name="route">
            <el-form-item label="Included routes" prop="route_include">
              <el-row class="msg-info">
                <el-col :span="18">Enter the CIDR format such as: 192.168.1.0/24</el-col>
                <el-col :span="2">
                  <el-button size="mini" type="success" icon="el-icon-plus" circle
                             @click.prevent="addDomain(ruleForm.route_include)"></el-button>
                </el-col>
                <el-col :span="4">
                  <el-button size="mini" type="info" icon="el-icon-edit" circle
                             @click.prevent="openIpListDialog('route_include')"></el-button>
                </el-col>
              </el-row>
              <templete v-if="activeTab == 'route'">
                <el-row v-for="(item,index) in ruleForm.route_include"
                        :key="index" style="margin-bottom: 5px" :gutter="10">
                  <el-col :span="10">
                    <el-input v-model="item.val"></el-input>
                  </el-col>
                  <el-col :span="12">
                    <el-input v-model="item.note" placeholder="Note"></el-input>
                  </el-col>
                  <el-col :span="2">
                    <el-button size="mini" type="danger" icon="el-icon-minus" circle
                               @click.prevent="removeDomain(ruleForm.route_include,index)"></el-button>
                  </el-col>
                </el-row>
              </templete>
            </el-form-item>

            <el-form-item label="Exclude routes" prop="route_exclude">
              <el-row class="msg-info">
                <el-col :span="18">Enter the CIDR format such as: 192.168.2.0/24</el-col>
                <el-col :span="2">
                  <el-button size="mini" type="success" icon="el-icon-plus" circle
                             @click.prevent="addDomain(ruleForm.route_exclude)"></el-button>
                </el-col>
                <el-col :span="4">
                  <el-button size="mini" type="info" icon="el-icon-edit" circle
                             @click.prevent="openIpListDialog('route_exclude')"></el-button>
                </el-col>
              </el-row>
              <templete v-if="activeTab == 'route'">
                <el-row v-for="(item,index) in ruleForm.route_exclude"
                        :key="index" style="margin-bottom: 5px" :gutter="10">
                  <el-col :span="10">
                    <el-input v-model="item.val"></el-input>
                  </el-col>
                  <el-col :span="12">
                    <el-input v-model="item.note" placeholder="Note"></el-input>
                  </el-col>
                  <el-col :span="2">
                    <el-button size="mini" type="danger" icon="el-icon-minus" circle
                               @click.prevent="removeDomain(ruleForm.route_exclude,index)"></el-button>
                  </el-col>
                </el-row>
              </templete>
            </el-form-item>
          </el-tab-pane>
          <el-tab-pane label="Permission control" name="link_acl">
            <el-form-item label="Permission control" prop="link_acl">
              <el-row class="msg-info">
                <el-col :span="22">Enter the CIDR format such as: 192.168.3.0/24
                  Protocol support all,tcp,udp,icmp
                  Port 0 means all ports, multiple ports: 80, 443, continuous ports: 1234-5678
                </el-col>
                <el-col :span="2">
                  <el-button size="mini" type="success" icon="el-icon-plus" circle
                             @click.prevent="addDomain(ruleForm.link_acl)"></el-button>
                </el-col>
              </el-row>

              <!--  Add drag functionality  -->
              <draggable v-model="ruleForm.link_acl" handle=".drag-handle" @end="onEnd">

              <el-row v-for="(item,index) in ruleForm.link_acl"
                      :key="index" style="margin-bottom: 5px" :gutter="1">

                <el-col :span="1" class="drag-handle">
                <i class="el-icon-rank"></i>
                </el-col>

                <el-col :span="9">
                  <el-input placeholder="Please enter a CIDR address" v-model="item.val">
                    <el-select v-model="item.action" slot="prepend">
                      <el-option label="Allow" value="allow"></el-option>
                      <el-option label="Deny" value="deny"></el-option>
                    </el-select>
                  </el-input>
                </el-col>

                <el-col :span="3">
                    <el-input placeholder="Protocol" v-model="item.protocol">
                </el-col>

                <el-col :span="6">
                  <!--  type="textarea" :autosize="{ minRows: 1, maxRows: 2}"  -->
                  <el-input v-model="item.port" placeholder="Multi-port, number separation"></el-input>
                </el-col>
                <el-col :span="3">
                  <el-input v-model="item.note" placeholder="Note"></el-input>
                </el-col>

                <el-col :span="2">
                  <el-button size="mini" type="danger" icon="el-icon-minus" circle
                             @click.prevent="removeDomain(ruleForm.link_acl,index)"></el-button>
                </el-col>
              </el-row>
              </draggable>

            </el-form-item>
          </el-tab-pane>

          <el-tab-pane label="Domain split tunneling" name="ds_domains">
            <el-form-item label="Include domain name" prop="ds_include_domains">
              <el-input type="textarea" :rows="5" v-model="ruleForm.ds_include_domains"
                        placeholder="Enter the domain name separated by , and by default it matches all subdomains, such as google.com, yahoo.com"></el-input>
            </el-form-item>
            <el-form-item label="Exclude domains" prop="ds_exclude_domains">
              <el-input type="textarea" :rows="5" v-model="ruleForm.ds_exclude_domains"
                        placeholder="Enter the domain name separated by , and by default it matches all subdomains, such as google.com, yahoo.com"></el-input>
              <div class="msg-info">Note: Domain split tunneling is only supported on AnyConnect desktop clients for Windows and MacOS, not on mobile clients.</div>
            </el-form-item>
          </el-tab-pane>
          <el-form-item>
            <templete v-if="activeTab == 'authtype' && ruleForm.auth.type != 'local'">
              <el-button @click="openAuthLoginDialog()" style="margin-right:10px">Test Login</el-button>
            </templete>
            <el-button type="primary" @click="submitForm('ruleForm')">Save</el-button>
            <el-button @click="closeDialog">Cancel</el-button>
          </el-form-item>
        </el-tabs>
      </el-form>
    </el-dialog>
    <!--Test user login pop-up box-->
    <el-dialog
        :close-on-click-modal="false"
        title="Test user login"
        :visible.sync="authLoginDialog"
        width="600px"
        custom-class="valgin-dialog"
        center>
      <el-form :model="authLoginForm" :rules="authLoginRules" ref="authLoginForm" label-width="100px">
        <el-form-item label="Account" prop="name">
          <el-input v-model="authLoginForm.name" ref="authLoginFormName"
                    @keydown.enter.native="testAuthLogin"></el-input>
        </el-form-item>
        <el-form-item label="Password" prop="pwd">
          <el-input type="password" v-model="authLoginForm.pwd" @keydown.enter.native="testAuthLogin"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="testAuthLogin()" :loading="authLoginLoading">Log in</el-button>
          <el-button @click="authLoginDialog = false">Cancel</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>
    <!--Edit mode popup-->
    <el-dialog
        :close-on-click-modal="false"
        title="Edit mode"
        :visible.sync="ipListDialog"
        width="650px"
        custom-class="valgin-dialog"
        center>
      <el-form ref="ipEditForm" label-width="80px">
        <el-form-item label="Routing table" prop="ip_list">
          <el-input type="textarea" :rows="10" v-model="ipEditForm.ip_list"
                    placeholder="One route per line, for example: 192.168.1.0/24, note or 192.168.1.0/24"></el-input>
          <div class="msg-info">Current total
            {{ ipEditForm.ip_list.trim() === '' ? 0 : ipEditForm.ip_list.trim().split("\n").length }}
            (Note: AnyConnect client supports a maximum of {{ this.maxRouteRows }} routes)
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="ipEdit()" :loading="ipEditLoading">Renew</el-button>
          <el-button @click="ipListDialog = false">Cancel</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>
  </div>
</template>

<script>
import axios from "axios";
import draggable from 'vuedraggable'

export default {
  name: "List",
  components: {draggable},
  mixins: [],
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['Groups', 'User groups'])
  },
  mounted() {
    this.getData(1);
    this.setAuthData();
  },
  data() {
    return {
      page: 1,
      tableData: [],
      count: 10,
      activeTab: "general",
      readMore: {},
      readMinRows: 5,
      maxRouteRows: 2500,
      defAuth: {
        type: 'local',
        radius: {addr: "", secret: "", nasip: ""},
        ldap: {
          addr: "",
          tls: false,
          base_dn: "",
          object_class: "person",
          search_attr: "sAMAccountName",
          member_of: "",
          bind_name: "",
          bind_pwd: "",
        },
      },
      ruleForm: {
        bandwidth: 0,
        bandwidth_format: '0',
        status: 1,
        allow_lan: true,
        client_dns: [{val: '114.114.114.114', note: 'Default DNS'}],
        split_dns: [],
        route_include: [{val: 'all', note: 'Default global proxy'}],
        route_exclude: [],
        link_acl: [],
        auth: {},
      },
      authLoginDialog: false,
      ipListDialog: false,
      authLoginLoading: false,
      authLoginForm: {
        name: "",
        pwd: "",
      },
      ipEditForm: {
        ip_list: "",
        type: "",
      },
      ipEditLoading: false,
      authLoginRules: {
        name: [
          {required: true, message: 'Please enter username', trigger: 'blur'},
        ],
        pwd: [
          {required: true, message: 'Please enter password', trigger: 'blur'},
          {min: 6, message: 'At least 6 characters in length', trigger: 'blur'}
        ],
      },
      rules: {
        name: [
          {required: true, message: 'Please enter a group name', trigger: 'blur'},
          {max: 30, message: 'Maximum 30 characters long', trigger: 'blur'}
        ],
        bandwidth_format: [
          {required: true, message: 'Please enter bandwidth limit', trigger: 'blur'},
          {type: 'string', message: 'Bandwidth limit must be a numeric value'}
        ],
        status: [
          {required: true}
        ],
        "auth.radius.addr": [
          {required: true, message: 'Please enter Radius server address', trigger: 'blur'}
        ],
        "auth.radius.secret": [
          {required: true, message: 'Please enter Radius server secret key', trigger: 'blur'}
        ],
        "auth.ldap.addr": [
          {required: true, message: 'Please enter the server address (including port)', trigger: 'blur'}
        ],
        "auth.ldap.bind_name": [
          {required: true, message: 'Please enter Bind DN', trigger: 'blur'}
        ],
        "auth.ldap.bind_pwd": [
          {required: true, message: 'Please enter Bind DN password', trigger: 'blur'}
        ],
        "auth.ldap.base_dn": [
          {required: true, message: 'Please enter Base DN', trigger: 'blur'}
        ],
        "auth.ldap.object_class": [
          {required: true, message: 'Please enter user object class', trigger: 'blur'}
        ],
        "auth.ldap.search_attr": [
          {required: true, message: 'Please enter user search attribute', trigger: 'blur'}
        ],
      },
    }
  },
  methods: {
    onEnd: function() {
       window.console.log("onEnd", this.ruleForm.link_acl);
    },
    setAuthData(row) {
      if (!row) {
        this.ruleForm.auth = JSON.parse(JSON.stringify(this.defAuth));
        return;
      }
      if (row.auth.type == "ldap" && !row.auth.ldap.object_class) {
        row.auth.ldap.object_class = this.defAuth.ldap.object_class;
      }
      this.ruleForm.auth = Object.assign(JSON.parse(JSON.stringify(this.defAuth)), row.auth);
    },
    handleDel(row) {
      axios.post('/group/del?id=' + row.id).then(resp => {
        const rdata = resp.data;
        if (rdata.code === 0) {
          this.$message.success(rdata.msg);
          this.getData(1);
        } else {
          this.$message.error(rdata.msg);
        }
        //console.log(rdata);
      }).catch(error => {
        this.$message.error('Oh, request error');
        console.log(error);
      });
    },
    handleEdit(row) {
      !this.$refs['ruleForm'] || this.$refs['ruleForm'].resetFields();
      //console.log(row)
      this.user_edit_dialog = true
      if (!row) {
        this.setAuthData(row)
        return;
      }
      axios.get('/group/detail', {
        params: {
          id: row.id,
        }
      }).then(resp => {
        resp.data.data.bandwidth_format = this.convertBandwidth(resp.data.data.bandwidth, 'BYTE', 'Mbps').toString();
        this.ruleForm = resp.data.data;
        this.setAuthData(resp.data.data);
      }).catch(error => {
        this.$message.error('Oh, request error');
        console.log(error);
      });
    },
    pageChange(p) {
      this.getData(p)
    },
    getData(page) {
      this.page = page
      axios.get('/group/list', {
        params: {
          page: page,
        }
      }).then(resp => {
        const rdata = resp.data.data;
        //console.log(rdata);
        this.tableData = rdata.datas;
        this.count = rdata.count
      }).catch(error => {
        this.$message.error('Oh, request error');
        console.log(error);
      });
    },
    removeDomain(arr, index) {
      console.log(index)
      if (index >= 0 && index < arr.length) {
        arr.splice(index, 1)
      }
      // let index = arr.indexOf(item);
      // if (index !== -1 && arr.length > 1) {
      //   arr.splice(index, 1)
      // }
      // arr.pop()
    },
    addDomain(arr) {
      //console.log("arr", arr)
      arr.push({protocol:"all", val: "", action: "allow", port: "0", note: ""});
    },
    submitForm(formName) {
      this.$refs[formName].validate((valid) => {
        if (!valid) {
          //console.log('error submit!!');
          return false;
        }
        this.ruleForm.bandwidth = this.convertBandwidth(this.ruleForm.bandwidth_format, 'Mbps', 'BYTE');
        axios.post('/group/set', this.ruleForm).then(resp => {
          const rdata = resp.data;
          if (rdata.code === 0) {
            this.$message.success(rdata.msg);
            this.getData(1);
            this.user_edit_dialog = false
          } else {
            this.$message.error(rdata.msg);
          }
          //console.log(rdata);
        }).catch(error => {
          this.$message.error('Oh, request error');
          console.log(error);
        });
      });
    },
    testAuthLogin() {
      this.$refs["authLoginForm"].validate((valid) => {
        if (!valid) {
          //console.log('error submit!!');
          return false;
        }
        this.authLoginLoading = true;
        axios.post('/group/auth_login', {
          name: this.authLoginForm.name,
          pwd: this.authLoginForm.pwd,
          auth: this.ruleForm.auth
        }).then(resp => {
          const rdata = resp.data;
          if (rdata.code === 0) {
            this.$message.success("Login successful");
          } else {
            this.$message.error(rdata.msg);
          }
          this.authLoginLoading = false;
          //console.log(rdata);
        }).catch(error => {
          this.$message.error('Oh, request error');
          console.log(error);
          this.authLoginLoading = false;
        });
      });
    },
    openAuthLoginDialog() {
      this.$refs["ruleForm"].validate((valid) => {
        if (!valid) {
          //console.log('error submit!!');
          return false;
        }
        this.authLoginDialog = true;
        // set authLoginFormName focus
        this.$nextTick(() => {
          this.$refs['authLoginFormName'].focus();
        });
      });
    },
    openIpListDialog(type) {
      this.ipListDialog = true;
      this.ipEditForm.type = type;
      this.ipEditForm.ip_list = this.ruleForm[type].map(item => item.val + (item.note ? "," + item.note : "")).join("\n");
    },
    ipEdit() {
      this.ipEditLoading = true;
      let ipList = [];
      if (this.ipEditForm.ip_list.trim() !== "") {
        ipList = this.ipEditForm.ip_list.trim().split("\n");
      }
      let arr = [];
      for (let i = 0; i < ipList.length; i++) {
        let item = ipList[i];
        if (item.trim() === "") {
          continue;
        }
        let ip = item.split(",");
        if (ip.length > 2) {
          ip[1] = ip.slice(1).join(",");
        }
        let note = ip[1] ? ip[1] : "";
        const pushToArr = () => {
          arr.push({val: ip[0], note: note});
        };
        if (this.ipEditForm.type == "route_include" && ip[0] == "all") {
          pushToArr();
          continue;
        }
        let valid = this.isValidCIDR(ip[0]);
        if (!valid.valid) {
          this.$message.error("Error: CIDR format is incorrect, it is recommended to change " + ip[0] + " to " + valid.suggestion);
          this.ipEditLoading = false;
          return;
        }
        pushToArr();
      }
      this.ruleForm[this.ipEditForm.type] = arr;
      this.ipEditLoading = false;
      this.ipListDialog = false;
    },
    isValidCIDR(input) {
      const cidrRegex = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)\/([12]?\d|3[0-2])$/;
      if (!cidrRegex.test(input)) {
        return {valid: false, suggestion: null};
      }
      const [ip, mask] = input.split('/');
      const maskNum = parseInt(mask);
      const ipParts = ip.split('.').map(part => parseInt(part));
      const binaryIP = ipParts.map(part => part.toString(2).padStart(8, '0')).join('');
      for (let i = maskNum; i < 32; i++) {
        if (binaryIP[i] === '1') {
          const binaryNetworkPart = binaryIP.substring(0, maskNum).padEnd(32, '0');
          const networkIPParts = [];
          for (let j = 0; j < 4; j++) {
            const octet = binaryNetworkPart.substring(j * 8, (j + 1) * 8);
            networkIPParts.push(parseInt(octet, 2));
          }
          const suggestedIP = networkIPParts.join('.');
          return {valid: false, suggestion: `${suggestedIP}/${mask}`};
        }
      }
      return {valid: true, suggestion: null};
    },
    resetForm(formName) {
      this.$refs[formName].resetFields();
    },
    toggleMore(id) {
      if (this.readMore[id]) {
        this.$set(this.readMore, id, false);
      } else {
        this.$set(this.readMore, id, true);
      }
    },
    authTypeChange() {
      this.$refs['ruleForm'].clearValidate();
    },
    beforeTabLeave() {
      var isSwitch = true
      if (!this.user_edit_dialog) {
        return isSwitch;
      }
      this.$refs['ruleForm'].validate((valid) => {
        if (!valid) {
          this.$message.error("Error: You have missed a required field.")
          isSwitch = false;
          return false;
        }
      });
      return isSwitch;
    },
    closeDialog() {
      this.user_edit_dialog = false;
      this.activeTab = "general";
    },
    convertBandwidth(bandwidth, fromUnit, toUnit) {
      const units = {
        bps: 1,
        Kbps: 1000,
        Mbps: 1000000,
        Gbps: 1000000000,
        BYTE: 8,
      };
      const result = bandwidth * units[fromUnit] / units[toUnit];
      const fixedResult = result.toFixed(2);
      return parseFloat(fixedResult);
    }
  },
}
</script>

<style scoped>
.msg-info {
  background-color: #f4f4f5;
  color: #909399;
  padding: 0 5px;
  margin: 0;
  box-sizing: border-box;
  border-radius: 4px;
  font-size: 12px;
}

.el-select {
  width: 80px;
}

::v-deep .valgin-dialog {
  display: flex;
  flex-direction: column;
  margin: 0 !important;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  max-height: calc(100% - 30px);
  max-width: calc(100% - 30px);
}

::v-deep .valgin-dialog .el-dialog__body {
  flex: 1;
  overflow: auto;
}

.drag-handle {
  cursor: move;
}

</style>