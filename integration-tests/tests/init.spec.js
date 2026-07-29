import assert from 'assert'
import { Builder, By, until } from 'selenium-webdriver'
import chrome from 'selenium-webdriver/chrome.js'

const base = process.env.FARIDOON_BASE_URL || 'http://localhost:8080'

describe('Faridoon Init / SPA', function () {
  this.timeout(30000)
  let driver

  before(async function () {
    const options = new chrome.Options()
    options.addArguments('--headless=new', '--no-sandbox', '--disable-dev-shm-usage')
    driver = await new Builder().forBrowser('chrome').setChromeOptions(options).build()
  })

  after(async function () {
    if (driver) await driver.quit()
  })

  it('loads the SPA footer with Faridoon branding', async function () {
    await driver.get(base + '/')
    const footer = await driver.wait(until.elementLocated(By.css('footer')), 10000)
    const text = await footer.getText()
    assert.match(text, /Powered by/)
    assert.match(text, /Faridoon/)
  })
})
