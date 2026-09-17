// Validate the explicitly public-domain Mutopia inputs and the documented
// piano-to-gameplay reduction. Reference files are bundled with their PD notice.
import { createHash } from "node:crypto";

const root = import.meta.dir + "/../../music/beethoven";
const manifest = await Bun.file(root + "/sources.json").json();
const score = await Bun.file(root + "/score.json").json();
for (const file of manifest.sources) {
  const bytes = new Uint8Array(
    await Bun.file(root + "/" + file.file).arrayBuffer(),
  );
  if (createHash("sha256").update(bytes).digest("hex") !== file.sha256)
    throw Error("Mutopia source hash mismatch: " + file.file);
}
const declaration = await Bun.file(
  root + "/sources/mutopia/beethoven_rondo_op129.ly",
).text();
if (
  !declaration.includes('license = "Public Domain"') ||
  !declaration.includes("free to distribute, modify, and perform")
)
  throw Error("Missing explicit score rights declaration");
const b = Buffer.from(
  await Bun.file(
    root + "/sources/mutopia/beethoven_rondo_op129.mid",
  ).arrayBuffer(),
);
if (
  b.toString("ascii", 0, 4) !== "MThd" ||
  b.readUInt16BE(8) !== 1 ||
  b.readUInt16BE(12) !== 384
)
  throw Error("Unexpected MIDI format/division");
const tracks: Array<Array<{ at: number; key: number; beats: number }>> = [];
let pos = 8 + b.readUInt32BE(4);
while (pos < b.length) {
  if (b.toString("ascii", pos, pos + 4) !== "MTrk")
    throw Error("MIDI track chunk");
  const end = pos + 8 + b.readUInt32BE(pos + 4);
  pos += 8;
  let ticks = 0,
    running = 0;
  const active = new Map<string, { ticks: number; key: number }>();
  const notes: Array<{ at: number; key: number; beats: number }> = [];
  function vlq() {
    let value = 0,
      x;
    do {
      x = b[pos++];
      value = value * 128 + (x & 127);
    } while (x & 128);
    return value;
  }
  while (pos < end) {
    ticks += vlq();
    let status = b[pos];
    if (status >= 128) {
      pos++;
      if (status < 240) running = status;
    } else {
      if (!running) throw Error("Invalid running status");
      status = running;
    }
    if (status === 255) {
      pos++;
      const length = vlq();
      pos += length;
      continue;
    }
    if (status === 240 || status === 247) {
      const length = vlq();
      pos += length;
      running = 0;
      continue;
    }
    const kind = status >> 4,
      channel = status & 15,
      d1 = b[pos++],
      d2 = kind === 12 || kind === 13 ? 0 : b[pos++];
    const id = channel + ":" + d1;
    if (kind === 9 && d2 > 0) active.set(id, { ticks, key: d1 });
    else if (kind === 8 || (kind === 9 && d2 === 0)) {
      const note = active.get(id);
      if (note) {
        notes.push({
          at: note.ticks / 384,
          key: note.key,
          beats: (ticks - note.ticks) / 384,
        });
        active.delete(id);
      }
    }
  }
  if (pos !== end || active.size) throw Error("Unclosed MIDI track/notes");
  notes.sort((a, b) => a.at - b.at || a.key - b.key);
  tracks.push(notes);
}
if (tracks.length !== 3) throw Error("Expected control/RH/LH tracks");
const groups = new Map<number, (typeof tracks)[number]>();
for (const note of tracks[1])
  if (note.at < 112) {
    const group = groups.get(note.at) ?? [];
    group.push(note);
    groups.set(note.at, group);
  }
const melody = [];
for (const [at, group] of [...groups].sort((a, b) => a[0] - b[0])) {
  if (at >= 111.5) continue; // Omit the pickup into the next episode; close on G.
  const top = group.reduce((a, b) => (a.key > b.key ? a : b));
  melody.push({
    at,
    key: top.key,
    beats: Math.min(top.beats, 112 - at),
    instrument: top.key >= 65 ? "flute" : "pluck",
  });
}
const bass = [],
  pluck = [];
function fold(key: number, low: number, high: number) {
  while (key < low) key += 12;
  while (key > high) key -= 12;
  return key;
}
for (let at = 0; at < 112; at += 0.5) {
  const active = tracks[2]
    .filter((n) => n.at <= at + 1e-8 && n.at + n.beats > at + 1e-8)
    .map((n) => n.key)
    .sort((a, b) => a - b);
  if (!active.length) continue;
  if (Number.isInteger(at))
    bass.push({ at, key: fold(active[0], 36, 54), beats: 0.65 });
  const upper = [
    ...new Set(
      active.slice(active.length > 1 ? 1 : 0).map((n) => fold(n, 60, 76)),
    ),
  ].sort((a, b) => a - b);
  pluck.push({ at, key: upper[Math.round(at * 2) % upper.length], beats: 0.5 });
}
for (const [name, derived] of Object.entries({ melody, bass, pluck })) {
  if (JSON.stringify(score[name]) !== JSON.stringify(derived))
    throw Error("Score reduction differs: " + name);
}
if (score.bpm !== 126 || JSON.stringify(score.form) !== '["A","A","B","B"]')
  throw Error("Unapproved tempo/form");
console.log(
  `PASS: ${manifest.sources.length} PD source files verified; ${melody.length} lead events and accompaniment reproduced; 126 BPM/AABB.`,
);
