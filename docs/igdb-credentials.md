# Credenciais do IGDB (Client ID e Client Secret)

O IGDB não tem cadastro próprio: a API é autenticada pela Twitch. As duas credenciais saem do console de desenvolvedor da Twitch e são lidas pelo conector `webscrappings/connectors/igdb.py`.

| Variável | O que é |
| --- | --- |
| `IGDB_TWITCH_CLIENT_ID` | Client ID do app registrado na Twitch |
| `IGDB_TWITCH_CLIENT_SECRET` | Client Secret do mesmo app |

## Passo a passo

1. **Ative a autenticação em dois fatores na conta Twitch.** Em [twitch.tv](https://www.twitch.tv), abra *Settings → Security and Privacy → Two-Factor Authentication*. Sem 2FA o console não permite registrar um app.
2. Abra o console: <https://dev.twitch.tv/console/apps> e clique em **Register Your Application**.
3. Preencha o formulário:
   - **Name**: qualquer nome único, sem a palavra "twitch" (ex.: `hexascore-ingest`).
   - **OAuth Redirect URLs**: `http://localhost` (o campo é obrigatório, mas o fluxo usado aqui não redireciona).
   - **Category**: `Application Integration`.
   - **Client Type**: `Confidential`. Um app `Public` não tem Client Secret e o tipo não pode ser trocado depois: nesse caso, registre outro app.
4. Clique em **Create**. Na lista de apps, clique em **Manage** no app criado.
5. Copie o **Client ID** exibido na página.
6. Clique em **New Secret** e copie o **Client Secret**. Ele aparece uma única vez; se perder, gere outro (o anterior deixa de valer).
7. Grave as duas variáveis no ambiente da ingestão (localmente no `.env`; no GitHub Actions como secrets do repositório):

   ```dotenv
   IGDB_TWITCH_CLIENT_ID=<client id>
   IGDB_TWITCH_CLIENT_SECRET=<client secret>
   ```

## Access token

Não é preciso gerar nem guardar o access token. A cada execução o conector pede um token de app (fluxo *client credentials*) e o usa nas chamadas ao IGDB:

```http
POST https://id.twitch.tv/oauth2/token?client_id=<id>&client_secret=<secret>&grant_type=client_credentials
```

As chamadas ao IGDB levam os cabeçalhos `Client-ID: <client id>` e `Authorization: Bearer <access token>`.

## Testar as credenciais

```sh
curl -X POST "https://id.twitch.tv/oauth2/token?client_id=$IGDB_TWITCH_CLIENT_ID&client_secret=$IGDB_TWITCH_CLIENT_SECRET&grant_type=client_credentials"
```

Resposta esperada: JSON com `access_token`, `expires_in` e `token_type: "bearer"`. `400 invalid client` indica Client ID errado; `403 invalid client secret` indica secret errado ou revogado.

## Cuidados

- O Client Secret é credencial: nunca vai para o repositório nem para variáveis `NEXT_PUBLIC_*`.
- Limite do IGDB: 4 requisições por segundo por app (o `HttpClient` já respeita via `HOST_INTERVALS`).
- Referência oficial: <https://api-docs.igdb.com/#account-creation>.
