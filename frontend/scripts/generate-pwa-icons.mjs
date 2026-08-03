#!/usr/bin/env node
/**
 * Generate PWA PNG icons from frontend/public/icon.svg.
 * Run after updating icon.svg so pwa-192x192.png and pwa-512x512.png stay in sync.
 *
 * Usage: node scripts/generate-pwa-icons.mjs   (from frontend/)
 *    or: npm run generate-pwa-icons
 */

import { readFileSync, writeFileSync } from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'
import { createRequire } from 'module'

const require = createRequire(import.meta.url)
const { Resvg } = require('@resvg/resvg-js')

const __dirname = dirname(fileURLToPath(import.meta.url))
const publicDir = join(__dirname, '..', 'public')
const svgPath = join(publicDir, 'icon.svg')
const sizes = [192, 512]

const svg = readFileSync(svgPath)

for (const size of sizes) {
  const resvg = new Resvg(svg, { fitTo: { mode: 'width', value: size } })
  const png = resvg.render()
  const outPath = join(publicDir, `pwa-${size}x${size}.png`)
  writeFileSync(outPath, png.asPng())
  console.log(`Wrote ${outPath}`)
}
