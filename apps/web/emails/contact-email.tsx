import { Body, Container, Head, Heading, Html, Preview, Text } from "@react-email/components";

type Props = { name: string; email: string; subject: string; message: string };

export function ContactEmail({ name, email, subject, message }: Props) {
	return (
		<Html>
			<Head />
			<Preview>{`Contact (${subject}) from ${name}`}</Preview>
			<Body style={{ fontFamily: "system-ui, sans-serif", backgroundColor: "#f6f6f4" }}>
				<Container style={{ maxWidth: 560, padding: 24, backgroundColor: "#ffffff" }}>
					<Heading as="h1" style={{ fontSize: 20 }}>
						New contact message: {subject}
					</Heading>
					<Text>
						From: {name} &lt;{email}&gt;
					</Text>
					<Text style={{ whiteSpace: "pre-wrap" }}>{message}</Text>
				</Container>
			</Body>
		</Html>
	);
}

export default ContactEmail;
