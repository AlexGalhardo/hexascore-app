import { Body, Button, Container, Head, Heading, Html, Preview, Text } from "@react-email/components";

type Props = { appName: string; heading: string; body: string; actionLabel: string; url: string };

export function ActionEmail({ appName, heading, body, actionLabel, url }: Props) {
	return (
		<Html>
			<Head />
			<Preview>{heading}</Preview>
			<Body style={{ fontFamily: "system-ui, sans-serif", backgroundColor: "#f6f6f4" }}>
				<Container style={{ maxWidth: 480, padding: 24, backgroundColor: "#ffffff" }}>
					<Text style={{ fontWeight: 700, fontSize: 18 }}>{appName}</Text>
					<Heading as="h1" style={{ fontSize: 22 }}>
						{heading}
					</Heading>
					<Text>{body}</Text>
					<Button
						href={url}
						style={{ backgroundColor: "#1d1d24", color: "#ffffff", padding: "12px 20px", borderRadius: 6 }}
					>
						{actionLabel}
					</Button>
					<Text style={{ color: "#6b6b76", fontSize: 12 }}>{url}</Text>
				</Container>
			</Body>
		</Html>
	);
}

export default ActionEmail;
