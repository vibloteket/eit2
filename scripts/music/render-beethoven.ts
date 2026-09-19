// Selected Beethoven gameplay G1, 126 BPM. Reproduces the approved PCM unchanged.
import { mkdir } from "node:fs/promises";
import { createHash } from "node:crypto";
const root = import.meta.dir + "/../..";
const sourceDir = root + "/music/beethoven";
const output = root + "/dist/audio";
const cacheRoot = root + "/.tmp/music-sources";
const config = await Bun.file(sourceDir + "/render-config.json").json();
const manifest = await Bun.file(
  root + "/music/badinerie/sample-sources.json",
).json();
const hashes = new Map(manifest.files.map((f) => [f.path, f]));
const digest = (b: Uint8Array) => createHash("sha256").update(b).digest("hex");
await mkdir(output, { recursive: true });
// Public-domain Mutopia Beethoven source arranged in the established B2 palette.
// See sources.json for the score declaration and pinned sample hashes.
const referenceRms = config.referenceRms;
const fluteTrims = new Map<number, number>();
const trimReport: any[] = [];
await mkdir(cacheRoot, { recursive: true });
async function get(path: string) {
  const entry = hashes.get(path) as { url: string; sha256: string } | undefined;
  if (!entry) throw Error("Undeclared sample source: " + path);
  let file = Bun.file(cacheRoot + "/" + path);
  if (!(await file.exists())) {
    const response = await fetch(entry.url);
    if (!response.ok) throw Error(path + ": HTTP " + response.status);
    const data = new Uint8Array(await response.arrayBuffer());
    if (digest(data) !== entry.sha256)
      throw Error("Downloaded source hash mismatch: " + path);
    await Bun.write(cacheRoot + "/" + path, data);
    file = Bun.file(cacheRoot + "/" + path);
  }
  if (digest(new Uint8Array(await file.arrayBuffer())) !== entry.sha256)
    throw Error("Cached source hash mismatch: " + path);
  return file;
}
await get("LICENSE");
type Region = {
  sample: string;
  center: number;
  lo: number;
  hi: number;
  velLo: number;
  velHi: number;
};
async function instrument(name: string, velocity: number) {
  const text = await (await get(name + ".sfz")).text();
  const path = text
    .match(/default_path=(.*)/)![1]
    .trim()
    .replaceAll("\\", "/");
  const out: Region[] = [];
  for (const block of text.split("<region>").slice(1)) {
    const val = (k: string, d = 0) =>
      Number(block.match(new RegExp("\\b" + k + "=([^\\s]+)"))?.[1] ?? d);
    const sample = block.match(/sample=([^\r\n]+)/)?.[1].trim();
    if (!sample) continue;
    const r = {
      sample: path + sample.replaceAll("\\", "/"),
      center: val("pitch_keycenter"),
      lo: val("lokey"),
      hi: val("hikey"),
      velLo: val("lovel"),
      velHi: val("hivel", 127),
    };
    if (
      velocity >= r.velLo &&
      velocity <= r.velHi &&
      !out.some((x) => x.center === r.center)
    )
      out.push(r);
  }
  if (!out.length) throw new Error("no regions");
  return out;
}
function wavRead(buffer: ArrayBuffer) {
  const b = Buffer.from(buffer);
  if (b.toString("ascii", 0, 4) !== "RIFF") throw new Error("not RIFF");
  let fmt: any, data: Buffer | undefined;
  for (let p = 12; p + 8 <= b.length; ) {
    const id = b.toString("ascii", p, p + 4),
      n = b.readUInt32LE(p + 4);
    if (id === "fmt ")
      fmt = {
        kind: b.readUInt16LE(p + 8),
        channels: b.readUInt16LE(p + 10),
        rate: b.readUInt32LE(p + 12),
        bits: b.readUInt16LE(p + 22),
      };
    if (id === "data") data = b.subarray(p + 8, p + 8 + n);
    p += 8 + n + (n % 2);
  }
  if (!fmt || !data) throw new Error("bad WAV");
  const { kind, channels, rate, bits } = fmt;
  const bytes = bits / 8,
    n = data.length / bytes / channels;
  const samples = new Float32Array(n);
  for (let i = 0; i < n; i++) {
    let v = 0;
    for (let c = 0; c < channels; c++) {
      const o = (i * channels + c) * bytes;
      v +=
        kind === 3
          ? data.readFloatLE(o)
          : bits === 16
            ? data.readInt16LE(o) / 32768
            : bits === 24
              ? data.readIntLE(o, 3) / 8388608
              : bits === 32
                ? data.readInt32LE(o) / 2147483648
                : (data[o] - 128) / 128;
    }
    samples[i] = v / channels;
  }
  return { samples, rate };
}
const cache = new Map<string, ReturnType<typeof wavRead>>();
const flute = await instrument("FluteStac", 80),
  cello = await instrument("CelloEnsPizz", 50);
const score = await Bun.file(sourceDir + "/score.json").json();
const baseChanges = config.baseCalibration;
const notes = [score.melody.filter((n) => n.instrument === "flute")];
const sr = 44100,
  bpm = 126,
  beat = 60 / bpm,
  phraseBeats = 224;
const cycleFrames = Math.round((phraseBeats * sr * 60) / bpm);
if (!Number.isInteger(cycleFrames)) throw Error("Nonintegral loop frames");
const seconds = 2 * phraseBeats * beat + 0.5;
// Correct only a quarter of the sample-to-sample level variation, capped at +/-0.8dB.
const keyLevels: any[] = [];
for (const key of [...new Set(notes.flat().map((n) => n.key))]) {
  const r = flute.find((r) => key >= r.lo && key <= r.hi);
  if (!r) throw new Error("Missing flute key");
  let wav = cache.get(r.sample);
  if (!wav) {
    wav = wavRead(await (await get(r.sample)).arrayBuffer());
    cache.set(r.sample, wav);
  }
  let peak = 0;
  for (const x of wav.samples) peak = Math.max(peak, Math.abs(x));
  const ratio = (2 ** ((key - r.center) / 12) * wav.rate) / sr;
  let energy = 0;
  const count = Math.min(
    Math.floor(wav.samples.length / ratio) - 1,
    Math.floor(0.16 * sr),
  );
  for (let i = 0; i < count; i++) {
    const pos = i * ratio,
      k = Math.floor(pos),
      f = pos - k;
    const x =
      (wav.samples[k] * (1 - f) + wav.samples[k + 1] * f) /
      Math.max(peak, 0.01);
    energy += x * x;
  }
  keyLevels.push({
    key,
    normalizedRms: Math.sqrt(energy / count),
    sample: r.sample,
  });
}
const sorted = baseChanges.fluteTrims
    .map((k) => k.normalizedRms)
    .sort((a, b) => a - b),
  median = sorted[Math.floor(sorted.length / 2)];
for (const item of keyLevels) {
  const trimDB =
    baseChanges.fluteTrims.find((k) => k.key === item.key)?.trimDB ??
    Math.max(
      -0.8,
      Math.min(0.8, 20 * Math.log10(median / item.normalizedRms) * 0.25),
    );
  fluteTrims.set(item.key, 10 ** (trimDB / 20));
  trimReport.push({ ...item, trimDB });
}
async function sampled(
  L: Float32Array,
  R: Float32Array,
  regions: Region[],
  key: number,
  start: number,
  dur: number,
  gain: number,
  pan: number,
) {
  let r =
    regions.find((r) => key >= r.lo && key <= r.hi) ??
    regions.reduce((a, b) =>
      Math.abs(a.center - key) < Math.abs(b.center - key) ? a : b,
    );
  let wav = cache.get(r.sample);
  if (!wav) {
    wav = wavRead(await (await get(r.sample)).arrayBuffer());
    cache.set(r.sample, wav);
  }
  let peak = 0;
  for (const x of wav.samples) peak = Math.max(peak, Math.abs(x));
  const ratio = (2 ** ((key - r.center) / 12) * wav.rate) / sr;
  const isFlute = regions === flute,
    releaseTime = isFlute ? 0.06 : 0.055;
  if (isFlute) gain *= fluteTrims.get(key) ?? 1;
  // Remove 50ms of low-flute lead-in, as in the preferred comparison. Same pitches, envelopes and gain.
  const sampleOffset =
    isFlute && r.sample.endsWith("LDFlute_stac_A3_v2_rr1.wav")
      ? Math.round(0.05 * wav.rate)
      : 0;
  let begin = Math.round(start * sr),
    count = Math.min(
      Math.floor((wav.samples.length - sampleOffset) / ratio) - 1,
      Math.floor((dur + releaseTime) * sr),
    );
  const lg = Math.sqrt((1 - pan) / 2),
    rg = Math.sqrt((1 + pan) / 2);
  for (let i = 0; i < count && begin + i < L.length; i++) {
    let pos = sampleOffset + i * ratio,
      k = Math.floor(pos),
      f = pos - k,
      t = i / sr;
    const phase = Math.min(1, Math.max(0, (t - dur) / releaseTime));
    const release = isFlute ? 0.5 + 0.5 * Math.cos(Math.PI * phase) : 1 - phase;
    const attack = isFlute
      ? 0.5 - 0.5 * Math.cos(Math.PI * Math.min(1, t / 0.005))
      : Math.min(1, t / 0.004);
    const v =
      (((wav.samples[k] * (1 - f) + wav.samples[k + 1] * f) * gain) /
        Math.max(peak, 0.01)) *
      release *
      attack;
    L[begin + i] += v * lg;
    R[begin + i] += v * rg;
  }
}
function pluck(
  L: Float32Array,
  R: Float32Array,
  key: number,
  start: number,
  gain: number,
) {
  const freq = 440 * 2 ** ((key - 69) / 12),
    begin = Math.round(start * sr);
  for (let i = 0; i < sr * 0.35 && begin + i < L.length; i++) {
    const t = i / sr;
    const v =
      (Math.sin(2 * Math.PI * freq * t) +
        0.24 * Math.sin(4 * Math.PI * freq * t) +
        0.1 * Math.sin(6 * Math.PI * freq * t)) *
      Math.exp(-t * 15) *
      Math.min(1, t / 0.003) *
      gain;
    L[begin + i] += v * 0.58;
    R[begin + i] += v * 0.82;
  }
}

function ambience(l: Float32Array, r: Float32Array, circular: boolean) {
  for (const [a, b] of [
    [l, r],
    [r, l],
  ]) {
    const dry = a.slice();
    for (const d of [0.037, 0.061, 0.089]) {
      const offset = Math.floor(d * sr);
      for (let i = circular ? 0 : offset; i < a.length; i++)
        b[i] += dry[(i - offset + a.length) % a.length] * 0.035;
    }
  }
}

function wav(l: Float32Array, r: Float32Array, master: number) {
  const data = Buffer.alloc(44 + l.length * 4);
  data.write("RIFF", 0);
  data.writeUInt32LE(data.length - 8, 4);
  data.write("WAVEfmt ", 8);
  data.writeUInt32LE(16, 16);
  data.writeUInt16LE(1, 20);
  data.writeUInt16LE(2, 22);
  data.writeUInt32LE(sr, 24);
  data.writeUInt32LE(sr * 4, 28);
  data.writeUInt16LE(4, 32);
  data.writeUInt16LE(16, 34);
  data.write("data", 36);
  data.writeUInt32LE(l.length * 4, 40);
  for (let i = 0; i < l.length; i++)
    for (const [c, a] of [l, r].entries()) {
      const v = a[i] * master;
      if (!Number.isFinite(v) || Math.abs(v) >= 1)
        throw Error("Clipping/nonfinite sample");
      data.writeInt16LE(Math.round(v * 32767), 44 + i * 4 + c * 2);
    }
  return data;
}

type Buffers = { L: Float32Array; R: Float32Array };
const mix = { flute: 0.245, lowMelodyPluck: 0.05, bass: 0.1, pluck: 0.028 };
const frameAt = (b: number) =>
  Math.round((Math.round(b * 384) * sr * 60) / (bpm * 384));
const sec = (b: number) => frameAt(b) / sr;
const usage = { flute: 0, lowMelody: 0, bass: 0, pluck: 0 };
async function section(
  buffer: Buffers,
  kind: "A" | "B",
  at: number,
  repeat: boolean,
) {
  const from = kind === "A" ? 0 : 48,
    to = kind === "A" ? 48 : 112;
  for (const n of score.melody) {
    if (n.at < from || n.at >= to) continue;
    const start = sec(at + n.at - from),
      beats = Math.min(n.beats, to - n.at);
    if (n.instrument === "flute") {
      await sampled(
        buffer.L,
        buffer.R,
        flute,
        n.key,
        start,
        beats * beat * (beats <= 0.25 ? 0.88 : 0.78),
        mix.flute * (Number.isInteger(n.at) ? 1 : 0.93),
        -0.08,
      );
      usage.flute++;
    } else {
      pluck(buffer.L, buffer.R, n.key, start, mix.lowMelodyPluck);
      usage.lowMelody++;
    }
  }
  for (const n of score.bass) {
    if (n.at < from || n.at >= to) continue;
    if (
      repeat &&
      Math.floor((n.at - from) / 2) % 4 === 3 &&
      (n.at - from) % 2 === 1
    )
      continue;
    await sampled(
      buffer.L,
      buffer.R,
      cello,
      n.key,
      sec(at + n.at - from),
      0.22,
      mix.bass * (n.at % 2 === 0 ? 1 : 0.74),
      0.15,
    );
    usage.bass++;
  }
  for (const n of score.pluck) {
    if (n.at < from || n.at >= to) continue;
    if (
      repeat &&
      Math.floor((n.at - from) / 2) % 4 === 3 &&
      (n.at - from) % 2 === 1.5
    )
      continue;
    pluck(
      buffer.L,
      buffer.R,
      n.key,
      sec(at + n.at - from),
      mix.pluck * (repeat ? 0.97 : 1),
    );
    usage.pluck++;
  }
  return at + to - from;
}
const frames = cycleFrames * 2 + sr,
  L = new Float32Array(frames),
  R = new Float32Array(frames);
const form = [];
for (let pass = 0; pass < 2; pass++) {
  let at = pass * phraseBeats;
  for (let i = 0; i < 4; i++) {
    const before = at,
      kind = score.form[i];
    at = await section({ L, R }, kind, at, i % 2 === 1);
    if (pass === 0)
      form.push({
        section: kind,
        repeat: i % 2 === 1,
        fromSeconds: before * beat,
        toSeconds: at * beat,
      });
  }
  if (at !== (pass + 1) * phraseBeats) throw Error("Wrong form length");
}
const dryL = L.slice(0, cycleFrames),
  dryR = R.slice(0, cycleFrames);
for (let i = 0; i < sr; i++) {
  dryL[i] += L[2 * cycleFrames + i];
  dryR[i] += R[2 * cycleFrames + i];
}
let dryError = 0;
for (let i = 0; i < cycleFrames; i++)
  dryError = Math.max(
    dryError,
    Math.abs(dryL[i] - L[cycleFrames + i]),
    Math.abs(dryR[i] - R[cycleFrames + i]),
  );
if (dryError > 2e-7) throw Error("Dry cycle mismatch: " + dryError);
ambience(L, R, false);
ambience(dryL, dryR, true);
const loopL = L.slice(cycleFrames, 2 * cycleFrames),
  loopR = R.slice(cycleFrames, 2 * cycleFrames);
let error = 0,
  peak = 0,
  sum = 0;
for (let i = 0; i < cycleFrames; i++) {
  error = Math.max(
    error,
    Math.abs(loopL[i] - dryL[i]),
    Math.abs(loopR[i] - dryR[i]),
  );
  peak = Math.max(peak, Math.abs(loopL[i]), Math.abs(loopR[i]));
  sum += loopL[i] * loopL[i] + loopR[i] * loopR[i];
}
if (error > 2e-7) throw Error("Room tail mismatch: " + error);
const rawRms = Math.sqrt(sum / (cycleFrames * 2));
// Preserve the exact G1 master gain: tempo is the only musical/mix change.
const gain = config.masterGain;
if (peak * gain >= 0.9)
  throw Error("Insufficient headroom with unchanged master gain");
const loop = wav(loopL, loopR, gain);
if (digest(loop) !== config.expectedLoopSHA256)
  throw Error("Gameplay loop hash differs; game asset not replaced");
await Bun.write(output + "/beethoven-gameplay-G1-126-loop.wav", loop);
// Short audition covers both contrasting sections once, with an actual ending tail.
const previewFrames = frameAt(112) + sr,
  p = {
    L: new Float32Array(previewFrames),
    R: new Float32Array(previewFrames),
  };
let at = await section(p, "A", 0, false);
at = await section(p, "B", at, false);
if (at !== 112) throw Error("Preview length");
ambience(p.L, p.R, false);
const preview = wav(p.L, p.R, gain);
if (digest(preview) !== config.expectedPreviewSHA256)
  throw Error("Gameplay preview hash differs; game asset not replaced");
await Bun.write(output + "/beethoven-gameplay-G1-126-preview.wav", preview);
await Bun.write(root + "/dist/audio/legacy/gameplay-beethoven.wav", loop);
const report = {
  title: score.title,
  bpm,
  loopBeats: 224,
  loopSeconds: cycleFrames / sr,
  previewSeconds: previewFrames / sr,
  form,
  mix,
  melodyEvents: score.melody.length,
  notesRoutedToPluck: score.melody.filter((n) => n.instrument === "pluck")
    .length,
  uniqueSamples: cache.size,
  dryError,
  circularError: error,
  peakDBFS: 20 * Math.log10(peak * gain),
  rmsDBFS: 20 * Math.log10(rawRms * gain),
  gain,
  loopSHA256: digest(loop),
  previewSHA256: digest(preview),
  sourceLicense: "Mutopia score explicitly Public Domain; VSCO samples CC0",
  expectedHashesVerified: true,
  userSelectedPCM: true,
  assistantAuditioned: false,
};
await Bun.write(
  output + "/beethoven-render-report.json",
  JSON.stringify(report, null, 2),
);
console.log(JSON.stringify(report, null, 2));
