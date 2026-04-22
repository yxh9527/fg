import fs from 'node:fs/promises'
import path from 'node:path'
import { createHash } from 'node:crypto'

const inputUrl = process.argv[2]
const outputRoot = process.argv[3]
const webToken = process.argv[4] || ''

if (!inputUrl || !outputRoot) {
  console.error('Usage: node tools/download-site.mjs <url> <output-dir> [webToken]')
  process.exit(1)
}

const startUrl = new URL(inputUrl)
const hostRoot = path.resolve(outputRoot, startUrl.host)
const visited = new Set()
const queue = [startUrl.href]

const binaryExts = new Set([
  '.png', '.jpg', '.jpeg', '.gif', '.webp', '.svg', '.ico', '.bmp',
  '.woff', '.woff2', '.ttf', '.eot', '.otf', '.mp4', '.webm', '.mp3',
  '.wav', '.ogg', '.pdf', '.zip', '.gz', '.map'
])

function sanitizeUrl(raw, base) {
  if (!raw) return null
  const value = raw.trim()
  if (!value || value.startsWith('#')) return null
  if (
    value.startsWith('data:') ||
    value.startsWith('javascript:') ||
    value.startsWith('mailto:') ||
    value.startsWith('tel:')
  ) {
    return null
  }
  try {
    return new URL(value, base)
  } catch {
    return null
  }
}

function shouldFetch(url) {
  return url.origin === startUrl.origin
}

function toLocalPath(url) {
  const pathname = decodeURIComponent(url.pathname)
  let targetPath = pathname
  if (!targetPath || targetPath === '/') {
    targetPath = '/index.html'
  } else if (targetPath.endsWith('/')) {
    targetPath += 'index.html'
  } else {
    const ext = path.extname(targetPath)
    if (!ext) targetPath += '/index.html'
  }

  if (url.search) {
    const hash = createHash('md5').update(url.search).digest('hex').slice(0, 8)
    const parsed = path.parse(targetPath)
    targetPath = path.join(parsed.dir, `${parsed.name}__${hash}${parsed.ext}`)
  }

  return path.join(hostRoot, targetPath.replace(/^\/+/, ''))
}

function isHtmlResponse(contentType, url) {
  if (contentType && contentType.includes('text/html')) return true
  const ext = path.extname(url.pathname).toLowerCase()
  return !ext
}

function isCssResponse(contentType, url) {
  if (contentType && contentType.includes('text/css')) return true
  return path.extname(url.pathname).toLowerCase() === '.css'
}

function isBinary(url, contentType) {
  if (contentType) {
    if (
      contentType.startsWith('image/') ||
      contentType.startsWith('font/') ||
      contentType.startsWith('audio/') ||
      contentType.startsWith('video/') ||
      contentType.includes('application/octet-stream') ||
      contentType.includes('application/pdf') ||
      contentType.includes('application/zip')
    ) {
      return true
    }
  }
  return binaryExts.has(path.extname(url.pathname).toLowerCase())
}

function collectMatches(text, regex, baseUrl, found) {
  let match
  while ((match = regex.exec(text)) !== null) {
    const raw = match[1] || match[2] || match[3]
    const url = sanitizeUrl(raw, baseUrl)
    if (url && shouldFetch(url)) found.add(url.href)
  }
}

function findLinks(text, baseUrl, contentType) {
  const found = new Set()

  if (isHtmlResponse(contentType, baseUrl)) {
    collectMatches(text, /\b(?:src|href|poster)=["']?([^"'#\s>]+)["']?/gi, baseUrl, found)
    collectMatches(text, /\bsrcset=["']?([^"'>]+)["']?/gi, baseUrl, found)
    collectMatches(text, /url\((['"]?)([^)'"]+)\1\)/gi, baseUrl, found)

    const srcsetRegex = /\bsrcset=["']([^"']+)["']/gi
    let srcsetMatch
    while ((srcsetMatch = srcsetRegex.exec(text)) !== null) {
      for (const candidate of srcsetMatch[1].split(',')) {
        const raw = candidate.trim().split(/\s+/)[0]
        const url = sanitizeUrl(raw, baseUrl)
        if (url && shouldFetch(url)) found.add(url.href)
      }
    }
  }

  if (isCssResponse(contentType, baseUrl)) {
    collectMatches(text, /@import\s+(?:url\()?['"]?([^"')]+)['"]?\)?/gi, baseUrl, found)
    collectMatches(text, /url\((['"]?)([^)'"]+)\1\)/gi, baseUrl, found)
  }

  return [...found]
}

async function saveFile(filePath, body) {
  await fs.mkdir(path.dirname(filePath), { recursive: true })
  await fs.writeFile(filePath, body)
}

async function download(urlString) {
  if (visited.has(urlString)) return
  visited.add(urlString)

  const url = new URL(urlString)
  const res = await fetch(url, {
    redirect: 'follow',
    headers: {
      'user-agent': 'Mozilla/5.0 (compatible; site-downloader/1.0)',
      ...(webToken ? { Webtoken: webToken } : {})
    }
  })

  if (!res.ok) {
    console.warn(`Skip ${url.href}: HTTP ${res.status}`)
    return
  }

  const contentType = (res.headers.get('content-type') || '').toLowerCase()
  const finalUrl = new URL(res.url)
  const filePath = toLocalPath(finalUrl)

  if (isBinary(finalUrl, contentType)) {
    const body = Buffer.from(await res.arrayBuffer())
    await saveFile(filePath, body)
    return
  }

  const text = await res.text()
  await saveFile(filePath, text)

  for (const next of findLinks(text, finalUrl, contentType)) {
    if (!visited.has(next)) queue.push(next)
  }
}

await fs.mkdir(hostRoot, { recursive: true })

while (queue.length > 0) {
  const current = queue.shift()
  try {
    await download(current)
  } catch (error) {
    console.warn(`Skip ${current}: ${error.message}`)
  }
}

console.log(`Downloaded ${visited.size} URLs into ${hostRoot}`)
