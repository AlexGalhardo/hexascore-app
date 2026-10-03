# Frontend e rotas


Locale: EN sem prefixo, `/pt-br/…`, `/es/…`. Slugs de rota em inglês nos 3 idiomas. Tudo mobile-first, light/dark.

| Rota | Conteúdo |
| --- | --- |
| `/` | navbar com busca global (autocomplete) + blocos por seção (melhores, lançamentos, mais avaliados) |
| `/games` `/movies` `/series` `/books` | catálogo denso com filtros na URL |
| `/{type}/[slug]` | página do item |
| `/search?q=` | resultados entre tipos |
| `/u/[nickname]`, `/u/[nickname]/lists/[id]` | perfil público, listas |
| `/feed` | avaliações de quem você segue |
| `/pricing` | planos da API |
| `/docs/*` | guias + Scalar |
| `/login` `/signup` `/recover-password` `/reset-password` `/verify-email` | auth |
| `/profile` | dados, senha (definir/mudar), 2FA, passkeys, excluir conta |
| `/profile/api` · `/profile/api/checkout` | chaves, uso, Scalar · assinar plano/comprar créditos |
| `/contact` `/policy` `/terms` | contato, privacidade, termos e cookies |
| `/admin/*` | contato, denúncias, afiliados, edição/merge de itens, status da ingestão, banir usuário (só EN) |

**Página do item**: cabeçalho (título, ano, tipo, capa pequena, gêneros/plataformas) · metascore grande + radar hexagonal das 6 categorias + nº de avaliações · botão "ver notas dos usuários" (modal paginado, só perfis públicos) · formulário com 6 controles (slider + campo numérico, acessível por teclado, média ao vivo) · notas externas (MC, RT, IMDb, HLTB) como links com atribuição · até 3 trailers (`youtube-nocookie`, carregam só no clique) · tabela de metadados · status + adicionar à lista · "onde comprar" (livros/jogos).

**Formulários** (validação zod idêntica cliente/servidor):
- Senha: 8–32, ≥ 1 maiúscula, 1 minúscula, 1 número, 1 especial; botão olho. Rodapé do login: "Não tem conta ainda? Criar conta"; do signup: "Já tem conta? Entrar". Ambos com Google.
- Nickname: 4–16 `[a-zA-Z0-9_]`, único; avatar placeholder mostra a 1ª letra em MAIÚSCULA enquanto digita.
- Contato: nome + e-mail preenchidos e `disabled` se logado (servidor usa a sessão); select "Dúvidas" / "Sugestões & Feedback" / "Bugs e Problemas Técnicos"; textarea 5 linhas, 32–512 caracteres com contador `0/512`.
- Excluir conta: modal explicando os 30 dias, botão confirmar desabilitado por 10s com contagem.

**Auth e conta (0.2, implementado)**:
- Login/signup/2FA/passkeys/troca de senha usam o client do Better Auth (`lib/auth-client.ts`); tudo que o Better Auth só expõe no servidor (`setPassword`) e as escritas próprias (nome, perfil privado, exclusão, contato) são Server Actions em `actions/` com retorno `ActionResult` (`lib/action-result.ts`), traduzido pelo namespace `actions`. Erros do Better Auth viram chaves `auth.errors.*` via `lib/auth-errors.ts` (429 → `rateLimited`, falha de rede → `network`).
- Links de e-mail usam `callbackURL`/`redirectTo` já localizados (`getPathname`): verificação cai em `/verify-email` (já logado, `autoSignInAfterVerification`), reset cai em `/reset-password?token=`. Em dev sem Resend, o link sai no log do servidor (`[email:dev] ... link=`).
- `?next=` do login só aceita caminho relativo (`lib/safe-next.ts`). `/profile` sem sessão redireciona para `/login?next=`.
- Banner de exclusão agendada (`components/account/deletion-banner.tsx`) é renderizado no layout para qualquer página quando `deletionScheduledAt` está setado.
- Namespaces novos: `auth`, `account` (inclui `actions`), `contact`. Textos legais ficam em `messages/<locale>/legal.json`, **fora** de `NAMESPACES`, carregados só por `/policy` e `/terms` (`lib/legal.ts`) para não irem ao browser em toda página. `messages/messages.test.ts` garante as mesmas chaves nos 3 idiomas.
- Primitivos de formulário em `components/ui/form.tsx` (`Field`, `Input`, `Notice`, `buttonClass`); Base UI `Dialog` (excluir conta) e `Menu` (menu do usuário).

**Preferências**: tema, consentimento e locale em cookies próprios; tema lido no servidor (sem flash, sem lib). Banner de cookies: só essenciais (Analytics é cookieless) + consentimento para embeds do YouTube.

**UI**: Base UI + Tailwind 4 (tokens por seção), `next-intl`, radar em SVG puro.
