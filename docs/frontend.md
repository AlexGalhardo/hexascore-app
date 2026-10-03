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
- Nickname: 4–16 letras minúsculas `[a-z]` (sem números, espaços ou símbolos), único; o input normaliza enquanto digita (`normalizeNickname`) e o plugin username do Better Auth valida no servidor (`usernameValidator`). No signup, e-mail tem no máximo 48 caracteres (`signupEmailSchema`, também checado no hook de `/sign-up/email`) e senha 32; contas antigas com e-mail maior continuam logando e recuperando senha.
- **Navbar** (`components/layout/navbar.tsx` + `nav-shell.tsx`, fixa no topo): marca "HexaScore" ("Hexa" verde, "Score" azul); no desktop busca ocupa o espaço até os botões (tema, Entrar ou menu da conta, idioma por último) e os links com ícone (Jogos, Filmes, Séries, Livros, API, Preços) ficam numa linha `justify-between`; no mobile só marca + lupa (abre a busca) + menu hambúrguer (links, ações e idioma). Idioma: `locale-menu.tsx` mostra só as siglas (EN / PT-BR / ES), com a atual marcada; o nome nativo fica no `aria-label`. Footer: só "© 2026 HexaScore", Termos, Privacidade e Contato.
- Contato: nome + e-mail preenchidos e `disabled` se logado (servidor usa a sessão); select "Dúvidas" / "Sugestões & Feedback" / "Bugs e Problemas Técnicos"; textarea 5 linhas, 32–512 caracteres com contador `0/512`.
- Excluir conta: modal explicando os 30 dias, botão confirmar desabilitado por 10s com contagem.

**Auth e conta (0.2, implementado)**:
- Login/signup/2FA/passkeys/troca de senha usam o client do Better Auth (`lib/auth-client.ts`); tudo que o Better Auth só expõe no servidor (`setPassword`) e as escritas próprias (nome, perfil privado, exclusão, contato) são Server Actions em `actions/` com retorno `ActionResult` (`lib/action-result.ts`), traduzido pelo namespace `actions`. Erros do Better Auth viram chaves `auth.errors.*` via `lib/auth-errors.ts` (429 → `rateLimited`, falha de rede → `network`).
- Links de e-mail usam `callbackURL`/`redirectTo` já localizados (`getPathname`): verificação cai em `/verify-email` (já logado, `autoSignInAfterVerification`), reset cai em `/reset-password?token=`. Em dev sem Resend, o link sai no log do servidor (`[email:dev] ... link=`).
- `?next=` do login só aceita caminho relativo (`lib/safe-next.ts`). `/profile` sem sessão redireciona para `/login?next=`.
- Banner de exclusão agendada (`components/account/deletion-banner.tsx`) é renderizado no layout para qualquer página quando `deletionScheduledAt` está setado.
- Namespaces novos: `auth`, `account` (inclui `actions`), `contact`. Textos legais ficam em `messages/<locale>/legal.json`, **fora** de `NAMESPACES`, carregados só por `/policy` e `/terms` (`lib/legal.ts`) para não irem ao browser em toda página. `messages/messages.test.ts` garante as mesmas chaves nos 3 idiomas (e os mesmos parágrafos por seção nos textos legais).
- **Marca**: o nome final está pendente, então nenhum texto de UI cita o codinome. Componentes usam `APP_NAME` (`lib/config.ts`); os textos legais escrevem `{appName}`, trocado em `loadLegal`; o teste de mensagens falha se algum JSON contiver "hexascore". Exceções conscientes: `/admin` (só inglês), os exemplos de código de `/docs` (domínio `hexascore.app`/`HEXASCORE_API_KEY`, ficam no TSX) e o `openapi.json`, que mudam junto com o domínio.
- **Guias de `/docs`** (quickstart, authentication, errors, pagination, rate-limits, changelog) são traduzidos pelo namespace `docs` (`messages/<locale>/docs.json`, `docs.<página>.*`; `lead` também vira a meta description). Markup inline fica no TSX via `t.rich` com as tags `<code>` (`richCode`) e links (`richLink`) de `components/docs/prose.tsx` — nunca `dangerouslySetInnerHTML`. Ficam em inglês: blocos de código, cabeçalhos HTTP, nomes de campos/parâmetros, valores de `code` de erro e URLs; valores com `{`/`<` (ex.: `{ data, next_cursor }`) entram como variáveis de interpolação, não no texto ICU.
- **E-mails de auth** (verificação, redefinição de senha) usam `auth.emails` no idioma de `localeFromRequest` (`i18n/locale.ts`: cookie `locale` → prefixo do Referer → `en`), porque `/api/auth/*` fica fora do escopo do next-intl. O lembrete PIX continua só em pt-BR.
- **Página do item**: links de afiliado com `rel="sponsored"` e aviso visível; sinopse com licença + fonte; crédito da capa (`meta.coverAttribution`) abaixo dos detalhes.
- Primitivos de formulário em `components/ui/form.tsx` (`Field`, `Input`, `Notice`, `buttonClass`); Base UI `Dialog` (excluir conta) e `Menu` (menu do usuário).

**Preferências**: tema, consentimento e locale em cookies próprios; tema lido no servidor (sem flash, sem lib). Banner de cookies: só essenciais (Analytics é cookieless) + consentimento para embeds do YouTube.

**UI**: Base UI + Tailwind 4 (tokens por seção), `next-intl`, radar em SVG puro.

**Busca de dados**:
- Server Components são o caminho padrão: páginas buscam no servidor (repositórios em `lib/`) e escritas são Server Actions. Não converta página server-rendered em fetch no cliente.
- Fetch disparado no browser usa **TanStack Query** (`@tanstack/react-query`), nunca `useEffect` + `fetch` à mão. `QueryProvider` (`components/layout/query-provider.tsx`) fica no layout `[locale]` dentro do `NextIntlClientProvider`: um `QueryClient` por sessão do browser (`useState`), `staleTime` 30 s, `retry` 1, sem `refetchOnWindowFocus`.
- O `queryFn` passa `AbortSignal.any([signal, AbortSignal.timeout(ms)])` ao `fetch`, checa `response.ok` e valida a resposta com zod antes de devolver.
- Query keys: array começando por um nome kebab-case do recurso, seguido de todo parâmetro que muda a resposta (`["search-suggest", q, locale]`, `["item-ratings", itemId]`, `["pix-charge", chargeId]`).
- Usos atuais: autocomplete da busca (`useQuery` com debounce de 200 ms, `keepPreviousData`, ativo com q ≥ 2), modal de notas dos usuários (`useInfiniteQuery` por cursor), status do PIX (`useQuery` com `refetchInterval` de 4 s até `paid`/`expired`/`failed`).
- Server Actions continuam com `useTransition`; não embrulhe em `useMutation` sem ganho claro.

## Implementação do catálogo (0.1)

- **Rotas**: `app/[locale]/[section]/page.tsx` atende `/games` `/movies` `/series` `/books` (seção validada por `typeFromSection` em `lib/sections.ts`; qualquer outra → `notFound()`), `app/[locale]/[section]/[slug]/page.tsx` é a página do item, `app/[locale]/search/page.tsx` a busca. Rotas estáticas (`/login`, `/profile`…) têm precedência sobre `[section]`.
- **Dados**: páginas leem só de `lib/catalog/cached.ts` (wrappers `unstable_cache` sobre `lib/catalog/repository.ts`). O repositório é puro (recebe `db` + parâmetros, devolve DTOs JSON) e será reaproveitado pela API pública. Tags: `catalog` (tudo) e `item:<id>` (uma página de item); revalidação por `POST /api/cron/revalidate` `{"tags": [...]}` com `Authorization: Bearer $CRON_SECRET` (comparação em tempo constante), chamado pela ingestão Python.
- **Por que `unstable_cache`**: no Next 16 `"use cache"`/`cacheTag` exigem `cacheComponents: true`, que muda o modelo de renderização do app inteiro (todo acesso dinâmico precisa de `<Suspense>`). Migração planejada quando auth/perfil estiverem prontos; até lá o Data Cache com tags faz o mesmo papel.
- **Filtros**: formulário GET puro (funciona sem JS). Parâmetros: `genre`, `platform` (só jogos), `year`, `min_score`, `max_score`, `min_<categoria>` (ex.: `min_gameplay`), `sort` (`-popularity` padrão, `±score`, `±release_date`, `±rating_count`, `±popularity`), `cursor`. Valor inválido é ignorado (`lib/catalog/search-params.ts`), cursor inválido volta à página 1.
- **Paginação**: keyset sobre `(expressão de ordenação, id)` com cursor opaco base64url que carrega a ordenação e a direção (`next`/`prev`); "Anterior/Próxima" são links comuns. Itens sem nota visível / sem data ficam sempre por último.
- **Nota visível**: `score` e médias por categoria só aparecem com `rating_count >= 5` (`MIN_RATINGS_FOR_SCORE`); abaixo disso o DTO traz `null` e a UI mostra "avaliações insuficientes". Filtros de nota também exigem o limiar.
- **Busca**: `pg_trgm` `word_similarity` no título (limiar 0.2 via `set_config` local na transação, tolera erros como "zleda" → "Zelda") OU `tsvector` (`websearch_to_tsquery('simple')`) em título + sinopse, em todos os idiomas; resultado no idioma pedido com fallback para inglês.
- **Página do item**: `generateMetadata` (título, descrição, OpenGraph, `alternates` hreflang); `<RatingPanel />` é o slot do formulário de avaliação (0.3) e hoje só linka para login; modal "ver notas dos usuários" busca `GET /api/items/[id]/ratings?cursor=` (só perfis públicos, não banidos, sem exclusão agendada); trailers `youtube-nocookie` só carregam com cookie `consent=all` ou clique; links de afiliado com `rel="sponsored nofollow noopener"` e aviso; sinopse com licença + fonte e aviso quando o idioma caiu para inglês.
- **Autocomplete**: `components/layout/search-autocomplete.tsx` (Base UI `Autocomplete`, debounce 200 ms, `GET /api/search/suggest?q=&locale=`, 8 resultados por relevância com capa, tipo, ano e nota final; ↑/↓ percorrem, Enter abre). É um `<form action="/<locale>/search">` com `name="q"`: sem JS, ou com Enter sem sugestão destacada, vira busca normal.
- **jsonb**: `items.meta`/`items.external_ids` são validados com zod na leitura (`lib/catalog/jsonb.ts`) de forma leniente — campo inválido é descartado e logado, a página não quebra. Ids externos têm formato restrito porque viram URLs; URLs vindas do banco só passam se forem `https:` (`lib/safe-url.ts`).

## Avaliação e social (0.3)

- **Página do item**: `<RatingPanel />` (`components/item/rating-panel.tsx`, server) escolhe o estado pela sessão: deslogado → "Entrar para avaliar" (`/login?next=<item>`), sem nickname (Google) → `/onboarding?next=`, e-mail não verificado → aviso + link para `/verify-email`; senão `RatingForm` (6 `ScoreInput`, nota geral ao vivo via `overall()`, pré-preenchido, salvar/excluir otimista com erros do namespace `actions`), `ItemStatusSelect` (rótulos por tipo: jogar/assistir/ler) e `AddToList` (`<details>` com checkboxes + "nova lista com este item"); em jogos com status `completed`, `TimeToBeatForm` (3 campos de horas opcionais — história principal, principal + extras, completista — passo 0,5, salvar/excluir via `actions/time-to-beat.ts`). Dados do usuário são lidos sem cache.
- **Tempo para zerar** (`components/item/time-to-beat.tsx`, só jogos): bloco "Time to beat" entre a nota da comunidade e o painel do usuário, com mediana em horas e nº de relatos por estilo, ou "relatos insuficientes" abaixo de 3. Textos em `item.timeToBeat`.
- **Escrita**: `saveRating`/`deleteRating` chamam `updateTag("item:<id>")` (quem avaliou vê o metascore novo na hora) + `revalidateTag("catalog", "max")`. As demais ações sociais chamam `refresh()`.
- **`/u/[nickname]`**: busca case-insensitive por `username`; banido/anonimizado/inexistente → 404; privado (ou exclusão agendada) → página mínima "Este perfil é privado" com 200 e só o nickname, exceto para o dono. Abas `?tab=ratings|status|lists`, ordenação `?sort=recent|best|worst`, paginação `?page=` (offset, 20). Seguir exige perfil público; sem botão no próprio perfil.
- **`/u/[nickname]/lists/[id]`**: visível ao dono sempre; aos outros só se pública, não oculta pela moderação e com dono de perfil público — qualquer outro caso é 404. O dono edita/exclui e reordena (↑/↓ trocam posições numa transação com lock na lista). Perfil (aba listas) e página da lista mostram a categoria (`components/lists/list-type-badge.tsx`: ponto com o accent da seção + rótulo `itemTypes`). Criar lista no perfil exige escolher a categoria (select com os 4 tipos) e mostra a cota da categoria (`lists.quota`, ex. "Game lists: 3/10"); no limite o envio fica desabilitado com `lists.limitReached`. No item, "Adicionar à lista" mostra só as listas do tipo do item, a cota desse tipo e "nova lista com este item" cria uma lista desse tipo.
- **`/feed`**: exige login (+ nickname); avaliações de quem você segue (só perfis públicos), cursor keyset `(created_at, user_id, item_id)`; cursor inválido volta ao início; vazio mostra "quem mais avalia".
- **`/onboarding`**: contas Google nascem sem `username`; `/profile`, `/feed` e o painel de avaliação mandam para cá. `setNickname` usa `auth.api.updateUser` (o plugin username valida 4–16, normaliza e garante unicidade) e só funciona enquanto o nickname é nulo.
- **Denúncias**: `ReportDialog` (Base UI `Dialog`) em perfis e listas de terceiros; usuário é denunciado por nickname (o browser nunca vê ids de usuário).
- **`/admin/*`**: UI só em inglês, sem i18n, servida sob qualquer prefixo de locale (o middleware do next-intl redireciona `/admin` para o locale do cookie; um 404 por locale trancaria admins em pt-BR/es). `requireAdmin()` (`app/[locale]/admin/guard.ts`) no layout **e** em cada página → 404 para não-admin; cada action de `actions/admin.ts` reverifica. Formulários simples via `components/admin/action-form.tsx` (`useActionState`, mostra o código do resultado). Páginas: visão geral, `reports`, `contact`, `items` (busca por id/slug/título → `items/[id]` com afiliados e merge), `ingestion`.
- Namespaces novos: `ratings`, `social`, `lists`. Constantes usadas no client (sem drizzle) em `lib/social/constants.ts`.

## SEO

- `/robots.txt` (`app/robots.ts`): só produção (`VERCEL_ENV=production`) é indexável; previews e local respondem `Disallow: /`. Em produção bloqueia `/api/` e, em todos os idiomas, `/admin`, `/profile`, `/feed`, `/onboarding`, `/search`, `/verify-email`, `/reset-password`, `/recover-password` (`lib/seo.ts`).
- `/sitemap-index.xml` (`app/sitemap-index.xml/route.ts`): índice para enviar ao Search Console, aponta para `/sitemap/<id>.xml`. `/sitemap/0.xml` = páginas estáticas; `/sitemap/1..n.xml` = itens do catálogo em blocos de 40k (`SITEMAP_CHUNK`), ordenados por id. Cada URL leva `hreflang` dos 3 idiomas + `x-default`. Contagem e páginas vêm do Data Cache (`catalog`, 1 dia); sem banco no build, o sitemap cai para só o bloco estático.
- Search Console: propriedade de domínio `hexascore.app` verificada por TXT no DNS da Vercel. CLI `bun run seo:gsc <sites|sitemaps|submit [url]|queries [dias] [query|page|country|device]>` (em `apps/web`, `scripts/search-console.ts`, sem SDK: JWT assinado → token OAuth). Credencial: JSON da service account `hexascore@hexascore.iam.gserviceaccount.com` na raiz do repo (`GSC_KEY_FILE`, ignorado por `.gitignore` — nunca commitar); ela precisa ser usuária da propriedade com permissão Completa.
