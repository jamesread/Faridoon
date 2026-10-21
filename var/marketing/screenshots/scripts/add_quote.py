"""Types a fictional quote into the Add Quote form to show the live preview."""
import time

from selenium.webdriver.common.by import By

QUOTE = (
    "<rook> i wrote the docs for the new feature\n"
    "<juniper> wow, how long are they?\n"
    "<rook> the feature or the docs?"
)


def run(driver):
    box = driver.find_element(By.CSS_SELECTOR, "main textarea")
    box.clear()
    box.send_keys(QUOTE)
    time.sleep(1.5)
    driver.execute_script("document.activeElement.blur(); window.scrollTo(0, 0);")
