import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test } from "vitest";

import Markdown from "./Markdown";

afterEach(cleanup);

test("renders the parsed subset and keeps markup as text", () => {
	const { container } = render(
		<Markdown
			nodes={[
				{
					t: "p",
					c: [
						{ t: "text", text: "<script>alert(1)</script> " },
						{ t: "strong", c: [{ t: "text", text: "bold" }] },
						{ t: "br" },
						{ t: "a", href: "https://example.com/", c: [{ t: "text", text: "site" }] },
					],
				},
				{ t: "ul", c: [{ t: "li", c: [{ t: "code", text: "x" }] }] },
				{
					t: "blockquote",
					c: [{ t: "p", c: [{ t: "em", c: [{ t: "text", text: "q" }] }] }],
				},
			]}
		/>,
	);

	expect(container.querySelector("script")).toBeNull();
	expect(screen.getByText(/<script>alert\(1\)<\/script>/)).toBeTruthy();
	expect(container.querySelector("strong")?.textContent).toBe("bold");
	const link = screen.getByRole("link", { name: "site" });
	expect(link.getAttribute("href")).toBe("https://example.com/");
	expect(link.getAttribute("rel")).toBe("nofollow ugc noopener");
	expect(container.querySelector("ul li code")?.textContent).toBe("x");
	expect(container.querySelector("blockquote em")?.textContent).toBe("q");
});

test("drops a link whose href is not http(s)", () => {
	render(
		<Markdown
			nodes={[{ t: "a", href: "javascript:alert(1)", c: [{ t: "text", text: "x" }] }]}
		/>,
	);
	expect(screen.queryByRole("link")).toBeNull();
	expect(screen.getByText("x")).toBeTruthy();
});
