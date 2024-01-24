<template>
  <el-container style="height: 100%;">
    <!--Sidebar menu-->
    <el-aside :width="is_active?'200':'64'">
      <LayoutAside :is_active="is_active" :route_path="route_path"/>
    </el-aside>

    <el-container>
      <!--Header content-->
      <el-header>
        <!--Listen to variable events of child components-->
        <LayoutHeader :is_active.sync="is_active" :route_name="route_name"/>
      </el-header>
      <!--Text content-->
      <!--style="background-color: rgb(240, 242, 245);"-->
      <el-main style="background-color: #fbfbfb">
        <!--The corresponding component content is rendered into router-view-->
        <!--Subcomponent reports route information-->
        <router-view :route_path.sync="route_path" :route_name.sync="route_name"></router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script>
import LayoutAside from "@/layout/LayoutAside";
import LayoutHeader from "@/layout/LayoutHeader";

export default {
  name: "Layout",
  components: {LayoutHeader, LayoutAside},
  data() {
    return {
      is_active: true,
      route_path: '/index',
      route_name: ['Home'],
    }
  },
  watch: {
    route_path: function (val) {
      // var w = document.getElementById('layout-menu').clientWidth;
      window.console.log('is_active', val)
    },
  },
  created() {
    window.console.log('layout-route', this.$route)
  },
}
</script>

<style>
.el-header {
  background-color: #fff;
  /*box-shadow: 0 1px 4px rgba(0, 21, 41, .08);*/
  color: #333;
  line-height: 60px;
  /*width: 100%;*/

  border-bottom: 1px solid #d8dce5;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, .12), 0 0 3px 0 rgba(0, 0, 0, .04);
}

</style>
