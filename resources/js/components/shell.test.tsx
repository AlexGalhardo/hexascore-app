import { render } from "@testing-library/react";
import { expect, test } from "vitest";

import CoverThumb from "./CoverThumb";
import { dialogPopup } from "./Dialog";
import { jsonLdText, truncate } from "./Seo";

test("dialogPopup centres every size above navbar and cookie banner", () => {
	for (const size of ["sm", "md", "lg"] as const) {
		const cls = dialogPopup(size);
		expect(cls).toContain("top-1/2");
		expect(cls).toContain("left-1/2");
		expect(cls).toContain("-translate-x-1/2");
		expect(cls).toContain("-translate-y-1/2");
		expect(cls).toContain("w-[calc(100%-2rem)]");
		expect(cls).toContain("z-[60]");
	}
});

test("truncate keeps titles and descriptions within their limit", () => {
	expect(truncate("short", 60)).toBe("short");
	expect(truncate("x".repeat(100), 60)).toHaveLength(60);
	expect(truncate("x".repeat(100), 60).endsWith("…")).toBe(true);
});

test("jsonLdText never contains a literal '<'", () => {
	const text = jsonLdText([{ name: "</script><script>alert(1)</script>" }]);
	expect(text).not.toContain("<");
	expect(JSON.parse(text).name).toBe("</script><script>alert(1)</script>");
});

const COVER = { thumb: "/a-96.jpg", full: "/a-600.jpg" };

test("CoverThumb renders a sized placeholder without a cover", () => {
	const r = render(<CoverThumb cover={null} alt="Cover" size="md" />);
	expect(r.queryByAltText("Cover")).toBeNull();
	const box = r.getByTestId("cover-placeholder");
	expect(box.style.width).toBe("44px");
	expect(box.style.height).toBe("64px");
});

test("CoverThumb is lazy by default and eager with priority", () => {
	const lazy = render(<CoverThumb cover={COVER} alt="Cover" />);
	const img = lazy.getByAltText("Cover");
	expect(img.getAttribute("loading")).toBe("lazy");
	expect(img.getAttribute("srcset")).toBe("/a-96.jpg 96w, /a-600.jpg 600w");
	expect(img.getAttribute("width")).toBe("32");
	lazy.unmount();

	const eager = render(<CoverThumb cover={COVER} alt="Cover" size="md" priority />);
	const big = eager.getByAltText("Cover");
	expect(big.getAttribute("loading")).toBe("eager");
	expect(big.getAttribute("fetchpriority")).toBe("high");
	expect(big.getAttribute("height")).toBe("64");
});
