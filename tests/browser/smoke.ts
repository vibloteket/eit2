import { chromium } from "playwright";
import { createHash } from "node:crypto";
import { mkdir } from "node:fs/promises";

const root = import.meta.dir + "/../..";
const output = root + "/dist/browser-test";
await mkdir(output, { recursive: true });
const allowed = new Set([
  "index.html",
  "eit2.wasm",
  "wasm_exec.js",
  "favicon.svg",
  "favicon-32.png",
  "LICENSE",
  "NOTICE.md",
  "ASSETS.md",
  "CREDITS.md",
  "MUSIC-SOURCES.md",
]);
const server = process.env.BASE_URL
  ? null
  : Bun.serve({
      hostname: "127.0.0.1",
      port: 0,
      fetch(req) {
        const path = new URL(req.url).pathname.slice(1) || "index.html";
        if (
          !allowed.has(path) &&
          !(path.startsWith("LICENSES/") && !path.includes(".."))
        )
          return new Response("Not found", { status: 404 });
        const file = Bun.file(root + "/dist/web/" + path);
        return new Response(file);
      },
    });
const url = process.env.BASE_URL || `http://127.0.0.1:${server!.port}/`;
const browser = await chromium.launch({
  headless: true,
  args: [
    "--no-sandbox",
    "--use-angle=swiftshader",
    "--enable-unsafe-swiftshader",
  ],
});
const errors: string[] = [],
  checks: string[] = [],
  audio: any[] = [];
try {
  const page = await browser.newPage({
    viewport: { width: 1280, height: 720 },
    hasTouch: true,
  });
  page.on("pageerror", (e) => errors.push(String(e)));
  await page.addInitScript(() => {
    const w = window as any;
    w.__audioProbe = { contexts: [], analysers: [] };
    const connect = AudioNode.prototype.connect;
    (AudioNode.prototype as any).connect = function (
      destination: any,
      ...args: any[]
    ) {
      if (destination instanceof AudioDestinationNode) {
        const analyser = this.context.createAnalyser();
        analyser.fftSize = 2048;
        w.__audioProbe.analysers.push(analyser);
        connect.call(analyser, destination);
        return (connect as any).call(this, analyser, ...args);
      }
      return (connect as any).call(this, destination, ...args);
    };
    w.AudioContext = new Proxy(w.AudioContext, {
      construct(target, args) {
        const context = Reflect.construct(target, args);
        w.__audioProbe.contexts.push(context);
        return context;
      },
    });
    w.__testPads = [];
    Object.defineProperty(navigator, "getGamepads", {
      value: () => w.__testPads,
    });
  });
  await page.goto(url, { waitUntil: "load", timeout: 90000 });
  await page.locator("canvas").waitFor({ timeout: 60000 });
  await page.waitForFunction(
    () => (window as any).__audioProbe.analysers.length > 0,
    undefined,
    { timeout: 60000 },
  );
  await page.waitForTimeout(1500);
  if ((await page.locator("body").innerText()).includes("Could not start"))
    throw Error("Game startup failed");
  async function key(k: string) {
    await page.keyboard.press(k, { delay: 100 });
    await page.waitForTimeout(250);
  }
  async function click(x: number, y: number) {
    await page.mouse.click(x, y, { delay: 100 });
    await page.waitForTimeout(600);
  }
  async function capture() {
    return Buffer.from(await page.screenshot());
  }
  const hash = (b: Uint8Array) => createHash("sha256").update(b).digest("hex");
  async function audioCheck(name: string, sound: boolean) {
    const measure = () =>
      page.evaluate(async () => {
        const p = (window as any).__audioProbe;
        let squares = 0,
          count = 0;
        for (let i = 0; i < 10; i++) {
          for (const a of p.analysers) {
            const data = new Float32Array(a.fftSize);
            a.getFloatTimeDomainData(data);
            for (const x of data) {
              squares += x * x;
              count++;
            }
          }
          await new Promise((r) => setTimeout(r, 40));
        }
        return {
          rms: Math.sqrt(squares / Math.max(1, count)),
          contexts: p.contexts.map((c) => c.state),
          analysers: p.analysers.length,
        };
      });
    let result = await measure();
    let settlingWindows = 1;
    // MUSIC OFF leaves effects enabled. Opening Credits deliberately plays a
    // select effect; on a cold/busy browser it can finish after click()'s fixed
    // delay. Require a full quiet measurement window after bounded settling,
    // rather than treating that transient effect as background music.
    while (!sound && result.rms > 0.00001 && settlingWindows < 4) {
      await page.waitForTimeout(250);
      result = await measure();
      settlingWindows++;
    }
    if (
      !result.analysers ||
      (sound ? result.rms < 0.0001 : result.rms > 0.00001)
    )
      throw Error(`Audio check failed: ${name}: ${JSON.stringify(result)}`);
    audio.push({ name, expectedSound: sound, settlingWindows, ...result });
  }
  await audioCheck("before gesture", false);
  await click(1100, 50);
  await audioCheck("lobby after gesture", true);

  // Establish the separate full screen and a stable empty-lobby comparison.
  await click(890, 660);
  const credits = await capture(),
    creditsHash = hash(credits);
  await Bun.write(output + "/credits.png", credits);
  await audioCheck("credits keeps lobby audio", true);
  await click(1100, 100);
  const emptyLobbyHash = hash(await capture());
  if (emptyLobbyHash === creditsHash)
    throw Error("Credits did not replace/return to lobby view");
  await Bun.write(output + "/lobby.png", await capture());
  async function assertCredits(label: string) {
    if (hash(await capture()) !== creditsHash) throw Error(label);
    checks.push(label);
  }
  async function assertEmptyLobby(label: string) {
    if (hash(await capture()) !== emptyLobbyHash) throw Error(label);
    checks.push(label);
  }

  await page.keyboard.down("Enter");
  await page.waitForTimeout(600);
  await assertCredits(
    "keyboard opens Credits and held opening key does not dismiss",
  );
  await page.keyboard.up("Enter");
  await page.waitForTimeout(200);
  await key("1");
  await assertEmptyLobby("any fresh key returns without joining a player");

  await page.mouse.move(890, 660);
  await page.mouse.down();
  await page.waitForTimeout(600);
  await assertCredits("held opening mouse button does not dismiss");
  await page.mouse.up();
  await page.waitForTimeout(200);
  await key("Escape");
  await assertEmptyLobby("Escape returns without changing lobby");

  // Hold touches across several game ticks, as for the keyboard/mouse tests.
  const cdp = await page.context().newCDPSession(page);
  async function tap(x: number, y: number) {
    await cdp.send("Input.dispatchTouchEvent", {
      type: "touchStart",
      touchPoints: [{ x, y }],
    });
    await page.waitForTimeout(150);
    await cdp.send("Input.dispatchTouchEvent", {
      type: "touchEnd",
      touchPoints: [],
    });
    await page.waitForTimeout(500);
  }
  await tap(890, 660);
  await assertCredits("touch opens separate Credits view");
  await tap(1100, 100);
  await assertEmptyLobby("touch returns without joining a player");

  // An unjoined controller can leave Credits; stick motion alone is not a press.
  await click(890, 660);
  await page.evaluate(() => {
    (window as any).__testPads = [
      {
        id: "Eit browser test controller",
        index: 0,
        connected: true,
        mapping: "standard",
        timestamp: performance.now(),
        axes: [0, 0, 0, 0],
        buttons: Array.from({ length: 17 }, () => ({
          pressed: false,
          touched: false,
          value: 0,
        })),
      },
    ];
  });
  await page.waitForTimeout(400);
  await page.evaluate(() => {
    (window as any).__testPads[0].axes[0] = 1;
  });
  await page.waitForTimeout(400);
  await assertCredits("stick movement does not dismiss Credits");
  await page.evaluate(() => {
    const p = (window as any).__testPads[0];
    p.axes[0] = 0;
    p.buttons[0] = { pressed: true, touched: true, value: 1 };
  });
  await page.waitForTimeout(500);
  await assertEmptyLobby(
    "unjoined gamepad button returns without joining or starting",
  );
  await page.evaluate(() => {
    (window as any).__testPads[0].buttons[0] = {
      pressed: false,
      touched: false,
      value: 0,
    };
  });
  await page.waitForTimeout(250);
  async function padButton(button: number) {
    await page.evaluate((button) => {
      const p = (window as any).__testPads[0];
      p.buttons[button] = { pressed: true, touched: true, value: 1 };
      p.timestamp = performance.now();
    }, button);
    await page.waitForTimeout(300);
    await page.evaluate((button) => {
      const p = (window as any).__testPads[0];
      p.buttons[button] = { pressed: false, touched: false, value: 0 };
      p.timestamp = performance.now();
    }, button);
    await page.waitForTimeout(500);
  }
  await padButton(0); // Join only; focus moves to Start.
  await padButton(15);
  await padButton(15); // Start -> Debug -> Credits.
  const joinedLobbyHash = hash(await capture());
  await padButton(0);
  await assertCredits("joined gamepad navigates to and opens Credits");
  await padButton(1);
  if (hash(await capture()) !== joinedLobbyHash)
    throw Error("Gamepad return leaked into lobby leave action");
  checks.push("gamepad return preserves its joined player");
  await padButton(1); // A fresh B press in the lobby deliberately leaves.
  await page.evaluate(() => {
    (window as any).__testPads = [];
  });
  await page.waitForTimeout(300);
  await click(890, 660);
  await key("Escape");
  await assertEmptyLobby(
    "lobby remains usable after gamepad Credits round trip",
  );

  // Visible utility-row keyboard navigation must reach Credits from Debug.
  await key("ArrowLeft");
  await key("ArrowRight");
  await key("Enter");
  await assertCredits("keyboard utility navigation reaches Credits");
  await key("Space");
  await assertEmptyLobby("Space returns with Credits focus restored");

  await click(510, 660);
  await audioCheck("music off", false);
  await click(890, 660);
  await audioCheck("credits respects music off", false);
  await key("Escape");
  await click(510, 660);
  await audioCheck("music on", true);
  await click(350, 660);
  await audioCheck("master mute", false);
  await click(890, 660);
  await audioCheck("credits respects mute", false);
  await key("Escape");
  await click(350, 660);
  await audioCheck("unmute", true);
  await key("1");
  await click(640, 580);
  await audioCheck("match track", true);
  await key("Escape");
  await key("ArrowDown");
  await key("ArrowDown");
  await key("Enter");
  await page.waitForTimeout(500);
  await audioCheck("back to lobby", true);

  for (const path of [
    "CREDITS.md",
    "MUSIC-SOURCES.md",
    "LICENSES/CC0-1.0-VSCO2.txt",
  ]) {
    const response = await page.request.get(new URL(path, url).toString());
    if (!response.ok() || (await response.body()).length < 100)
      throw Error("Missing public notice: " + path);
  }
  if (errors.length) throw Error(errors.join("\n"));
  const result = {
    passed: true,
    baseURL: url,
    browser: "Chromium / SwiftShader",
    subjectivelyAuditioned: false,
    checks,
    audio,
    errors,
  };
  await Bun.write(output + "/report.json", JSON.stringify(result, null, 2));
  console.log(JSON.stringify(result, null, 2));
} finally {
  await browser.close();
  server?.stop(true);
}
