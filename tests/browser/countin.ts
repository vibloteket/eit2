import { chromium } from "playwright";
import { PNG } from "pngjs";
import { mkdir } from "node:fs/promises";
const root = import.meta.dir + "/../..",
  output = root + "/dist/countin-test";
await mkdir(output, { recursive: true });
const server = process.env.BASE_URL
  ? null
  : Bun.serve({
      hostname: "127.0.0.1",
      port: 0,
      fetch(req) {
        const p = new URL(req.url).pathname.slice(1) || "index.html";
        if (p.includes("..")) return new Response("", { status: 400 });
        return new Response(Bun.file(root + "/dist/web/" + p));
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
const checks: string[] = [],
  errors: string[] = [],
  measurements: any[] = [];
function check(ok: boolean, name: string) {
  if (!ok) throw Error(name);
  checks.push(name);
}
try {
  const page = await browser.newPage({
    viewport: { width: 1280, height: 720 },
    hasTouch: true,
  });
  page.on("pageerror", (e) => errors.push(String(e)));
  // An audio-thread tap retains every short click even if software rendering
  // stalls the main JS thread. Its output is silence, so it cannot alter the mix.
  await page.addInitScript(() => {
    const w = window as any;
    w.__countSamples = [];
    w.__countReady = false;
    const source = `class CountProbe extends AudioWorkletProcessor {
   constructor(){super();this.n=0;this.en=0;this.peak=0;}
   process(inputs){const a=inputs[0];if(a&&a.length){for(let i=0;i<a[0].length;i++){let e=0;for(const ch of a){const v=ch[i];e+=v*v;this.peak=Math.max(this.peak,Math.abs(v));}this.en+=e/a.length;this.n++;if(this.n>=Math.round(sampleRate*.01)){this.port.postMessage({t:currentTime+i/sampleRate,rms:Math.sqrt(this.en/this.n),peak:this.peak});this.n=0;this.en=0;this.peak=0;}}}return true;}
  }registerProcessor('eit-count-probe',CountProbe);`;
    const url = URL.createObjectURL(
      new Blob([source], { type: "text/javascript" }),
    );
    const connect = AudioNode.prototype.connect;
    let attached = false;
    (AudioNode.prototype as any).connect = function (
      dest: any,
      ...args: any[]
    ) {
      const result = (connect as any).call(this, dest, ...args);
      if (!attached && dest instanceof AudioDestinationNode) {
        attached = true;
        const input = this,
          ctx = this.context;
        ctx.audioWorklet
          .addModule(url)
          .then(() => {
            const tap = new AudioWorkletNode(ctx, "eit-count-probe");
            tap.port.onmessage = (e) => w.__countSamples.push(e.data);
            connect.call(input, tap);
            connect.call(tap, ctx.destination);
            w.__countContext = ctx;
            w.__countReady = true;
          })
          .catch((e) => (w.__countError = String(e)));
      }
      return result;
    };
  });
  await page.goto(url, { waitUntil: "load", timeout: 90000 });
  await page.locator("canvas").waitFor({ timeout: 60000 });
  await page.waitForFunction(
    () => (window as any).__countReady || (window as any).__countError,
    undefined,
    { timeout: 60000 },
  );
  check(
    !(await page.evaluate(() => (window as any).__countError)),
    "audio-thread probe ready",
  );
  await page.waitForTimeout(1200);
  const key = async (k: string) => {
    await page.keyboard.press(k, { delay: 100 });
    await page.waitForTimeout(250);
  };
  const click = async (x: number, y: number) => {
    await page.mouse.click(x, y, { delay: 100 });
    await page.waitForTimeout(400);
  };
  const mark = () =>
    page.evaluate(() => (window as any).__countContext.currentTime);
  const since = (t: number) =>
    page.evaluate(
      (t) =>
        (window as any).__countSamples
          .filter((s) => s.t >= t)
          .map((s) => ({ ...s, t: s.t - t })),
      t,
    );
  const settleAudio = async () => {
    const started = await mark();
    for (let i = 0; i < 40; i++) {
      await page.waitForTimeout(100);
      const now = await mark();
      const elapsed = now - started;
      if (elapsed < 0.3) continue;
      const recent = (await since(started)).filter(
        (s) => s.t >= elapsed - 0.25,
      );
      if (recent.every((s) => s.rms < 0.0001)) return;
    }
    throw Error("Audio did not become quiet after Settings");
  };
  const mean = (a: any[], lo: number, hi: number) => {
    const x = a.filter((v) => v.t >= lo && v.t <= hi);
    if (!x.length) throw Error("empty audio window");
    return x.reduce((s, v) => s + v.rms, 0) / x.length;
  };
  const groups = (a: any[]) => {
    const out: number[] = [];
    let last = -1;
    for (const s of a) {
      if (s.rms > 0.001) {
        if (s.t - last > 0.04) out.push(s.t);
        last = s.t;
      }
    }
    return out;
  };
  const capture = async () =>
    PNG.sync.read(Buffer.from(await page.screenshot()));
  const pixel = (img: any, x: number, y: number) =>
    [
      ...img.data.subarray(
        (y * img.width + x) * 4,
        (y * img.width + x) * 4 + 3,
      ),
    ].join(",");
  const card = (img: any) => pixel(img, 458, 260) === "44,82,91";
  const hasInk = (img: any, x0: number, y0: number, x1: number, y1: number) => {
    for (let y = y0; y < y1; y++)
      for (let x = x0; x < x1; x++)
        if (pixel(img, x, y) !== "242,238,226") return true;
    return false;
  };
  const back = async () => {
    await key("Escape");
    await key("ArrowDown");
    await key("ArrowDown");
    await key("Enter");
    await page.waitForTimeout(350);
  };
  const openSettings = async () => {
    await click(640, 660);
  };
  const closeSettings = async () => {
    await key("Escape");
    await page.waitForTimeout(600); // Let the settings select effect decay.
  };
  const toggleMusic = async (expectQuiet = false) => {
    await openSettings();
    await click(640, 256);
    await closeSettings();
    if (expectQuiet) await settleAudio();
  };
  const toggleSound = async (expectQuiet = false) => {
    await openSettings();
    await click(640, 192);
    await closeSettings();
    if (expectQuiet) await settleAudio();
  };
  await key("1");
  await page.waitForTimeout(400);
  await toggleMusic(true);
  await page.waitForTimeout(400); // Music off, effects on.
  let start = await mark();
  await page.mouse.click(640, 580, { delay: 60 });
  await page.waitForFunction(
    (t) => (window as any).__countSamples.some((s) => s.t > t && s.rms > 0.001),
    start,
    { timeout: 5000 },
  );
  await page.keyboard.down("ArrowUp");
  await page.keyboard.down("ShiftLeft");
  await page.keyboard.down("KeyS");
  await page.waitForTimeout(3300);
  let a = await since(start);
  measurements.push({
    name: "music-off count-in with held gameplay keys",
    samples: a,
  });
  let onsets = groups(a);
  await Bun.write(
    output + "/first-trace.json",
    JSON.stringify({ onsets, samples: a }, null, 2),
  );
  check(
    onsets.length === 3,
    "three preparation clicks; no queued rotate/drop sound",
  );
  check(
    onsets[1] - onsets[0] > 0.06 &&
      onsets[1] - onsets[0] < 1 &&
      onsets[2] - onsets[1] > 0.06 &&
      onsets[2] - onsets[1] < 1,
    "three distinct audio clicks without a simultaneous burst",
  );
  check(
    mean(a, onsets[2] + 0.35, onsets[2] + 1) < 0.0001,
    "GO preserves music-off setting",
  );
  await page.keyboard.up("ArrowUp");
  await page.keyboard.up("ShiftLeft");
  await page.keyboard.up("KeyS");
  check(!card(await capture()), "count-in clears and gameplay starts");
  start = await mark();
  await key("ShiftLeft");
  await page.waitForTimeout(500);
  a = await since(start);
  check(
    Math.max(...a.map((s) => s.peak)) > 0.06,
    "fresh drop works after preparation",
  );
  await key("Escape");
  await key("Enter");
  check(!card(await capture()), "ordinary Resume does not start a count-in");
  // Visual checks are separate from timing measurements: compositor screenshots
  // themselves can stall a software renderer and must not define the tempo oracle.
  await key("Escape");
  await key("ArrowDown");
  await key("Enter");
  const ready = await capture();
  check(card(ready), "Restart shows baton preparation");
  check(
    hasInk(ready, 900, 15, 1240, 45),
    "now-playing label visible during preparation",
  );
  await Bun.write(output + "/countin.png", PNG.sync.write(ready));
  await page.waitForTimeout(2500);
  await back();
  await toggleMusic();
  await toggleSound(true);
  await page.waitForTimeout(350); // Music on; master mute.
  start = await mark();
  await page.mouse.click(640, 580, { delay: 50 });
  const mutedImage = await capture();
  check(card(mutedImage), "master mute keeps visible preparation");
  await page.waitForTimeout(2500);
  a = await since(start);
  measurements.push({ name: "master-muted count-in", samples: a });
  check(
    Math.max(...a.map((s) => s.peak)) < 0.0001,
    "master mute silences preparation and music",
  );
  await back();
  await toggleSound();
  await page.waitForTimeout(400);
  start = await mark();
  await page.mouse.click(640, 580, { delay: 50 });
  await page.waitForTimeout(4000);
  a = await since(start);
  measurements.push({ name: "normal count-in and music", samples: a });
  // Locate the quiet gap between the last click and the musical downbeat without
  // assuming zero input/output latency in a particular browser/audio backend.
  onsets = groups(a);
  check(onsets.length >= 4, "three clicks followed by gameplay music");
  const first = onsets[0];
  check(
    onsets[3] > onsets[2] + 0.08,
    "music begins on the downbeat, not during preparation",
  );
  check(
    mean(a, onsets[2] + 0.08, onsets[3] - 0.02) < 0.0001,
    "music held between last click and GO",
  );
  check(
    mean(a, onsets[3] + 0.1, onsets[3] + 0.4) > 0.0003,
    "music playing after GO",
  );
  await key("Escape");
  await key("ArrowDown");
  await key("Enter");
  await key("Escape");
  const before = await page.screenshot();
  await page.waitForTimeout(850);
  const after = await page.screenshot();
  check(
    Buffer.compare(Buffer.from(before), Buffer.from(after)) === 0,
    "pause freezes preparation and board",
  );
  await key("ArrowDown");
  await key("ArrowDown");
  await key("Enter");
  await page.waitForTimeout(400);
  check(!card(await capture()), "return to lobby cancels preparation");
  start = await mark();
  await page.waitForTimeout(500);
  a = await since(start);
  check(mean(a, 0.1, 0.4) > 0.0003, "cancel restores lobby music");
  check(errors.length === 0, "no browser errors");
  const result = {
    passed: true,
    baseURL: url,
    checks,
    measurements,
    errors,
    subjectivelyAuditioned: false,
  };
  await Bun.write(output + "/report.json", JSON.stringify(result, null, 2));
  console.log(JSON.stringify({ passed: true, checks, errors }, null, 2));
} finally {
  await browser.close();
  server?.stop(true);
}
