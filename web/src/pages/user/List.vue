<template>
  <div>
    <el-card>
      <el-form :inline="true">
        <el-form-item>
          <el-button size="small" type="primary" icon="el-icon-plus" @click="handleEdit('')">Add
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-dropdown size="small" placement="bottom">
            <el-upload class="uploaduser" action="uploaduser" accept=".xlsx, .xls" :http-request="upLoadUser" :limit="1"
              :show-file-list="false">
              <el-button size="small" icon="el-icon-upload2" type="primary">Batch add</el-button>
            </el-upload>
            <el-dropdown-menu slot="dropdown">
              <el-dropdown-item>
                <el-link style="font-size:12px;" type="success" href="Adding_user_templates_in_batches.xlsx"><i
                    class="el-icon-download"></i>Download template
                </el-link>
              </el-dropdown-item>
            </el-dropdown-menu>
          </el-dropdown>
        </el-form-item>
        <el-form-item label="Username or Name or Email:">
          <el-input size="small" v-model="searchData" placeholder="Please enter content"
            @keydown.enter.native="searchEnterFun"></el-input>
        </el-form-item>

        <el-form-item>
          <el-button size="small" type="primary" icon="el-icon-search" @click="handleSearch()">Search
          </el-button>
          <el-button size="small" icon="el-icon-refresh" @click="reset">Reset search
          </el-button>
        </el-form-item>
      </el-form>

      <el-table ref="multipleTable" :data="tableData" border>

        <el-table-column sortable="true" prop="id" label="ID" width="60">
        </el-table-column>

        <el-table-column prop="username" label="Username" width="150">
        </el-table-column>

        <el-table-column prop="nickname" label="Name" width="100">
        </el-table-column>

        <el-table-column prop="email" label="Email">
        </el-table-column>
        <el-table-column prop="otp_secret" label="OTP key" width="110">
          <template slot-scope="scope">
            <el-button v-if="!scope.row.disable_otp" type="text" icon="el-icon-view" @click="getOtpImg(scope.row)">
              {{ scope.row.otp_secret.substring(0, 6) }}
            </el-button>
          </template>
        </el-table-column>

        <el-table-column prop="groups" label="User groups">
          <template slot-scope="scope">
            <el-row v-for="item in scope.row.groups" :key="item">{{ item }}</el-row>
          </template>
        </el-table-column>

        <el-table-column prop="status" label="State" width="70">
          <template slot-scope="scope">
            <el-tag v-if="scope.row.status === 1" type="success">Available</el-tag>
            <el-tag v-if="scope.row.status === 0" type="danger">Deactivate</el-tag>
            <el-tag v-if="scope.row.status === 2">Expired</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="updated_at" label="Update time" :formatter="tableDateFormat">
        </el-table-column>

        <el-table-column label="Operate" width="210">
          <template slot-scope="scope">
            <el-button size="mini" type="primary" @click="handleEdit(scope.row)">Edit
            </el-button>

            <!--            <el-popconfirm
                            class="m-left-10"
                            @onConfirm="handleClick('reset',scope.row)"
                            title="Are you sure you want to reset your user password and key?">
                          <el-button
                              slot="reference"
                              size="mini"
                              type="warning">Reset
                          </el-button>
                        </el-popconfirm>-->

            <el-popconfirm class="m-left-10" @confirm="handleDel(scope.row)" title="Are you sure you want to delete user?">
              <el-button slot="reference" size="mini" type="danger">Delete
              </el-button>
            </el-popconfirm>

          </template>
        </el-table-column>
      </el-table>

      <div class="sh-20"></div>

      <el-pagination background layout="prev, pager, next" :pager-count="11" @current-change="pageChange"
        :current-page="page" :total="count">
      </el-pagination>

    </el-card>

    <el-dialog title="OTP key" :visible.sync="otpImgData.visible" width="350px" center>
      <div style="text-align: center">{{ otpImgData.username }} : {{ otpImgData.nickname }}</div>
      <img :src="otpImgData.base64Img" alt="otp-img" />
    </el-dialog>

    <!--Add or modify pop-up boxes-->
    <el-dialog :close-on-click-modal="false" title="User" :visible="user_edit_dialog" @close="disVisible" width="650px"
      center>

      <el-form :model="ruleForm" :rules="rules" ref="ruleForm" label-width="100px" class="ruleForm">
        <el-form-item label="User ID" prop="id">
          <el-input v-model="ruleForm.id" disabled></el-input>
        </el-form-item>
        <el-form-item label="Username" prop="username">
          <el-input v-model="ruleForm.username" :disabled="ruleForm.id > 0"></el-input>
        </el-form-item>
        <el-form-item label="Name" prop="nickname">
          <el-input v-model="ruleForm.nickname"></el-input>
        </el-form-item>
        <el-form-item label="Email" prop="email">
          <el-input v-model="ruleForm.email"></el-input>
        </el-form-item>

        <el-form-item label="PIN" prop="pin_code">
          <el-input v-model="ruleForm.pin_code" placeholder="Leave blank and the system will automatically generate"></el-input>
        </el-form-item>

        <el-form-item label="Expiration time" prop="limittime">
          <el-date-picker v-model="ruleForm.limittime" type="date" size="small" align="center" style="width:130px"
            :picker-options="pickerOptions" placeholder="Select date">
          </el-date-picker>
        </el-form-item>

        <el-form-item label="Disable OTP" prop="disable_otp">
          <el-switch v-model="ruleForm.disable_otp" active-text="After OTP is enabled, the user password is the PIN code, and the OTP password is the dynamic code generated after scanning the code.">
          </el-switch>
        </el-form-item>

        <el-form-item label="OTP key" prop="otp_secret" v-if="!ruleForm.disable_otp">
          <el-input v-model="ruleForm.otp_secret" placeholder="Leave blank and the system will automatically generate"></el-input>
        </el-form-item>

        <el-form-item label="User groups" prop="groups">
          <el-checkbox-group v-model="ruleForm.groups">
            <el-checkbox v-for="(item) in grouNames" :key="item" :label="item" :name="item"></el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <el-form-item label="Send email" prop="send_email">
          <el-switch v-model="ruleForm.send_email">
          </el-switch>
        </el-form-item>

        <el-form-item label="State" prop="status">
          <el-radio-group v-model="ruleForm.status">
            <el-radio :label="1" border>Enable</el-radio>
            <el-radio :label="0" border>Disable</el-radio>
            <el-radio :label="2" border>Expired</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="submitForm('ruleForm')">Save</el-button>
          <!--          <el-button @click="resetForm('ruleForm')">Reset</el-button>-->
          <el-button @click="disVisible">取消</el-button>
        </el-form-item>
      </el-form>

    </el-dialog>

  </div>
</template>

<script>
import axios from "axios";

export default {
  name: "List",
  components: {},
  mixins: [],
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['User information', 'User list'])
  },
  mounted() {
    this.getGroups();
    this.getData(1)
  },

  data() {
    return {
      page: 1,
      grouNames: [],
      tableData: [],
      count: 10,
      pickerOptions: {
        disabledDate(time) {
          return time.getTime() < Date.now();
        }
      },
      searchData: '',
      otpImgData: { visible: false, username: '', nickname: '', base64Img: '' },
      ruleForm: {
        send_email: true,
        status: 1,
        groups: [],
      },
      rules: {
        username: [
          { required: true, message: 'Please enter your username', trigger: 'blur' },
          { max: 50, message: 'Less than 50 characters in length', trigger: 'blur' }
        ],
        nickname: [
          { required: true, message: 'Please enter your user name', trigger: 'blur' }
        ],
        email: [
          { required: true, message: 'Please enter the user's email address', trigger: 'blur' },
          { type: 'email', message: 'Please enter a valid email address', trigger: ['blur', 'change'] }
        ],
        password: [
          { min: 6, message: 'Greater than 6 characters', trigger: 'blur' }
        ],
        pin_code: [
          { min: 6, message: 'PIN is longer than 6 characters', trigger: 'blur' }
        ],
        date1: [
          { type: 'date', required: true, message: 'Please select date', trigger: 'change' }
        ],
        groups: [
          { type: 'array', required: true, message: 'Please select at least one group', trigger: 'change' }
        ],
        status: [
          { required: true }
        ],
      },
    }
  },

  methods: {
    upLoadUser(item) {
      const formData = new FormData();
      formData.append("file", item.file);
      axios.post('/user/uploaduser', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      }).then(resp => {
        if (resp.data.code === 0) {
          this.$message.success(resp.data.data);
          this.getData(1);
        } else {
          this.$message.error(resp.data.msg);
          this.getData(1);
        }
        console.log(resp.data);
      })
    },
    getOtpImg(row) {
      // this.base64Img = Buffer.from(data).toString('base64');
      this.otpImgData.visible = true
      axios.get('/user/otp_qr', {
        params: {
          id: row.id,
          b64: '1',
        }
      }).then(resp => {
        var rdata = resp.data;
        // console.log(resp);
        this.otpImgData.username = row.username;
        this.otpImgData.nickname = row.nickname;
        this.otpImgData.base64Img = 'data:image/png;base64,' + rdata
      }).catch(error => {
        this.$message.error('Oops, request error');
        console.log(error);
      });
    },
    handleDel(row) {
      axios.post('/user/del?id=' + row.id).then(resp => {
        var rdata = resp.data
        if (rdata.code === 0) {
          this.$message.success(rdata.msg);
          this.getData(1);
        } else {
          this.$message.error(rdata.msg);
        }
        console.log(rdata);
      }).catch(error => {
        this.$message.error('Oops, request error');
        console.log(error);
      });
    },
    handleEdit(row) {
      !this.$refs['ruleForm'] || this.$refs['ruleForm'].resetFields();
      console.log(row)
      this.user_edit_dialog = true
      if (!row) {
        return;
      }

      axios.get('/user/detail', {
        params: {
          id: row.id,
        }
      }).then(resp => {
        var data = resp.data.data
        // Modify the default not to send emails
        data.send_email = false
        this.ruleForm = data
      }).catch(error => {
        this.$message.error('Oops, request error');
        console.log(error);
      });
    },
    handleSearch() {
      console.log(this.searchData)
      this.getData(1, this.searchData)
    },
    pageChange(p) {
      this.getData(p)
    },
    getData(page, prefix) {
      this.page = page
      axios.get('/user/list', {
        params: {
          page: page,
          prefix: prefix || '',
        }
      }).then(resp => {
        var data = resp.data.data
        console.log(data);
        this.tableData = data.datas;
        this.count = data.count
      }).catch(error => {
        this.$message.error('Oops, request error');
        console.log(error);
      });
    },
    getGroups() {
      axios.get('/group/names', {}).then(resp => {
        var data = resp.data.data
        console.log(data.datas);
        this.grouNames = data.datas;
      }).catch(error => {
        this.$message.error('Oops, request error');
        console.log(error);
      });
    },
    submitForm(formName) {
      this.$refs[formName].validate((valid) => {
        if (!valid) {
          console.log('error submit!!');
          return false;
        }

        // alert('submit!');
        axios.post('/user/set', this.ruleForm).then(resp => {
          var data = resp.data
          if (data.code === 0) {
            this.$message.success(data.msg);
            this.getData(1);
            this.user_edit_dialog = false
          } else {
            this.$message.error(data.msg);
          }
          console.log(data);
        }).catch(error => {
          this.$message.error('Oops, request error');
          console.log(error);
        });
      });
    },
    resetForm(formName) {
      this.$refs[formName].resetFields();
    },
    searchEnterFun(e) {
      var keyCode = window.event ? e.keyCode : e.which;
      if (keyCode == 13) {
        this.handleSearch()
      }
    },
    reset() {
      this.searchData = "";
      this.handleSearch();
    },
  },
}
</script>

<style scoped></style>