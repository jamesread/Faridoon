"""Logs in as the fictional admin and pins the picocrank theme.

repo-helper reuses one browser session for every capture, so the theme is
pinned in every section and the login only happens once.
"""
import time

LOGIN_JS = """
const done = arguments[arguments.length - 1];
fetch('/faridoon.v1.FaridoonService/Login', {
  method: 'POST',
  credentials: 'same-origin',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({username: 'morgan', password: 'marketing-demo'}),
}).then((r) => done(r.ok)).catch(() => done(false));
"""

ME_JS = """
const done = arguments[arguments.length - 1];
fetch('/faridoon.v1.FaridoonService/GetCurrentUser', {
  method: 'POST',
  credentials: 'same-origin',
  headers: {'Content-Type': 'application/json'},
  body: '{}',
}).then((r) => r.json()).then((j) => done(!!(j.user && j.user.username))).catch(() => done(false));
"""


def prepare_session(driver, theme):
    changed = False
    if not driver.execute_async_script(ME_JS):
        if not driver.execute_async_script(LOGIN_JS):
            raise RuntimeError("marketing login failed")
        changed = True
    if driver.execute_script("return localStorage.getItem('picocrank-theme');") != theme:
        driver.execute_script("localStorage.setItem('picocrank-theme', arguments[0]);", theme)
        changed = True
    if changed:
        driver.refresh()
        time.sleep(3)
