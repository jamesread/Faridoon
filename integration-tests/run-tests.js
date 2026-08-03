// Faridoon integration tests
// Starts the Go service with -configdir and runs mocha.

import { spawn } from 'child_process'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'
import { createInterface } from 'readline'

const __dirname = dirname(fileURLToPath(import.meta.url))
const repoRoot = join(__dirname, '..')
const testsDir = join(__dirname, 'tests')
const serviceBin = join(repoRoot, 'service', 'faridoon-service')
const testPort = process.env.FARIDOON_TEST_PORT || '18080'
const baseUrl = process.env.FARIDOON_BASE_URL || `http://localhost:${testPort}`

function dbEnv() {
  return {
    DB_HOST: process.env.DB_HOST || 'mysql',
    DB_USER: process.env.DB_USER || process.env.DB_USERNAME || 'user',
    DB_PASS: process.env.DB_PASS || process.env.DB_PASSWORD || 'password',
    DB_NAME: process.env.DB_NAME || process.env.DB_DATABASE || 'faridoon',
  }
}

async function main() {
  const env = {
    ...process.env,
    ...dbEnv(),
    PORT: testPort,
    FARIDOON_STATIC_DIR: join(repoRoot, 'frontend', 'dist'),
  }

  const child = spawn(serviceBin, ['-configdir', testsDir], {
    cwd: join(repoRoot, 'service'),
    env,
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
    }, 8000)
  })
  await waitForServer

  const mocha = spawn(
    'npx',
    ['mocha', '--timeout', '30000', 'tests/**/*.spec.js'],
    {
      cwd: __dirname,
      env: {
        ...env,
        FARIDOON_BASE_URL: baseUrl,
      },
      stdio: 'inherit',
      shell: false,
    },
  )
  const code = await new Promise((resolve) => mocha.on('close', resolve))
  child.kill('SIGTERM')
  process.exit(code ?? 1)
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
