import Vue from "vue";
import VueRouter from "vue-router";
import { getToken } from "./token";

Vue.use(VueRouter)


const routes = [
    { path: '/login', component: () => import('@/pages/Login') },
    {
        path: '/admin',
        component: () => import('@/layout/Layout'),
        redirect: '/admin/home',
        children: [
            { path: 'home', component: () => import('@/pages/Home') },

            { path: 'set/system', component: () => import('@/pages/set/System') },
            { path: 'set/soft', component: () => import('@/pages/set/Soft') },
            { path: 'set/other', component: () => import('@/pages/set/Other') },
            { path: 'set/audit', component: () => import('@/pages/set/Audit') },

            { path: 'user/list', component: () => import('@/pages/user/List') },
            { path: 'user/policy', component: () => import('@/pages/user/Policy') },
            { path: 'user/online', component: () => import('@/pages/user/Online') },
            { path: 'user/ip_map', component: () => import('@/pages/user/IpMap') },
            { path: 'user/lockmanager', component: () => import('@/pages/user/LockManager') },

            { path: 'group/list', component: () => import('@/pages/group/List') },

        ],
    },

    { path: '*', redirect: '/admin/home' },
]

// 3. Create a router instance and pass the `routes` configuration
// You can also pass other configuration parameters, but let's keep it simple for now.
const router = new VueRouter({
    routes
})

//Route guard
router.beforeEach((to, from, next) => {
    // Determine whether the route to be entered requires authentication

    const token = getToken();

    console.log("beforeEach", from.path, to.path, token)
    // console.log(from)

    // If there is no token, all jump to login
    if (!token) {
        if (to.path === "/login") {
            next();
            return;
        }

        next({
            path: '/login',
            query: {
                redirect: to.path
            }
        });
        return;
    }

    if (to.path === "/login") {
        next({ path: '/admin/home' });
        return;
    }

    // With token
    next();
});

export default router;

