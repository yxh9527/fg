<template>
  <div>
    <Button @click="showAdd" type="primary">新建</Button>
    <Row>
      <Col span="4">
        <h3>KEY</h3>
      </Col>
      <Col span="12">
        <h3>CONFIG</h3>
      </Col>
      <Col span="4">
        <h3>REMARK</h3>
      </Col>
      <Col span="4">
        <h3>EDIT</h3>
      </Col>
    </Row>
    <div>
      <template v-for="config in configList">
        <div class="listItem" :key="config.id">
          <Row>
            <Col span="4">{{ config.id }}</Col>
            <Col style="color: blue" span="12">
              <pre>{{ formatJson(config.config) }}</pre>
            </Col>
            <Col span="4">{{ config.remark }}</Col>
            <Col span="4">
              <Button @click="editConfig(config)" style="margin-right: 10px"
                >修改</Button
              >
              <Button type="error" @click="deleteConfig(config.id)"
                >删除</Button
              >
            </Col>
          </Row>
        </div>
      </template>
    </div>
    <Modal
      v-model="addConfigModel"
      :title="isEdit ? '编辑配置' : '新建配置'"
      :width="800"
      @on-ok="confirmConfig"
    >
      <div>
        <Form :model="formItem" :label-width="240">
          <FormItem label="KEY" :required="true">
            <Select
              :disabled="isEdit"
              v-model="formItem.id"
              style="width: 200px"
            >
              <Option
                v-for="item in ['mysql', 'redis', 'etcd']"
                :value="item"
                :key="item"
                >{{ item }}</Option
              >
            </Select>
          </FormItem>
          <!-- 配置项 -->
          <template v-if="formItem.id == 'mysql'">
            <FormItem label="host">
              <Input v-model="mysqlConfig.host" placeholder></Input>
            </FormItem>
            <FormItem label="port">
              <Input v-model="mysqlConfig.port" placeholder></Input>
            </FormItem>
            <FormItem label="user">
              <Input v-model="mysqlConfig.user" placeholder></Input>
            </FormItem>
            <FormItem label="password">
              <Input v-model="mysqlConfig.password" placeholder></Input>
            </FormItem>
            <FormItem label="database">
              <Input v-model="mysqlConfig.database" placeholder></Input>
            </FormItem>
            <FormItem label="ObjectTimeout">
              <Input
                v-model="mysqlConfig.pool.ObjectTimeout"
                placeholder
              ></Input>
            </FormItem>
            <FormItem label="IntervalCheckTime">
              <Input
                v-model="mysqlConfig.pool.IntervalCheckTime"
                placeholder
              ></Input>
            </FormItem>
            <FormItem label="MaxIdleTime">
              <Input v-model="mysqlConfig.pool.MaxIdleTime" placeholder></Input>
            </FormItem>
            <FormItem label="MaxObjectNum">
              <Input
                v-model="mysqlConfig.pool.MaxObjectNum"
                placeholder
              ></Input>
            </FormItem>
            <FormItem label="MinObjectNum">
              <Input
                v-model="mysqlConfig.pool.MinObjectNum"
                placeholder
              ></Input>
            </FormItem>
          </template>
          <template v-if="formItem.id == 'redis'">
            <FormItem label="default.host">
              <Input v-model="redisConfig.default.host" placeholder></Input>
            </FormItem>
            <FormItem label="default.port">
              <Input v-model="redisConfig.default.port" placeholder></Input>
            </FormItem>
            <FormItem label="default.password">
              <Input v-model="redisConfig.default.password" placeholder></Input>
            </FormItem>
            <FormItem label="default.database">
              <Input v-model="redisConfig.default.database" placeholder></Input>
            </FormItem>
            <FormItem label="default.max_active">
              <Input
                v-model="redisConfig.default.max_active"
                placeholder
              ></Input>
            </FormItem>
            <FormItem label="default.max_idle">
              <Input v-model="redisConfig.default.max_idle" placeholder></Input>
            </FormItem>
            <FormItem label="default.pool.MaxObjectNum">
              <Input
                v-model="redisConfig.default.pool.MaxObjectNum"
                placeholder
              ></Input>
            </FormItem>
            <FormItem label="default.pool.MinObjectNum">
              <Input
                v-model="redisConfig.default.pool.MinObjectNum"
                placeholder
              ></Input>
            </FormItem>

            <FormItem label="webRedis.host">
              <Input v-model="redisConfig.webRedis.host" placeholder></Input>
            </FormItem>
            <FormItem label="webRedis.port">
              <Input v-model="redisConfig.webRedis.port" placeholder></Input>
            </FormItem>
            <FormItem label="webRedis.password">
              <Input
                v-model="redisConfig.webRedis.password"
                placeholder
              ></Input>
            </FormItem>
            <FormItem label="webRedis.database">
              <Input
                v-model="redisConfig.webRedis.database"
                placeholder
              ></Input>
            </FormItem>
            <FormItem label="webRedis.max_active">
              <Input
                v-model="redisConfig.webRedis.max_active"
                placeholder
              ></Input>
            </FormItem>
            <FormItem label="webRedis.max_idle">
              <Input
                v-model="redisConfig.webRedis.max_idle"
                placeholder
              ></Input>
            </FormItem>
            <FormItem label="webRedis.pool.MaxObjectNum">
              <Input
                v-model="redisConfig.webRedis.pool.MaxObjectNum"
                placeholder
              ></Input>
            </FormItem>
            <FormItem label="webRedis.pool.MinObjectNum">
              <Input
                v-model="redisConfig.webRedis.pool.MinObjectNum"
                placeholder
              ></Input>
            </FormItem>
          </template>
          <template v-if="formItem.id == 'etcd'">
            <FormItem label="ETCD_HOST">
              <Input v-model="etcdConfig.ETCD_HOST" placeholder></Input>
            </FormItem>
            <FormItem label="ETCD_VERSION">
              <Input v-model="etcdConfig.ETCD_VERSION" placeholder></Input>
            </FormItem>
            <FormItem label="ETCD_USERNAME">
              <Input v-model="etcdConfig.ETCD_USERNAME" placeholder></Input>
            </FormItem>
            <FormItem label="ETCD_PASSWORD">
              <Input v-model="etcdConfig.ETCD_PASSWORD" placeholder></Input>
            </FormItem>
            <FormItem label="maxIdleTime">
              <Input v-model="etcdConfig.POOL.maxIdleTime" placeholder></Input>
            </FormItem>
            <FormItem label="maxObjectNum">
              <Input v-model="etcdConfig.POOL.maxObjectNum" placeholder></Input>
            </FormItem>
            <FormItem label="minObjectNum">
              <Input v-model="etcdConfig.POOL.minObjectNum" placeholder></Input>
            </FormItem>
          </template>
          <!-- 配置项 -->
          <FormItem label="备注" :required="true">
            <Input v-model="formItem.remark" placeholder="输入备注"></Input>
          </FormItem>
        </Form>
      </div>
    </Modal>
  </div>
</template>

<script>
import axios from '@/libs/api.request'
import { getToken } from '@/libs/util'

export default {
  components: {},
  data () {
    return {
      // 配置列表
      configList: [],
      // 添加配置对话框
      addConfigModel: false,
      // mysql配置
      mysqlConfig: {
        host: '',
        port: '',
        user: '',
        password: '',
        database: '',
        pool: {
          ObjectTimeout: '',
          IntervalCheckTime: '',
          MaxIdleTime: '',
          MaxObjectNum: '',
          MinObjectNum: ''
        }
      },
      // redis配置
      redisConfig: {
        default: {
          host: '',
          port: '',
          password: '',
          database: '',
          max_active: '',
          max_idle: '',
          pool: {
            MaxObjectNum: '',
            MinObjectNum: ''
          }
        },
        webRedis: {
          host: '',
          port: '',
          password: '',
          database: '',
          max_active: '',
          max_idle: '',
          pool: {
            MaxObjectNum: '',
            MinObjectNum: ''
          }
        }
      },
      // etcd配置
      etcdConfig: {
        ETCD_HOST: '',
        ETCD_VERSION: '',
        ETCD_USERNAME: '',
        ETCD_PASSWORD: '',
        POOL: {
          maxIdleTime: '',
          maxObjectNum: '',
          minObjectNum: ''
        }
      },
      // 新建配置对象
      formItem: {
        id: '',
        config: '',
        remark: ''
      },
      // 编辑模式
      isEdit: false
    }
  },
  methods: {
    /**
     * 显示添加
     */
    showAdd () {
      this.isEdit = false
      this.addConfigModel = true
      // 重置默认配置
      this.formItem = {
        id: '',
        config: '',
        remark: ''
      }
      this.mysqlConfig = {
        host: '',
        port: '',
        user: '',
        password: '',
        database: '',
        pool: {
          ObjectTimeout: '',
          IntervalCheckTime: '',
          MaxIdleTime: '',
          MaxObjectNum: '',
          MinObjectNum: ''
        }
      }
      this.etcdConfig = {
        ETCD_HOST: '',
        ETCD_VERSION: '',
        ETCD_USERNAME: '',
        ETCD_PASSWORD: '',
        POOL: {
          maxIdleTime: '',
          maxObjectNum: '',
          minObjectNum: ''
        }
      }
      this.redisConfig = {
        default: {
          host: '',
          port: '',
          password: '',
          database: '',
          max_active: '',
          max_idle: '',
          pool: {
            MaxObjectNum: '',
            MinObjectNum: ''
          }
        },
        webRedis: {
          host: '',
          port: '',
          password: '',
          database: '',
          max_active: '',
          max_idle: '',
          pool: {
            MaxObjectNum: '',
            MinObjectNum: ''
          }
        }
      }
    },

    /**
     * 格式化JSON
     */
    formatJson (msg) {
      let obj
      try {
        obj = JSON.parse(msg)
      } catch (error) {
        return msg
      }
      var jsonStr = JSON.stringify(obj, null, 6)
      return `${jsonStr}`
    },

    /**
     *编辑配置
     */
    editConfig (config) {
      this.isEdit = true
      this.addConfigModel = true
      this.formItem = config
      switch (config.id) {
        case 'mysql':
          this.mysqlConfig = JSON.parse(config.config).mysql
          break
        case 'redis':
          this.redisConfig = JSON.parse(config.config)
          break
        case 'etcd':
          this.etcdConfig = JSON.parse(config.config)
          break
      }
    },

    /**
     * 删除配置
     */
    async deleteConfig (id) {
      let data = await axios.request({
        url: 'v1/default-config/delete',
        method: 'POST',
        params: {
          token: getToken(),
          id
        }
      })

      if (data && data.data.code == 200) {
        this.$Message.info('删除成功')
        this.refreshConfig()
      }
    },

    /**
     * 提交修改
     */
    async confirmEdit () {
      switch (this.formItem.id) {
        case 'mysql':
          this.formItem.config = JSON.stringify({ mysql: this.mysqlConfig })
          break
        case 'redis':
          this.formItem.config = JSON.stringify(this.redisConfig)
          break
        case 'etcd':
          this.formItem.config = JSON.stringify(this.etcdConfig)
          break
      }

      let data = await axios.request({
        url: 'v1/default-config/update',
        method: 'post',
        params: {
          token: getToken(),
          id: this.formItem.id,
          config: this.formItem.config,
          remark: this.formItem.remark
        }
      })

      if (data && data.data.code == 200) {
        this.$Message.info('修改成功')
        this.refreshConfig()
      }
    },

    /**
     * 提交配置
     */
    async confirmConfig () {
      if (this.isEdit) {
        this.confirmEdit()
        return
      }

      if (this.configList.find((x) => x.id == this.formItem.id)) {
        this.$Message.error('配置KEY存在')
        return
      }
      switch (this.formItem.id) {
        case 'mysql':
          this.formItem.config = JSON.stringify({ mysql: this.mysqlConfig })
          break
        case 'redis':
          this.formItem.config = JSON.stringify(this.redisConfig)
          break
        case 'etcd':
          this.formItem.config = JSON.stringify(this.etcdConfig)
          break
      }

      if (this.formItem.id && this.formItem.config) {
        let data = await axios.request({
          url: 'v1/default-config/add',
          method: 'post',
          params: {
            token: getToken(),
            id: this.formItem.id,
            config: this.formItem.config,
            remark: this.formItem.remark
          }
        })

        if (data && data.data.code == 200) {
          this.$Message.info('添加成功')
          this.refreshConfig()
        }
      } else {
        this.$Message.error('请输入正确配置')
      }
    },

    /**
     * 获取配置
     */
    async refreshConfig () {
      let data = await axios.request({
        url: 'v1/default-config/list',
        method: 'get',
        params: {
          token: getToken()
        }
      })
      if (data && data.data.code == 200) {
        this.configList = data.data.data
      }
    }
  },
  mounted () {
    this.refreshConfig()
  }
}
</script>

<style scoped lang="less">
.listItem {
  border: 1px solid #dedede;
  border-bottom: none;
  padding: 10px;
  background: white;
}

.listItem:last-child {
  border-bottom: 1px solid #dedede;
}

pre {
  margin: 0;
}
</style>
