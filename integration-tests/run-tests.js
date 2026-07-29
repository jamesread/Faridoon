# Faridoon integration tests
# Starts the Go service with -configdir and runs mocha.

import { spawn } from 'child_process'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'
import { createInterface } from 'readline'

const __dirname = dirname(fileURLToPath(import.meta.url))
const repoRoot = join(__dirname, '..')
const testsDir = join(__dirname, 'tests')
const serviceBin = join(repoRoot, 'service', 'faridoon-service')

async function main() {
  const child = spawn(serviceBin, ['-configdir', testsDir], {
    cwd: join(repoRoot, 'service'),
    env: { ...process.env, FARIDOON_STATIC_DIR: join(repoRoot, 'frontend', 'dist') },
    stdio: ['ignore', 'pipe', 'pipe'],
  })
  let resolved = false
  const waitForServer = new Promise((resolve) => {
    const rl = createInterface({ input: child.stdout })
    rl.on('line', (line) => {
      process.stdout.write(line + '\n')
      if (line.includes('Starting Faridoon') && !resolved) {
        resolved = true
        rl.close()
        resolve()
      }
    })
    child.stderr.on('data', (d) => {
      const s = d.toString()
      process.stderr.write(s)
      if (s.includes('Starting Faridoon') && !resolved) {
        resolved = true
        resolve()
      }
    })
    setTimeout(() => {
      if (!resolved) {
        resolved = true
        resolve()
      }
    }, 3000)
  })
  await waitForServer
  const mocha = spawn('npx', ['mocha', '--timeout', '30000', 'tests/init.spec.js'], {
    cwd: __dirname,
    env: { ...process.env, FARIDOON_BASE_URL: 'http://localhost:8080' },
    stdio: 'inherit',
  })
  const code = await new Promise((resolve) => mocha.on('close', resolve))
  child.kill('SIGTERM')
  process.exit(code)
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
