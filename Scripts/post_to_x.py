import sys, json, subprocess, time, urllib.request
from playwright.sync_api import sync_playwright

CHROME_PATH = r"C:\Program Files\Google\Chrome\Application\chrome.exe"
USER_DATA_DIR = r"C:\Users\JohnVictor\AppData\Local\Google\Chrome\User Data"
PROFILE_DIR = "Profile 1"
DEBUG_PORT = 9222


def wait_for_chrome(port: int, proc: subprocess.Popen, timeout: int = 20):
    for i in range(timeout):
        if proc.poll() is not None:
            print(f"  Chrome process exited with code {proc.returncode}")
            return False
        try:
            urllib.request.urlopen(f"http://127.0.0.1:{port}/json/version", timeout=2)
            return True
        except Exception:
            print(f"  waiting for Chrome debugger... ({i+1}s)")
            time.sleep(1)
    return False


def post(message: str):
    print("Starting Chrome...")
    proc = subprocess.Popen([
        CHROME_PATH,
        f"--user-data-dir={USER_DATA_DIR}",
        f"--profile-directory={PROFILE_DIR}",
        f"--remote-debugging-port={DEBUG_PORT}",
        "--no-first-run",
        "--no-default-browser-check",
    ])

    if not wait_for_chrome(DEBUG_PORT, proc):
        print("Chrome didn't open debug port in time")
        try:
            proc.terminate()
        except:
            pass
        return

    print("Chrome ready, connecting via CDP...")
    with sync_playwright() as p:
        browser = p.chromium.connect_over_cdp(f"http://127.0.0.1:{DEBUG_PORT}")
        context = browser.contexts[0]
        page = context.new_page()

        page.goto("https://x.com/compose/post", wait_until="networkidle")
        page.wait_for_selector('[data-testid="tweetTextarea_0"]', timeout=30000)
        print("Typing...")

        page.evaluate(f"""() => {{
            const ta = document.querySelector('[data-testid="tweetTextarea_0"]');
            ta.focus();
            document.execCommand('insertText', false, {json.dumps(message)});
        }}""")

        page.wait_for_timeout(1000)
        page.click('[data-testid="tweetButtonInline"]')
        print("Posted!")
        page.wait_for_timeout(2000)
        browser.close()

    print("Done")
    proc.terminate()


if __name__ == "__main__":
    msg = " ".join(sys.argv[1:]) if len(sys.argv) > 1 else ""
    if not msg:
        msg = input("Post text: ")
    post(msg)
