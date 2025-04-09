<template>
    <div id="lock-manager">
        <el-card>
            <div slot="header">
                <el-button type="primary" @click="getLocks">Refresh</el-button>
            </div>
            <el-table :data="locksInfo" style="width: 100%" border>
                <el-table-column type="index" label="Serial" width="60"></el-table-column>
                <el-table-column prop="description" label="Description"></el-table-column>
                <el-table-column prop="username" label="Username"></el-table-column>
                <el-table-column prop="ip" label="IP address"></el-table-column>
                <el-table-column prop="state.locked" label="State" width="100">
                    <template slot-scope="scope">
                        <el-tag :type="scope.row.state.locked ? 'danger' : 'success'">
                            {{ scope.row.state.locked ? 'Locked' : 'Unlocked' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="state.attempts" label="Number of failures"></el-table-column>
                <el-table-column prop="state.lock_time" label="Lock deadline">
                    <template slot-scope="scope">
                        {{ formatDate(scope.row.state.lock_time) }}
                    </template>
                </el-table-column>
                <el-table-column prop="state.lastAttempt" label="Last attempt time">
                    <template slot-scope="scope">
                        {{ formatDate(scope.row.state.lastAttempt) }}
                    </template>
                </el-table-column>
                <el-table-column label="Operate">
                    <template slot-scope="scope">
                        <div class="button">
                            <el-button size="small" type="danger" @click="unlock(scope.row)">
                                Unlock
                            </el-button>
                        </div>
                    </template>
                </el-table-column>
            </el-table>
        </el-card>
    </div>
</template>

<script>
import axios from 'axios';

export default {
    name: 'LockManager',
    data() {
        return {
            locksInfo: []
        };
    },
    methods: {
        getLocks() {
            axios.get('/locksinfo/list')
                .then(response => {
                    this.locksInfo = response.data.data;
                })
                .catch(error => {
                    console.error('Failed to get locks info: ', error);
                    this.$message.error('Unable to obtain lock information, please try again later.');
                });
        },
        unlock(lock) {
            const lockInfo = {
                state: { locked: false },
                username: lock.username,
                ip: lock.ip,
                description: lock.description
            };

            axios.post('/locksinfo/unlok', lockInfo)
                .then(() => {
                    this.$message.success('Unlocked successfully!');
                    this.getLocks();
                })
                .catch(error => {
                    console.error('Failed to unlock: ', error);
                    this.$message.error('Unlock failed, please try again later.');
                });
        },
        formatDate(dateString) {
            if (!dateString) return '';
            const date = new Date(dateString);
            return new Intl.DateTimeFormat('zh-CN', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit',
                hour12: false
            }).format(date);
        }
    },
    created() {
        this.getLocks();
    }
};
</script>

<style scoped>
.button {
    display: flex;
    justify-content: center;
    align-items: center;
}
</style>