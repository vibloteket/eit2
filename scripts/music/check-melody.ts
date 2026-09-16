// Independent symbolic check only: fetch a pinned reference into memory.
// Do not redistribute the reference engraving or infer a license for its file.
import { createHash } from "node:crypto";

const root = import.meta.dir + "/../..";
const audit = await Bun.file(
  root + "/music/badinerie/melody-crosscheck.json",
).json();
const score = await Bun.file(root + "/music/badinerie/score.json").json();
const url = audit.referenceURL
  .replace("https://github.com/", "https://raw.githubusercontent.com/")
  .replace("/blob/", "/");
const response = await fetch(url);
if (!response.ok) throw new Error(`Reference HTTP ${response.status}`);
const bytes = new Uint8Array(await response.arrayBuffer());
if (
  createHash("sha256").update(bytes).digest("hex") !== audit.referenceSHA256
) {
  throw new Error(
    "Reference hash mismatch; do not silently accept different notation",
  );
}
const text = new TextDecoder()
  .decode(bytes)
  .replace(/%[^\n]*/g, "")
  .replace(/\\acciaccatura\s*\{[^}]*\}/g, " ");
const notes: Array<{ key: number; beats: number }> = [];
for (const m of text.matchAll(
  /(?<![a-zA-Z\\])([a-h](?:is|es)?)([',]*)(\d+)(\.*)/g,
)) {
  const base = (
    { c: 48, d: 50, e: 52, f: 53, g: 55, a: 57, h: 59, b: 58 } as Record<
      string,
      number
    >
  )[m[1][0]];
  const octave = [...m[2]].reduce((n, c) => n + (c === "'" ? 1 : -1), 0);
  notes.push({
    key:
      base +
      (m[1].endsWith("is") ? 1 : m[1].endsWith("es") ? -1 : 0) +
      12 * octave,
    beats: (4 / Number(m[3])) * (2 - 2 ** -m[4].length),
  });
}
const ours: Array<{ key: number; beats: number }> = [];
let stolen = 0;
for (const measure of [...score.sections.A, ...score.sections.B]) {
  for (const note of measure.notes) {
    if (note.ornament) {
      stolen += note.beats;
      continue;
    }
    ours.push({ key: note.key, beats: note.beats + stolen });
    stolen = 0;
  }
}
if (ours.length !== 230 || JSON.stringify(ours) !== JSON.stringify(notes)) {
  throw new Error(
    "Principal melody differs from independently checked public-domain note sequence",
  );
}
console.log(
  "PASS: all 230 principal pitches/durations match; grace interpretation is documented separately.",
);
