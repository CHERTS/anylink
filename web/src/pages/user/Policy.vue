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
            prop="username"
            label="Username">
        </el-table-column>
        <el-table-column
            prop="allow_lan"
            label="Local network">
          <template slot-scope="scope">
            <el-switch
                v-model="scope.row.allow_lan"
                disabled>
            </el-switch>
          </template>
        </el-table-column>

        <el-table-column
            prop="client_dns"
            label="Client DNS"
            width="160">
          <template slot-scope="scope">
            <el-row v-for="(item,inx) in scope.row.client_dns" :key="inx">{{ item.val }}</el-row>
          </template>
        </el-table-column>

        <el-table-column
            prop="route_include"
            label="Include route"
            width="200">
          <template slot-scope="scope">
            <el-row v-for="(item,inx) in scope.row.route_include.slice(0, readMinRows)" :key="inx">{{ item.val }}</el-row>
            <div v-if="scope.row.route_include.length > readMinRows">
              <div v-if="readMore[`ri_${ scope.row.id }`]">
                <el-row v-for="(item,inx) in scope.row.route_include.slice(readMinRows)" :key="inx">{{ item.val }}</el-row>              
              </div>
              <el-button size="mini" type="text" @click="toggleMore(`ri_${ scope.row.id }`)">{{ readMore[`ri_${ scope.row.id }`] ? "▲ Collapse" : "▼ More" }}</el-button>              
            </div>            
          </template>
        </el-table-column>

        <el-table-column
            prop="route_exclude"
            label="Exclude route"
            width="200">
          <template slot-scope="scope">
            <el-row v-for="(item,inx) in scope.row.route_exclude.slice(0, readMinRows)" :key="inx">{{ item.val }}</el-row>
            <div v-if="scope.row.route_exclude.length > readMinRows">
              <div v-if="readMore[`re_${ scope.row.id }`]">
                <el-row v-for="(item,inx) in scope.row.route_exclude.slice(readMinRows)" :key="inx">{{ item.val }}</el-row>              
              </div>
              <el-button size="mini" type="text" @click="toggleMore(`re_${ scope.row.id }`)">{{ readMore[`re_${ scope.row.id }`] ? "▲ Collapse" : "▼ More" }}</el-button>              
            </div>
          </template>
        </el-table-column>
        <el-table-column
            prop="status"
            label="Status"
            width="70">
          <template slot-scope="scope">
            <el-tag v-if="scope.row.status === 1" type="success">Enable</el-tag>
            <el-tag v-else type="danger">Disable</el-tag>
          </template>

        </el-table-column>

        <el-table-column
            prop="updated_at"
            label="Update time"
            :formatter="tableDateFormat">
        </el-table-column>

        <el-table-column
            label="Actions"
            width="150">
          <template slot-scope="scope">
            <el-button
                size="mini"
                type="primary"
                @click="handleEdit(scope.row)">Edit
            </el-button>

            <el-popconfirm
                style="margin-left: 10px"
                @confirm="handleDel(scope.row)"
                title="Are you sure you want to delete the user policy?">
              <el-button
                  slot="reference"
                  size="mini"
                  type="danger">Delete
              </el-button>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

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
        title="Adding policy"
        :visible.sync="user_edit_dialog"
        width="750px"
        top="50px"
        @close='closeDialog'
        center>

      <el-form :model="ruleForm" :rules="rules" ref="ruleForm" label-width="100px" class="ruleForm">
        <el-tabs v-model="activeTab" :before-leave="beforeTabLeave">
           <el-tab-pane label="General" name="general">
                <el-form-item label="ID" prop="id">
                <el-input v-model="ruleForm.id" disabled></el-input>
                </el-form-item>

                <el-form-item label="Username" prop="username">
                <el-input v-model="ruleForm.username" :disabled="ruleForm.id > 0"></el-input>
                </el-form-item>

                <el-form-item label="Local network" prop="allow_lan">
                  <el-switch
                      v-model="ruleForm.allow_lan">
                  </el-switch>
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
                        <el-input v-model="item.note" placeholder="Description"></el-input>
                        </el-col>
                        <el-col :span="2">
                        <el-button size="mini" type="danger" icon="el-icon-minus" circle
                                    @click.prevent="removeDomain(ruleForm.client_dns,index)"></el-button>
                        </el-col>
                    </el-row>
                </el-form-item>
                <el-form-item label="Status" prop="status">
                    <el-radio-group v-model="ruleForm.status">
                        <el-radio :label="1" border>Enable</el-radio>
                        <el-radio :label="0" border>Disable</el-radio>
                    </el-radio-group>
                </el-form-item>                
            </el-tab-pane>

            <el-tab-pane label="Routing" name="route">
                <el-form-item label="Include route" prop="route_include">
                    <el-row class="msg-info">
                        <el-col :span="20">Enter CIDR format such as: 192.168.1.0/24</el-col>
                        <el-col :span="4">
                        <el-button size="mini" type="success" icon="el-icon-plus" circle
                                    @click.prevent="addDomain(ruleForm.route_include)"></el-button>
                        </el-col>
                    </el-row>
                    <el-row v-for="(item,index) in ruleForm.route_include"
                            :key="index" style="margin-bottom: 5px" :gutter="10">
                        <el-col :span="10">
                        <el-input v-model="item.val"></el-input>
                        </el-col>
                        <el-col :span="12">
                        <el-input v-model="item.note" placeholder="Description"></el-input>
                        </el-col>
                        <el-col :span="2">
                        <el-button size="mini" type="danger" icon="el-icon-minus" circle
                                    @click.prevent="removeDomain(ruleForm.route_include,index)"></el-button>
                        </el-col>
                    </el-row>      
                </el-form-item>

                <el-form-item label="Exclude route" prop="route_exclude">
                    <el-row class="msg-info">
                        <el-col :span="20">Enter CIDR format such as: 192.168.2.0/24</el-col>
                        <el-col :span="4">
                        <el-button size="mini" type="success" icon="el-icon-plus" circle
                                    @click.prevent="addDomain(ruleForm.route_exclude)"></el-button>
                        </el-col>
                    </el-row>
                    <el-row v-for="(item,index) in ruleForm.route_exclude"
                            :key="index" style="margin-bottom: 5px" :gutter="10">
                        <el-col :span="10">
                        <el-input v-model="item.val"></el-input>
                        </el-col>
                        <el-col :span="12">
                        <el-input v-model="item.note" placeholder="Description"></el-input>
                        </el-col>
                        <el-col :span="2">
                        <el-button size="mini" type="danger" icon="el-icon-minus" circle
                                    @click.prevent="removeDomain(ruleForm.route_exclude,index)"></el-button>
                        </el-col>
                    </el-row>
                </el-form-item> 
            </el-tab-pane>
            
            <el-tab-pane label="Domains" name="ds_domains">
                <el-form-item label="Include domain name" prop="ds_include_domains">
                    <el-input type="textarea" :rows="5" v-model="ruleForm.ds_include_domains"></el-input>
                </el-form-item>                
                <el-form-item label="Exclude domain name" prop="ds_exclude_domains">
                    <el-input type="textarea" :rows="5" v-model="ruleForm.ds_exclude_domains"></el-input>
                </el-form-item>
            </el-tab-pane>
        </el-tabs>
        <el-form-item>
            <el-button type="primary" @click="submitForm('ruleForm')">Save</el-button>
            <el-button @click="disVisible">Cancel</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>
</div>
  
</template>

<script>
import axios from "axios";

export default {
  name: "Policy",
  components: {},
  mixins: [],
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['Users', 'Policy'])
  },
  mounted() {
    this.getData(1)
  },
  data() {
    return {
      page: 1,
      tableData: [],
      count: 10,
      activeTab : "general",
      readMore: {},
      readMinRows : 5,      
      ruleForm: {
        bandwidth: 0,
        status: 1,
        allow_lan: true,
        client_dns: [{val: '1.1.1.1'}],
        route_include: [{val: 'all', note: 'Default global proxy'}],
        route_exclude: [],
        re_upper_limit : 0,        
      },
      rules: {
        username: [
          {required: true, message: 'Please enter username', trigger: 'blur'},
          {max: 30, message: 'Less than 30 characters long', trigger: 'blur'}
        ],
        bandwidth: [
          {required: true, message: 'Please enter bandwidth limit', trigger: 'blur'},
          {type: 'number', message: 'Bandwidth must be a numeric value'}
        ],
        status: [
          {required: true}
        ],       
      },
    }
  },
  methods: {
    handleDel(row) {
      axios.post('/user/policy/del?id=' + row.id).then(resp => {
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
      this.activeTab = "general"
      this.user_edit_dialog = true
      if (!row) {
        return;
      }

      axios.get('/user/policy/detail', {
        params: {
          id: row.id,
        }
      }).then(resp => {
        this.ruleForm = resp.data.data
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
      axios.get('/user/policy/list', {
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
      //console.log(index)
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
      arr.push({val: "", action: "allow", port: 0});
    },
    submitForm(formName) {
      this.$refs[formName].validate((valid) => {
        if (!valid) {
          //console.log('Error submit!');
          return false;
        }

        axios.post('/user/policy/set', this.ruleForm).then(resp => {
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
    beforeTabLeave() {
      var isSwitch = true
      if (! this.user_edit_dialog) {
        return isSwitch;
      }      
      this.$refs['ruleForm'].validate((valid) => {
        if (!valid) {
          this.$message.error("Error: You have omitted a required field.")
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
</style>
