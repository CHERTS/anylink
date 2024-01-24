<template>
  <el-card>
      <el-table
          :data="soft_data"
          border>

        <el-table-column
            prop="info"
            label="Information"
            width="260">
        </el-table-column>

        <el-table-column
            prop="name"
            label="Configuration"
            width="200">
        </el-table-column>

        <el-table-column
            prop="env"
            label="Environment variables"
            width="220">
        </el-table-column>

        <el-table-column
            prop="data"
            label="Data">
          <template slot-scope="scope">
            {{ scope.row.data }}
          </template>
        </el-table-column>

      </el-table>
  </el-card>
</template>

<script>
import axios from "axios";

export default {
  name: "Soft",
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['Basic', 'Configuration'])
  },
  mounted() {
    this.getSoftInfo()
  },
  data() {
    return {soft_data: []}
  },

  methods: {
    getSoftInfo() {
      axios.get('/set/soft', {}).then(resp => {
        var data = resp.data
        //console.log(data);
        this.soft_data = data.data;
      }).catch(error => {
        this.$message.error('Oh, request error');
        console.log(error);
      });
    }
  },
}
</script>

<style scoped>

</style>
