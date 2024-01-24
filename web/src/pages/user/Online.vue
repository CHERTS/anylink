<template>
  <div>
    <el-card>
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
            label="Group">
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
            prop="status"
            label="Real time up/down"
            width="220">
          <template slot-scope="scope">
            <el-tag type="success">{{ scope.row.bandwidth_up }}</el-tag>
            <el-tag class="m-left-10">{{ scope.row.bandwidth_down }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column
            prop="status"
            label="Total up/down"
            width="200">
          <template slot-scope="scope">
            <el-tag effect="dark" type="success">{{ scope.row.bandwidth_up_all }}</el-tag>
            <el-tag class="m-left-10" effect="dark">{{ scope.row.bandwidth_down_all }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column
            prop="last_login"
            label="Login time"
            :formatter="tableDateFormat">
        </el-table-column>

        <el-table-column
            label="Actions"
            width="150">
          <template slot-scope="scope">
            <el-button
                size="mini"
                type="primary"
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
    getData() {
      axios.get('/user/online').then(resp => {
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

</style>
