// Proof-of-work solver (ADR 0016): finds n so that SHA-256(c + n) has `bits`
// leading zero bits. Runs in a worker so the page stays responsive; the
// synchronous hash avoids one promise per attempt.

const K = new Uint32Array([
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
	0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
	0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
	0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
	0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
	0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
	0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
	0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
]);

const rotr = (x: number, n: number): number => (x >>> n) | (x << (32 - n));

const IV = [
	0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
];

// Scratch space shared by every call: the solver hashes millions of messages, so the hot path allocates nothing.
const w = new Uint32Array(64);

// Hashes the already padded message `p` (a multiple of 64 bytes) into `h`.
function compress(p: Uint8Array, h: Uint32Array): void {
	h.set(IV);
	for (let off = 0; off < p.length; off += 64) {
		for (let i = 0; i < 16; i++) {
			const o = off + i * 4;
			w[i] = (p[o] << 24) | (p[o + 1] << 16) | (p[o + 2] << 8) | p[o + 3];
		}
		for (let i = 16; i < 64; i++) {
			const x = w[i - 15];
			const y = w[i - 2];
			const s0 = rotr(x, 7) ^ rotr(x, 18) ^ (x >>> 3);
			const s1 = rotr(y, 17) ^ rotr(y, 19) ^ (y >>> 10);
			w[i] = (w[i - 16] + s0 + w[i - 7] + s1) | 0;
		}
		let a = h[0];
		let b = h[1];
		let c = h[2];
		let d = h[3];
		let e = h[4];
		let f = h[5];
		let g = h[6];
		let hh = h[7];
		for (let i = 0; i < 64; i++) {
			const t1 =
				(hh +
					(rotr(e, 6) ^ rotr(e, 11) ^ rotr(e, 25)) +
					((e & f) ^ (~e & g)) +
					K[i] +
					w[i]) |
				0;
			const t2 =
				((rotr(a, 2) ^ rotr(a, 13) ^ rotr(a, 22)) + ((a & b) ^ (a & c) ^ (b & c))) | 0;
			hh = g;
			g = f;
			f = e;
			e = (d + t1) | 0;
			d = c;
			c = b;
			b = a;
			a = (t1 + t2) | 0;
		}
		h[0] += a;
		h[1] += b;
		h[2] += c;
		h[3] += d;
		h[4] += e;
		h[5] += f;
		h[6] += g;
		h[7] += hh;
	}
}

// A zeroed buffer holding `len` message bytes plus the SHA-256 padding; the caller fills bytes [0, len).
function padded(len: number): Uint8Array {
	const p = new Uint8Array(((len + 9 + 63) >> 6) << 6);
	p[len] = 0x80;
	// Messages here are far below 2^29 bytes, so the high length word stays 0.
	new DataView(p.buffer).setUint32(p.length - 4, len * 8, false);
	return p;
}

export function sha256(data: Uint8Array): Uint32Array {
	const p = padded(data.length);
	p.set(data);
	const h = new Uint32Array(8);
	compress(p, h);
	return h;
}

export function leadingZeroBits(digest: Uint32Array): number {
	let total = 0;
	for (const word of digest) {
		if (word === 0) {
			total += 32;
			continue;
		}
		return total + Math.clz32(word);
	}
	return total;
}

export function solve(c: string, bits: number): string {
	const prefix = new TextEncoder().encode(c);
	const h = new Uint32Array(8);
	let digits = 0;
	let buf = padded(0);
	for (let n = 0; ; n++) {
		if (n === 0 || String(n).length !== digits) {
			digits = String(n).length;
			buf = padded(prefix.length + digits);
			buf.set(prefix);
		}
		for (let i = digits - 1, v = n; i >= 0; i--, v = (v / 10) | 0) {
			buf[prefix.length + i] = 48 + (v % 10);
		}
		compress(buf, h);
		if (leadingZeroBits(h) >= bits) return String(n);
	}
}

self.onmessage = (e: MessageEvent<{ c: string; bits: number }>) => {
	self.postMessage(solve(e.data.c, e.data.bits));
};
