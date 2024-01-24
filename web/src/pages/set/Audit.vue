<template>
  <div>
    <el-card>    
    <el-tabs v-model="activeName" @tab-click="handleClick">
        <el-tab-pane label="User activity log" name="act_log">
            <AuditActLog ref="auditActLog"></AuditActLog>
        </el-tab-pane>        
        <el-tab-pane label="User access log" name="access_audit">
            <AuditAccess ref="auditAccess"></AuditAccess>
        </el-tab-pane>
    </el-tabs>
    </el-card>      
  </div>
</template>

<script>
import AuditAccess from "../../components/audit/Access";
import AuditActLog from "../../components/audit/ActLog";

export default {
  name: "Audit",
  components:{
    AuditAccess,
    AuditActLog
  },
  mixins: [],
  mounted() {    
    this.upTab();
  },  
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['Basic', 'Audit'])        
  },
  data() {
    return {
      activeName: "act_log",
    }
  },
  methods: {  
    upTab() {
      var tabname = this.$route.query.tabname
      if (tabname) {
        this.activeName = tabname
      }
      this.handleClick(this.activeName)      
    },
    handleClick() {
        switch (this.activeName) {
        case "access_audit":
            this.$refs.auditAccess.setSearchData()
            this.$refs.auditAccess.getData(1)            
            break
        case "act_log":
            this.$refs.auditActLog.getData(1)
            break          
        }
        this.$router.push({path: this.$route.path, query: {tabname: this.activeName}})
    },    
  }
}
</script>
