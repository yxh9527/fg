import fs from 'node:fs/promises'
import path from 'node:path'

const appMapPath = process.argv[2] || 'client-record2/maps/app.b98df11551009e75d7e7.js.map'
const cssMapPath = process.argv[3] || 'client-record2/maps/app.ea95501c28a86deeecb2da5c0f369522.css.map'
const outRoot = path.resolve(process.argv[4] || 'client-record-restored')

const appMap = JSON.parse(await fs.readFile(appMapPath, 'utf8'))
const cssMap = JSON.parse(await fs.readFile(cssMapPath, 'utf8'))

const commonVueNames = new Set([
  'slot_record.vue',
  'slot_account.vue',
  'slot_detail.vue',
  'chess_record.vue',
  'chess_account.vue',
  'chess_detail.vue',
  'fruit_record.vue',
  'fruit_account.vue',
  'fruit_detail.vue',
  'hunting_record.vue',
  'hunting_account.vue',
  'hunting_detail.vue',
])

function normalizeContent(content) {
  return String(content ?? '').replace(/\r\n/g, '\n')
}

function fromWebpackSrc(source) {
  return source.replace(/^webpack:\/\/\/\.?\//, '')
}

function mapAppSource(source) {
  if (source.startsWith('webpack:///./src/') && !source.includes('?')) {
    return fromWebpackSrc(source)
  }

  if (source === 'webpack:///App.vue') {
    return 'src/App.vue'
  }

  if (source === 'webpack:///Home.vue' || source === 'webpack:///Login.vue' || source === 'webpack:///refresh.vue') {
    return `src/components/${path.basename(source.replace('webpack:///', ''))}`
  }

  const bareName = source.replace('webpack:///', '')
  if (commonVueNames.has(bareName)) {
    return `src/components/Common/${bareName}`
  }

  return null
}

function mapCssSource(source) {
  if (source.startsWith('webpack:///./src/assets/css/')) {
    return fromWebpackSrc(source)
  }

  return null
}

async function ensureWrite(relativePath, content, origin, manifest) {
  const targetPath = path.join(outRoot, relativePath)
  const normalized = normalizeContent(content)

  await fs.mkdir(path.dirname(targetPath), { recursive: true })

  try {
    const existing = normalizeContent(await fs.readFile(targetPath, 'utf8'))
    if (existing === normalized) {
      manifest.duplicates.push({ path: relativePath, origin })
      return
    }
    manifest.conflicts.push({ path: relativePath, origin })
    return
  } catch {
    // File does not exist yet.
  }

  await fs.writeFile(targetPath, normalized)
  manifest.written.push({ path: relativePath, origin })
}

const manifest = {
  appMapPath,
  cssMapPath,
  written: [],
  duplicates: [],
  conflicts: [],
  skipped: [],
}

for (let i = 0; i < appMap.sources.length; i += 1) {
  const source = appMap.sources[i]
  const content = normalizeContent(appMap.sourcesContent[i])
  if (
    content.startsWith('var __vue_exports__, __vue_options__') ||
    content.startsWith('module.exports={render:function')
  ) {
    manifest.skipped.push({ source, reason: 'compiled-vue-loader-artifact' })
    continue
  }
  const target = mapAppSource(source)
  if (!target) {
    manifest.skipped.push({ source, reason: 'no-app-target' })
    continue
  }
  await ensureWrite(target, content, source, manifest)
}

for (let i = 0; i < cssMap.sources.length; i += 1) {
  const source = cssMap.sources[i]
  const target = mapCssSource(source)
  if (!target) {
    manifest.skipped.push({ source, reason: 'no-css-target' })
    continue
  }
  await ensureWrite(target, cssMap.sourcesContent[i], source, manifest)
}

const notes = `# client-record-restored

This directory was reconstructed from public webpack source maps.

Recovered:
- Vue single-file components under \`src/components\`
- App entry files under \`src/\`
- Vuex files under \`src/vuex\`
- Utility scripts under \`src/assets/js\`
- Language packs under \`src/components/Common/lang\`
- Stylesheets under \`src/assets/css\`

Not recovered exactly:
- Original build config such as webpack config, Babel config, and npm lockfiles
- Exact dependency versions
- Build-time constants such as \`HOST\`
- Webpack aliases such as the \`assets\` import alias
- Server-side API responses and runtime localStorage/session data

Source maps used:
- \`${appMapPath}\`
- \`${cssMapPath}\`
`

await fs.mkdir(outRoot, { recursive: true })
await fs.writeFile(path.join(outRoot, 'README.md'), notes.replace(/\n/g, '\r\n'))
await fs.writeFile(
  path.join(outRoot, 'recovery-manifest.json'),
  JSON.stringify(manifest, null, 2).replace(/\n/g, '\r\n'),
)

console.log(`Restored ${manifest.written.length} files into ${outRoot}`)
console.log(`Duplicates skipped: ${manifest.duplicates.length}`)
console.log(`Conflicts: ${manifest.conflicts.length}`)
