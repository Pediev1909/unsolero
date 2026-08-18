import { readFile, readdir } from 'node:fs/promises'
import process from 'node:process'
import { URL } from 'node:url'
import { gzipSync } from 'node:zlib'

const budgets = JSON.parse(
  await readFile(
    new URL('../performance-budgets.json', import.meta.url),
    'utf8',
  ),
)
const manifest = JSON.parse(
  await readFile(
    new URL('../dist/.vite/manifest.json', import.meta.url),
    'utf8',
  ),
)
const distURL = new URL('../dist/', import.meta.url)
const entry = manifest['index.html']
if (!entry?.file) throw new Error('Vite manifest does not contain index.html')

async function gzipBytes(file) {
  return gzipSync(await readFile(new URL(file, distURL)), { level: 9 })
    .byteLength
}

const initialFiles = new Set()
function addInitial(record) {
  if (!record || initialFiles.has(record.file)) return
  initialFiles.add(record.file)
  for (const css of record.css ?? []) initialFiles.add(css)
  for (const imported of record.imports ?? []) addInitial(manifest[imported])
}
addInitial(entry)

const assets = await readdir(new URL('../dist/assets/', import.meta.url))
const checks = []
const entryGzip = await gzipBytes(entry.file)
checks.push(['entry JavaScript', entryGzip, budgets.entryJavaScriptGzipBytes])

for (const asset of assets.filter((file) => file.endsWith('.js'))) {
  if (initialFiles.has(`assets/${asset}`)) continue
  checks.push([
    `lazy chunk ${asset}`,
    await gzipBytes(`assets/${asset}`),
    budgets.lazyChunkGzipBytes,
  ])
}
let cssGzip = 0
for (const asset of assets.filter((file) => file.endsWith('.css'))) {
  cssGzip += await gzipBytes(`assets/${asset}`)
}
checks.push(['global CSS', cssGzip, budgets.globalCssGzipBytes])

let initialTransfer = 0
for (const file of initialFiles) initialTransfer += await gzipBytes(file)
checks.push([
  'initial JS/CSS transfer',
  initialTransfer,
  budgets.initialTransferGzipBytes,
])

let failed = false
for (const [name, actual, maximum] of checks) {
  const status = actual <= maximum ? 'PASS' : 'FAIL'
  process.stdout.write(
    `${status} ${name}: ${actual} <= ${maximum} gzip bytes\n`,
  )
  failed ||= actual > maximum
}
if (failed) process.exitCode = 1
