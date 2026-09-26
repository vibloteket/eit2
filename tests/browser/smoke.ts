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

  async function openSettings() {
    await click(640, 660);
  }
  async function openCredits() {
    await openSettings();
    await click(640, 512); // About & Credits
  }
  async function closeCreditsToSettings() {
    await click(1100, 100);
  }
  async function closeSettingsToLobby() {
    await key("Escape");
  }

  // Establish the separate About & Credits screen and a stable empty-lobby comparison.
  await openCredits();
  const credits = await capture(),
    creditsHash = hash(credits);
  await Bun.write(output + "/credits.png", credits);
  await audioCheck("credits keeps lobby audio", true);
  await closeCreditsToSettings();
  await closeSettingsToLobby();
  const emptyLobbyHash = hash(await capture());
  if (emptyLobbyHash === creditsHash)
    throw Error(
      "Credits did not replace/return through Settings to lobby view",
    );
  await Bun.write(output + "/lobby.png", await capture());
  async function assertCredits(label: string) {
    if (hash(await capture()) !== creditsHash) throw Error(label);
    checks.push(label);
  }
  async function assertEmptyLobby(label: string) {
    if (hash(await capture()) !== emptyLobbyHash) throw Error(label);
    checks.push(label);
  }

  // Keyboard: Settings is already focused after returning. Open it, move to
  // About, and hold Enter to prove the opening press cannot dismiss Credits.
  await key("Enter");
  for (let i = 0; i < 5; i++) await key("ArrowDown");
  await page.keyboard.down("Enter");
  await page.waitForTimeout(600);
  await assertCredits(
    "keyboard opens Credits through Settings and held Enter does not dismiss",
  );
  await page.keyboard.up("Enter");
  await page.waitForTimeout(200);
  await key("1");
  await closeSettingsToLobby();
  await assertEmptyLobby("fresh key returns through Settings without joining");

  await openSettings();
  await page.mouse.move(640, 512);
  await page.mouse.down();
  await page.waitForTimeout(600);
  await assertCredits("held opening mouse button does not dismiss");
  await page.mouse.up();
  await page.waitForTimeout(200);
  await key("Escape");
  await closeSettingsToLobby();
  await assertEmptyLobby(
    "Escape returns through Settings without changing lobby",
  );

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
  await tap(640, 660);
  await tap(640, 512);
  await assertCredits("touch opens separate Credits view through Settings");
  await tap(1100, 100);
  await tap(640, 576); // Settings Back
  await assertEmptyLobby(
    "touch returns through Settings without joining a player",
  );

  // An unjoined controller can leave Credits; stick motion alone is not a press.
  await openCredits();
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
  await padButton(1); // Credits -> Settings, without joining.
  await assertEmptyLobby("unjoined gamepad B returns through Settings");

  await padButton(0); // Join only; focus moves to Start.
  await padButton(15); // Start -> Settings.
  const joinedLobbyHash = hash(await capture());
  await padButton(0); // Open Settings.
  for (let i = 0; i < 5; i++) await padButton(13); // Down to About & Credits.
  await padButton(0);
  await assertCredits("joined gamepad navigates through Settings to Credits");
  await padButton(1); // Credits -> Settings.
  await padButton(1); // Settings -> Lobby.
  if (hash(await capture()) !== joinedLobbyHash)
    throw Error("Gamepad return leaked into lobby leave action");
  checks.push("gamepad return preserves its joined player");
  await padButton(1); // A fresh B press in the lobby deliberately leaves.
  await page.evaluate(() => {
    (window as any).__testPads = [];
  });
  await page.waitForTimeout(300);
  await openCredits();
  await key("Escape");
  await closeSettingsToLobby();
  await assertEmptyLobby(
    "lobby remains usable after gamepad Credits round trip",
  );

  // Keyboard navigation reaches Credits through the Settings menu.
  await key("Enter");
  for (let i = 0; i < 5; i++) await key("ArrowDown");
  await key("Enter");
  await assertCredits("keyboard utility navigation reaches Settings Credits");
  await key("Space");
  await closeSettingsToLobby();
  await assertEmptyLobby("Space returns through Settings with focus restored");

  async function toggleSound() {
    await openSettings();
    await click(640, 192);
    await closeSettingsToLobby();
  }
  async function toggleMusic() {
    await openSettings();
    await click(640, 256);
    await closeSettingsToLobby();
  }
  async function openCreditsAndReturn() {
    await openCredits();
    await closeCreditsToSettings();
    await closeSettingsToLobby();
  }

  await toggleMusic();
  await audioCheck("music off from Settings", false);
  await openCredits();
  await audioCheck("credits respects music off", false);
  await closeCreditsToSettings();
  await closeSettingsToLobby();
  await toggleMusic();
  await audioCheck("music on from Settings", true);

  await openSettings();
  await key("ArrowDown");
  await key("ArrowDown");
  for (let i = 0; i < 10; i++) await key("ArrowLeft");
  await closeSettingsToLobby();
  await audioCheck("music volume zero from Settings", false);
  await openSettings();
  await key("ArrowDown");
  await key("ArrowDown");
  for (let i = 0; i < 10; i++) await key("ArrowRight");
  await closeSettingsToLobby();
  await audioCheck("music volume restored from Settings", true);

  await toggleSound();
  await audioCheck("master mute from Settings", false);
  await openCredits();
  await audioCheck("credits respects mute", false);
  await closeCreditsToSettings();
  await closeSettingsToLobby();
  await toggleSound();
  await audioCheck("unmute from Settings", true);
  await key("1");
  await click(640, 580);
  await page.waitForTimeout(3200); // Let the musical count-in finish.
  await audioCheck("match track", true);
  // Capture a real hard drop while the new background track is playing.
  // Keyboard layout 1 uses Left Shift for drop. The music-only peak is below
  // .9 * .08; a higher observed transient proves effects remain in the mix.
  const dropMeasurement = page.evaluate(async () => {
    const probe = (window as any).__audioProbe;
    let peak = 0;
    for (let i = 0; i < 30; i++) {
      for (const analyser of probe.analysers) {
        const samples = new Float32Array(analyser.fftSize);
        analyser.getFloatTimeDomainData(samples);
        for (const x of samples) peak = Math.max(peak, Math.abs(x));
      }
      await new Promise((resolve) => setTimeout(resolve, 40));
    }
    return peak;
  });
  await key("ShiftLeft");
  const dropPeak = await dropMeasurement;
  if (dropPeak <= 0.072 || dropPeak >= 1)
    throw Error("Hard drop/music mix peak is unexpected: " + dropPeak);
  checks.push(
    "hard-drop transient exceeds music-only ceiling without clipping",
  );
  audio.push({
    name: "match music with hard drop",
    peak: dropPeak,
    expectedSound: true,
  });
  await key("Escape");
  await key("ArrowDown");
  await key("Enter"); // Pause menu: Restart.
  await page.waitForTimeout(3200);
  await audioCheck("match music playing after Restart", true);
  checks.push("pause-menu Restart resumes match playback");
  await key("Escape");
  await key("Enter"); // Resume is not Restart.
  await audioCheck("match music playing after Resume", true);
  await key("Escape");
  await key("ArrowDown");
  await key("ArrowDown");
  await key("Enter");
  await page.waitForTimeout(500);
  await audioCheck("back to lobby", true);

  await toggleMusic();
  await audioCheck("music off before restart test", false);
  await click(640, 580);
  await page.waitForTimeout(3200); // Let the musical count-in finish.
  await key("Escape");
  await key("ArrowDown");
  await key("Enter");
  await page.waitForTimeout(3200);
  await audioCheck("Restart respects music off", false);
  checks.push("Restart does not enable disabled music");
  await key("Escape");
  await key("ArrowDown");
  await key("ArrowDown");
  await key("Enter");
  await toggleMusic();
  await audioCheck("music restored after disabled restart", true);

  // A then B were selected above. Exercise C, D, A and B with audio enabled.
  // Restart must preserve each selection, and the four-track playlist wraps.
  for (const name of [
    "Bach sonata C",
    "Vivaldi D",
    "Bourrees A after playlist wrap",
    "Handel B",
  ]) {
    await click(640, 580);
    await page.waitForTimeout(3200); // Let the musical count-in finish.
    await audioCheck(name, true);
    await key("Escape");
    await key("ArrowDown");
    await key("Enter");
    await page.waitForTimeout(3200);
    await audioCheck(name + " after Restart", true);
    await key("Escape");
    await key("ArrowDown");
    await key("ArrowDown");
    await key("Enter");
  }
  checks.push("four-track new-match playlist and same-track Restart exercised");
  await audioCheck("unchanged lobby after four-track rotation", true);

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
