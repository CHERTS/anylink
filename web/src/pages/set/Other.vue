<template>
  <el-card>
    <el-tabs v-model="activeName" @tab-click="handleClick">
      <el-tab-pane label="Email" name="dataSmtp">
        <el-form
            :model="dataSmtp"
            ref="dataSmtp"
            :rules="rules"
            label-width="100px"
            class="tab-one">
          <el-form-item label="Server" prop="host">
            <el-input v-model="dataSmtp.host"></el-input>
          </el-form-item>
          <el-form-item label="Port" prop="port">
            <el-input v-model.number="dataSmtp.port"></el-input>
          </el-form-item>
          <el-form-item label="Username" prop="username">
            <el-input v-model="dataSmtp.username"></el-input>
          </el-form-item>
          <el-form-item label="Password" prop="password">
            <el-input
                type="password"
                v-model="dataSmtp.password"
                placeholder="If the password is empty, it will not be modified."
            ></el-input>
          </el-form-item>
          <el-form-item label="Encryption" prop="encryption">
            <el-radio-group v-model="dataSmtp.encryption">
              <el-radio label="None">None</el-radio>
              <el-radio label="SSLTLS">SSLTLS</el-radio>
              <el-radio label="STARTTLS">STARTTLS</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="From" prop="from">
            <el-input v-model="dataSmtp.from"></el-input>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="submitForm('dataSmtp')">Save</el-button>
            <el-button @click="resetForm('dataSmtp')">Reset</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="Audit log" name="dataAuditLog">
        <el-form
            :model="dataAuditLog"
            ref="dataAuditLog"
            :rules="rules"
            label-width="100px"
            class="tab-one">
          <el-form-item label="Audit interval" prop="audit_interval">
            <el-input-number
                v-model="dataAuditLog.audit_interval"
                :min="-1"
                size="small"
                label="Second"
                :disabled="true"
            ></el-input-number>
            Second
            <p class="input_tip">
              Please manually modify the audit_interval parameter in the configuration file before restarting the service.
              <strong style="color: #ea3323">-1 means turning off the audit log</strong>
            </p>
          </el-form-item>
          <el-form-item label="Storage duration" prop="life_day">
            <el-input-number
                v-model="dataAuditLog.life_day"
                :min="0"
                :max="365"
                size="small"
                label="Days"
            ></el-input-number>
            Days
            <p class="input_tip">
              Range: 0 ~ 365 days
              <strong style="color: #ea3323">0 means a permanent storage</strong>
            </p>
          </el-form-item>
          <el-form-item label="Cleanup time" prop="clear_time">
            <el-time-select
                v-model="dataAuditLog.clear_time"
                :picker-options="{
                start: '00:00',
                step: '01:00',
                end: '23:00',
              }"
                :editable="false"
                size="small"
                placeholder="Please choose"
                style="width: 130px"
            >
            </el-time-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="submitForm('dataAuditLog')">Save</el-button>
            <el-button @click="resetForm('dataAuditLog')">Reset</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
      <el-tab-pane label="Certificate" name="datacertManage">
        <el-tabs
            tab-position="left"
            v-model="datacertManage"
            @tab-click="handleClick">
          <el-tab-pane label="Custom certificate" name="customCert">
            <el-form
                ref="customCert"
                :model="customCert"
                label-width="100px"
                size="small"
                class="tab-one">
              <el-form-item>
                <el-upload
                    class="uploadCert"
                    :before-upload="beforeCertUpload"
                    :action="certUpload"
                    :limit="1">
                  <el-button size="mini" icon="el-icon-plus" slot="trigger">Certificate file</el-button>
                  <el-tooltip
                    class="item"
                    effect="dark"
                    content="Please upload the cert file in .pem format"
                    placement="top">
                    <i class="el-icon-info"></i>
                  </el-tooltip>
                </el-upload>
              </el-form-item>
              <el-form-item>
                <el-upload
                  class="uploadCert"
                  :before-upload="beforeKeyUpload"
                  :action="certUpload"
                  :limit="1">
                  <el-button size="mini" icon="el-icon-plus" slot="trigger">Private key file</el-button>
                  <el-tooltip
                    class="item"
                    effect="dark"
                    content="Please upload the key file in .pem format"
                    placement="top">
                    <i class="el-icon-info"></i>
                  </el-tooltip>
                </el-upload>
              </el-form-item>
              <el-form-item>
                <el-button
                  size="small"
                  icon="el-icon-upload"
                  type="primary"
                  @click="submitForm('customCert')">Save</el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>
          <el-tab-pane label="Let's Encrypt certificate" name="letsCert">
            <el-form
              :model="letsCert"
              ref="letsCert"
              :rules="rules"
              label-width="120px"
              size="small"
              class="tab-one">
              <el-form-item label="Domain name" prop="domain">
                <el-input v-model="letsCert.domain"></el-input>
              </el-form-item>
              <el-form-item label="Email" prop="legomail">
                <el-input v-model="letsCert.legomail"></el-input>
              </el-form-item>
              <el-form-item label="Service provider" prop="name">
                <el-radio-group v-model="letsCert.name">
                  <el-radio label="aliyun">Ali Cloud</el-radio>
                  <el-radio label="txcloud">Tencent Cloud</el-radio>
                  <el-radio label="cfcloud">Cloudflare</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item
                v-for="component in dnsProvider[letsCert.name]"
                :key="component.prop"
                :label="component.label"
                :rules="component.rules">
                <component
                  :is="component.component"
                  :type="component.type"
                  v-model="letsCert[letsCert.name][component.prop]"></component>
              </el-form-item>
              <el-form-item>
                <el-switch
                  style="display: block"
                  v-model="letsCert.renew"
                  active-color="#13ce66"
                  inactive-color="#ff4949"
                  inactive-text="Automatic renewal">
                </el-switch>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="submitForm('letsCert')">Save</el-button>
                <el-button @click="resetForm('letsCert')">Reset</el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </el-tab-pane>
      <el-tab-pane label="Other" name="dataOther">
        <el-form
          :model="dataOther"
          ref="dataOther"
          :rules="rules"
          label-width="100px"
          class="tab-one">
          <el-form-item label="VPN address" prop="link_addr">
            <el-input placeholder="Please enter content" v-model="dataOther.link_addr">
            </el-input>
          </el-form-item>

          <el-form-item label="Banner info" prop="banner">
            <el-input
              type="textarea"
              :rows="5"
              placeholder="Please enter content"
              v-model="dataOther.banner">
            </el-input>
          </el-form-item>

          <el-form-item label="Custom homepage status code" prop="homecode">
            <el-input-number
                v-model="dataOther.homecode"
                :min="0"
                :max="1000"
            ></el-input-number>
          </el-form-item>

          <el-form-item label="Customize home page" prop="homeindex">
            <el-input
                type="textarea"
                :rows="10"
                placeholder="Please enter content"
                v-model="dataOther.homeindex"
            >
            </el-input>
            <el-tooltip content="For customized content, please refer to the files in the index_template directory." placement="top">
              <i class="el-icon-question"></i>
            </el-tooltip>
          </el-form-item>

          <el-form-item label="Account opening email template" prop="account_mail">
            <el-input
                type="textarea"
                :rows="10"
                placeholder="Please enter content"
                v-model="dataOther.account_mail"
            >
            </el-input>
          </el-form-item>

          <el-form-item label="Email">
            <iframe
                width="500px"
                height="300px"
                :srcdoc="dataOther.account_mail"
            >
            </iframe>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="submitForm('dataOther')">Save</el-button>
            <el-button @click="resetForm('dataOther')">Reset</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script>
import axios from "axios";

export default {
  name: "Other",
  created() {
    this.$emit("update:route_path", this.$route.path);
    this.$emit("update:route_name", ["Basic", "Other"]);
  },
  mounted() {
    this.getSmtp();
  },
  data() {
    return {
      activeName: "dataSmtp",
      datacertManage: "customCert",
      dataSmtp: {},
      dataAuditLog: {},
      letsCert: {
        domain: ``,
        legomail: ``,
        name: "",
        renew: "",
        aliyun: {
          apiKey: "",
          secretKey: "",
        },
        txcloud: {
          secretId: "",
          secretKey: "",
        },
        cfcloud: {
          authToken: "",
        },
      },
      customCert: {cert: "", key: ""},
      dataOther: {},
      rules: {
        host: {required: true, message: "Please enter server address", trigger: "blur"},
        port: [
          {required: true, message: "Please enter server port", trigger: "blur"},
          {
            type: "number",
            message: "Please enter correct server port",
            trigger: ["blur", "change"],
          },
        ],
        issuer: {required: true, message: "Please enter system name", trigger: "blur"},
        domain: {
          required: true,
          message: "Please enter the domain name for which you need to apply for a certificate",
          trigger: "blur",
        },
        legomail: {
          required: true,
          message: "Please enter the email address for applying for the certificate",
          trigger: "blur",
        },
        name: {required: true, message: "Please select a domain name service provider", trigger: "blur"},
      },
      certUpload: "/set/other/customcert",
      dnsProvider: {
        aliyun: [
          {
            label: "APIKey",
            prop: "apiKey",
            component: "el-input",
            type: "password",
            rules: {
              required: true,
              message: "Please enter the correct API key",
              trigger: "blur",
            },
          },
          {
            label: "SecretKey",
            prop: "secretKey",
            component: "el-input",
            type: "password",
            rules: {
              required: true,
              message: "Please enter the correct Secret key",
              trigger: "blur",
            },
          },
        ],
        txcloud: [
          {
            label: "SecretID",
            prop: "secretId",
            component: "el-input",
            type: "password",
            rules: {
              required: true,
              message: "Please enter the correct API key",
              trigger: "blur",
            },
          },
          {
            label: "SecretKey",
            prop: "secretKey",
            component: "el-input",
            type: "password",
            rules: {
              required: true,
              message: "Please enter the correct Secret key",
              trigger: "blur",
            },
          },
        ],
        cfcloud: [
          {
            label: "AuthToken",
            prop: "authToken",
            component: "el-input",
            type: "password",
            rules: {
              required: true,
              message: "Please enter the correct Auth token",
              trigger: "blur",
            },
          },
        ],
      },
    };
  },
  methods: {
    handleClick(tab, event) {
      window.console.log(tab.name, event);
      switch (tab.name) {
        case "dataSmtp":
          this.getSmtp();
          break;
        case "dataAuditLog":
          this.getAuditLog();
          break;
        case "letsCert":
          this.getletsCert();
          break;
        case "dataOther":
          this.getOther();
          break;
      }
    },
    beforeCertUpload(file) {
      // if (file.type !== 'application/x-pem-file') {
      //   this.$message.error('Only certificate files in .pem format can be uploaded')
      //   return false
      // }
      this.customCert.cert = file;
    },
    beforeKeyUpload(file) {
      // if (file.type !== 'application/x-pem-file') {
      //   this.$message.error('Only private key files in .pem format can be uploaded')
      //   return false
      // }
      this.customCert.key = file;
    },
    getSmtp() {
      axios
        .get("/set/other/smtp")
        .then((resp) => {
          let rdata = resp.data;
          //console.log(rdata);
          if (rdata.code !== 0) {
            this.$message.error(rdata.msg);
            return;
          }
          this.dataSmtp = rdata.data;
        })
        .catch((error) => {
          this.$message.error("Oh, request error");
          console.log(error);
        });
    },
    getAuditLog() {
      axios
        .get("/set/other/audit_log")
        .then((resp) => {
          let rdata = resp.data;
          //console.log(rdata);
          if (rdata.code !== 0) {
            this.$message.error(rdata.msg);
            return;
          }
          this.dataAuditLog = rdata.data;
        })
        .catch((error) => {
          this.$message.error("Oh, request error");
          console.log(error);
        });
    },
    getletsCert() {
      axios
        .get("/set/other/getcertset")
        .then((resp) => {
          let rdata = resp.data;
          //console.log(rdata);
          if (rdata.code !== 0) {
            this.$message.error(rdata.msg);
            return;
          }
          this.letsCert = Object.assign({}, this.letsCert, rdata.data);
        })
        .catch((error) => {
          this.$message.error("Oh, request error");
          console.log(error);
        });
    },
    getOther() {
      axios
        .get("/set/other")
        .then((resp) => {
          let rdata = resp.data;
          //console.log(rdata);
          if (rdata.code !== 0) {
            this.$message.error(rdata.msg);
            return;
          }
          this.dataOther = rdata.data;
        })
        .catch((error) => {
          this.$message.error("Oh, request error");
          console.log(error);
        });
    },
    submitForm(formName) {
      this.$refs[formName].validate((valid) => {
        if (!valid) {
          alert("error submit!");
        }

        switch (formName) {
          case "dataSmtp":
            axios.post("/set/other/smtp/edit", this.dataSmtp).then((resp) => {
              var rdata = resp.data;
              //console.log(rdata);
              if (rdata.code === 0) {
                this.$message.success(rdata.msg);
              } else {
                this.$message.error(rdata.msg);
              }
            });
            break;
          case "dataAuditLog":
            axios
              .post("/set/other/audit_log/edit", this.dataAuditLog)
              .then((resp) => {
                var rdata = resp.data;
                //console.log(rdata);
                if (rdata.code === 0) {
                  this.$message.success(rdata.msg);
                } else {
                  this.$message.error(rdata.msg);
                }
              });
            break;
          case "letsCert":
            var loading = this.$loading({
              lock: true,
              text: "Certificate application in progress...",
              spinner: "el-icon-loading",
              background: "rgba(0, 0, 0, 0.7)",
            });
            axios.post("/set/other/createcert", this.letsCert).then((resp) => {
              var rdata = resp.data;
              //console.log(rdata);
              if (rdata.code === 0) {
                loading.close();
                this.$message.success(rdata.msg);
              } else {
                loading.close();
                this.$message.error(rdata.msg);
              }
            });
            break;
          case "customCert":
            var formData = new FormData();
            formData.append("cert", this.customCert.cert);
            formData.append("key", this.customCert.key);
            axios.post(this.certUpload, formData).then((resp) => {
              var rdata = resp.data;
              //console.log(rdata);
              if (rdata.code === 0) {
                this.$message.success(rdata.msg);
              } else {
                this.$message.error(rdata.msg);
              }
            });
            break;
          case "dataOther":
            axios.post("/set/other/edit", this.dataOther).then((resp) => {
              var rdata = resp.data;
              //console.log(rdata);
              if (rdata.code === 0) {
                this.$message.success(rdata.msg);
              } else {
                this.$message.error(rdata.msg);
              }
            });
            break;
        }
      });
    },
    resetForm(formName) {
      this.$refs[formName].resetFields();
    },
  },
};
</script>

<style scoped>
.tab-one {
  width: 700px;
}

.input_tip {
  line-height: 1.428;
  margin: 2px 0 0 0;
}
</style>
