import sys, json, subprocess, time, urllib.request
from playwright.sync_api import sync_playwright

CHROME_PATH = r"C:\Program Files\Google\Chrome\Application\chrome.exe"
USER_DATA_DIR = r"C:\Users\JohnVictor\AppData\Local\Google\Chrome\User Data"
PROFILE_DIR = "Profile 1"
DEBUG_PORT = 9223


def wait_for_chrome(port: int, timeout: int = 20):
    for i in range(timeout):
        try:
            urllib.request.urlopen(f"http://127.0.0.1:{port}/json/version", timeout=2)
            return True
        except Exception:
            print(f"  waiting for Chrome debugger... ({i+1}s)")
            time.sleep(1)
    return False


def post(message: str):
    proc = subprocess.Popen([
        CHROME_PATH,
        f"--user-data-dir={USER_DATA_DIR}",
        f"--profile-directory={PROFILE_DIR}",
        f"--remote-debugging-port={DEBUG_PORT}",
        f"--remote-debugging-address=127.0.0.1",
        "--no-first-run",
        "--no-default-browser-check",
    ], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

    if not wait_for_chrome(DEBUG_PORT):
        print("Chrome didn't start in time")
        proc.terminate()
        return

    with sync_playwright() as p:
        browser = p.chromium.connect_over_cdp(f"http://127.0.0.1:{DEBUG_PORT}")
        context = browser.contexts[0]
        page = context.new_page()

        page.goto("https://bsky.app/", wait_until="networkidle")
        page.locator('[data-testid="newPostButton"]').wait_for(timeout=15000)
        page.locator('[data-testid="newPostButton"]').click()
        print("New post dialog opened")
        page.wait_for_timeout(2000)

        page.evaluate(f"""() => {{
            const editor = document.querySelector('div[contenteditable="true"][role="textbox"]');
            if (!editor) return;
            editor.focus();
            editor.innerText = {json.dumps(message)};
            editor.dispatchEvent(new InputEvent('input', {{ bubbles: true, inputType: 'insertText' }}));
        }}""")
        print("Text entered")

        page.wait_for_timeout(1000)
        page.locator('[data-testid="composerPublishButton"]').click()
        print("Posted!")
        page.wait_for_timeout(2000)
        browser.close()

    proc.terminate()


if __name__ == "__main__":
    msg = " ".join(sys.argv[1:]) if len(sys.argv) > 1 else ""
    if not msg:
        msg = input("Post text: ")
    post(msg)
