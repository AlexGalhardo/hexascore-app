# Webhooks

Receba uma requisição HTTP assinada quando os itens mudarem, em vez de fazer polling.

## Configuração

Os webhooks estão disponíveis nos planos Hobby (2 endpoints) e PRO (10 endpoints). Crie um endpoint com `POST /v1/webhooks`:

| Campo | Significado |
| --- | --- |
| `url` | Uma URL `https` que responda em poucos segundos. Endereços privados e internos são recusados. |
| `events` | Um ou mais entre `item.created`, `item.updated`, `item.deleted`. |
| `itemTypes` | Um ou mais entre `games`, `movies`, `series`, `books`. |

A resposta contém o `secret` de assinatura, que começa com `whsec_`. Ele é exibido uma única vez: guarde-o no seu gerenciador de segredos. Use `GET /v1/webhooks` para listar os endpoints, `DELETE /v1/webhooks/{id}` para removê-los e `POST /v1/webhooks/{id}/test` para enviar uma entrega de teste. Gerenciar webhooks não conta na sua cota.

## O que enviamos

Enviamos um lote de até 100 eventos em um único `POST` com corpo JSON. Os payloads trazem apenas ids e nomes de campos; busque o item de novo para ler os dados dele.

```json
{
  "type": "item.batch",
  "deliveredAt": "2026-11-02T10:15:00Z",
  "events": [
    {
      "id": "chg_1048576",
      "type": "item.updated",
      "itemType": "games",
      "itemId": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a6b71",
      "fields": ["score", "ratingCount"],
      "changedAt": "2026-11-02T10:14:12Z"
    },
    {
      "id": "chg_1048577",
      "type": "item.deleted",
      "itemType": "movies",
      "itemId": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a5521",
      "mergedInto": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a5520",
      "changedAt": "2026-11-02T10:14:30Z"
    }
  ]
}
```

Os eventos do mesmo item dentro de uma janela de entrega são agrupados. Os ids de evento são únicos, então use-os para ignorar uma repetição.

## Verifique a assinatura

As entregas seguem o padrão [Standard Webhooks](https://www.standardwebhooks.com). Toda requisição tem três cabeçalhos: `webhook-id`, `webhook-timestamp` (segundos Unix) e `webhook-signature` (um ou mais valores `v1,<base64>` separados por espaço).

Para verificar uma entrega:

- Leia o corpo bruto da requisição. Não o interprete nem o serialize de novo.
- Rejeite a requisição quando `webhook-timestamp` estiver a mais de 5 minutos do seu relógio.
- Calcule o HMAC-SHA256 em base64 de `<webhook-id>.<webhook-timestamp>.<corpo bruto>`, usando como chave o segredo sem o prefixo `whsec_`, decodificado de base64.
- Compare-o com cada valor `v1` em tempo constante. Aceite a requisição quando um deles coincidir.

```ts
import { createHmac, timingSafeEqual } from "node:crypto";

export function verifyWebhook(secret: string, headers: Headers, rawBody: string): boolean {
	const id = headers.get("webhook-id");
	const timestamp = headers.get("webhook-timestamp");
	const signatures = headers.get("webhook-signature");
	if (!id || !timestamp || !signatures) return false;
	if (Math.abs(Date.now() / 1000 - Number(timestamp)) > 300) return false;

	const key = Buffer.from(secret.replace(/^whsec_/, ""), "base64");
	const expected = createHmac("sha256", key).update(`${id}.${timestamp}.${rawBody}`).digest();
	return signatures.split(" ").some((entry) => {
		const [version, value] = entry.split(",");
		if (version !== "v1" || !value) return false;
		const got = Buffer.from(value, "base64");
		return got.length === expected.length && timingSafeEqual(got, expected);
	});
}
```

```python
import base64
import binascii
import hashlib
import hmac
import time


def verify_webhook(secret: str, headers: dict[str, str], raw_body: bytes) -> bool:
    try:
        msg_id = headers["webhook-id"]
        timestamp = headers["webhook-timestamp"]
        signatures = headers["webhook-signature"]
        if abs(time.time() - int(timestamp)) > 300:
            return False
    except (KeyError, ValueError):
        return False

    key = base64.b64decode(secret.removeprefix("whsec_"))
    signed = f"{msg_id}.{timestamp}.".encode() + raw_body
    expected = hmac.new(key, signed, hashlib.sha256).digest()
    for entry in signatures.split(" "):
        version, _, value = entry.partition(",")
        try:
            got = base64.b64decode(value)
        except binascii.Error:
            continue
        if version == "v1" and hmac.compare_digest(got, expected):
            return True
    return False
```

## Novas tentativas e falhas

Responda com qualquer status `2xx` em poucos segundos e faça o trabalho de verdade depois. Repetimos as outras respostas e os timeouts 8 vezes, com intervalos crescentes a partir de 30 segundos. Um `410 Gone` interrompe a entrega definitivamente. Depois de três entregas seguidas esgotarem as tentativas, o endpoint é desativado e enviamos um e-mail ao dono da conta. As entregas ficam registradas por 30 dias.

A entrega é pelo menos uma vez e sem ordem estrita. Os webhooks ainda podem perder eventos, então chame [`/v1/changes`](/docs/changes) com o `since` do último evento processado para preencher qualquer lacuna.
