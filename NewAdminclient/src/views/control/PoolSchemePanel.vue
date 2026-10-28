<template>
  <div class="scheme-panel">
    <el-alert title="默认方案持续应用到现有及新增代理。各代理的水池配置和余额独立；专属配置不会被默认方案更新覆盖。" type="info" :closable="false" show-icon />
    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
    <el-alert v-if="catalogError" :title="'游戏列表加载失败：' + catalogError" type="error" :closable="false" show-icon />
    <el-radio-group v-model="view" size="small">
      <el-radio-button label="default">默认合并方案</el-radio-button>
      <el-radio-button label="agents">指定代理处理</el-radio-button>
      <el-radio-button label="test">测试代理</el-radio-button>
    </el-radio-group>

    <el-card v-if="view === 'default'" key="default-scheme-card" shadow="never">
      <div class="scheme-toolbar">
        <el-radio-group v-model="scope" :disabled="saving" @change="changeScope">
          <el-radio label="all">全部站点全部代理</el-radio><el-radio label="site">指定站点全部代理</el-radio>
        </el-radio-group>
        <el-select v-if="scope === 'site'" v-model="scopeSiteId" filterable placeholder="明确选择适用站点" :disabled="saving" @change="changeScope">
          <el-option v-for="site in sites" :key="site.id" :value="Number(site.id)" :label="site.name" />
        </el-select>
        <el-tag>已发布版本 {{ version }}</el-tag><el-tag v-if="dirty" type="warning">有未发布修改</el-tag>
      </div>
      <p class="scheme-note">适用范围：{{ scopeLabel }}。已单独配置的代理保持专属设置；现有分组存在差异的代理需要明确选择保留或恢复默认。</p>
      <p class="scheme-note">下方列表是完整的默认方案。新增组保存后先进入待发布草稿，点击“预检查并发布”会发布列表内全部 {{ groups.length }} 个组，无需勾选组。</p>
      <p v-if="savedScopes.length" class="scheme-note">已保存范围：{{ savedScopes.map(item => siteLabel(item.webId)).join('、') }}。适用范围不能交叉覆盖。</p>
      <div class="scheme-toolbar">
        <el-button :loading="loadingScheme" :disabled="saving" @click="reloadScheme">重新加载方案</el-button>
        <el-button v-if="!gamesReady" @click="fetchCatalog">重新获取游戏</el-button>
        <el-button :disabled="saving || loadingScheme || !scopeReady" @click="editGroup()">新增水池组</el-button>
        <el-button type="primary" :disabled="saving || loadingScheme || !scopeReady || !gamesReady" :loading="previewLoading" @click="openDefaultPreview">预检查并发布</el-button>
      </div>
      <el-table key="default-scheme-table" v-loading="loadingScheme" :data="groups" empty-text="暂无合并组；发布空方案会让跟随代理恢复独立水池">
        <el-table-column prop="name" label="组名称" min-width="140" />
        <el-table-column key="draft-publication-state" label="发布状态" width="130"><template slot-scope="row"><el-tag :type="draftGroupStatus(row.row) === '已发布' ? 'success' : 'warning'" size="small">{{ draftGroupStatus(row.row) }}</el-tag></template></el-table-column>
        <el-table-column label="成员游戏" min-width="260"><template slot-scope="row"><el-tag v-for="symbol in row.row.symbols" :key="symbol" size="small" class="game-tag">{{ gameLabel(symbol) }}</el-tag></template></el-table-column>
        <el-table-column label="配置来源" min-width="160"><template slot-scope="row">{{ gameLabel(row.row.sourceSymbol) }}</template></el-table-column>
        <el-table-column label="操作" width="140"><template slot-scope="row"><el-button type="text" :disabled="saving" @click="editGroup(row.row)">编辑</el-button><el-button type="text" :disabled="saving" @click="removeGroup(row.row)">删除</el-button></template></el-table-column>
      </el-table>
      <p class="scheme-note">每组 1–10 个游戏，配置来源按首次加入顺序确定。新建实际组时读取各代理自己的有效配置，编辑现有组保留原组配置和余额。</p>
    </el-card>

    <el-card v-show="view === 'agents'" key="agent-scheme-card" shadow="never">
      <el-alert title="在下方列表勾选正式代理后，再预检查。测试代理（ID 0）请使用“测试代理”入口。" type="info" :closable="false" show-icon />
      <div class="scheme-toolbar">
        <el-button type="text" @click="openTestAgent">管理测试代理（ID 0）</el-button>
      </div>
      <div class="scheme-toolbar">
        <el-select v-model="agentWebId" @change="filterAgents"><el-option :value="0" label="全部站点" /><el-option v-for="site in sites" :key="site.id" :value="Number(site.id)" :label="site.name" /></el-select>
        <el-input v-model.trim="search" placeholder="搜索代理名称" clearable class="agent-search" @keyup.enter.native="filterAgents" />
        <el-select v-model="modeFilter" clearable placeholder="全部配置模式" @change="filterAgents"><el-option value="default" label="跟随默认" /><el-option value="custom" label="专属配置" /><el-option value="migration" label="待接入确认" /></el-select>
        <el-button :loading="loadingAgents" @click="filterAgents">查询</el-button>
        <el-button :disabled="saving || previewLoading" @click="resetAgentFilters">清除筛选</el-button>
        <el-button @click="clearSelection">清空选择（{{ selectedAgentIds.length }}）</el-button>
      </div>
      <el-alert v-if="agentLoadError" :title="agentLoadError" type="error" :closable="false" show-icon />
      <el-table key="agent-scheme-table" ref="agentTable" v-loading="loadingAgents" :data="agents" row-key="agentId" @selection-change="selectAgents">
        <template slot="empty">
          <div class="agent-empty">
            <span>{{ agentEmptyNotice }}</span>
            <el-button v-if="agentLoadError" type="text" @click="fetchAgents">重新加载代理</el-button>
            <el-button v-else-if="hasAgentFilters" type="text" @click="resetAgentFilters">清除筛选并查看全部正式代理</el-button>
            <el-button type="text" @click="openTestAgent">管理测试代理（ID 0）</el-button>
          </div>
        </template>
        <el-table-column key="agent-selection" type="selection" :reserve-selection="true" width="55" align="center" />
        <el-table-column prop="agentId" label="代理 ID" width="95" /><el-table-column prop="name" label="代理名称" min-width="150" />
        <el-table-column label="配置模式" width="130"><template slot-scope="row">{{ modeLabel(row.row.mode) }}</template></el-table-column>
        <el-table-column label="生效 / 目标版本" width="160"><template slot-scope="row">{{ row.row.appliedRevision || '-' }} / {{ row.row.targetRevision || '-' }}</template></el-table-column>
        <el-table-column label="状态" width="140"><template slot-scope="row">{{ row.row.initializing ? '初始化中 · ' : '' }}{{ statusLabel(row.row.status) }}</template></el-table-column>
        <el-table-column prop="error" label="说明" min-width="230" />
        <el-table-column key="agent-actions" label="操作" width="160"><template slot-scope="row">
          <el-button size="mini" :type="isAgentSelected(row.row) ? 'primary' : 'default'" :aria-pressed="isAgentSelected(row.row)" @click.stop="toggleAgentSelection(row.row)">{{ isAgentSelected(row.row) ? '取消选择' : '选择' }}</el-button>
          <el-button type="text" @click="showAgent(row.row)">查看</el-button>
        </template></el-table-column>
      </el-table>
      <el-pagination :current-page.sync="agentPage" :page-size="30" :total="agentTotal" layout="total, prev, pager, next" @current-change="fetchAgents" />
      <el-divider />
      <el-form label-width="110px" :disabled="saving || previewLoading">
        <el-form-item label="执行对象">已选 {{ selectedAgentIds.length }} 个代理（保留跨页选择）。预检查会逐个展示代理和结果。</el-form-item>
        <el-form-item label="处理方式"><el-radio-group v-model="agentAction"><el-radio label="custom">修改专属配置</el-radio><el-radio label="restore">恢复最新默认</el-radio><el-radio label="keep">保留现状为专属</el-radio></el-radio-group></el-form-item>
        <template v-if="agentAction === 'custom'">
          <el-form-item label="分组操作"><el-radio-group v-model="agentKind" @change="resetAgentForm"><el-radio label="create">新增组</el-radio><el-radio label="update">编辑组</el-radio><el-radio label="delete">删除组</el-radio></el-radio-group></el-form-item>
          <el-form-item v-if="agentKind !== 'create'" label="选择分组"><el-select v-model="agentForm.key" filterable placeholder="从第一个选中代理的生效配置选择" @change="chooseAgentGroup"><el-option v-for="group in agentGroupOptions" :key="group.key" :value="group.key" :label="group.name" /></el-select><el-button type="text" :loading="loadingAgentGroups" @click="loadAgentGroups">刷新分组</el-button></el-form-item>
          <template v-if="agentKind !== 'delete'">
            <el-form-item label="组名称"><el-input v-model.trim="agentForm.name" maxlength="100" /></el-form-item>
            <el-form-item label="成员游戏"><el-select v-model="agentForm.symbols" multiple filterable :multiple-limit="10" style="width:100%" @change="agentSymbolsChanged"><el-option v-for="game in catalog" :key="game.symbol" :value="game.symbol" :label="gameLabel(game.symbol)" /></el-select></el-form-item>
          </template>
        </template>
        <el-form-item>
          <el-button type="primary" :disabled="!!agentPreviewDisabledReason" :loading="previewLoading" @click="openAgentPreview">预检查指定代理</el-button>
          <div v-if="agentPreviewDisabledReason" class="scheme-note" role="status">{{ agentPreviewDisabledReason }}</div>
          <div v-if="requiresAgentCatalog && !gamesReady">
            <span v-if="catalogError" class="error-text">{{ catalogError }}</span>
            <el-button type="text" :loading="loadingCatalog" @click="fetchCatalog">重新获取游戏</el-button>
          </div>
        </el-form-item>
      </el-form>
      <p class="scheme-note">专属配置按代理完整保存，其余分组保留；以后默认方案新增的分组也不会自动加入。恢复默认会同步该代理所属范围的最新方案。</p>
    </el-card>

    <pool-group-panel v-if="view === 'test'" :agent-id="0" :site-id="null" :games="catalog" :initial-draft="testAgentDraft" />

    <el-card shadow="never">
      <div class="scheme-toolbar"><strong>同步批次</strong><el-checkbox v-model="includeAutomatic" @change="filterBatches">显示新增代理和补偿批次</el-checkbox><el-button size="small" @click="fetchBatches">刷新</el-button></div>
      <el-table :data="batches" empty-text="暂无同步批次">
        <el-table-column prop="id" label="批次" width="85" /><el-table-column label="类型" width="140"><template slot-scope="row">{{ batchKind(row.row.kind) }}</template></el-table-column>
        <el-table-column prop="revisionId" label="目标版本" width="100" /><el-table-column label="进度" min-width="310"><template slot-scope="row">{{ batchProgress(row.row) }}</template></el-table-column>
        <el-table-column label="操作" width="170"><template slot-scope="row"><el-button type="text" @click="showBatch(row.row)">明细</el-button><el-button type="text" :disabled="!row.row.counts.failed" @click="retry(row.row)">重试失败项</el-button></template></el-table-column>
      </el-table>
      <el-pagination :current-page.sync="batchPage" :page-size="10" :total="batchTotal" layout="total, prev, pager, next" @current-change="fetchBatches" />
    </el-card>

    <el-dialog title="编辑默认方案水池组" :visible.sync="groupDialog" width="650px" :close-on-click-modal="false">
      <el-form label-width="100px"><el-form-item label="组名称" required><el-input v-model.trim="groupForm.name" maxlength="100" /></el-form-item>
        <el-form-item label="成员游戏" required><el-select v-model="groupForm.symbols" multiple filterable :multiple-limit="10" style="width:100%" @change="groupSymbolsChanged"><el-option v-for="game in catalog" :key="game.symbol" :value="game.symbol" :label="gameLabel(game.symbol)" :disabled="occupied(game.symbol)" /></el-select></el-form-item>
        <el-form-item label="配置来源">{{ gameLabel(groupForm.sourceSymbol || groupForm.symbols[0]) }}</el-form-item>
      </el-form>
      <span slot="footer"><el-button @click="groupDialog = false">取消</el-button><el-button type="primary" @click="saveGroupDraft">保存到待发布方案</el-button></span>
    </el-dialog>

    <el-dialog :title="previewType === 'default' ? '发布默认方案预检查' : '指定代理预检查'" :visible.sync="previewDialog" width="1000px" :close-on-click-modal="false" :show-close="!saving" :close-on-press-escape="!saving">
      <el-alert v-if="previewType === 'default'" :title="'适用范围：' + previewScopeLabel + '，现有代理 ' + previewTotal + ' 个。后续新增代理也会自动应用；专属配置保留，冲突代理需后续单独处理。'" type="warning" :closable="false" />
      <el-alert v-else :title="'共检查 ' + previewRows.length + ' 个代理，可执行 ' + executableRows.length + ' 个；提交按钮只包含可执行项。'" type="warning" :closable="false" />
      <section v-if="previewType === 'default'" data-testid="default-publish-groups" class="publish-groups">
        <strong>本次发布的完整方案：{{ previewDefaultGroupCount }} 个水池组</strong>
        <p class="scheme-note">请先核对组名称和成员游戏。下方代理表用于检查应用结果，不是水池组选择列表；点击发布后会提交这里展示的完整方案。</p>
        <el-alert v-if="!previewDefaultGroupCount" title="本次发布空方案，将删除跟随代理的合并组并恢复独立水池。" type="warning" :closable="false" show-icon />
        <el-table key="publish-group-snapshot" :data="previewGroupRows" max-height="300" empty-text="目标方案没有合并组">
          <el-table-column prop="publicationStatus" label="本次变更" width="130" />
          <el-table-column prop="name" label="组名称" min-width="140" />
          <el-table-column label="成员游戏" min-width="280"><template slot-scope="row"><el-tag v-for="symbol in row.row.symbols" :key="symbol" size="small" class="game-tag">{{ gameLabel(symbol) }}</el-tag></template></el-table-column>
          <el-table-column label="配置来源" min-width="150"><template slot-scope="row">{{ gameLabel(row.row.sourceSymbol) }}</template></el-table-column>
        </el-table>
        <p v-if="previewRemovedGroups.length" class="scheme-note">将移除的组：{{ previewRemovedGroups.map(group => group.name).join('、') }}。对应成员恢复独立累计。</p>
        <strong>代理应用预检查</strong>
      </section>
      <el-table key="publish-agent-checks" v-loading="previewLoading" :data="previewRows"><el-table-column prop="agentId" label="代理 ID" width="95" /><el-table-column prop="name" label="代理名称" min-width="130" /><el-table-column label="变化" min-width="300"><template slot-scope="row"><div v-for="(change, index) in row.row.changes || []" :key="index">{{ change }}</div></template></el-table-column><el-table-column prop="error" label="冲突或失败原因" min-width="240" /></el-table>
      <el-pagination v-if="previewType === 'default'" :current-page.sync="previewPage" :page-size="30" :total="previewTotal" layout="total, prev, pager, next" @current-change="fetchDefaultPreview" />
      <span slot="footer"><el-button :disabled="saving" @click="previewDialog = false">取消</el-button><el-button type="primary" :loading="saving" :disabled="previewLoading || (previewType === 'agents' && !executableRows.length)" @click="submitPreview">{{ previewType === 'default' ? '发布完整方案（' + previewDefaultGroupCount + '组）并同步' : '提交可执行的 ' + executableRows.length + ' 个代理' }}</el-button></span>
    </el-dialog>

    <el-dialog title="批次明细" :visible.sync="batchDialog" width="900px">
      <p v-if="batchDetail">{{ batchProgress(batchDetail) }}</p>
      <el-table :data="batchItems"><el-table-column prop="agentId" label="代理 ID" width="100" /><el-table-column label="状态" width="150"><template slot-scope="row">{{ statusLabel(row.row.status) }}</template></el-table-column><el-table-column prop="operationId" label="执行任务" width="100" /><el-table-column prop="error" label="说明" min-width="300" /></el-table>
      <el-pagination :current-page.sync="batchItemPage" :page-size="30" :total="batchItemTotal" layout="total, prev, pager, next" @current-change="fetchBatchDetail" />
    </el-dialog>
    <el-dialog :title="'代理 ' + detailAgentId + ' 的实际配置'" :visible.sync="agentDialog" width="90%">
      <p v-if="agentDetail">{{ modeLabel(agentDetail.binding.mode) }} · 生效版本 {{ agentDetail.binding.appliedRevision || '-' }} · 目标版本 {{ agentDetail.binding.targetRevision || '-' }}</p>
      <p v-if="agentDetail && agentDetail.binding.error" class="error-text">{{ agentDetail.binding.error }}</p>
      <p v-if="agentDetail && agentDetail.hasDifference" class="scheme-note">当前实际配置与待应用配置不同，详情中的余额对应实际生效组。</p>
      <pool-group-panel v-if="agentDialog" :key="detailAgentId" :agent-id="detailAgentId" :site-id="null" :games="catalog" read-only />
    </el-dialog>
  </div>
</template>

<script>
import * as poolSchemes from '@/api/poolScheme';
import PoolGroupPanel from './PoolGroupPanel.vue';

const clone = value => JSON.parse(JSON.stringify(value));
const emptyGroup = () => ({ key: '', name: '', symbols: [], sourceSymbol: '' });
const randomId = () => { const words = new Uint32Array(4); window.crypto.getRandomValues(words); return Array.from(words, value => value.toString(16).padStart(8, '0')).join(''); };
const failureText = error => error?.response?.msg || error?.message || '请求失败，请重试';
const ordered = (previous, values) => [...previous.filter(value => values.includes(value)), ...values.filter(value => !previous.includes(value))];
const publicationStatus = (group, publishedGroups) => {
  const previous = publishedGroups.find(item => item.key === group.key);
  if (!previous) return '新增待发布';
  return previous.name === group.name && previous.sourceSymbol === group.sourceSymbol && JSON.stringify(previous.symbols) === JSON.stringify(group.symbols) ? '已发布' : '修改待发布';
};

export default {
  name: 'PoolSchemePanel',
  components: { PoolGroupPanel },
  props: { sites: { type: Array, default: () => [] }, games: { type: Array, default: () => [] } },
  data() {
    return { view: 'default', scope: 'all', scopeSiteId: null, version: 0, groups: [], savedGroups: '[]', savedScopes: [], catalog: [], gamesReady: false, loadingCatalog: false, catalogError: '', error: '', testAgentDraft: null,
      loadingScheme: false, loadingAgents: false, saving: false, schemeSequence: 0, agentSequence: 0, disposed: false, timer: null,
      groupDialog: false, groupForm: emptyGroup(), groupOrder: [], editingKey: '',
      agentWebId: 0, search: '', modeFilter: '', agents: [], agentPage: 1, agentTotal: 0, agentLoadError: '', selectedAgentIds: [], agentAction: 'custom', agentKind: 'create', agentForm: emptyGroup(), agentOrder: [], agentGroupOptions: [], loadingAgentGroups: false, agentGroupSequence: 0,
      previewDialog: false, previewType: 'default', previewLoading: false, previewRequest: null, previewPublishedGroups: [], previewRows: [], previewTotal: 0, previewPage: 1, previewSequence: 0, previewScopeLabel: '', requestSignature: '', requestId: '',
      batches: [], batchPage: 1, batchTotal: 0, includeAutomatic: false, batchSequence: 0, batchDialog: false, batchDetail: null, batchItems: [], batchItemPage: 1, batchItemTotal: 0, detailBatchId: 0, batchDetailSequence: 0,
      agentDialog: false, detailAgentId: null, agentDetail: null, detailSequence: 0 };
  },
  computed: {
    scopeReady() { return this.scope === 'all' || Number(this.scopeSiteId) > 0; },
    scopeWebId() { return this.scope === 'all' ? 0 : Number(this.scopeSiteId); },
    scopeLabel() { return this.scopeReady ? this.siteLabel(this.scopeWebId) : '尚未选择站点'; },
    dirty() { return JSON.stringify(this.groups) !== this.savedGroups; },
    executableRows() { return this.previewRows.filter(row => !row.error); },
    publishedGroups() { return JSON.parse(this.savedGroups); },
    previewDefaultGroupCount() { return this.previewType === 'default' && this.previewRequest ? this.previewRequest.groups.length : 0; },
    previewGroupRows() { return this.previewType === 'default' && this.previewRequest ? this.previewRequest.groups.map(group => ({ ...group, publicationStatus: publicationStatus(group, this.previewPublishedGroups) === '已发布' ? '保留' : publicationStatus(group, this.previewPublishedGroups) })) : []; },
    previewRemovedGroups() { return this.previewType === 'default' && this.previewRequest ? this.previewPublishedGroups.filter(group => !this.previewRequest.groups.some(item => item.key === group.key)) : []; },
    hasAgentFilters() { return Number(this.agentWebId) > 0 || !!this.search || !!this.modeFilter; },
    agentEmptyNotice() {
      if (this.agentLoadError) return '代理列表加载失败，请重新加载。';
      if (this.loadingAgents) return '正在加载代理列表…';
      return this.hasAgentFilters ? '没有符合当前站点、名称和配置模式的正式代理，请清除筛选后重试。' : '暂无正式代理。测试代理（ID 0）请使用“测试代理”入口。';
    },
    requiresAgentCatalog() { return this.agentAction === 'custom' && this.agentKind !== 'delete'; },
    agentPreviewDisabledReason() {
      if (this.saving) return '正在提交配置，请稍候。';
      if (this.previewLoading) return '正在预检查，请稍候。';
      if (!this.selectedAgentIds.length) return '请先勾选至少一个正式代理。搜索名称不会自动选中代理；测试代理 ID 0 请使用上方测试代理入口。';
      if (this.selectedAgentIds.length > 1000) return '一次最多选择1000个正式代理。';
      if (this.requiresAgentCatalog && !this.gamesReady) return this.loadingCatalog ? '正在加载游戏列表，请稍候。' : '游戏列表未加载成功，请点击“重新获取游戏”。';
      return '';
    },
  },
  watch: {
    view(value) { if (value === 'agents') this.$nextTick(() => { if (this.$refs.agentTable) this.$refs.agentTable.doLayout(); }); },
  },
  methods: {
    siteLabel(id) { return Number(id) === 0 ? '全部站点' : (this.sites.find(site => Number(site.id) === Number(id)) || {}).name || '站点 ' + id; },
    draftGroupStatus(group) { return publicationStatus(group, this.publishedGroups); },
    gameLabel(symbol) { const game = this.catalog.find(item => item.symbol === symbol); return game ? (game.nameZH ? game.name + ' [' + game.nameZH + ']' : game.name || symbol) : symbol || '请选择游戏'; },
    modeLabel(mode) { return { default: '跟随默认', custom: '专属配置', migration: '待接入确认' }[mode] || '尚未接入'; },
    statusLabel(status) { return { pending: '待执行', waiting: '等待结算', applying: '执行中', applied: '已生效', failed: '失败待重试', conflict: '需处理冲突', superseded: '已由新版本替代', excluded: '已排除' }[status] || '-'; },
    batchKind(kind) { return { publish: '默认方案发布', custom: '专属配置修改', restore: '恢复默认', keep: '保留现状为专属', initialize: '初始化 / 补偿' }[kind] || kind; },
    batchProgress(batch) { const counts = batch.counts || {}; return (batch.dispatchDone ? '已派发' : '派发中') + ' · 共 ' + batch.total + ' · 生效 ' + (counts.applied || 0) + ' · 等待 ' + ((counts.pending || 0) + (counts.waiting || 0) + (counts.applying || 0)) + ' · 失败 ' + (counts.failed || 0) + ' · 冲突 ' + (counts.conflict || 0) + ' · 排除/替代 ' + ((counts.excluded || 0) + (counts.superseded || 0)); },
    async fetchGroups() { await Promise.all([this.dirty ? Promise.resolve() : this.fetchScheme(), this.fetchAgents(), this.fetchBatches()]); },
    async fetchScheme() {
      const sequence = ++this.schemeSequence;
      if (!this.scopeReady) { this.loadingScheme = false; return; }
      this.loadingScheme = true;
      try { const response = await poolSchemes.getScheme({ webId: this.scopeWebId }); if (this.disposed || sequence !== this.schemeSequence) return;
        const data = response.data.data; this.groups = data.groups || []; this.savedGroups = JSON.stringify(this.groups); this.savedScopes = data.schemes || []; this.version = data.scheme.version || 0; this.$emit('count', this.groups.length); this.error = '';
      } catch (error) { if (sequence === this.schemeSequence) this.error = failureText(error); }
      finally { if (sequence === this.schemeSequence) this.loadingScheme = false; }
    },
    changeScope() { this.schemeSequence++; this.previewSequence++; this.previewDialog = false; this.groupDialog = false; this.groups = []; this.savedGroups = '[]'; this.version = 0; this.fetchScheme(); },
    async reloadScheme() { if (this.dirty) { try { await this.$confirm('重新加载将放弃尚未发布的草稿，是否继续？', '重新加载'); } catch (_) { return; } } await this.fetchScheme(); },
    occupied(symbol) { return this.groups.some(group => group.key !== this.editingKey && group.symbols.includes(symbol)); },
    editGroup(group) { this.editingKey = group ? group.key : ''; this.groupForm = group ? clone(group) : emptyGroup(); this.groupOrder = [...this.groupForm.symbols]; this.groupDialog = true; },
    groupSymbolsChanged(values) { this.groupForm.symbols = ordered(this.groupOrder, values); this.groupOrder = [...this.groupForm.symbols]; },
    saveGroupDraft() {
      if (!this.groupForm.name || !this.groupForm.symbols.length || this.groupForm.symbols.length > 10) { this.$message.warning('请填写组名称并选择1至10个游戏'); return; }
      if (this.groupForm.symbols.some(symbol => this.occupied(symbol))) { this.$message.warning('同一游戏不能重复归组'); return; }
      const group = clone(this.groupForm); if (!group.key) { group.key = 'g:' + randomId(); group.sourceSymbol = group.symbols[0]; }
      const index = this.groups.findIndex(item => item.key === group.key); if (index < 0) this.groups.push(group); else this.$set(this.groups, index, group);
      this.groupDialog = false; this.$emit('count', this.groups.length);
    },
    async removeGroup(group) { try { await this.$confirm('从默认方案草稿移除“' + group.name + '”？发布后跟随代理的对应组将被删除，原成员恢复独立累计。', '删除组'); } catch (_) { return; } this.groups = this.groups.filter(item => item.key !== group.key); this.$emit('count', this.groups.length); },
    defaultRequest() { return { scope: this.scope, webId: this.scopeWebId, version: this.version, groups: clone(this.groups) }; },
    async openDefaultPreview() {
      if (!this.scopeReady) return;
      this.previewType = 'default'; this.previewRequest = this.defaultRequest(); this.previewPublishedGroups = clone(this.publishedGroups); this.previewScopeLabel = this.scopeLabel; this.previewPage = 1; this.previewRows = []; this.previewTotal = 0; this.previewDialog = true;
      await this.fetchDefaultPreview();
    },
    async fetchDefaultPreview() {
      const sequence = ++this.previewSequence; this.previewLoading = true;
      try { const response = await poolSchemes.previewScheme(this.previewRequest, { page: this.previewPage, pageSize: 30 }); if (this.disposed || sequence !== this.previewSequence) return; this.previewRows = response.data.data.rows || []; this.previewTotal = response.data.data.total || 0; }
      catch (error) { if (sequence === this.previewSequence) { this.error = failureText(error); this.previewDialog = false; } }
      finally { if (sequence === this.previewSequence) this.previewLoading = false; }
    },
    nextRequestId(payload) { const signature = JSON.stringify(payload); if (signature !== this.requestSignature) { this.requestSignature = signature; this.requestId = randomId(); } return this.requestId; },
    async submitPreview() {
      if (this.saving || this.previewLoading || !this.previewRequest) return;
      let payload = clone(this.previewRequest);
      if (this.previewType === 'agents') { if (!this.executableRows.length) return; payload.agentIds = this.executableRows.map(row => row.agentId); payload.expected = Object.fromEntries(this.executableRows.map(row => [String(row.agentId), row.generation])); }
      payload.requestId = this.nextRequestId(payload); this.saving = true;
      try { const response = await (this.previewType === 'default' ? poolSchemes.publishScheme(payload) : poolSchemes.submitAgents(payload)); this.previewDialog = false; this.$message.success('批次 #' + response.data.data.id + ' 已提交，实际生效情况请查看同步进度');
        if (this.previewType === 'default') await this.fetchScheme();
        else if (payload.action === 'custom' && payload.kind === 'create') this.resetAgentForm();
        await Promise.all([this.fetchAgents(), this.fetchBatches()]); this.$emit('changed');
      } catch (error) { this.error = failureText(error); }
      finally { this.saving = false; }
    },
    async filterAgents() { this.agentPage = 1; await this.fetchAgents(); },
    async resetAgentFilters() { this.agentWebId = 0; this.search = ''; this.modeFilter = ''; await this.filterAgents(); },
    openTestAgent() {
      this.testAgentDraft = this.view === 'agents' && this.agentAction === 'custom' && this.agentKind === 'create' && (this.agentForm.name || this.agentForm.symbols.length)
        ? { name: this.agentForm.name, symbols: [...this.agentForm.symbols] } : null;
      this.view = 'test';
    },
    async fetchAgents() {
      const sequence = ++this.agentSequence; this.loadingAgents = true;
      try { const response = await poolSchemes.getAgents({ webId: this.agentWebId, search: this.search, mode: this.modeFilter || undefined, page: this.agentPage, pageSize: 30 }); if (this.disposed || sequence !== this.agentSequence) return; this.agents = response.data.data.rows || []; this.agentTotal = response.data.data.total || 0; this.agentLoadError = ''; }
      catch (error) { if (sequence === this.agentSequence) this.agentLoadError = failureText(error); }
      finally { if (sequence === this.agentSequence) this.loadingAgents = false; }
    },
    selectAgents(rows) { const first = this.selectedAgentIds[0]; this.selectedAgentIds = rows.map(row => row.agentId); if (this.selectedAgentIds[0] !== first) this.loadAgentGroups(); },
    isAgentSelected(agent) { return this.selectedAgentIds.some(id => Number(id) === Number(agent.agentId)); },
    toggleAgentSelection(agent) { if (this.$refs.agentTable) this.$refs.agentTable.toggleRowSelection(agent, !this.isAgentSelected(agent)); },
    clearSelection() { if (this.$refs.agentTable) this.$refs.agentTable.clearSelection(); this.selectedAgentIds = []; this.resetAgentForm(); },
    resetAgentForm() { this.agentForm = emptyGroup(); this.agentOrder = []; },
    async loadAgentGroups() {
      const sequence = ++this.agentGroupSequence; this.agentGroupOptions = []; if (this.agentKind !== 'create') this.resetAgentForm(); if (!this.selectedAgentIds.length) { this.loadingAgentGroups = false; return; }
      this.loadingAgentGroups = true;
      try { const response = await poolSchemes.getAgentDetail({ agentId: this.selectedAgentIds[0] }); if (this.disposed || sequence !== this.agentGroupSequence) return; this.agentGroupOptions = response.data.data.groups || []; }
      catch (error) { if (sequence === this.agentGroupSequence) this.error = failureText(error); }
      finally { if (sequence === this.agentGroupSequence) this.loadingAgentGroups = false; }
    },
    chooseAgentGroup(key) { const group = this.agentGroupOptions.find(item => item.key === key); if (group) { this.agentForm = clone(group); this.agentOrder = [...group.symbols]; } },
    agentSymbolsChanged(values) { this.agentForm.symbols = ordered(this.agentOrder, values); this.agentOrder = [...this.agentForm.symbols]; },
    async openAgentPreview() {
      if (!this.selectedAgentIds.length || this.selectedAgentIds.length > 1000) { this.$message.warning('请选择1至1000个代理'); return; }
      const group = clone(this.agentForm);
      if (this.agentAction === 'custom') {
        if (this.agentKind === 'create') { if (!group.key) { group.key = 'g:' + randomId(); this.agentForm.key = group.key; } group.sourceSymbol = group.symbols[0]; }
        if (!group.key || (this.agentKind !== 'delete' && (!group.name || !group.symbols.length || group.symbols.length > 10))) { this.$message.warning('请完整填写分组信息'); return; }
      }
      const payload = { agentIds: [...this.selectedAgentIds], action: this.agentAction, kind: this.agentKind, group };
      const sequence = ++this.previewSequence; this.previewType = 'agents'; this.previewRequest = payload; this.previewRows = []; this.previewLoading = true; this.previewDialog = true;
      try { const response = await poolSchemes.previewAgents(payload); if (this.disposed || sequence !== this.previewSequence) return; this.previewRows = response.data.data || []; }
      catch (error) { if (sequence === this.previewSequence) { this.error = failureText(error); this.previewDialog = false; } }
      finally { if (sequence === this.previewSequence) this.previewLoading = false; }
    },
    filterBatches() { this.batchPage = 1; this.fetchBatches(); },
    async fetchBatches() { const sequence = ++this.batchSequence; try { const response = await poolSchemes.getBatches({ page: this.batchPage, pageSize: 10, includeAutomatic: this.includeAutomatic }); if (this.disposed || sequence !== this.batchSequence) return; this.batches = response.data.data.rows || []; this.batchTotal = response.data.data.total || 0; } catch (error) { if (sequence === this.batchSequence) this.error = failureText(error); } },
    async showBatch(batch) { this.detailBatchId = batch.id; this.batchDetail = null; this.batchItems = []; this.batchItemPage = 1; this.batchDialog = true; await this.fetchBatchDetail(); },
    async fetchBatchDetail() { const sequence = ++this.batchDetailSequence; try { const response = await poolSchemes.getBatch({ id: this.detailBatchId, page: this.batchItemPage, pageSize: 30 }); if (this.disposed || sequence !== this.batchDetailSequence) return; const data = response.data.data; this.batchDetail = data.batch; this.batchItems = data.rows || []; this.batchItemTotal = data.total || 0; } catch (error) { if (sequence === this.batchDetailSequence) this.error = failureText(error); } },
    async retry(batch) { try { await poolSchemes.retryBatch({ id: batch.id }); this.$message.success('失败项已安排重试'); await this.fetchBatches(); } catch (error) { this.error = failureText(error); } },
    async showAgent(agent) { const sequence = ++this.detailSequence; this.detailAgentId = agent.agentId; this.agentDetail = null; this.agentDialog = true; try { const response = await poolSchemes.getAgentDetail({ agentId: agent.agentId }); if (this.disposed || sequence !== this.detailSequence) return; this.agentDetail = response.data.data; } catch (error) { if (sequence === this.detailSequence) this.error = failureText(error); } },
    async poll() { try { await Promise.all([this.fetchAgents(), this.fetchBatches(), this.batchDialog ? this.fetchBatchDetail() : Promise.resolve()]); } finally { if (!this.disposed) this.timer = setTimeout(() => this.poll(), 5000); } },
    async fetchCatalog() {
      if (this.loadingCatalog) return;
      this.loadingCatalog = true;
      try { const response = await poolSchemes.getSchemeGames(); if (!this.disposed) { this.catalog = response.data.data || []; this.gamesReady = true; this.catalogError = ''; } }
      catch (error) { if (!this.disposed) this.catalogError = failureText(error); }
      finally { this.loadingCatalog = false; }
    },
    async initialize() { await this.fetchCatalog(); await this.fetchScheme(); if (!this.disposed) this.poll(); },
  },
  mounted() { this.initialize(); },
  beforeDestroy() { this.disposed = true; clearTimeout(this.timer); },
};
</script>

<style scoped>
.scheme-panel { display:flex; flex-direction:column; gap:16px; }
.scheme-toolbar { display:flex; align-items:center; flex-wrap:wrap; gap:12px; margin-bottom:14px; }
.scheme-note { color:#606266; line-height:1.8; }
.agent-search { width:220px; }
.game-tag { margin:3px 5px 3px 0; }
.error-text { color:#c45656; }
.agent-empty { display:flex; flex-direction:column; align-items:center; gap:4px; padding:16px; line-height:1.6; }
.publish-groups { margin:18px 0; }
.el-pagination { margin-top:16px; }
</style>
