<template>
  <div class="tab-stack pool-group-panel">
    <el-alert v-if="!hasSelectedAgent" title="请先选择代理，再维护该代理的水池合并配置。" type="info" :closable="false" />
    <el-alert v-else title="同一代理、同一组内的游戏共用水池；编辑成员保留组余额，新加入游戏的原独立池运行数据会清除。" type="info" :closable="false" show-icon />
    <el-alert v-if="loadError" :title="loadError" type="error" :closable="false" />
    <el-alert v-if="hasSelectedAgent && hasPending" :title="pendingNotice" type="warning" :closable="false" show-icon />
    <el-card shadow="never" class="content-card section-card section-card--table">
      <div class="table-toolbar">
        <div>
          <div class="panel-title panel-title--sm">水池组</div>
          <div class="table-toolbar__note">每组 1–10 个游戏。新组从初始 Base 重新累计，重置周期沿用原有设置。</div>
        </div>
        <div class="toolbar-actions">
          <el-button :disabled="!hasSelectedAgent" :loading="loading" @click="fetchGroups">刷新</el-button>
          <el-button type="primary" :disabled="!hasSelectedAgent || saving" @click="openDialog()">新增水池组</el-button>
        </div>
      </div>
      <el-table v-loading="loading" :data="groups" empty-text="暂无水池组" style="width: 100%">
        <el-table-column prop="id" label="组 ID" width="90" />
        <el-table-column prop="name" label="组名称" min-width="140" />
        <el-table-column label="成员游戏" min-width="250">
          <template slot-scope="scope"><el-tag v-for="symbol in scope.row.symbols" :key="symbol" size="small" class="game-tag">{{ gameLabel(symbol) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="数量" width="70"><template slot-scope="scope">{{ scope.row.symbols.length }}</template></el-table-column>
        <el-table-column prop="balance" label="组余额" min-width="120" align="right" />
        <el-table-column label="初始 Base" min-width="110" align="right"><template slot-scope="scope">{{ configItem(scope.row.config).base }}</template></el-table-column>
        <el-table-column label="配置来源" min-width="140"><template slot-scope="scope">{{ gameLabel(scope.row.sourceSymbol) }}</template></el-table-column>
        <el-table-column prop="version" label="版本" width="70" />
        <el-table-column label="更新时间" width="170"><template slot-scope="scope">{{ dateTime(scope.row.updatedAt) }}</template></el-table-column>
        <el-table-column label="操作" width="140" fixed="right"><template slot-scope="scope">
          <el-button type="text" :disabled="hasPending" @click="openDialog(scope.row)">编辑</el-button>
          <el-button type="text" :disabled="hasPending" @click="deleteGroup(scope.row)">删除</el-button>
        </template></el-table-column>
      </el-table>
    </el-card>
    <el-card v-if="hasSelectedAgent" shadow="never" class="content-card section-card section-card--table">
      <div class="table-toolbar"><div>
        <div class="panel-title panel-title--sm">执行记录</div>
        <div class="table-toolbar__note">提交后需等待开奖服务执行。等待结算期间暂停受影响范围的新下注，已接受订单继续结算；失败任务自动重试。</div>
      </div></div>
      <el-table :data="operations" empty-text="暂无操作记录">
        <el-table-column prop="id" label="操作 ID" width="95" />
        <el-table-column label="操作" width="100"><template slot-scope="scope">{{ kindLabel(scope.row.kind) }}</template></el-table-column>
        <el-table-column prop="groupId" label="组 ID" width="85" />
        <el-table-column label="状态" width="150"><template slot-scope="scope"><el-tag :type="statusType(scope.row.status)" size="small">{{ statusLabel(scope.row.status) }}</el-tag></template></el-table-column>
        <el-table-column prop="error" label="执行说明" min-width="240" />
        <el-table-column label="更新时间" width="170"><template slot-scope="scope">{{ dateTime(scope.row.updatedAt) }}</template></el-table-column>
      </el-table>
    </el-card>
    <el-dialog :title="form.groupId ? '编辑水池组' : '新增水池组'" :visible.sync="dialogVisible" width="700px" :close-on-click-modal="false" :close-on-press-escape="!saving" :show-close="!saving">
      <el-alert v-if="hasPending" :title="pendingNotice" type="warning" :closable="false" show-icon />
      <div v-if="gameLoadError" class="game-load-error">
        <el-alert :title="gameLoadError" type="error" :closable="false" show-icon />
        <el-button size="small" :loading="loadingGames" @click="fetchDialogGames">重新获取游戏</el-button>
      </div>
      <el-form v-loading="loadingGames" label-width="110px" :disabled="saving || loadingGames">
        <el-form-item label="组名称" required><el-input v-model.trim="form.name" maxlength="100" show-word-limit /></el-form-item>
        <el-form-item label="成员游戏" required>
          <el-select v-model="form.symbols" multiple filterable :multiple-limit="10" placeholder="按加入顺序选择，最多10个" style="width: 100%" @change="preserveOrder">
            <el-option v-for="game in availableGames" :key="game.symbol" :label="optionLabel(game)" :value="game.symbol" :disabled="!form.groupId && !game.poolConfig" />
          </el-select>
          <div class="field-note">已选 {{ form.symbols.length }}/10。已加入其他组的游戏已过滤。</div>
        </el-form-item>
        <el-form-item label="配置来源">{{ gameLabel(sourceSymbol) || '请先选择游戏' }}</el-form-item>
        <el-form-item label="水池参数">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item v-for="item in configFields" :key="item.key" :label="item.label">{{ previewConfig[item.key] === undefined ? '-' : previewConfig[item.key] }}</el-descriptions-item>
          </el-descriptions>
        </el-form-item>
      </el-form>
      <el-alert :title="form.groupId ? '保留现有组余额和配置。新成员不带入旧池余额；移出成员从自己的初始配置重新累计。' : '默认复制第一个加入游戏的配置。新组从初始 Base 重新累计，原独立水池运行数据会清除。'" type="warning" :closable="false" show-icon />
      <span slot="footer"><el-button :disabled="saving" @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" :disabled="hasPending || loadingGames || !!gameLoadError" @click="saveGroup">提交配置</el-button></span>
    </el-dialog>
  </div>
</template>

<script>
import dayjs from 'dayjs';
import { withPoolGroupLogs } from '@/libs/poolGroupLog';
import { parseOptionalAgentId } from './controlHelpers';
import { getPoolGroups, getPoolGroupGames, createPoolGroup, updatePoolGroup, deletePoolGroup } from '@/api/data';

const freshForm = () => ({ groupId: 0, version: 0, name: '', symbols: [], sourceSymbol: '', config: null });
export default withPoolGroupLogs({
  name: 'PoolGroupPanel',
  props: { agentId: { default: null }, siteId: { default: null }, games: { type: Array, default: () => [] } },
  data() { return { groups: [], operations: [], pendingOperation: null, availableGames: [], loading: false, loadingGames: false, gameLoadError: '', saving: false, loadError: '', dialogVisible: false, dialogSequence: 0, form: freshForm(), selectionOrder: [], pollTimer: null, sequence: 0, disposed: false, lastSignature: '', requestId: '', revision: null,
    configFields: [{key:'base',label:'初始 Base'},{key:'revenue',label:'税收'},{key:'min',label:'最低水位'},{key:'max',label:'最高水位'},{key:'normal',label:'正常水位'},{key:'control',label:'控制参数'},{key:'minRate',label:'最低比例'},{key:'maxRate',label:'最高比例'},{key:'normalRate',label:'正常比例'}] }; },
  computed: {
    hasSelectedAgent() { return parseOptionalAgentId(this.agentId) !== null; },
    pendingTask() { return (this.pendingOperation && this.pendingOperation.status !== 'applied' ? this.pendingOperation : null) || this.operations.find(op => op.status !== 'applied') || null; },
    hasPending() { return !!this.pendingTask; },
    pendingNotice() { const op = this.pendingTask; return op ? '该代理有未完成的' + this.kindLabel(op.kind) + '任务 #' + op.id + '（' + this.statusLabel(op.status) + '）' + (op.error ? '：' + op.error : '') + '。可以先填写新增配置，任务完成后再提交。' : ''; },
    sourceSymbol() { return this.form.groupId ? this.form.sourceSymbol : (this.form.symbols[0] || ''); },
    previewConfig() { const source = this.availableGames.find(g => g.symbol === this.sourceSymbol); return this.configItem(this.form.groupId ? this.form.config : source && source.poolConfig); },
  },
  watch: {
    agentId() { this.resetScope(); },
    siteId() { this.resetScope(); },
  },
  methods: {
    configItem(config) { return config && config.pool && config.pool[1] || {}; },
    optionLabel(game) { return game.nameZH ? `${game.name} [${game.nameZH}]` : game.name || game.symbol; },
    gameLabel(symbol) { const game = [...this.availableGames, ...this.games].find(g => g.symbol === symbol); return game ? (game.label || this.optionLabel(game)) : symbol; },
    kindLabel(kind) { return {create:'新增',update:'编辑',delete:'删除',reset:'重置'}[kind] || kind; },
    statusLabel(value) { return {pending:'待执行',waiting:'等待结算',applying:'切换中',applied:'已生效',failed:'失败待重试'}[value] || value; },
    statusType(value) { return value === 'applied' ? 'success' : value === 'failed' ? 'danger' : 'warning'; },
    dateTime(value) { return value ? dayjs(Number(value) * 1000).format('YYYY-MM-DD HH:mm:ss') : '-'; },
    resetScope() { this.sequence++; this.dialogSequence++; this.groups = []; this.operations = []; this.pendingOperation = null; this.availableGames = []; this.dialogVisible = false; this.loadingGames = false; this.gameLoadError = ''; this.loadError = ''; this.form = freshForm(); this.revision = null; this.$emit('count', 0); this.fetchGroups(); },
    async fetchGroups() {
      clearTimeout(this.pollTimer);
      const sequence = ++this.sequence;
      if (!this.hasSelectedAgent) { this.groups = []; this.operations = []; this.pendingOperation = null; this.loading = false; this.$emit('count', 0); return; }
      this.loading = true;
      try {
        const response = await getPoolGroups({ agentId: this.agentId, webId: this.siteId || undefined });
        if (this.disposed || sequence !== this.sequence) return;
        const data = response.data.data || {}; this.groups = Array.isArray(data.groups) ? data.groups : []; this.operations = Array.isArray(data.operations) ? data.operations : []; this.pendingOperation = data.pendingOperation || null; this.loadError = '';
        if (this.revision !== null && this.revision !== data.revision) this.$emit('changed');
        this.revision = data.revision;
        this.$emit('count', this.groups.length);
      } catch (error) { if (sequence === this.sequence) this.loadError = '获取水池组失败，请刷新重试。'; }
      finally { if (!this.disposed && sequence === this.sequence) { this.loading = false; this.pollTimer = setTimeout(() => this.fetchGroups(), this.loadError ? 15000 : 4000); } }
    },
    async openDialog(row) {
      if (!this.hasSelectedAgent) { this.$message.warning('请先选择有效代理'); return; }
      this.dialogSequence++; this.availableGames = []; this.gameLoadError = '';
      this.form = row ? {groupId:row.id,version:row.version,name:row.name,symbols:[...row.symbols],sourceSymbol:row.sourceSymbol,config:row.config} : freshForm();
      this.selectionOrder = [...this.form.symbols]; this.requestId = ''; this.lastSignature = ''; this.dialogVisible = true;
      await this.fetchDialogGames();
    },
    async fetchDialogGames() {
      if (!this.hasSelectedAgent || !this.dialogVisible) return;
      const agentId = this.agentId; const groupId = this.form.groupId; const sequence = ++this.dialogSequence;
      this.loadingGames = true; this.gameLoadError = '';
      try {
        const response = await getPoolGroupGames({ agentId, webId: this.siteId || undefined, groupId: groupId || undefined });
        if (this.disposed || sequence !== this.dialogSequence || agentId !== this.agentId || !this.dialogVisible) return;
        this.availableGames = Array.isArray(response.data.data) ? response.data.data : [];
      } catch (error) {
        if (!this.disposed && sequence === this.dialogSequence && agentId === this.agentId && this.dialogVisible) this.gameLoadError = error?.response?.msg || error?.message || '获取可选游戏失败，请重试';
      } finally { if (!this.disposed && sequence === this.dialogSequence) this.loadingGames = false; }
    },
    preserveOrder(value) { this.form.symbols = [...this.selectionOrder.filter(s => value.includes(s)), ...value.filter(s => !this.selectionOrder.includes(s))]; this.selectionOrder = [...this.form.symbols]; },
    nextRequestId(payload) {
      const signature = JSON.stringify(payload);
      if (signature !== this.lastSignature) {
        const words = new Uint32Array(4); window.crypto.getRandomValues(words); this.requestId = Array.from(words, n => n.toString(16).padStart(8, '0')).join(''); this.lastSignature = signature;
      }
      return this.requestId;
    },
    async saveGroup() {
      if (!this.hasSelectedAgent) { this.$message.warning('请先选择有效代理'); return; }
      if (this.hasPending) { this.$message.warning(this.pendingNotice); return; }
      if (this.loadingGames || this.gameLoadError) { this.$message.warning('请等待可选游戏加载成功后再提交'); return; }
      if (!this.form.name || !this.form.symbols.length || this.form.symbols.length > 10) { this.$message.error('请填写组名称，并选择1至10个游戏'); return; }
      if (!this.previewConfig || this.previewConfig.base === undefined) { this.$message.error('来源游戏缺少有效水池配置'); return; }
      const payload = { agentId: Number(this.agentId), groupId: this.form.groupId, version: this.form.version, name: this.form.name, symbols: [...this.form.symbols] };
      payload.requestId = this.nextRequestId(payload); this.saving = true;
      try { await (payload.groupId ? updatePoolGroup(payload) : createPoolGroup(payload)); this.dialogVisible = false; this.$message.success('操作已提交，等待开奖服务执行'); await this.fetchGroups(); this.$emit('changed'); }
      finally { this.saving = false; }
    },
    async deleteGroup(row) {
      if (this.hasPending) { this.$message.warning(this.pendingNotice); return; }
      const agentId = this.agentId;
      try { await this.$confirm(`确认删除“${row.name}”？成员游戏将分别从自己的初始配置重新累计，组余额不会分摊。`, '删除水池组', {type:'warning'}); } catch (error) { return; }
      if (agentId !== this.agentId) return;
      const payload = {agentId:Number(agentId),groupId:row.id,version:row.version}; payload.requestId = this.nextRequestId(payload);
      await deletePoolGroup(payload); this.$message.success('删除操作已提交，等待执行'); await this.fetchGroups(); this.$emit('changed');
    },
  },
  mounted() { this.fetchGroups(); },
  beforeDestroy() { this.disposed = true; this.sequence++; this.dialogSequence++; clearTimeout(this.pollTimer); },
});
</script>

<style scoped>
.pool-group-panel { display: flex; flex-direction: column; gap: 16px; }
.game-tag { margin: 3px 5px 3px 0; }
.field-note { color: #909399; font-size: 12px; line-height: 24px; }
.game-load-error { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
</style>
