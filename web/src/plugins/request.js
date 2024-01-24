// http://www.zhangwj.com/
// Global axios defaults
import axios from "axios";
import {getToken, removeToken} from "./token";
// axios.defaults.headers.common['Jwt'] = AUTH_TOKEN;
axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded';

if (process.env.NODE_ENV !== 'production') {
    // development environment
    axios.defaults.baseURL = 'http://127.0.0.1:8800';
}

function request(vm) {
    // HTTP request interceptor
    axios.interceptors.request.use(config => {
        // What to do before sending a request
        // Get the token and add it to the headers request header
        const token = getToken();
        if (token) {
            config.headers.Jwt = token;
        }
        return config;
    });

    console.log(vm)

    // HTTP response interceptor
    // Unified processing of 401 status, token expiration processing, clearing token and jumping to login
    // Parameter 1, indicates a successful response
    axios.interceptors.response.use(null, err => {
        // No login or token expired
        if (err.response.status === 401) {
            // Logout, status and token
            // vm.$store.dispatch("logout");
            //Jump login page
            removeToken();
            vm.$router.push('/login');
            // Note: The vm instance here needs to be passed in externally
        }
        return Promise.reject(err);
    });
}

export default request




