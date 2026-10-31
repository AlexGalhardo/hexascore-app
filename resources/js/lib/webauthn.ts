export class PasskeyCancelled extends Error {
	constructor() {
		super("passkey cancelled");
		this.name = "PasskeyCancelled";
	}
}

interface CredentialJSON {
	toJSON: () => unknown;
}

interface CreationConstructors {
	parseCreationOptionsFromJSON: (options: unknown) => PublicKeyCredentialCreationOptions;
}

interface RequestConstructors {
	parseRequestOptionsFromJSON: (options: unknown) => PublicKeyCredentialRequestOptions;
}

function cancelled(err: unknown): boolean {
	return (
		err instanceof PasskeyCancelled ||
		(err instanceof DOMException && err.name === "NotAllowedError")
	);
}

export async function createPasskey(options: unknown): Promise<unknown> {
	if (typeof PublicKeyCredential === "undefined" || !navigator.credentials?.create) {
		throw new PasskeyCancelled();
	}

	try {
		const publicKey = (
			PublicKeyCredential as unknown as CreationConstructors
		).parseCreationOptionsFromJSON(options);
		const cred = await navigator.credentials.create({ publicKey });
		if (!cred) {
			throw new PasskeyCancelled();
		}

		return (cred as unknown as CredentialJSON).toJSON();
	} catch (err) {
		if (cancelled(err)) {
			throw new PasskeyCancelled();
		}

		throw err;
	}
}

export async function getPasskey(options: unknown): Promise<unknown> {
	if (typeof PublicKeyCredential === "undefined" || !navigator.credentials?.get) {
		throw new PasskeyCancelled();
	}

	try {
		const publicKey = (
			PublicKeyCredential as unknown as RequestConstructors
		).parseRequestOptionsFromJSON(options);
		const cred = await navigator.credentials.get({ publicKey });
		if (!cred) {
			throw new PasskeyCancelled();
		}

		return (cred as unknown as CredentialJSON).toJSON();
	} catch (err) {
		if (cancelled(err)) {
			throw new PasskeyCancelled();
		}

		throw err;
	}
}
