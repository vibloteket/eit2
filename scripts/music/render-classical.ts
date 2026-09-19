// Reproduce the four approved classical gameplay arrangements as periodic PCM loops.
// Historical score data and CC0 samples are pinned under music/classical.
import { mkdir } from "node:fs/promises";
import { createHash } from "node:crypto";
const repo = import.meta.dir + "/../..",
  root = repo + "/music/classical",
  out = repo + "/dist/audio/classical",
  sr = 44100,
  bpm = 112,
  beat = 60 / bpm;
await mkdir(out, { recursive: true });
const oldManifest = await Bun.file(
  repo + "/music/badinerie/sample-sources.json",
).json();
const priorManifest = await Bun.file(root + "/sample-sources.json").json();
const expected = await Bun.file(root + "/manifest.json").json();
if (priorManifest.revision !== oldManifest.revision)
  throw Error("Sample revisions disagree");
const declared = new Map(
  [...oldManifest.files, ...priorManifest.files].map((f: any) => [f.path, f]),
);
const used = new Map<string, any>(
    priorManifest.files.map((f: any) => [f.path, f]),
  ),
  cacheRoot = repo + "/.tmp/music-sources";
await mkdir(cacheRoot, { recursive: true });
const hash = (b: Uint8Array) => createHash("sha256").update(b).digest("hex");
for (const input of expected.inputs) {
  if (hash(await Bun.file(root + "/" + input.path).bytes()) !== input.sha256)
    throw Error("Notation/source hash mismatch: " + input.path);
}
async function source(path: string) {
  const url = `https://raw.githubusercontent.com/sgossner/VSCO-2-CE/${oldManifest.revision}/${path.split("/").map(encodeURIComponent).join("/")}`;
  const entry: any = declared.get(path),
    local = Bun.file(cacheRoot + "/" + path);
  if (!entry || entry.license !== "CC0-1.0")
    throw Error("Undeclared or unlicensed source " + path);
  let data: Uint8Array;
  if (await local.exists()) data = await local.bytes();
  else {
    const r = await fetch(url);
    if (!r.ok) throw Error("Sample fetch " + r.status + " " + path);
    data = new Uint8Array(await r.arrayBuffer());
    if (entry.sha256 !== hash(data))
      throw Error("Sample hash mismatch " + path);
    await Bun.write(local, data);
  }
  if (entry.sha256 !== hash(data)) throw Error("Sample hash mismatch " + path);
  used.set(path, {
    path,
    url,
    sha256: hash(data),
    bytes: data.length,
    license: "CC0-1.0",
  });
  return data;
}
function wavRead(bytes: Uint8Array) {
  const b = Buffer.from(bytes);
  if (
    b.toString("ascii", 0, 4) !== "RIFF" ||
    b.toString("ascii", 8, 12) !== "WAVE"
  )
    throw Error("Not WAV");
  let data!: Buffer,
    channels = 0,
    rate = 0,
    bits = 0,
    fmt = 0;
  for (let p = 12; p + 8 <= b.length; ) {
    let n = b.readUInt32LE(p + 4),
      id = b.toString("ascii", p, p + 4);
    if (id === "fmt ") {
      fmt = b.readUInt16LE(p + 8);
      channels = b.readUInt16LE(p + 10);
      rate = b.readUInt32LE(p + 12);
      bits = b.readUInt16LE(p + 22);
    }
    if (id === "data") data = b.subarray(p + 8, p + 8 + n);
    p += 8 + n + (n % 2);
  }
  if (!data || ![1, 3].includes(fmt)) throw Error("Unsupported WAV");
  const out = new Float32Array(data.length / (bits / 8) / channels);
  let peak = 0;
  for (let i = 0; i < out.length; i++) {
    let v = 0;
    for (let c = 0; c < channels; c++) {
      const o = ((i * channels + c) * bits) / 8;
      v +=
        fmt === 3
          ? data.readFloatLE(o)
          : bits === 16
            ? data.readInt16LE(o) / 32768
            : bits === 24
              ? data.readIntLE(o, 3) / 8388608
              : bits === 32
                ? data.readInt32LE(o) / 2147483648
                : (data[o] - 128) / 128;
    }
    out[i] = v / channels;
    peak = Math.max(peak, Math.abs(out[i]));
  }
  return { samples: out, rate, peak };
}
type Region = { lo: number; hi: number; center: number; path: string };
async function instrument(name: string, velocity: number) {
  const text = new TextDecoder().decode(await source(name + ".sfz"));
  const base =
    text
      .match(/default_path=(.*)/)?.[1]
      .trim()
      .replaceAll("\\", "/") ?? "";
  const regions: Region[] = [];
  for (const chunk of text.split("<region>").slice(1)) {
    const kv = Object.fromEntries(
      [...chunk.matchAll(/^(\w+)=(.*?)\s*$/gm)].map((m) => [m[1], m[2]]),
    );
    if (
      !kv.sample ||
      velocity < +(kv.lovel ?? 0) ||
      velocity > +(kv.hivel ?? 127)
    )
      continue;
    const center = +(kv.pitch_keycenter ?? kv.key);
    if (regions.some((r) => r.center === center)) continue;
    regions.push({
      lo: +(kv.lokey ?? kv.key),
      hi: +(kv.hikey ?? kv.key),
      center,
      path: base + kv.sample.replaceAll("\\", "/"),
    });
  }
  if (!regions.length) throw Error("Empty instrument " + name);
  return regions;
}
await source("LICENSE");
const flute = await instrument("FluteStac", 80),
  sustain = await instrument("FluteSusNV", 80),
  cello = await instrument("CelloEnsPizz", 50);
const samples = new Map<string, ReturnType<typeof wavRead>>();
async function sample(r: Region) {
  if (!samples.has(r.path)) samples.set(r.path, wavRead(await source(r.path)));
  return samples.get(r.path)!;
}
function region(list: Region[], key: number) {
  return (
    list.find((r) => key >= r.lo && key <= r.hi) ??
    list.reduce((a, b) =>
      Math.abs(a.center - key) <= Math.abs(b.center - key) ? a : b,
    )
  );
}
function midi(s: string): number | null {
  if (s === "r") return null;
  const m = s.match(/^([A-G])([#b]*)(-?\d+)$/);
  if (!m) throw Error("Invalid pitch " + s);
  return (
    12 * (+m[3] + 1) +
    ({ C: 0, D: 2, E: 4, F: 5, G: 7, A: 9, B: 11 } as any)[m[1]] +
    [...m[2]].reduce((s, c) => s + (c === "#" ? 1 : -1), 0)
  );
}
const score: any = {};
for (const id of ["A1", "A2", "B", "C", "D"])
  score[id] = await Bun.file(root + "/scores/" + id + ".json").json();
const bars = (id: string) => score[id].segments[0].bars;
const selections = [
  {
    id: "A",
    file: "gameplay-bach-bourrees.wav",
    title: "Bach — BWV1067 Bourrée I/II",
    tonic: 47,
    sequence: [...bars("A1"), ...bars("A2"), ...bars("A2"), ...bars("A1")],
  },
  {
    id: "B",
    file: "gameplay-handel-allegro.wav",
    title: "Handel — HWV369 IV Allegro",
    tonic: 41,
    sequence: [...bars("B"), ...bars("B")],
  },
  {
    id: "C",
    file: "gameplay-bach-sonata.wav",
    title: "Bach — BWV1035 II Allegro",
    tonic: 40,
    sequence: [
      ...bars("C").filter((b: any) => b.n !== 0),
      ...bars("C").filter((b: any) => b.n !== 0),
      ...bars("C").filter((b: any) => b.n !== 0),
    ],
  },
  {
    id: "D",
    file: "gameplay-vivaldi-allegro.wav",
    title: "Vivaldi — RV428 III Allegro, solo episode",
    tonic: 38,
    sequence: [...bars("D"), ...bars("D")],
  },
];
type BufferPair = { L: Float32Array; R: Float32Array };
const buffer = (frames: number): BufferPair => ({
  L: new Float32Array(frames),
  R: new Float32Array(frames),
});
const reports: any[] = [];
for (const spec of selections) {
  if (Bun.argv[2] && spec.id !== Bun.argv[2]) continue;
  const beats = spec.sequence.reduce((s: number, b: any) => s + b.beats, 0),
    seconds = beats * beat,
    cycleFrames = Math.round(seconds * sr),
    frames = cycleFrames + Math.ceil(0.8 * sr),
    lead = buffer(frames),
    back = buffer(frames);
  if (beats !== 96 || cycleFrames !== 2268000)
    throw Error("Unexpected musical loop length");
  const events: {
      key: number;
      at: number;
      dur: number;
      bar: number;
      offset: number;
    }[] = [],
    bassEvents: { key: number | null; at: number; dur: number }[] = [];
  let at = 0;
  for (const bar of spec.sequence) {
    for (const voice of ["lead", "bass"]) {
      let offset = 0;
      for (const [p, d] of bar[voice]) {
        const key = midi(p);
        if (voice === "lead" && key !== null)
          events.push({ key, at: at + offset, dur: d, bar: bar.n, offset });
        if (voice === "bass") bassEvents.push({ key, at: at + offset, dur: d });
        offset += d;
      }
      if (Math.abs(offset - bar.beats) > 1e-8)
        throw Error("Bar duration mismatch");
    }
    at += bar.beats;
  }
  // Only the explicitly transcribed cross-bar tie; repeated same-pitch attacks stay separate.
  if (spec.id === "C")
    for (let i = 0; i < events.length - 1; i++) {
      const a = events[i],
        b = events[i + 1];
      if (
        a.bar === 5 &&
        b.bar === 6 &&
        a.key === b.key &&
        Math.abs(a.at + a.dur - b.at) < 1e-8
      ) {
        a.dur += b.dur;
        events.splice(i + 1, 1);
      }
    }
  const stats = {
      staccato: 0,
      sustained: 0,
      bass: 0,
      pluck: 0,
      fluteSampleShortages: 0,
    },
    articulation: any[] = [];
  async function play(
    dst: BufferPair,
    key: number,
    start: number,
    dur: number,
    kind: "flute" | "bass",
    gain: number,
  ) {
    let long = kind === "flute" && dur >= 0.38,
      r = region(kind === "bass" ? cello : long ? sustain : flute, key),
      s = await sample(r);
    let offset =
      kind === "flute" && !long && r.path.endsWith("LDFlute_stac_A3_v2_rr1.wav")
        ? Math.round(0.05 * s.rate)
        : 0;
    let ratio = (2 ** ((key - r.center) / 12) * s.rate) / sr,
      release = kind === "flute" ? 0.06 : 0.065;
    if (
      kind === "flute" &&
      !long &&
      (s.samples.length - offset) / ratio / sr < dur + release
    ) {
      long = true;
      r = region(sustain, key);
      s = await sample(r);
      offset = 0;
      ratio = (2 ** ((key - r.center) / 12) * s.rate) / sr;
    }
    if (long) {
      let onset = 0;
      while (
        onset < Math.min(s.samples.length, 0.15 * s.rate) &&
        Math.abs(s.samples[onset]) < s.peak * 0.03
      )
        onset++;
      offset = Math.max(0, onset - Math.round(0.008 * s.rate));
      gain *= 0.73;
    }
    const available = (s.samples.length - offset - 2) / ratio / sr;
    if (kind === "flute" && available < dur + release) {
      stats.fluteSampleShortages++;
      throw Error(
        `Exhausted flute sample ${r.path}: ${available}<${dur + release}`,
      );
    }
    if (kind === "flute") {
      stats[long ? "sustained" : "staccato"]++;
      articulation.push({
        key,
        start,
        duration: dur,
        available,
        sample: r.path,
        sustain: long,
      });
    } else stats.bass++;
    const begin = Math.round(start * sr),
      count = Math.min(
        Math.floor((dur + release) * sr),
        Math.floor((s.samples.length - offset - 2) / ratio),
      ),
      pan = kind === "flute" ? -0.08 : -0.18,
      lg = Math.sqrt((1 - pan) / 2),
      rg = Math.sqrt((1 + pan) / 2);
    for (let i = 0; i < count && begin + i < frames; i++) {
      const p = offset + i * ratio,
        k = Math.floor(p),
        f = p - k,
        t = i / sr,
        rel =
          0.5 +
          0.5 *
            Math.cos(Math.PI * Math.min(1, Math.max(0, (t - dur) / release))),
        attack = Math.min(1, t / (long ? 0.012 : 0.005));
      const v =
        ((s.samples[k] * (1 - f) + s.samples[k + 1] * f) /
          Math.max(0.01, s.peak)) *
        gain *
        attack *
        rel;
      dst.L[begin + i] += v * lg;
      dst.R[begin + i] += v * rg;
    }
  }
  function pluck(key: number, start: number, gain: number) {
    const hz = 440 * 2 ** ((key - 69) / 12),
      begin = Math.round(start * sr);
    for (let i = 0; i < sr * 0.48 && begin + i < frames; i++) {
      const t = i / sr,
        v =
          (Math.sin(2 * Math.PI * hz * t) +
            0.24 * Math.sin(4 * Math.PI * hz * t) +
            0.1 * Math.sin(6 * Math.PI * hz * t)) *
          Math.exp(-11 * t) *
          Math.min(1, t / 0.003) *
          gain;
      back.L[begin + i] += v * 0.58;
      back.R[begin + i] += v * 0.82;
    }
    stats.pluck++;
  }
  for (const e of events)
    await play(
      lead,
      e.key,
      e.at * beat,
      e.dur * beat * (e.dur >= 0.75 ? 0.96 : 0.9),
      "flute",
      0.245 * (e.offset === 0 ? 1 : 0.94),
    );
  let lastBass = spec.tonic + 12;
  for (let t = 0; t < beats - 1e-8; t += 0.5) {
    const b = bassEvents.find(
      (e) => e.at <= t + 1e-8 && e.at + e.dur > t + 1e-8,
    );
    if (b?.key !== null && b?.key !== undefined) lastBass = b.key;
    let low = lastBass - 12;
    while (low < 36) low += 12;
    while (low > 59) low -= 12;
    const bassBeat =
      spec.id === "B"
        ? Math.abs(t / 1.5 - Math.round(t / 1.5)) < 1e-8
        : Number.isInteger(t);
    if (bassBeat)
      await play(
        back,
        low,
        t * beat,
        0.26,
        "bass",
        0.075 * (Number.isInteger(t / 2) ? 1 : 0.87),
      );
    const e = events.find((e) => e.at <= t + 1e-8 && e.at + e.dur > t + 1e-8);
    let upper = Math.round(t * 2) % 2 && e ? e.key - 12 : lastBass;
    while (upper < 55) upper += 12;
    while (upper > 71) upper -= 12;
    pluck(upper, t * beat, 0.026 * (Math.round(t * 2) % 2 ? 0.84 : 1));
  }
  // Fold every release/pluck tail into the next cycle; no fade or zero padding in loops.
  for (const pair of [lead, back])
    for (const a of [pair.L, pair.R])
      for (let i = cycleFrames; i < frames; i++) a[i - cycleFrames] += a[i];
  const mix = buffer(cycleFrames);
  for (let i = 0; i < cycleFrames; i++) {
    mix.L[i] = lead.L[i] + back.L[i];
    mix.R[i] = lead.R[i] + back.R[i];
  }
  for (const [a, b] of [
    [mix.L, mix.R],
    [mix.R, mix.L],
  ]) {
    const dry = a.slice();
    for (const d of [0.037, 0.061, 0.089]) {
      const off = Math.round(d * sr);
      for (let i = 0; i < cycleFrames; i++)
        b[i] += dry[(i - off + cycleFrames) % cycleFrames] * 0.035;
    }
  }
  let energy = 0,
    peak = 0;
  for (let i = 0; i < cycleFrames; i++) {
    peak = Math.max(peak, Math.abs(mix.L[i]), Math.abs(mix.R[i]));
    energy += (mix.L[i] ** 2 + mix.R[i] ** 2) / 2;
  }
  const rms = Math.sqrt(energy / Math.floor(seconds * sr)),
    gain = Math.min(0.06 / rms, 0.85 / peak),
    wav = Buffer.alloc(44 + cycleFrames * 4);
  wav.write("RIFF");
  wav.writeUInt32LE(wav.length - 8, 4);
  wav.write("WAVEfmt ", 8);
  wav.writeUInt32LE(16, 16);
  wav.writeUInt16LE(1, 20);
  wav.writeUInt16LE(2, 22);
  wav.writeUInt32LE(sr, 24);
  wav.writeUInt32LE(sr * 4, 28);
  wav.writeUInt16LE(4, 32);
  wav.writeUInt16LE(16, 34);
  wav.write("data", 36);
  wav.writeUInt32LE(cycleFrames * 4, 40);
  for (let i = 0; i < cycleFrames; i++)
    for (const [c, a] of [mix.L, mix.R].entries()) {
      const v = a[i] * gain;
      if (!Number.isFinite(v) || Math.abs(v) >= 1)
        throw Error("Clipping/nonfinite");
      wav.writeInt16LE(Math.round(v * 32767), 44 + i * 4 + c * 2);
    }
  let minimumBackingRms = Infinity;
  for (
    let begin = Math.round(0.1 * sr);
    begin + 4410 < seconds * sr;
    begin += 2205
  ) {
    let en = 0;
    for (let i = begin; i < begin + 4410; i++)
      en += (back.L[i] ** 2 + back.R[i] ** 2) / 2;
    minimumBackingRms = Math.min(
      minimumBackingRms,
      Math.sqrt(en / 4410) * gain,
    );
  }
  const entry = expected.tracks.find((t) => t.id === spec.id);
  if (!entry || entry.sha256 !== hash(wav))
    throw Error("Rendered hash differs from accepted loop " + spec.id);
  await Bun.write(repo + "/internal/sound/audio/" + spec.file, wav);
  const report = {
    id: spec.id,
    title: spec.title,
    file: spec.file,
    bpm,
    beats,
    seconds: cycleFrames / sr,
    frames: cycleFrames,
    activeSeconds: seconds,
    rms: rms * gain,
    peak: peak * gain,
    minimumBacking100msRms: minimumBackingRms,
    stats,
    sha256: hash(wav),
    articulation,
  };
  await Bun.write(
    out + "/" + spec.id + "-verification.json",
    JSON.stringify(report, null, 2),
  );
  reports.push({ ...report, articulation: undefined });
  console.log(JSON.stringify(reports.at(-1)));
}
await Bun.write(
  out + "/render-report.json",
  JSON.stringify({ tracks: reports, sampleFiles: [...used.values()] }, null, 2),
);
