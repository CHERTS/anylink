<template>
  <div>
    <el-row :gutter="10" class="mb10">
      <el-col :span="8">
        <el-card v-if="system.cpu" body-style="text-align: center;">
          <div slot="header">
            <span>CPU utilization</span>
          </div>

          <el-progress type="circle" :percentage="system.cpu.percent" style="margin-bottom: 20px"/>

          <Cell left="CPU frequency" :right="system.cpu.ghz" divider/>
          <Cell left="System load" :right="system.sys.load"/>
        </el-card>
      </el-col>


      <el-col :span="8">
        <el-card v-if="system.mem" body-style="text-align: center;">
          <div slot="header">
            <span>Memory utilization</span>
          </div>

          <el-progress type="circle" :percentage="system.mem.percent" style="margin-bottom: 20px"/>

          <Cell left="Total memory" :right="system.mem.total" divider/>
          <Cell left="Free memory" :right="system.mem.free"/>
        </el-card>
      </el-col>


      <el-col :span="8">
        <el-card v-if="system.disk" body-style="text-align: center;">
          <div slot="header">
            <span>Disk information</span>
          </div>

          <el-progress type="circle" :percentage="system.disk.percent" style="margin-bottom: 20px"/>

          <Cell left="Total space" :right="system.disk.total" divider/>
          <Cell left="Free space" :right="system.disk.free"/>
        </el-card>
      </el-col>

    </el-row>

    <el-card v-if="system.sys" style="margin-top: 10px">
      <div slot="header">
        <span>Product information</span>
      </div>
      <Cell left="Version" :right="system.sys.appVersion" divider/>
      <Cell left="CommitId" :right="system.sys.appCommitId" divider/>
      <Cell left="Go system" :right="system.sys.goOs" divider/>
      <Cell left="Go arch" :right="system.sys.goArch" divider/>
      <Cell left="Go version" :right="system.sys.goVersion" divider/>
      <Cell left="Goroutine" :right="system.sys.goroutine"/>
    </el-card>


    <el-card v-if="system.sys" style="margin-top: 10px">
      <div slot="header">
        <span>Server information</span>
      </div>

      <Cell left="Server hostname" :right="system.sys.hostname" divider/>
      <Cell left="Operating system" :right="system.sys.platform" divider/>
      <Cell left="Kernel version" :right="system.sys.kernel" divider/>
      <Cell left="CPU core" :right="system.cpu.core" divider/>
      <Cell left="CPU model" :right="system.cpu.modelName"/>
    </el-card>

  </div>
</template>

<script>

import Cell from "@/components/Cell";
import axios from "axios";

export default {
  name: 'Monitor',
  components: {Cell},
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['Basic', 'System'])
  },
  mounted() {
    this.getData();

    // const chatTimer = setInterval(() => {
    //   this.getData();
    // }, 2000);
    //
    // this.$once('hook:beforeDestroy', () => {
    //   clearInterval(chatTimer);
    // });

  },
  data() {
    return {system: {}}
  },
  methods: {
    getData() {
      axios.get('/set/system', {}).then(resp => {
        var data = resp.data.data
        //console.log(data);
        this.system = data;
      }).catch(error => {
        this.$message.error('Oh, request error');
        console.log(error);
      });
    }
  }
}
</script>

<style scoped>
.monitor {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.monitor-left {
  font-size: 14px;
}

.monitor-right {
  font-size: 12px;
  color: #909399;
}

</style>


