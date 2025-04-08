<template>
  <div>
    <el-card>
      <el-form :inline="true">
        <el-form-item>
          <el-select
              v-model="searchCate"
              style="width: 86px;"                      
              @change="handleSearch">
            <el-option
                label="Username"
                value="username">
            </el-option>
            <el-option
                label="Login group"
                value="group">
            </el-option>            
            <el-option
                label="MAC address"
                value="mac_addr">
            </el-option>
            <el-option
                label="IP address"
                value="ip">
            </el-option>
            <el-option
                label="Remote address"
                value="remote_addr">
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input
              v-model="searchText"
              placeholder="Enter search content"
              @input="handleSearch">
          </el-input>
        </el-form-item>        
        <el-form-item>
          Show dormant users:
            <el-switch
                v-model="showSleeper"
                @change="handleSearch">
            </el-switch>
        </el-form-item>
        <el-form-item>
          <el-button
              class="extra-small-button"
              type="danger"
              size="mini"
              :loading="loadingOneOffline"
              @click="handleOneOffline">
            One-click offline
          </el-button>
      </el-form>

      <el-table
          ref="multipleTable"
          :data="tableData"
          border>

        <el-table-column
            sortable="true"
            type="index"
            label="Serial"
            width="60">
        </el-table-column>

        <el-table-column
            prop="username"
            label="Username">
        </el-table-column>

        <el-table-column
            prop="group"
            label="Login group">
        </el-table-column>

        <el-table-column
            prop="mac_addr"
            label="MAC">
        </el-table-column>

        <el-table-column
            prop="unique_mac"
            label="Unique MAC">
            <template slot-scope="scope">
                <el-tag v-if="scope.row.unique_mac" type="success">Yes</el-tag>
                <el-tag v-else type="info">No</el-tag>
            </template>
        </el-table-column>

        <el-table-column
            prop="ip"
            label="IP address"
            width="140">
        </el-table-column>

        <el-table-column
            prop="remote_addr"
            label="Remote address">
        </el-table-column>
        <el-table-column
            prop="transport_protocol"
            label="Transport protocol">
        </el-table-column>
        <el-table-column
            prop="tun_name"
            label="TUN name">
        </el-table-column>

        <el-table-column
            prop="mtu"
            label="MTU">
        </el-table-column>

        <el-table-column
            prop="is_mobile"
            label="Mobile">
          <template slot-scope="scope">
            <i v-if="scope.row.client === 'mobile'" class="el-icon-mobile-phone" style="font-size: 20px;color: red"></i>
            <i v-else class="el-icon-s-platform" style="font-size: 20px;color: blue"></i>
          </template>
        </el-table-column>

        <el-table-column
            label="Real time up/down"
            width="220">
          <template slot-scope="scope">
            <el-tag type="success">{{ scope.row.bandwidth_up }}</el-tag>
            <el-tag class="m-left-10">{{ scope.row.bandwidth_down }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column
            label="Total up/down"
            width="200">
          <template slot-scope="scope">
            <el-tag effect="dark" type="success">{{ scope.row.bandwidth_up_all }}</el-tag>
            <el-tag class="m-left-10" effect="dark">{{ scope.row.bandwidth_down_all }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column
            prop="last_login"
            label="Log in time"
            :formatter="tableDateFormat">
        </el-table-column>

        <el-table-column
            label="Actions"
            width="150">
          <template slot-scope="scope">
            <el-button
                size="mini"
                type="primary"
                v-if="scope.row.remote_addr !== ''"
                @click="handleReline(scope.row)">Reconnect
            </el-button>

            <el-popconfirm
                class="m-left-10"
                @confirm="handleOffline(scope.row)"
                title="Are you sure you want to disconnect the user?">
              <el-button
                  slot="reference"
                  size="mini"
                  type="danger">Disconnect
              </el-button>
            </el-popconfirm>

          </template>
        </el-table-column>
      </el-table>

    </el-card>
  </div>
</template>

<script>
import axios from "axios";
import { MessageBox } from 'element-ui';

export default {
  name: "Online",
  components: {},
  mixins: [],
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['User', 'Online'])
  },
  mounted() {
    this.getData();

    const chatTimer = setInterval(() => {
      this.getData();
    }, 10000);

    this.$once('hook:beforeDestroy', () => {
      clearInterval(chatTimer);
    })

  },
  data() {
    return {
      tableData: [],
      searchCate: 'username',
      searchText: '',
      showSleeper: false,
      loadingOneOffline: false,
    }
  },
  methods: {
    handleOffline(row) {
      axios.post('/user/offline?token=' + row.token).then(resp => {
        var data = resp.data
        if (data.code === 0) {
          this.$message.success(data.msg);
          this.getData();
        } else {
          this.$message.error(data.msg);
        }
        //console.log(data);
      }).catch(error => {
        this.$message.error('Oh, request error');
        console.log(error);
      });
    },

    handleReline(row) {
      axios.post('/user/reline?token=' + row.token).then(resp => {
        var data = resp.data
        if (data.code === 0) {
          this.$message.success(data.msg);
          this.getData();
        } else {
          this.$message.error(data.msg);
        }
        //console.log(data);
      }).catch(error => {
        this.$message.error('Oh, request error');
        console.log(error);
      });
    },

    handleEdit(a, row) {
      console.log(a, row)
    },
    handleOneOffline() {
        if (this.tableData === null || this.tableData.length === 0) {
            this.$message.error('Error: The current online user table is empty and one-click offline operation cannot be performed!');
            return;
        }
        MessageBox.confirm('All users under the current search conditions will be "offline". Are you sure you want to execute this?', 'Danger', {
            confirmButtonText: 'Sure',
            cancelButtonText: 'Cancel',
            type: 'danger'
        }).then(() => {   
            try {
                this.loadingOneOffline = true;
                this.getData();        
                this.$message.success('Successful operation');
                this.loadingOneOffline = false;
                // Clear current form
                this.tableData = [];
            } catch (error) {
                this.loadingOneOffline = false;
                this.$message.error('operation failed');
            }
        });        
    },
    handleSearch() {
        this.getData();
    },
    getData() {
      axios.get('/user/online', 
        {
          params: {            
            search_cate: this.searchCate,
            search_text: this.searchText,
            show_sleeper: this.showSleeper,
            one_offline: this.loadingOneOffline
          }
        }
      ).then(resp => {
        var data = resp.data.data
        //console.log(data);
        this.tableData = data.datas;
        this.count = data.count
      }).catch(error => {
        this.$message.error('Oh, request error');
        console.log(error);
      });
    },
  },
}
</script>

<style scoped>
/deep/ .el-form .el-form-item__label,
/deep/ .el-form .el-form-item__content,
/deep/ .el-form .el-input,
/deep/ .el-form .el-select,
/deep/ .el-form .el-button,
/deep/ .el-form .el-select-dropdown__item {
  font-size: 11px;
}
.el-select-dropdown .el-select-dropdown__item {
    font-size: 11px;
    padding: 0 10px;
}
/deep/ .el-input__inner{
    height: 30px;
    padding: 0 10px;
}
.extra-small-button {
  padding: 5px 10px;
}
</style>
