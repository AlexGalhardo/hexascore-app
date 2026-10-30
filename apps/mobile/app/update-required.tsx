import type { ReactElement } from "react";
import { storeListingUrl } from "../src/platform/store";
import { Placeholder } from "../src/ui/Placeholder";

export default function UpdateRequiredScreen(): ReactElement {
	return <Placeholder title={`Update required ${storeListingUrl()}`} />;
}
