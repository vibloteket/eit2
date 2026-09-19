// Independent checks of committed source data and the four embedded periodic loops.
import { createHash } from "node:crypto";
const repo = import.meta.dir + "/../..",
  dir = repo + "/music/classical",
  hash = (b: Uint8Array) => createHash("sha256").update(b).digest("hex");
const manifest = await Bun.file(dir + "/manifest.json").json();
let checks = 0;
function check(ok: boolean, why: string) {
  checks++;
  if (!ok) throw Error(why);
}
check(
  manifest.bpm === 112 &&
    manifest.quarterBeats === 96 &&
    manifest.frames === 2268000,
  "loop timing contract",
);
check(manifest.tracks.length === 4, "four tracks");
for (const input of manifest.inputs)
  check(
    hash(await Bun.file(dir + "/" + input.path).bytes()) === input.sha256,
    "source hash " + input.path,
  );
for (const id of ["A1", "A2", "B", "C", "D"]) {
  const s = await Bun.file(dir + "/scores/" + id + ".json").json();
  check(!s.uncertain?.length, id + " unresolved notation");
  for (const seg of s.segments)
    for (const bar of seg.bars)
      for (const voice of ["lead", "bass"]) {
        let sum = 0;
        for (const [p, d] of bar[voice]) {
          check(p === "r" || /^[A-G][#b]*-?\d+$/.test(p), id + " pitch");
          check(Number.isFinite(d) && d > 0, id + " duration");
          sum += d;
        }
        check(
          Math.abs(sum - bar.beats) < 1e-8,
          `${id} ${bar.n} ${voice} bar sum`,
        );
      }
}
const historical = await Bun.file(
    dir + "/audits/handel-historical.json",
  ).json(),
  hand = await Bun.file(dir + "/scores/B.json").json();
check(
  historical.uncertain.length === 0 && historical.bars.length === 8,
  "Handel historical transcription",
);
for (let i = 0; i < 8; i++)
  for (const v of ["lead", "bass"])
    check(
      JSON.stringify(historical.bars[i][v]) ===
        JSON.stringify(hand.segments[0].bars[i][v]),
      "Handel production must use independent historical data",
    );
const names = [
  "gameplay-bach-bourrees.wav",
  "gameplay-handel-allegro.wav",
  "gameplay-bach-sonata.wav",
  "gameplay-vivaldi-allegro.wav",
];
for (const [index, t] of manifest.tracks.entries()) {
  check(t.file === names[index], "track order");
  const b = Buffer.from(
    await Bun.file(repo + "/internal/sound/audio/" + t.file).arrayBuffer(),
  );
  check(hash(b) === t.sha256, t.file + " render hash");
  check(
    b.toString("ascii", 0, 4) === "RIFF" &&
      b.toString("ascii", 8, 12) === "WAVE",
    "WAV header",
  );
  check(
    b.readUInt16LE(20) === 1 &&
      b.readUInt16LE(22) === 2 &&
      b.readUInt32LE(24) === 44100 &&
      b.readUInt16LE(34) === 16,
    "PCM format",
  );
  check(
    b.length === 44 + 2268000 * 4 && b.readUInt32LE(40) === 2268000 * 4,
    "no listening tail/padding embedded",
  );
  let en = 0,
    peak = 0,
    dc = 0,
    minWindow = Infinity;
  for (let offset = 44; offset < b.length; offset += 2) {
    let v = b.readInt16LE(offset) / 32768;
    en += v * v;
    dc += v;
    peak = Math.max(peak, Math.abs(v));
  }
  const count = (b.length - 44) / 2,
    rms = Math.sqrt(en / count);
  check(Math.abs(rms - 0.06) < 0.0002, "RMS match");
  check(peak < 0.85, "headroom");
  check(Math.abs(dc / count) < 0.001, "DC offset");
  for (let c = 0; c < 2; c++) {
    const d = Math.abs(
      b.readInt16LE(44 + c * 2) - b.readInt16LE(b.length - 4 + c * 2),
    );
    check(d <= 328, "periodic seam delta " + t.id);
  }
  for (let at = 0; at < 2268000; at += 2205) {
    let e = 0;
    for (let j = 0; j < 4410; j++) {
      const k = (at + j) % 2268000;
      for (let c = 0; c < 2; c++) {
        const x = b.readInt16LE(44 + k * 4 + c * 2) / 32768;
        e += x * x;
      }
    }
    minWindow = Math.min(minWindow, Math.sqrt(e / 8820));
  }
  check(minWindow > 0.0005, "silence/gap including wrap");
  check(t.stats.fluteSampleShortages === 0, "flute sample coverage");
  check(t.minimumBacking100msRms > 0.001, "continuous backing audit");
  console.log(
    `${t.id}: ${t.frames} frames, RMS ${rms.toFixed(6)}, peak ${peak.toFixed(4)}, minimum100ms ${minWindow.toFixed(5)}`,
  );
}
check(
  hash(
    await Bun.file(repo + "/internal/sound/audio/lobby-badinerie.wav").bytes(),
  ) === "d41d29242ea46c8df0014eb84910dc72ced90f1ecd7149d06346e943d95361b5",
  "approved title/lobby/Credits PCM changed",
);
for (const file of ["gameplay-beethoven.wav", "gameplay-beethoven-g2.wav"])
  check(
    !(await Bun.file(repo + "/internal/sound/audio/" + file).exists()),
    "legacy Beethoven still embedded",
  );
const samples = await Bun.file(dir + "/sample-sources.json").json();
for (const f of samples.files) {
  check(f.license === "CC0-1.0", "sample license");
  check(
    hash(await Bun.file(repo + "/.tmp/music-sources/" + f.path).bytes()) ===
      f.sha256,
    "sample provenance " + f.path,
  );
}
console.log(
  `PASS ${checks} classical-music checks; no subjective audition implied.`,
);
